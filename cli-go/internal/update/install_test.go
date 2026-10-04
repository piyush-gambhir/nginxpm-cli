package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type archiveEntry struct {
	name     string
	body     string
	typeflag byte // tar only; 0 means regular file
	symlink  bool // zip only
}

func tarGz(t *testing.T, entries ...archiveEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Mode: 0o755, Typeflag: tar.TypeReg, Size: int64(len(e.body))}
		if e.typeflag != 0 && e.typeflag != tar.TypeReg {
			hdr.Typeflag, hdr.Size, hdr.Linkname = e.typeflag, 0, "/etc/passwd"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if hdr.Size > 0 {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func zipArchive(t *testing.T, entries ...archiveEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		fh := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		mode := os.FileMode(0o755)
		if e.symlink {
			mode = os.ModeSymlink | 0o777
		}
		fh.SetMode(mode)
		w, err := zw.CreateHeader(fh)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, e.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// releaseServer serves one release's archive and checksums.txt the way
// GitHub release downloads are laid out.
func releaseServer(t *testing.T, version, archiveName string, archive []byte, checksums string) *httptest.Server {
	t.Helper()
	prefix := fmt.Sprintf("/%s/releases/download/v%s/", Repo, version)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case prefix + archiveName:
			_, _ = w.Write(archive)
		case prefix + "checksums.txt":
			_, _ = io.WriteString(w, checksums)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	setGitHubURL(t, srv.URL)
	return srv
}

func sha(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// installedExe creates a stand-in for the running executable.
func installedExe(t *testing.T, name string) string {
	t.Helper()
	exe := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(exe, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	return exe
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestArchiveNameMatchesGoReleaser(t *testing.T) {
	for goos, want := range map[string]string{
		"darwin":  "nginxpm-cli_darwin_arm64.tar.gz",
		"linux":   "nginxpm-cli_linux_arm64.tar.gz",
		"windows": "nginxpm-cli_windows_arm64.zip",
	} {
		if got := ArchiveName(goos, "arm64"); got != want {
			t.Errorf("ArchiveName(%s) = %q, want %q", goos, got, want)
		}
	}
}

func TestInstallReplacesUnixExecutable(t *testing.T) {
	archive := tarGz(t, archiveEntry{name: "README.md", body: "docs"}, archiveEntry{name: "nginxpm", body: "new binary"})
	name := ArchiveName("linux", "amd64")
	releaseServer(t, "0.2.0", name, archive, sha(archive)+"  "+name+"\n")
	exe := installedExe(t, "nginxpm")
	if err := os.Chmod(exe, 0o700); err != nil {
		t.Fatal(err)
	}

	in := &Installer{GOOS: "linux", GOARCH: "amd64", ExecPath: exe, Out: io.Discard}
	if err := in.Install(context.Background(), "0.2.0"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, exe); got != "new binary" {
		t.Fatalf("executable = %q, want the new binary", got)
	}
	if runtime.GOOS != "windows" {
		if fi, _ := os.Stat(exe); fi.Mode().Perm() != 0o755 {
			t.Errorf("mode = %v, want 0755", fi.Mode().Perm())
		}
	}
	entries, _ := os.ReadDir(filepath.Dir(exe))
	if len(entries) != 1 {
		t.Errorf("install left extra files next to the executable: %v", entries)
	}
}

func TestInstallWindowsMovesRunningExeAside(t *testing.T) {
	archive := zipArchive(t, archiveEntry{name: "LICENSE", body: "mit"}, archiveEntry{name: "nginxpm.exe", body: "new exe"})
	name := ArchiveName("windows", "amd64")
	releaseServer(t, "0.2.0", name, archive, sha(archive)+"  "+name+"\n")
	exe := installedExe(t, "nginxpm.exe")
	// A leftover from an earlier update must not block this one.
	if err := os.WriteFile(exe+".old", []byte("older"), 0o755); err != nil {
		t.Fatal(err)
	}

	in := &Installer{GOOS: "windows", GOARCH: "amd64", ExecPath: exe, Out: io.Discard}
	if err := in.Install(context.Background(), "v0.2.0"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, exe); got != "new exe" {
		t.Fatalf("nginxpm.exe = %q, want the new exe", got)
	}
	if got := readFile(t, exe+".old"); got != "old binary" {
		t.Fatalf("nginxpm.exe.old = %q, want the previous executable", got)
	}

	RemoveLeftoverBinary("linux", exe)
	if _, err := os.Stat(exe + ".old"); err != nil {
		t.Fatalf("cleanup ran on a non-Windows OS: %v", err)
	}
	RemoveLeftoverBinary("windows", exe)
	if _, err := os.Stat(exe + ".old"); !os.IsNotExist(err) {
		t.Fatalf("nginxpm.exe.old still present after cleanup: %v", err)
	}
	if got := readFile(t, exe); got != "new exe" {
		t.Fatalf("cleanup touched nginxpm.exe: %q", got)
	}
}

func TestReplaceWindowsRestoresOldExeOnFailure(t *testing.T) {
	exe := installedExe(t, "nginxpm.exe")
	src := filepath.Join(t.TempDir(), "nginxpm.exe")
	if err := os.WriteFile(src, []byte("new exe"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Fail the step that moves the staged exe into place, after the running
	// exe has already been moved aside.
	orig := rename
	t.Cleanup(func() { rename = orig })
	rename = func(from, to string) error {
		if to == exe && from != exe+".old" {
			return fmt.Errorf("injected failure")
		}
		return orig(from, to)
	}

	if err := replaceWindows(src, exe); err == nil || !strings.Contains(err.Error(), "injected failure") {
		t.Fatalf("replaceWindows error = %v, want the injected failure", err)
	}
	if got := readFile(t, exe); got != "old binary" {
		t.Fatalf("nginxpm.exe = %q, want the old binary restored", got)
	}
	if entries, _ := os.ReadDir(filepath.Dir(exe)); len(entries) != 1 {
		t.Errorf("rollback left files behind: %v", entries)
	}
}

func TestInstallRefusesChecksumProblems(t *testing.T) {
	archive := tarGz(t, archiveEntry{name: "nginxpm", body: "new binary"})
	name := ArchiveName("linux", "amd64")
	for label, checksums := range map[string]string{
		"mismatch":      strings.Repeat("0", 64) + "  " + name + "\n",
		"missing entry": sha(archive) + "  nginxpm-cli_darwin_amd64.tar.gz\n",
	} {
		t.Run(label, func(t *testing.T) {
			releaseServer(t, "0.2.0", name, archive, checksums)
			exe := installedExe(t, "nginxpm")
			in := &Installer{GOOS: "linux", GOARCH: "amd64", ExecPath: exe, Out: io.Discard}
			err := in.Install(context.Background(), "0.2.0")
			if err == nil || !strings.Contains(err.Error(), "checksum") {
				t.Fatalf("Install error = %v, want a checksum refusal", err)
			}
			if got := readFile(t, exe); got != "old binary" {
				t.Fatalf("executable = %q, want the old binary", got)
			}
		})
	}
}

func TestInstallUnwritableDirectoryKeepsOldBinary(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs a non-root Unix user for directory permissions")
	}
	exe := installedExe(t, "nginxpm")
	dir := filepath.Dir(exe)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	in := &Installer{GOOS: "linux", GOARCH: "amd64", ExecPath: exe, Out: io.Discard}
	err := in.Install(context.Background(), "0.2.0")
	if err == nil || !strings.Contains(err.Error(), "sudo") || !strings.Contains(err.Error(), "install script") {
		t.Fatalf("Install error = %v, want a not-writable message suggesting sudo or the install script", err)
	}
	if got := readFile(t, exe); got != "old binary" {
		t.Fatalf("executable = %q, want the old binary", got)
	}
}

func TestExtractBinaryRefusesUnsafeEntries(t *testing.T) {
	for _, tc := range []struct {
		label   string
		goos    string
		archive []byte
		wantErr string
	}{
		{"tar traversal", "linux", tarGz(t, archiveEntry{name: "../../nginxpm", body: "x"}), "unsafe path"},
		{"tar absolute", "linux", tarGz(t, archiveEntry{name: "/usr/local/bin/nginxpm", body: "x"}), "unsafe path"},
		{"tar traversal before binary", "linux", tarGz(t, archiveEntry{name: "../evil", body: "x"}, archiveEntry{name: "nginxpm", body: "x"}), "unsafe path"},
		{"tar symlink", "linux", tarGz(t, archiveEntry{name: "nginxpm", typeflag: tar.TypeSymlink}), "not a regular file"},
		{"tar hardlink", "linux", tarGz(t, archiveEntry{name: "nginxpm", typeflag: tar.TypeLink}), "not a regular file"},
		{"tar missing binary", "linux", tarGz(t, archiveEntry{name: "README.md", body: "x"}), "not found"},
		{"zip traversal", "windows", zipArchive(t, archiveEntry{name: `..\..\nginxpm.exe`, body: "x"}), "unsafe path"},
		{"zip symlink", "windows", zipArchive(t, archiveEntry{name: "nginxpm.exe", body: "target", symlink: true}), "not a regular file"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			dir := t.TempDir()
			archivePath := filepath.Join(dir, "release-archive")
			if err := os.WriteFile(archivePath, tc.archive, 0o600); err != nil {
				t.Fatal(err)
			}
			destDir := filepath.Join(dir, "out")
			if err := os.Mkdir(destDir, 0o755); err != nil {
				t.Fatal(err)
			}
			_, err := extractBinary(archivePath, destDir, tc.goos)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("extractBinary error = %v, want %q", err, tc.wantErr)
			}
			if entries, _ := os.ReadDir(dir); len(entries) != 2 {
				t.Errorf("extraction wrote outside destDir: %v", entries)
			}
		})
	}
}

func TestExtractBinaryWritesFixedName(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "release-archive")
	if err := os.WriteFile(archivePath, tarGz(t, archiveEntry{name: "./nginxpm", body: "bin"}), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := extractBinary(archivePath, dir, "darwin")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "nginxpm"); got != want {
		t.Fatalf("extractBinary path = %q, want %q", got, want)
	}
	if data := readFile(t, got); data != "bin" {
		t.Fatalf("extracted %q, want %q", data, "bin")
	}
}
