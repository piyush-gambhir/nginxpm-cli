package update

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofrs/flock"
)

// latestServer answers github.com/<repo>/releases/latest with a redirect to
// tag (or with status when non-zero) and counts requests.
func latestServer(t *testing.T, tag string, status int) *atomic.Int32 {
	t.Helper()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if r.URL.Path != "/"+Repo+"/releases/latest" {
			http.NotFound(w, r)
			return
		}
		if status != 0 {
			w.WriteHeader(status)
			return
		}
		w.Header().Set("Location", GitHubURL+"/"+Repo+"/releases/tag/"+tag)
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)
	setGitHubURL(t, srv.URL)
	return &hits
}

func setGitHubURL(t *testing.T, u string) {
	t.Helper()
	orig := GitHubURL
	GitHubURL = u
	t.Cleanup(func() { GitHubURL = orig })
}

func TestFetchLatestReadsTheRedirectWithoutFollowingIt(t *testing.T) {
	for _, tc := range []struct {
		label    string
		location string // "" sends no Location; "$" is replaced by the server URL
		want     string
	}{
		{"semver v tag", "$/" + Repo + "/releases/tag/v0.1.11", "0.1.11"},
		{"missing Location", "", ""},
		{"foreign host", "https://evil.example/" + Repo + "/releases/tag/v0.1.11", ""},
		{"other repository", "$/someone/else/releases/tag/v0.1.11", ""},
		{"non-semver tag", "$/" + Repo + "/releases/tag/latest", ""},
		{"tag without v", "$/" + Repo + "/releases/tag/0.1.11", ""},
		{"pre-release tag", "$/" + Repo + "/releases/tag/v0.1.11-rc1", ""},
		{"no releases yet", "$/" + Repo + "/releases", ""},
	} {
		t.Run(tc.label, func(t *testing.T) {
			var followed atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/"+Repo+"/releases/latest" {
					followed.Add(1)
					w.WriteHeader(http.StatusOK)
					return
				}
				if tc.location != "" {
					w.Header().Set("Location", strings.Replace(tc.location, "$", GitHubURL, 1))
				}
				w.WriteHeader(http.StatusFound)
			}))
			t.Cleanup(srv.Close)
			setGitHubURL(t, srv.URL)

			got, err := fetchLatest(BackgroundTimeout)
			if tc.want != "" && (err != nil || got != tc.want) {
				t.Fatalf("fetchLatest = %q, %v; want %q", got, err, tc.want)
			}
			if tc.want == "" && err == nil {
				t.Fatalf("fetchLatest = %q, want an error", got)
			}
			if followed.Load() != 0 {
				t.Fatal("fetchLatest followed the redirect")
			}
		})
	}
}

func TestFetchLatestRejectsNonRedirect(t *testing.T) {
	latestServer(t, "", http.StatusForbidden)
	if _, err := fetchLatest(BackgroundTimeout); err == nil || !strings.Contains(err.Error(), "status 403") {
		t.Fatalf("fetchLatest error = %v, want status 403", err)
	}
}

func setNow(t *testing.T, at time.Time) {
	t.Helper()
	orig := now
	now = func() time.Time { return at }
	t.Cleanup(func() { now = orig })
}

func TestNotifierDisabled(t *testing.T) {
	env := func(vars map[string]string) func(string) string {
		return func(k string) string { return vars[k] }
	}
	none := env(nil)
	if NotifierDisabled("0.1.9", none, true) {
		t.Fatal("notifier disabled for a release build in an interactive terminal")
	}
	for _, tc := range []struct {
		label   string
		version string
		getenv  func(string) string
		tty     bool
	}{
		{"stderr not a terminal", "0.1.9", none, false},
		{"CI set", "0.1.9", env(map[string]string{"CI": "true"}), true},
		{"CI set to 0", "0.1.9", env(map[string]string{"CI": "0"}), true},
		{"NGINXPM_NO_UPDATE_NOTIFIER", "0.1.9", env(map[string]string{"NGINXPM_NO_UPDATE_NOTIFIER": "1"}), true},
		{"NO_UPDATE_NOTIFIER", "0.1.9", env(map[string]string{"NO_UPDATE_NOTIFIER": "yes"}), true},
		{"dev build", "dev", none, true},
		{"empty version", "", none, true},
		{"commit hash version", "1b756a3", none, true},
	} {
		if !NotifierDisabled(tc.version, tc.getenv, tc.tty) {
			t.Errorf("%s: notifier enabled, want disabled", tc.label)
		}
	}
	if NotifierDisabled("v0.1.9-3-g1b756a3-dirty", none, true) {
		t.Error("notifier disabled for a `make install` build of a tag")
	}
}

