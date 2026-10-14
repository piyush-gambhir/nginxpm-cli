package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/build"
	"github.com/spf13/cobra"
)

func TestCommandsResolve(t *testing.T) {
	// Malformed config and a release build ensure resolution bypasses normal setup.
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	for _, key := range []string{"CI", "NGINXPM_NO_UPDATE_NOTIFIER", "NO_UPDATE_NOTIFIER", "NGINXPM_QUIET"} {
		t.Setenv(key, "")
	}
	configDir := filepath.Join(configHome, "nginxpm-cli")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	originalVersion, originalTTY, originalTransport := build.Version, stderrIsTerminal, http.DefaultTransport
	var networkHits atomic.Int32
	build.Version = "0.1.9"
	stderrIsTerminal = func() bool { return true }
	http.DefaultTransport = resolveTransport(func(*http.Request) (*http.Response, error) {
		networkHits.Add(1)
		return nil, fmt.Errorf("network forbidden during resolution")
	})
	t.Cleanup(func() {
		build.Version, stderrIsTerminal, http.DefaultTransport = originalVersion, originalTTY, originalTransport
	})
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("[invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, command, effect string
		args                  []string
		flags                 map[string]string
		positional            []string
	}{
		{"read", "nginxpm proxy list", "read", []string{"proxy", "list"}, map[string]string{}, []string{}},
		{"write with leading flag", "nginxpm proxy delete", "remote_write", []string{"-o", "json", "proxy", "delete", "1", "--confirm"}, map[string]string{"output": "json", "confirm": "true"}, []string{"1"}},
		{"alias", "nginxpm proxy get", "read", []string{"ph", "get", "2"}, map[string]string{}, []string{"2"}},
		{"target separator", "nginxpm proxy get", "read", []string{"proxy", "get", "--", "-1"}, map[string]string{}, []string{"-1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"-o", "yaml", "commands", "resolve", "--"}, tc.args...)
			stdout, stderr, err := runRoot(t, "", args...)
			if err != nil {
				t.Fatal(err)
			}
			var got resolution
			if err := json.Unmarshal([]byte(stdout), &got); err != nil {
				t.Fatalf("invalid JSON: %v: %s", err, stdout)
			}
			want := resolution{Command: tc.command, Effect: tc.effect, Flags: tc.flags, Args: tc.positional}
			if !reflect.DeepEqual(got, want) || stderr != "" || OutputFormat != "json" {
				t.Fatalf("got %+v, stderr %q, format %q; want %+v", got, stderr, OutputFormat, want)
			}
		})
	}
	for _, args := range [][]string{{"nonsense"}, {}, {"proxy"}, {"proxy", "nonsense"}, {"proxy", "list", "--unknown"}, {"proxy", "list", "-o"}} {
		stdout, _, err := runRoot(t, "", append([]string{"commands", "resolve", "--"}, args...)...)
		if err == nil || stdout != "" || OutputFormat != "json" {
			t.Errorf("%v: stdout %q, error %v, format %q", args, stdout, err, OutputFormat)
		}
	}
	backgroundChecks.Wait()
	if networkHits.Load() != 0 {
		t.Fatal("resolution started a network update check")
	}
}

func TestResolutionAnnotations(t *testing.T) {
	for _, tc := range []struct {
		name        string
		annotations map[string]string
		setFlag     bool
		effect      string
		conditional bool
	}{
		{"read", nil, false, "read", false},
		{"remote", map[string]string{"mutates": "true", "writes-local": "true"}, false, "remote_write", false},
		{"conditional remote", map[string]string{"mutates": "conditional"}, false, "remote_write", true},
		{"remote flag unset", map[string]string{"mutates-when-flag": "write"}, false, "read", false},
		{"remote flag set false", map[string]string{"mutates-when-flag": "write"}, true, "remote_write", false},
		{"local", map[string]string{"writes-local": "true"}, false, "local_write", false},
		{"conditional local", map[string]string{"writes-local": "conditional"}, false, "local_write", true},
		{"self update", map[string]string{"self-update": "true"}, false, "local_write", false},
		{"local flag unset", map[string]string{"writes-local-when-flag": "write"}, false, "read", false},
		{"local flag set false", map[string]string{"writes-local-when-flag": "write"}, true, "local_write", false},
		{"false annotation", map[string]string{"mutates": "false", "writes-local": "false"}, false, "read", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test", Annotations: tc.annotations}
			cmd.Flags().Bool("write", false, "Write changes")
			args := []string{}
			if tc.setFlag {
				args = append(args, "--write=false")
			}
			if err := cmd.ParseFlags(args); err != nil {
				t.Fatal(err)
			}
			got := describeResolution(cmd)
			if got.Effect != tc.effect || got.Conditional != tc.conditional {
				t.Fatalf("got %+v; want %s, conditional=%t", got, tc.effect, tc.conditional)
			}
		})
	}
	cmd := &cobra.Command{Use: "secret", Annotations: map[string]string{"interactive": "true", "secret-output": "true"}}
	got := describeResolution(cmd)
	if !got.Interactive || !got.Secret {
		t.Fatalf("missing interaction or secret annotations: %+v", got)
	}
}

type resolveTransport func(*http.Request) (*http.Response, error)

func (f resolveTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
