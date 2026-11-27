package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/config"
)

func TestLoginKeepsProfileReadOnly(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("CI", "1")
	for _, key := range []string{"NGINXPM_URL", "NGINXPM_EMAIL", "NGINXPM_PASSWORD", "NGINXPM_NO_INPUT", "NGINXPM_READ_ONLY"} {
		t.Setenv(key, "")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/":
			_, _ = w.Write([]byte(`{"status":"OK","setup":true}`))
		case "/api/tokens":
			_, _ = w.Write([]byte(`{"token":"t"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	dir := filepath.Join(home, "nginxpm-cli")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := "current_profile: prod\nprofiles:\n  prod:\n    url: " + srv.URL + "\n    email: old@example.com\n    password: old\n    read_only: true\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	origIn, origOut := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = stdinR, devNull
	t.Cleanup(func() { os.Stdin, os.Stdout = origIn, origOut; devNull.Close() })
	if _, err := stdinW.WriteString(srv.URL + "\nadmin@example.com\nnew-secret\nn\nprod\n"); err != nil {
		t.Fatal(err)
	}
	stdinW.Close()

	if _, _, err := runRoot(t, "", "login"); err != nil {
		t.Fatal(err)
	}
	saved, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	p := saved.Profiles["prod"]
	if !p.ReadOnly || p.Password != "new-secret" || p.Email != "admin@example.com" {
		t.Fatalf("profile after login: %+v", p)
	}
}
