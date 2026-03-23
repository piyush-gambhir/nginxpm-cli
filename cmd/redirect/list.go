package redirect

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdRedirectList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List redirection hosts",
		Long: `List all redirection hosts.

Shows ID, domains, forward target, HTTP code, preserve path, SSL status, and enabled state.

Examples:
  # List all redirection hosts
  nginxpm redirect list

  # Output as JSON
  nginxpm redirect list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			hosts, err := c.ListRedirectHosts(context.Background())
			if err != nil {
				return err
			}

			if len(hosts) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No redirection hosts found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, hosts, &output.TableDef{
				Headers: []string{"ID", "DOMAINS", "FORWARD TO", "HTTP CODE", "PRESERVE PATH", "SSL", "ENABLED"},
				RowFunc: func(item interface{}) []string {
					h := item.(client.RedirectHost)
					domains := strings.Join(h.DomainNames, ", ")
					forwardTo := fmt.Sprintf("%s://%s", h.ForwardScheme, h.ForwardDomainName)
					httpCode := fmt.Sprintf("%d", h.ForwardHTTPCode)
					preservePath := "no"
					if h.PreservePath {
						preservePath = "yes"
					}
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
						forwardTo,
						httpCode,
						preservePath,
						ssl,
						enabled,
					}
				},
			})
		},
	}
}
