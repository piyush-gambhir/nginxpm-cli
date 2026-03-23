package proxy

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdProxyGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a proxy host",
		Long: `Get a proxy host by ID.

Examples:
  # Get proxy host details
  nginxpm proxy get 1

  # Output as JSON
  nginxpm proxy get 1 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ID %q: %w", args[0], err)
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			host, err := c.GetProxyHost(context.Background(), id)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, host, &output.TableDef{
				Headers: []string{"ID", "DOMAINS", "FORWARD", "SSL", "ENABLED"},
				RowFunc: func(item interface{}) []string {
					h := item.(*client.ProxyHost)
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
