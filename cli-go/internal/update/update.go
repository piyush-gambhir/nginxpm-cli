// Package update checks GitHub for new nginxpm releases, prints the new-version
// notice, and installs releases (see install.go).
package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

const (
	// Repo is the GitHub repository that publishes nginxpm releases.
	Repo = "piyush-gambhir/nginxpm-cli"

	cacheDuration  = 24 * time.Hour
	noticeInterval = 24 * time.Hour
	cacheFileName  = "update-check.json"

	// BackgroundTimeout bounds the release lookup behind the new-version notice.
	BackgroundTimeout = 3 * time.Second
	// ForegroundTimeout bounds the release lookup that `nginxpm update` runs.
	ForegroundTimeout = 15 * time.Second

	// SelfUpdateCommand is how a release install updates itself.
	SelfUpdateCommand = "nginxpm update"
	// SourceUpdateCommand updates a build installed with `make install` into
	// a Go bin directory. `go install .../cli-go@latest` is not suggested: it
	// installs a binary named cli-go with version "dev".
	SourceUpdateCommand = "git pull && make install (in your nginxpm-cli/cli-go clone)"
)

// GitHubURL is where releases are looked up and downloaded; tests point it
// at an httptest server. Only github.com pages are used, never
// api.github.com, whose unauthenticated limit of 60 requests an hour per IP
// fails behind shared NAT (offices, CI runners, VPNs).
var GitHubURL = "https://github.com"

// now is the clock; tests override it.
var now = time.Now

