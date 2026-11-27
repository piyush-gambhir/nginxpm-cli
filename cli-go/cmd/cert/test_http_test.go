package cert

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/config"
)

func runTestHTTP(t *testing.T, format, response string, args ...string) (string, error) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/nginx/certificates/test-http" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(srv.Close)
	var out bytes.Buffer
	f := &cmdutil.Factory{
		IOStreams: cmdutil.IOStreams{In: strings.NewReader(""), Out: &out, ErrOut: &bytes.Buffer{}},
		Resolved:  &config.ResolvedConfig{Output: format},
		Client:    func() (*client.Client, error) { return client.NewClientWithToken(srv.URL, false), nil },
	}
	cmd := newCmdCertTestHTTP(f)
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	return out.String(), err
}

func TestCertTestHTTPFailsWhenADomainIsNotOK(t *testing.T) {
	out, err := runTestHTTP(t, "table", `{"good.example.com":"ok","bad.example.com":"no-host"}`, "good.example.com", "bad.example.com")
	if err == nil || !strings.Contains(err.Error(), "bad.example.com (no-host)") || strings.Contains(err.Error(), "good.example.com") || strings.Contains(out, "passed") {
		t.Fatalf("out %q, err %v", out, err)
	}
	out, err = runTestHTTP(t, "json", `{"bad.example.com":"404"}`, "bad.example.com")
	if err == nil || !strings.Contains(out, `"bad.example.com": "404"`) {
		t.Fatalf("json: out %q, err %v", out, err)
	}
	out, err = runTestHTTP(t, "table", `{"good.example.com":"ok"}`, "good.example.com")
	if err != nil || !strings.Contains(out, "HTTP reachability test passed.") {
		t.Fatalf("ok: out %q, err %v", out, err)
	}
}
