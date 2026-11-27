package audit

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/config"
)

func TestAuditListNotesTheServerCap(t *testing.T) {
	for _, n := range []int{100, 3} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rows := make([]string, n)
			for i := range rows {
				rows[i] = fmt.Sprintf(`{"id":%d,"action":"updated"}`, n-i)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("[" + strings.Join(rows, ",") + "]"))
		}))
		var out, errOut bytes.Buffer
		f := &cmdutil.Factory{
			IOStreams: cmdutil.IOStreams{In: strings.NewReader(""), Out: &out, ErrOut: &errOut},
			Resolved:  &config.ResolvedConfig{Output: "json"},
			Client:    func() (*client.Client, error) { return client.NewClientWithToken(srv.URL, false), nil },
		}
		cmd := newCmdAuditList(f)
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		srv.Close()
		noted := strings.Contains(errOut.String(), "at most the 100 newest audit entries")
		if err != nil || noted != (n == 100) || strings.Contains(out.String(), "Note:") {
			t.Fatalf("%d rows: err %v, stderr %q", n, err, errOut.String())
		}
	}
}
