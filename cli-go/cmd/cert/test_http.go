package cert

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/output"
)

func newCmdCertTestHTTP(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "test-http <domain> [domain...]",
		Short: "Test HTTP reachability for domains",
		Long: `Test HTTP reachability for one or more domain names.

The server reports a result per domain: ok, no-host, 404, wrong-data, failed,
or other:<detail>. The command exits non-zero and names each domain whose
result is not ok. With -o json or -o yaml it prints the per-domain results.

Examples:
  # Test a single domain
  nginxpm cert test-http example.com

  # Test multiple domains
  nginxpm cert test-http example.com www.example.com`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]interface{}{
				"domains": args,
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			results, err := c.TestHTTP(cmd.Context(), body)
			if err != nil {
				return err
			}

			var failed []string
			for _, domain := range args {
				if result := results[domain]; result != "ok" {
					if result == "" {
						result = "no result"
					}
					failed = append(failed, fmt.Sprintf("%s (%s)", domain, result))
				}
			}

			if f.Resolved.Output != "table" {
				if err := output.Print(f.IOStreams.Out, f.Resolved.Output, results, nil); err != nil {
					return err
				}
			}
			if len(failed) > 0 {
				return fmt.Errorf("HTTP reachability test failed: %s", strings.Join(failed, ", "))
			}
			if !f.Quiet && f.Resolved.Output == "table" {
				fmt.Fprintln(f.IOStreams.Out, "HTTP reachability test passed.")
			}
			return nil
		},
	}
}
