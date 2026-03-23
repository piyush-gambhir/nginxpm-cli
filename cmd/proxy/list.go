package proxy

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdProxyList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List proxy hosts",
		Long: `List all proxy hosts.

Shows ID, domains, forward target, SSL status, and enabled state.

Examples:
  # List all proxy hosts
  nginxpm proxy list

  # Output as JSON
  nginxpm proxy list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			hosts, err := c.ListProxyHosts(context.Background())
			if err != nil {
				return err
			}

			if len(hosts) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No proxy hosts found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, hosts, &output.TableDef{
				Headers: []string{"ID", "DOMAINS", "FORWARD", "SSL", "ENABLED"},
				RowFunc: func(item interface{}) []string {
					h := item.(client.ProxyHost)
					domains := strings.Join(h.DomainNames, ", ")
					forward := fmt.Sprintf("%s://%s:%d", h.ForwardScheme, h.ForwardHost, h.ForwardPort)
					ssl := "none"
					if h.CertificateID != nil && h.CertificateID != 0 && h.CertificateID != float64(0) {
						ssl = fmt.Sprintf("%v", h.CertificateID)
					}
					enabled := "no"
					if h.Enabled {
						enabled = "yes"
					}
					return []string{
						fmt.Sprintf("%d", h.ID),
						domains,
						forward,
						ssl,
						enabled,
					}
				},
			})
		},
	}
}
