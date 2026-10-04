package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/build"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/update"
)

const noticeHeadline = "A new version of nginxpm is available"

// updateEnv isolates one test: release build 0.1.9, an interactive stderr,
// a fake GitHub serving latest (and, when archive is set, its download), a
// temp config dir, and a stand-in executable.
type updateEnv struct {
	apiHits   atomic.Int32
	configDir string
	exe       string
}

func newUpdateEnv(t *testing.T, latest string, archive []byte) *updateEnv {
	t.Helper()
	env := &updateEnv{}
	xdg := t.TempDir()
	env.configDir = filepath.Join(xdg, "nginxpm-cli")
	for k, v := range map[string]string{
		"XDG_CONFIG_HOME": xdg, "HOME": t.TempDir(), "GOBIN": "", "GOPATH": t.TempDir(),
		"CI": "", "NGINXPM_NO_UPDATE_NOTIFIER": "", "NO_UPDATE_NOTIFIER": "",
		"NGINXPM_QUIET": "", "NGINXPM_NO_INPUT": "", "NGINXPM_READ_ONLY": "",
	} {
		t.Setenv(k, v)
	}

	archiveName := update.ArchiveName("linux", "amd64")
	download := fmt.Sprintf("/%s/releases/download/v%s/", update.Repo, latest)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/"+update.Repo+"/releases/latest":
			env.apiHits.Add(1)
			fmt.Fprintf(w, `{"tag_name": "v%s"}`, latest)
		case archive != nil && r.URL.Path == download+archiveName:
			_, _ = w.Write(archive)
		case archive != nil && r.URL.Path == download+"checksums.txt":
			sum := sha256.Sum256(archive)
			fmt.Fprintf(w, "%s  %s\n", hex.EncodeToString(sum[:]), archiveName)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	env.exe = filepath.Join(t.TempDir(), "nginxpm")
	if err := os.WriteFile(env.exe, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	origAPI, origDL, origVersion := update.APIBaseURL, update.DownloadBaseURL, build.Version
	origTTY, origStdin, origWait, origExe := stderrIsTerminal, stdinIsTerminal, updateNoticeWait, executablePath
	origGOOS, origGOARCH := goos, goarch
	t.Cleanup(func() {
		update.APIBaseURL, update.DownloadBaseURL, build.Version = origAPI, origDL, origVersion
		stderrIsTerminal, stdinIsTerminal, updateNoticeWait, executablePath = origTTY, origStdin, origWait, origExe
		goos, goarch = origGOOS, origGOARCH
	})
	update.APIBaseURL, update.DownloadBaseURL, build.Version = srv.URL, srv.URL, "0.1.9"
	stderrIsTerminal = func() bool { return true }
	stdinIsTerminal = func() bool { return false }
	updateNoticeWait = 5 * time.Second
	executablePath = func() (string, error) { return env.exe, nil }
	goos, goarch = "linux", "amd64"
	return env
}

func runRoot(t *testing.T, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	root := newRootCmd()
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)
	err = root.Execute()
	return out.String(), errOut.String(), err
}

