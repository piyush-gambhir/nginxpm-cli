package cert

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdCertDNSProviders(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "dns-providers",
		Short: "List DNS challenge providers",
		Long: `List available DNS challenge providers for Let's Encrypt certificates.

Examples:
  # List DNS providers
  nginxpm cert dns-providers`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			providers, err := c.ListDNSProviders(context.Background())
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, "json", providers, nil)
		},
	}
}