// semverPattern matches release versions (optionally with a "v" prefix and a
// pre-release or build suffix, as `make install` produces from git describe).
var semverPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+([-+][0-9A-Za-z.+-]+)?$`)

// releaseTagPattern matches the release tags this CLI publishes. Release
// versions end up in download URLs, so anything else is rejected.
var releaseTagPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)

// UpdateInfo holds the result of an update check.
type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
}

// cacheEntry is the on-disk update-check.json. A failed check is stored with
// an empty LatestVersion so it is not retried for 24 hours.
type cacheEntry struct {
	LastChecked     string `json:"last_checked"`
	LatestVersion   string `json:"latest_version,omitempty"`
	Error           string `json:"error,omitempty"`
	NotifiedVersion string `json:"notified_version,omitempty"`
	NotifiedAt      string `json:"notified_at,omitempty"`
}

// IsReleaseVersion reports whether v is a semver release build rather than a
// development build ("dev", empty, or a bare commit hash).
func IsReleaseVersion(v string) bool {
	return semverPattern.MatchString(v)
}

// NotifierDisabled reports whether the background check must not run at all:
// stderr is not a terminal, CI is set, the user opted out, or this is a
// development build. Callers also skip it for --quiet and for the update,
// version, completion, help, and __complete commands.
func NotifierDisabled(currentVersion string, getenv func(string) string, stderrIsTerminal bool) bool {
	if !stderrIsTerminal {
		return true
	}
	for _, name := range []string{"CI", "NGINXPM_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER"} {
		if getenv(name) != "" {
			return true
		}
	}
	return !IsReleaseVersion(currentVersion)
}

// ReleaseURL is the release notes page for version.
func ReleaseURL(version string) string {
	return fmt.Sprintf("https://github.com/%s/releases/tag/v%s", Repo, trimV(version))
}

// CachedResult reads the cache without touching the network. fresh is true
// when the last check (successful or not) is less than 24 hours old; info is
// nil when that check failed or nothing is cached.
func CachedResult(currentVersion, configDir string) (info *UpdateInfo, fresh bool) {
	entry, err := readCache(filepath.Join(configDir, cacheFileName))
	if err != nil {
		return nil, false
	}
	checked, err := time.Parse(time.RFC3339, entry.LastChecked)
	if err != nil || now().Sub(checked) >= cacheDuration {
		return nil, false
	}
	return newInfo(currentVersion, entry.LatestVersion), true
}

// CheckForUpdate returns the cached result when it is less than 24 hours old,
// and otherwise asks GitHub (3-second timeout) and caches the answer. Failed
// checks are cached too, so an offline machine does not retry on every run.
// It returns nil when the latest version is unknown.
func CheckForUpdate(currentVersion, configDir string) *UpdateInfo {
	if info, fresh := CachedResult(currentVersion, configDir); fresh {
		return info
	}
	latest, err := fetchLatest(BackgroundTimeout)
	recordCheck(configDir, latest, err)
	return newInfo(currentVersion, latest)
}

// CheckForUpdateFresh always asks GitHub, bypassing the cache, and refreshes
// the cache with the answer.
func CheckForUpdateFresh(currentVersion, configDir string) (*UpdateInfo, error) {
	latest, err := fetchLatest(ForegroundTimeout)
	if err != nil {
		return nil, err
	}
	recordCheck(configDir, latest, nil)
	return newInfo(currentVersion, latest), nil
}

// recordCheck stores a check result, keeping the notice fields that another
// process may have written while the request was in flight.
func recordCheck(configDir, latest string, checkErr error) {
	modifyCache(configDir, true, func(entry *cacheEntry) bool {
		entry.LastChecked = now().UTC().Format(time.RFC3339)
		entry.LatestVersion, entry.Error = latest, ""
		if checkErr != nil {
			entry.Error = checkErr.Error()
		}
		return true
	})
}

// CachedInfo returns what the cache knows about the latest release without
// touching the network, or nil when nothing is cached.
func CachedInfo(currentVersion, configDir string) *UpdateInfo {
	entry, err := readCache(filepath.Join(configDir, cacheFileName))
	if err != nil {
		return nil
	}
	return newInfo(currentVersion, entry.LatestVersion)
}

// ClearCache removes the cached check, for example after a successful update.
func ClearCache(configDir string) {
	_ = os.Remove(filepath.Join(configDir, cacheFileName))
}

// Notify prints the new-version notice to w when info reports an update that
// was not already announced in the last 24 hours, and records the
// announcement in the cache so a burst of commands shows it once. The claim
// is made under the cache lock; if another process holds it, this run skips
// the notice rather than wait.
func Notify(w io.Writer, info *UpdateInfo, configDir string, goInstall bool) {
	if info == nil || !info.Available {
		return
	}
	modifyCache(configDir, false, func(entry *cacheEntry) bool {
		if entry.NotifiedVersion == info.LatestVersion {
			if at, err := time.Parse(time.RFC3339, entry.NotifiedAt); err == nil && now().Sub(at) < noticeInterval {
				return false
			}
		}
		if entry.LastChecked == "" {
			entry.LastChecked, entry.LatestVersion = now().UTC().Format(time.RFC3339), info.LatestVersion
		}
		PrintNotice(w, info, goInstall)
		entry.NotifiedVersion = info.LatestVersion
		entry.NotifiedAt = now().UTC().Format(time.RFC3339)
		return true
	})
}

// PrintNotice writes the new-version notice, preceded by a blank line.
func PrintNotice(w io.Writer, info *UpdateInfo, goInstall bool) {
	fmt.Fprintf(w, "\nA new version of nginxpm is available: v%s -> v%s\n", trimV(info.CurrentVersion), trimV(info.LatestVersion))
	fmt.Fprintf(w, "Update with: %s\n", UpdateCommand(goInstall))
	fmt.Fprintf(w, "Release notes: %s\n", info.ReleaseURL)
}

// UpdateCommand is the command that updates this installation.
func UpdateCommand(goInstall bool) string {
	if goInstall {
		return SourceUpdateCommand
	}
	return SelfUpdateCommand
}

// IsGoInstall reports whether exePath sits in a Go bin directory ($GOBIN,
// $GOPATH/bin, or ~/go/bin), where the binary was built from source and must
// not replace itself.
func IsGoInstall(exePath string, getenv func(string) string, homeDir string) bool {
	if exePath == "" {
		return false
	}
	dir := cleanDir(filepath.Dir(exePath))
	var candidates []string
	if gobin := getenv("GOBIN"); gobin != "" {
		candidates = append(candidates, gobin)
	}
	for _, p := range filepath.SplitList(getenv("GOPATH")) {
		if p != "" {
			candidates = append(candidates, filepath.Join(p, "bin"))
		}
	}
	if homeDir != "" {
		candidates = append(candidates, filepath.Join(homeDir, "go", "bin"))
	}
	for _, c := range candidates {
		if cleanDir(c) == dir {
			return true
		}
	}
	return false
}

func cleanDir(dir string) string {
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		dir = resolved
	}
	return filepath.Clean(dir)
}

func newInfo(currentVersion, latest string) *UpdateInfo {
	if latest == "" {
		return nil
	}
	available, _ := isNewer(latest, currentVersion)
	return &UpdateInfo{
		Available:      available,
		CurrentVersion: currentVersion,
		LatestVersion:  latest,
		ReleaseURL:     ReleaseURL(latest),
	}
}

// fetchLatest reads the latest release from the redirect that
// github.com/<repo>/releases/latest answers with (to .../releases/tag/<tag>),
// without following it, and returns the version without the "v" prefix.
func fetchLatest(timeout time.Duration) (string, error) {
	latestURL := fmt.Sprintf("%s/%s/releases/latest", GitHubURL, Repo)
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(http.MethodGet, latestURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", "nginxpm-cli")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	if resp.StatusCode < 300 || resp.StatusCode > 399 {
		return "", fmt.Errorf("GET %s returned status %d, want a redirect to the latest release", latestURL, resp.StatusCode)
	}
	return tagFromLocation(resp.Header.Get("Location"))
}

// tagFromLocation takes the release version from a latest-release redirect,
// which must point at GitHubURL/<Repo>/releases/tag/v<X.Y.Z>.
func tagFromLocation(location string) (string, error) {
	if location == "" {
		return "", fmt.Errorf("the latest release redirect has no Location header")
	}
	base, err := url.Parse(GitHubURL)
	if err != nil {
		return "", err
	}
	loc, err := url.Parse(location)
	if err != nil || !strings.EqualFold(loc.Scheme, base.Scheme) || !strings.EqualFold(loc.Host, base.Host) || loc.User != nil {
		return "", fmt.Errorf("unexpected latest release location %q", location)
	}
	tag, ok := strings.CutPrefix(loc.Path, "/"+Repo+"/releases/tag/")
	if !ok || !strings.HasPrefix(tag, "v") || !releaseTagPattern.MatchString(tag) {
		return "", fmt.Errorf("latest release location %q does not name a vX.Y.Z release tag", location)
	}
	return trimV(tag), nil
}

// modifyCache runs a read-modify-write of the cache under a cross-process
// lock; fn returns whether to write. With wait=false it gives up when another
// process holds the lock. Errors are ignored: the cache is an optimization.
func modifyCache(configDir string, wait bool, fn func(*cacheEntry) bool) {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return
	}
	path := filepath.Join(configDir, cacheFileName)
	lock := flock.New(path + ".lock")
	if wait {
		if err := lock.Lock(); err != nil {
			return
		}
	} else if ok, err := lock.TryLock(); err != nil || !ok {
		return
	}
	defer lock.Unlock()
	entry, err := readCache(path)
	if err != nil {
		entry = &cacheEntry{}
	}
	if fn(entry) {
		writeCache(path, entry)
	}
}

func readCache(path string) (*cacheEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// writeCache replaces the cache file atomically so concurrent runs never read
// a partial file. Errors are ignored: the cache is an optimization.
func writeCache(path string, entry *cacheEntry) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return
	}
	tmp, err := os.CreateTemp(dir, ".update-check-*.tmp")
	if err != nil {
		return
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return
	}
	if err := tmp.Close(); err != nil {
		return
	}
	_ = os.Rename(tmp.Name(), path)
}

// isNewer returns true if latest is a higher semver than current.
func isNewer(latest, current string) (bool, error) {
	lMajor, lMinor, lPatch, err := parseSemver(trimV(latest))
	if err != nil {
		return false, err
	}
	cMajor, cMinor, cPatch, err := parseSemver(trimV(current))
	if err != nil {
		return false, err
	}
	if lMajor != cMajor {
		return lMajor > cMajor, nil
	}
	if lMinor != cMinor {
		return lMinor > cMinor, nil
	}
	return lPatch > cPatch, nil
}

func parseSemver(v string) (major, minor, patch int, err error) {
	// Strip any pre-release or metadata suffix (e.g. "1.2.3-rc1+build").
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid semver: %s", v)
	}
	if major, err = strconv.Atoi(parts[0]); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}
	if minor, err = strconv.Atoi(parts[1]); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}
	if patch, err = strconv.Atoi(parts[2]); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %s", parts[2])
	}
	return major, minor, patch, nil
}

func trimV(v string) string {
	return strings.TrimPrefix(v, "v")
}
