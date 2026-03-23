package cert

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdCertTestHTTP(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "test-http <domain> [domain...]",
		Short: "Test HTTP reachability for domains",
		Long: `Test HTTP reachability for one or more domain names.

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

			if err := c.TestHTTP(context.Background(), body); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintln(f.IOStreams.Out, "HTTP reachability test passed.")
			}
			return nil
		},
	}
}