func TestPrintNotice(t *testing.T) {
	info := newInfo("0.1.9", "0.1.10")
	for _, tc := range []struct {
		goInstall bool
		line      string
	}{
		{false, "Update with: nginxpm update\n"},
		{true, "Update with: git pull && make install (in your nginxpm-cli/cli-go clone)\n"},
	} {
		var buf bytes.Buffer
		PrintNotice(&buf, info, tc.goInstall)
		want := "\nA new version of nginxpm is available: v0.1.9 -> v0.1.10\n" +
			tc.line +
			"Release notes: https://github.com/piyush-gambhir/nginxpm-cli/releases/tag/v0.1.10\n"
		if buf.String() != want {
			t.Errorf("goInstall=%t notice:\n%q\nwant:\n%q", tc.goInstall, buf.String(), want)
		}
	}
}

func TestNotifyOncePerVersionPerDay(t *testing.T) {
	dir := t.TempDir()
	start := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)
	notify := func(at time.Time, latest string) string {
		setNow(t, at)
		var buf bytes.Buffer
		Notify(&buf, newInfo("0.1.9", latest), dir, false)
		return buf.String()
	}

	if out := notify(start, "0.1.10"); !strings.Contains(out, "v0.1.9 -> v0.1.10") {
		t.Fatalf("first notice missing, got %q", out)
	}
	if out := notify(start.Add(time.Hour), "0.1.10"); out != "" {
		t.Fatalf("same version announced twice within 24h: %q", out)
	}
	if out := notify(start.Add(2*time.Hour), "0.1.11"); !strings.Contains(out, "v0.1.11") {
		t.Fatalf("a newer version was not announced: %q", out)
	}
	if out := notify(start.Add(27*time.Hour), "0.1.11"); !strings.Contains(out, "v0.1.11") {
		t.Fatalf("notice not repeated after 24h: %q", out)
	}

	var buf bytes.Buffer
	Notify(&buf, newInfo("0.1.11", "0.1.11"), dir, false)
	if buf.Len() != 0 {
		t.Fatalf("notice printed with no update available: %q", buf.String())
	}
}

func TestCheckForUpdateCachesForADay(t *testing.T) {
	dir := t.TempDir()
	hits := latestServer(t, "v0.1.10", 0)
	start := time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)

	setNow(t, start)
	info := CheckForUpdate("0.1.9", dir)
	if info == nil || !info.Available || info.LatestVersion != "0.1.10" {
		t.Fatalf("CheckForUpdate = %+v, want 0.1.10 available", info)
	}
	setNow(t, start.Add(23*time.Hour))
	if info := CheckForUpdate("0.1.9", dir); info == nil || info.LatestVersion != "0.1.10" {
		t.Fatalf("cached CheckForUpdate = %+v", info)
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("GitHub requests within 24h = %d, want 1", got)
	}
	setNow(t, start.Add(25*time.Hour))
	CheckForUpdate("0.1.9", dir)
	if got := hits.Load(); got != 2 {
		t.Fatalf("GitHub requests after 24h = %d, want 2", got)
	}
}

func TestNotifySkipsWhileAnotherProcessHoldsTheCache(t *testing.T) {
	dir := t.TempDir()
	lock := flock.New(filepath.Join(dir, cacheFileName) + ".lock")
	if err := lock.Lock(); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	Notify(&buf, newInfo("0.1.9", "0.1.10"), dir, false)
	if buf.Len() != 0 {
		t.Fatalf("notice printed while another process held the cache lock: %q", buf.String())
	}
	if err := lock.Unlock(); err != nil {
		t.Fatal(err)
	}
	Notify(&buf, newInfo("0.1.9", "0.1.10"), dir, false)
	if !strings.Contains(buf.String(), "v0.1.10") {
		t.Fatalf("notice not printed once the lock was free: %q", buf.String())
	}
}

