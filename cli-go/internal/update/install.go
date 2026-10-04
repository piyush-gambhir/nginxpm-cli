package update

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	// maxReleaseArtifactBytes caps every download and the extracted binary.
	maxReleaseArtifactBytes int64 = 256 << 20
	maxChecksumsBytes       int64 = 1 << 20
	downloadTimeout               = 120 * time.Second
	projectName                   = "nginxpm-cli"
	binaryName                    = "nginxpm"
)

// rename is os.Rename; tests override it to fail a specific step.
var rename = os.Rename

// Installer downloads a release archive, verifies it against checksums.txt,
// and replaces ExecPath with the binary inside. GOOS and GOARCH pick the
// archive and the replacement strategy, so tests can run every platform's
// logic on any OS.
type Installer struct {
	GOOS     string
	GOARCH   string
	ExecPath string    // resolved path of the executable to replace
	Out      io.Writer // progress messages
}

// ArchiveName is the GoReleaser archive for goos/goarch
// (name_template "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}", zip on Windows).
func ArchiveName(goos, goarch string) string {
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("%s_%s_%s.%s", projectName, goos, goarch, ext)
}

// Install replaces in.ExecPath with release version. On any failure the old
// binary is left in place and working.
func (in *Installer) Install(ctx context.Context, version string) error {
	version = trimV(version)
	if !releaseTagPattern.MatchString(version) {
		return fmt.Errorf("invalid release version %q", version)
	}
	if err := checkWritable(filepath.Dir(in.ExecPath), in.GOOS); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "nginxpm-update-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archive := ArchiveName(in.GOOS, in.GOARCH)
	base := fmt.Sprintf("%s/%s/releases/download/v%s/", GitHubURL, Repo, version)
	archivePath := filepath.Join(tmpDir, "release-archive")
	fmt.Fprintf(in.Out, "Downloading %s...\n", base+archive)
	if err := download(ctx, archivePath, base+archive, maxReleaseArtifactBytes); err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}

	fmt.Fprintln(in.Out, "Verifying checksum...")
	checksumsPath := filepath.Join(tmpDir, "checksums.txt")
	if err := download(ctx, checksumsPath, base+"checksums.txt", maxChecksumsBytes); err != nil {
		return fmt.Errorf("downloading checksums: %w", err)
	}
	if err := verifyChecksum(archivePath, checksumsPath, archive); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	binPath, err := extractBinary(archivePath, tmpDir, in.GOOS)
	if err != nil {
		return fmt.Errorf("extracting update: %w", err)
	}

	fmt.Fprintf(in.Out, "Replacing %s...\n", in.ExecPath)
	if in.GOOS == "windows" {
		err = replaceWindows(binPath, in.ExecPath)
	} else {
		err = replaceUnix(binPath, in.ExecPath)
	}
	if err != nil {
		return fmt.Errorf("replacing binary: %w", err)
	}
	return nil
}

// RemoveLeftoverBinary deletes the <exe>.old that a Windows update leaves
// behind (a running .exe cannot delete itself). Best effort; no-op elsewhere.
func RemoveLeftoverBinary(goos, execPath string) {
	if goos != "windows" || execPath == "" {
		return
	}
	_ = os.Remove(execPath + ".old")
}

// checkWritable fails with a clear message when the update cannot be written
// next to the executable, before anything is downloaded.
func checkWritable(dir, goos string) error {
	f, err := os.CreateTemp(dir, ".nginxpm-write-test-*")
	if err != nil {
		if goos == "windows" {
			return fmt.Errorf("cannot write to %s (%w): re-run from an Administrator terminal, or move nginxpm.exe to a directory you can write to", dir, err)
		}
		return fmt.Errorf("cannot write to %s (%w): re-run with sudo, or reinstall with the install script into a writable directory (INSTALL_DIR=~/.local/bin)", dir, err)
	}
	name := f.Name()
	f.Close()
	_ = os.Remove(name)
	return nil
}