func releaseTarGz(t *testing.T, body string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: "nginxpm", Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func exeContents(t *testing.T, env *updateEnv) string {
	t.Helper()
	data, err := os.ReadFile(env.exe)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestUpdateNoticeShownOncePerVersion(t *testing.T) {
	env := newUpdateEnv(t, "0.1.10", nil)

	_, stderr, err := runRoot(t, "", "config", "list-profiles")
	if err != nil {
		t.Fatal(err)
	}
	want := "\n" + noticeHeadline + ": v0.1.9 -> v0.1.10\n" +
		"Update with: nginxpm update\n" +
		"Release notes: https://github.com/piyush-gambhir/nginxpm-cli/releases/tag/v0.1.10\n"
	if stderr != want {
		t.Fatalf("stderr:\n%q\nwant:\n%q", stderr, want)
	}

	_, stderr, err = runRoot(t, "", "config", "list-profiles")
	if err != nil {
		t.Fatal(err)
	}
	if stderr != "" {
		t.Fatalf("notice repeated on the next run: %q", stderr)
	}
	if got := env.apiHits.Load(); got != 1 {
		t.Fatalf("GitHub requests = %d, want 1 (second run uses the cache)", got)
	}
}

func TestUpdateNoticeGoInstallLine(t *testing.T) {
	env := newUpdateEnv(t, "0.1.10", nil)
	gobin := filepath.Dir(env.exe)
	t.Setenv("GOBIN", gobin)

	_, stderr, err := runRoot(t, "", "config", "list-profiles")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr, "Update with: git pull && make install") || strings.Contains(stderr, "nginxpm update") {
		t.Fatalf("notice for a Go bin install should point at make install:\n%s", stderr)
	}
}

func TestUpdateNoticeSuppressed(t *testing.T) {
	for _, tc := range []struct {
		label string
		setup func(t *testing.T)
		args  []string
	}{
		{"stderr not a terminal", func(*testing.T) { stderrIsTerminal = func() bool { return false } }, nil},
		{"CI", func(t *testing.T) { t.Setenv("CI", "true") }, nil},
		{"NGINXPM_NO_UPDATE_NOTIFIER", func(t *testing.T) { t.Setenv("NGINXPM_NO_UPDATE_NOTIFIER", "1") }, nil},
		{"NO_UPDATE_NOTIFIER", func(t *testing.T) { t.Setenv("NO_UPDATE_NOTIFIER", "1") }, nil},
		{"NGINXPM_QUIET", func(t *testing.T) { t.Setenv("NGINXPM_QUIET", "1") }, nil},
		{"--quiet", func(*testing.T) {}, []string{"--quiet"}},
		{"-q", func(*testing.T) {}, []string{"-q"}},
		{"dev build", func(*testing.T) { build.Version = "dev" }, nil},
		{"commit-hash build", func(*testing.T) { build.Version = "1b756a3" }, nil},
		{"version command", func(*testing.T) {}, []string{"version"}},
		{"help command", func(*testing.T) {}, []string{"help", "config"}},
		{"__complete", func(*testing.T) {}, []string{"__complete", "config", ""}},
		{"__completeNoDesc", func(*testing.T) {}, []string{"__completeNoDesc", "config", ""}},
	} {
		t.Run(tc.label, func(t *testing.T) {
			env := newUpdateEnv(t, "0.1.10", nil)
			tc.setup(t)
			args := tc.args
			if len(args) == 0 || strings.HasPrefix(args[0], "-") {
				args = append([]string{"config", "list-profiles"}, args...)
			}
			_, stderr, err := runRoot(t, "", args...)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(stderr, noticeHeadline) {
				t.Errorf("notice printed: %q", stderr)
			}
			if got := env.apiHits.Load(); got != 0 {
				t.Errorf("GitHub requests = %d, want 0", got)
			}
		})
	}
}

func TestSkipUpdateCheckCommands(t *testing.T) {
	newUpdateEnv(t, "0.1.10", nil)
	root := newRootCmd()
	for _, args := range [][]string{
		{"update"}, {"update", "--check"}, {"version"}, {"completion", "bash"},
	} {
		cmd, _, err := root.Find(args)
		if err != nil {
			t.Fatalf("%v: %v", args, err)
		}
		if !skipUpdateCheck(cmd) {
			t.Errorf("%v: update check not skipped", args)
		}
	}
	// A subcommand that happens to be named update still gets the notice.
	cmd, _, err := root.Find([]string{"proxy", "update"})
	if err != nil {
		t.Fatal(err)
	}
	if skipUpdateCheck(cmd) {
		t.Error("`proxy update` skips the update check")
	}
}

func TestUpdateCheckJSON(t *testing.T) {
	env := newUpdateEnv(t, "0.1.10", nil)
	for range 2 { // the second run must query GitHub again, not the cache
		stdout, _, err := runRoot(t, "", "update", "--check", "-o", "json")
		if err != nil {
			t.Fatal(err)
		}
		var got map[string]any
		if err := json.Unmarshal([]byte(stdout), &got); err != nil {
			t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
		}
		want := map[string]any{
			"current_version":  "0.1.9",
			"latest_version":   "0.1.10",
			"update_available": true,
			"release_url":      "https://github.com/piyush-gambhir/nginxpm-cli/releases/tag/v0.1.10",
			"install_method":   "self",
		}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("update --check -o json = %v, want %v", got, want)
		}
	}
	if got := env.apiHits.Load(); got != 2 {
		t.Fatalf("GitHub requests = %d, want 2 (--check bypasses the cache)", got)
	}

	t.Setenv("GOBIN", filepath.Dir(env.exe))
	stdout, _, err := runRoot(t, "", "update", "--check", "-o", "json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, `"install_method": "go"`) {
		t.Fatalf("Go bin install not reported: %s", stdout)
	}
}

func TestUpdateAlreadyLatest(t *testing.T) {
	env := newUpdateEnv(t, "0.1.9", nil)
	stdout, _, err := runRoot(t, "", "update", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "already the latest version") {
		t.Fatalf("stdout = %q", stdout)
	}
	if exeContents(t, env) != "old binary" {
		t.Fatal("executable changed")
	}
}