func TestRefreshKeepsNoticeWrittenDuringRequest(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, cacheFileName)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Another process announces 0.1.10 while this request is in flight.
		writeCache(path, &cacheEntry{LastChecked: "2026-10-05T08:00:00Z", LatestVersion: "0.1.10", NotifiedVersion: "0.1.10", NotifiedAt: "2026-10-05T08:00:00Z"})
		w.Header().Set("Location", GitHubURL+"/"+Repo+"/releases/tag/v0.1.10")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)
	setGitHubURL(t, srv.URL)
	setNow(t, time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC))

	CheckForUpdate("0.1.9", dir)
	entry, err := readCache(path)
	if err != nil {
		t.Fatal(err)
	}
	if entry.NotifiedVersion != "0.1.10" || entry.LastChecked != "2026-10-05T09:00:00Z" {
		t.Fatalf("cache after refresh = %+v, want the concurrent notice kept and the new check time", entry)
	}
}

func TestCheckForUpdateCachesFailures(t *testing.T) {
	dir := t.TempDir()
	hits := latestServer(t, "", http.StatusInternalServerError)
	setNow(t, time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC))
	for range 3 {
		if info := CheckForUpdate("0.1.9", dir); info != nil {
			t.Fatalf("CheckForUpdate after a failure = %+v, want nil", info)
		}
	}
	if got := hits.Load(); got != 1 {
		t.Fatalf("GitHub requests after a failed check = %d, want 1 (failure cached)", got)
	}
}

func TestCheckForUpdateRejectsOddTags(t *testing.T) {
	latestServer(t, "v1.0.0/../../evil", 0)
	if _, err := CheckForUpdateFresh("0.1.9", t.TempDir()); err == nil {
		t.Fatal("accepted a release tag that is not a version")
	}
}

func TestCachedInfoNeverUsesNetwork(t *testing.T) {
	dir := t.TempDir()
	hits := latestServer(t, "v0.1.10", 0)
	if info := CachedInfo("0.1.9", dir); info != nil {
		t.Fatalf("CachedInfo with no cache = %+v, want nil", info)
	}
	writeCache(filepath.Join(dir, cacheFileName), &cacheEntry{LastChecked: "2020-01-01T00:00:00Z", LatestVersion: "0.1.10"})
	if info := CachedInfo("0.1.9", dir); info == nil || !info.Available || info.LatestVersion != "0.1.10" {
		t.Fatalf("CachedInfo = %+v, want 0.1.10 available", info)
	}
	if hits.Load() != 0 {
		t.Fatal("CachedInfo contacted GitHub")
	}
	ClearCache(dir)
	if _, err := os.Stat(filepath.Join(dir, cacheFileName)); !os.IsNotExist(err) {
		t.Fatalf("ClearCache left the cache: %v", err)
	}
}

func TestIsGoInstall(t *testing.T) {
	home := t.TempDir()
	gopath := t.TempDir()
	gobin := t.TempDir()
	env := func(vars map[string]string) func(string) string {
		return func(k string) string { return vars[k] }
	}
	for _, tc := range []struct {
		label string
		exe   string
		env   map[string]string
		want  bool
	}{
		{"GOBIN", filepath.Join(gobin, "nginxpm"), map[string]string{"GOBIN": gobin}, true},
		{"GOPATH/bin", filepath.Join(gopath, "bin", "nginxpm"), map[string]string{"GOPATH": "/nonexistent" + string(filepath.ListSeparator) + gopath}, true},
		{"~/go/bin", filepath.Join(home, "go", "bin", "nginxpm"), nil, true},
		{"/usr/local/bin", "/usr/local/bin/nginxpm", map[string]string{"GOBIN": gobin, "GOPATH": gopath}, false},
		{"subdirectory of a Go bin dir", filepath.Join(home, "go", "bin", "sub", "nginxpm"), nil, false},
	} {
		if got := IsGoInstall(tc.exe, env(tc.env), home); got != tc.want {
			t.Errorf("%s: IsGoInstall = %t, want %t", tc.label, got, tc.want)
		}
	}
}