func download(ctx context.Context, dst, url string, limit int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "nginxpm-cli")
	resp, err := (&http.Client{Timeout: downloadTimeout}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if err := copyLimited(out, resp.Body, limit); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// copyLimited copies src to dst and fails if src holds more than limit bytes.
func copyLimited(dst io.Writer, src io.Reader, limit int64) error {
	written, err := io.CopyN(dst, src, limit+1)
	if err != nil && err != io.EOF {
		return err
	}
	if written > limit {
		return fmt.Errorf("release artifact exceeds %d MiB limit", limit>>20)
	}
	return nil
}

// verifyChecksum checks filePath against the "<sha256>  <name>" line for name
// in checksums.txt, refusing a missing entry or a mismatch.
func verifyChecksum(filePath, checksumsPath, name string) error {
	sums, err := os.Open(checksumsPath)
	if err != nil {
		return err
	}
	defer sums.Close()
	var expected string
	scanner := bufio.NewScanner(sums)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) == 2 && parts[1] == name {
			expected = parts[0]
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading checksums: %w", err)
	}
	if expected == "" {
		return fmt.Errorf("no checksum for %s in checksums.txt", name)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if err := copyLimited(h, f, maxReleaseArtifactBytes); err != nil {
		return err
	}
	if actual := hex.EncodeToString(h.Sum(nil)); !strings.EqualFold(actual, expected) {
		return fmt.Errorf("SHA-256 mismatch for %s: expected %s, got %s", name, expected, actual)
	}
	return nil
}

// extractBinary writes the release binary from the archive to a fixed name in
// destDir, so no part of an archive entry name reaches the filesystem (zip
// slip). Unsafe entry names and a non-regular binary entry are refused.
func extractBinary(archivePath, destDir, goos string) (string, error) {
	want := binaryName
	if goos == "windows" {
		want += ".exe"
	}
	outPath := filepath.Join(destDir, want)
	if goos == "windows" {
		return outPath, extractFromZip(archivePath, want, outPath)
	}
	return outPath, extractFromTarGz(archivePath, want, outPath)
}

// entryIsBinary reports whether an archive entry is the release binary, and
// refuses absolute paths and ".." components anywhere in the archive.
func entryIsBinary(name, want string) (bool, error) {
	name = strings.ReplaceAll(name, `\`, "/")
	if path.IsAbs(name) || filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return false, fmt.Errorf("unsafe path %q in archive", name)
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return false, fmt.Errorf("unsafe path %q in archive", name)
		}
	}
	return path.Clean(name) == want, nil
}

func extractFromTarGz(archivePath, want, outPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("opening gzip: %w", err)
	}
	defer gz.Close()

	// Bound the decompressed stream too, including entries that are skipped.
	tr := tar.NewReader(io.LimitReader(gz, 2*maxReleaseArtifactBytes))
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("%s not found in archive", want)
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}
		match, err := entryIsBinary(hdr.Name, want)
		if err != nil {
			return err
		}
		if !match {
			continue
		}
		if hdr.Typeflag != tar.TypeReg {
			return fmt.Errorf("%s in archive is not a regular file", want)
		}
		if hdr.Size > maxReleaseArtifactBytes {
			return fmt.Errorf("%s in archive exceeds %d MiB limit", want, maxReleaseArtifactBytes>>20)
		}
		return writeBinary(outPath, tr)
	}
}

func extractFromZip(archivePath, want, outPath string) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}
	defer zr.Close()
	for _, zf := range zr.File {
		match, err := entryIsBinary(zf.Name, want)
		if err != nil {
			return err
		}
		if !match {
			continue
		}
		if !zf.Mode().IsRegular() {
			return fmt.Errorf("%s in archive is not a regular file", want)
		}
		if zf.UncompressedSize64 > uint64(maxReleaseArtifactBytes) {
			return fmt.Errorf("%s in archive exceeds %d MiB limit", want, maxReleaseArtifactBytes>>20)
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		return writeBinary(outPath, rc)
	}
	return fmt.Errorf("%s not found in archive", want)
}

func writeBinary(outPath string, r io.Reader) error {
	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if err := copyLimited(out, r, maxReleaseArtifactBytes); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// stageNextTo copies src into a new temp file in dst's directory (so a rename
// onto dst stays on one filesystem) and returns its path.
func stageNextTo(src, dst string) (string, error) {
	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()
	pattern := ".nginxpm-update-*"
	if strings.HasSuffix(strings.ToLower(dst), ".exe") {
		pattern += ".exe"
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), pattern)
	if err != nil {
		return "", err
	}
	if err := copyLimited(tmp, in, maxReleaseArtifactBytes); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	if err := tmp.Chmod(0o755); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}

// replaceUnix renames a staged copy over the executable in one atomic step.
func replaceUnix(src, execPath string) error {
	staged, err := stageNextTo(src, execPath)
	if err != nil {
		return err
	}
	if err := rename(staged, execPath); err != nil {
		os.Remove(staged)
		return err
	}
	return nil
}

// replaceWindows moves the running executable aside to <exe>.old (Windows
// allows renaming, not overwriting, a running .exe) and moves the new one into
// place, restoring the old one if that fails. The next start removes .old.
func replaceWindows(src, execPath string) error {
	staged, err := stageNextTo(src, execPath)
	if err != nil {
		return err
	}
	old := execPath + ".old"
	_ = os.Remove(old)
	if err := rename(execPath, old); err != nil {
		os.Remove(staged)
		return fmt.Errorf("moving the running executable aside: %w", err)
	}
	if err := rename(staged, execPath); err != nil {
		os.Remove(staged)
		if restoreErr := rename(old, execPath); restoreErr != nil {
			return fmt.Errorf("installing new executable: %w (restoring %s also failed: %v)", err, old, restoreErr)
		}
		return fmt.Errorf("installing new executable: %w", err)
	}
	return nil
}
