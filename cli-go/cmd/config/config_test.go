package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/cmdutil"
)

func TestConfigViewMasksPasswords(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", home)
	dir := filepath.Join(home, "nginxpm-cli")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := "current_profile: prod\nprofiles:\n  prod:\n    url: https://npm.example.com\n    email: admin@example.com\n    password: s3cret\n    read_only: true\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	cmd := NewCmdConfig(&cmdutil.Factory{IOStreams: cmdutil.IOStreams{In: strings.NewReader(""), Out: &out, ErrOut: &bytes.Buffer{}}})
	cmd.SetArgs([]string{"view"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "s3cret") || !strings.Contains(out.String(), "password: '****'") || !strings.Contains(out.String(), "email: admin@example.com") {
		t.Fatalf("config view output:\n%s", out.String())
	}
	saved, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil || !strings.Contains(string(saved), "s3cret") {
		t.Fatalf("config view must not change the file: %v\n%s", err, saved)
	}
}