func TestUpdateInstalls(t *testing.T) {
	env := newUpdateEnv(t, "0.1.10", releaseTarGz(t, "new binary"))
	// Seed the cache so the test can check that a successful update clears it.
	if _, _, err := runRoot(t, "", "update", "--check"); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := runRoot(t, "", "update", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "Updated nginxpm v0.1.9 -> v0.1.10\nRelease notes: https://github.com/piyush-gambhir/nginxpm-cli/releases/tag/v0.1.10\n") {
		t.Fatalf("stdout = %q", stdout)
	}
	if exeContents(t, env) != "new binary" {
		t.Fatal("executable not replaced")
	}
	if _, err := os.Stat(filepath.Join(env.configDir, "update-check.json")); !os.IsNotExist(err) {
		t.Fatalf("update cache not cleared: %v", err)
	}
}

func TestUpdatePromptDefaultsToYes(t *testing.T) {
	env := newUpdateEnv(t, "0.1.10", releaseTarGz(t, "new binary"))
	stdinIsTerminal = func() bool { return true }
	stdout, _, err := runRoot(t, "\n", "update")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "v0.1.9 -> v0.1.10") || !strings.Contains(stdout, "Update now? [Y/n]") {
		t.Fatalf("stdout = %q", stdout)
	}
	if exeContents(t, env) != "new binary" {
		t.Fatal("pressing Enter did not install")
	}

	env = newUpdateEnv(t, "0.1.10", releaseTarGz(t, "new binary"))
	stdinIsTerminal = func() bool { return true }
	stdout, _, err = runRoot(t, "n\n", "update")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "Update cancelled.") || exeContents(t, env) != "old binary" {
		t.Fatalf("answering n did not cancel: %q", stdout)
	}
}

func TestUpdateNeedsYesWithoutAPrompt(t *testing.T) {
	for _, tc := range []struct {
		label string
		args  []string
		setup func(t *testing.T)
	}{
		{"--no-input", []string{"--no-input", "update"}, func(*testing.T) { stdinIsTerminal = func() bool { return true } }},
		{"NGINXPM_NO_INPUT", []string{"update"}, func(t *testing.T) {
			t.Setenv("NGINXPM_NO_INPUT", "1")
			stdinIsTerminal = func() bool { return true }
		}},
		{"stdin not a terminal", []string{"update"}, func(*testing.T) {}},
	} {
		t.Run(tc.label, func(t *testing.T) {
			env := newUpdateEnv(t, "0.1.10", releaseTarGz(t, "new binary"))
			tc.setup(t)
			_, _, err := runRoot(t, "y\n", tc.args...)
			if err == nil || !strings.Contains(err.Error(), "--yes") {
				t.Fatalf("error = %v, want a message saying to pass --yes", err)
			}
			if exeContents(t, env) != "old binary" {
				t.Fatal("executable changed without --yes")
			}
		})
	}
}

func TestUpdateReadOnly(t *testing.T) {
	for _, tc := range []struct {
		label string
		args  []string
		env   string
	}{
		{"--read-only", []string{"--read-only", "update", "--yes"}, ""},
		{"NGINXPM_READ_ONLY", []string{"update", "--yes"}, "true"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			env := newUpdateEnv(t, "0.1.10", releaseTarGz(t, "new binary"))
			t.Setenv("NGINXPM_READ_ONLY", tc.env)
			_, _, err := runRoot(t, "", tc.args...)
			if err == nil || !strings.Contains(err.Error(), "read-only") {
				t.Fatalf("error = %v, want read-only refusal", err)
			}
			if exeContents(t, env) != "old binary" {
				t.Fatal("read-only mode installed an update")
			}

			check := append(tc.args[:len(tc.args)-1:len(tc.args)-1], "--check")
			if _, _, err := runRoot(t, "", check...); err != nil {
				t.Fatalf("%v: %v", check, err)
			}
		})
	}
}

func TestUpdateGoInstallDoesNotSelfReplace(t *testing.T) {
	env := newUpdateEnv(t, "0.1.10", releaseTarGz(t, "new binary"))
	t.Setenv("GOBIN", filepath.Dir(env.exe))
	stdout, _, err := runRoot(t, "", "update", "--yes")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout, "Update with: git pull && make install") {
		t.Fatalf("stdout = %q", stdout)
	}
	if exeContents(t, env) != "old binary" {
		t.Fatal("Go bin install replaced itself")
	}
}

func TestVersionShowsCachedLatest(t *testing.T) {
	env := newUpdateEnv(t, "0.1.10", nil)
	stdout, _, err := runRoot(t, "", "version")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout, "latest") {
		t.Fatalf("version printed a latest version with no cache: %q", stdout)
	}

	if _, _, err := runRoot(t, "", "update", "--check"); err != nil {
		t.Fatal(err)
	}
	hits := env.apiHits.Load()
	stdout, _, err = runRoot(t, "", "version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stdout, "nginxpm-cli version 0.1.9\n") || !strings.Contains(stdout, "  latest: 0.1.10\n  update_available: true\n") {
		t.Fatalf("version output = %q", stdout)
	}
	if env.apiHits.Load() != hits {
		t.Fatal("version contacted GitHub")
	}
}
