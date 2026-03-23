package cert

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdCertList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List certificates",
		Long: `List all SSL certificates.

Shows ID, nice name, provider, domains, and expiry date.

Examples:
  # List all certificates
  nginxpm cert list

  # Output as JSON
  nginxpm cert list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			certs, err := c.ListCertificates(context.Background())
			if err != nil {
				return err
			}

			if len(certs) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No certificates found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, certs, &output.TableDef{
				Headers: []string{"ID", "NICE NAME", "PROVIDER", "DOMAINS", "EXPIRES"},
				RowFunc: func(item interface{}) []string {
					cert := item.(client.Certificate)
					domains := strings.Join(cert.DomainNames, ", ")
					expires := "N/A"
					if cert.ExpiresOn != "" {
						expires = cert.ExpiresOn
					}
					return []string{
						fmt.Sprintf("%d", cert.ID),
						cert.NiceName,
						cert.Provider,
						domains,
						expires,
					}
				},
			})
		},
	}
}
