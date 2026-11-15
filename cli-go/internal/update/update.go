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
	Repo = "piyush-gambhir/nginxpm-cli"

	cacheDuration  = 24 * time.Hour
	noticeInterval = 24 * time.Hour
	cacheFileName  = "update-check.json"

	BackgroundTimeout = 3 * time.Second
	ForegroundTimeout = 15 * time.Second

	SelfUpdateCommand   = "nginxpm update"
	SourceUpdateCommand = "git pull && make install (in your nginxpm-cli/cli-go clone)"
)

var GitHubURL = "https://github.com"

var now = time.Now

var semverPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+([-+][0-9A-Za-z.+-]+)?$`)

var releaseTagPattern = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)

type UpdateInfo struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
}

type cacheEntry struct {
	LastChecked     string `json:"last_checked"`
	LatestVersion   string `json:"latest_version,omitempty"`
	Error           string `json:"error,omitempty"`
	NotifiedVersion string `json:"notified_version,omitempty"`
	NotifiedAt      string `json:"notified_at,omitempty"`
}

func IsReleaseVersion(v string) bool {
	return semverPattern.MatchString(v)
}

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

func ReleaseURL(version string) string {
	return fmt.Sprintf("https://github.com/%s/releases/tag/v%s", Repo, trimV(version))
}

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

func CheckForUpdate(currentVersion, configDir string) *UpdateInfo {
	if info, fresh := CachedResult(currentVersion, configDir); fresh {
		return info
	}
	modifyCache(configDir, true, func(entry *cacheEntry) bool {
		entry.LastChecked = now().UTC().Format(time.RFC3339)
		return true
	})
	latest, err := fetchLatest(BackgroundTimeout)
	recordCheck(configDir, latest, err)
	return newInfo(currentVersion, latest)
}

func CheckForUpdateFresh(currentVersion, configDir string) (*UpdateInfo, error) {
	latest, err := fetchLatest(ForegroundTimeout)
	if err != nil {
		return nil, err
	}
	recordCheck(configDir, latest, nil)
	return newInfo(currentVersion, latest), nil
}

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

func CachedInfo(currentVersion, configDir string) *UpdateInfo {
	entry, err := readCache(filepath.Join(configDir, cacheFileName))
	if err != nil {
		return nil
	}
	return newInfo(currentVersion, entry.LatestVersion)
}

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

func PrintNotice(w io.Writer, info *UpdateInfo, goInstall bool) {
	fmt.Fprintf(w, "\nA new version of nginxpm is available: v%s -> v%s\n", trimV(info.CurrentVersion), trimV(info.LatestVersion))
	fmt.Fprintf(w, "Update with: %s\n", UpdateCommand(goInstall))
	fmt.Fprintf(w, "Release notes: %s\n", info.ReleaseURL)
}

func UpdateCommand(goInstall bool) string {
	if goInstall {
		return SourceUpdateCommand
	}
	return SelfUpdateCommand
}

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
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
	default:
		return "", fmt.Errorf("GET %s returned status %d, want a redirect to the latest release", latestURL, resp.StatusCode)
	}
	return tagFromLocation(resp.Header.Get("Location"))
}

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
