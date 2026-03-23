package redirect

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

func newCmdRedirectGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a redirection host",
		Long: `Get a redirection host by ID.

Examples:
  # Get redirection host details
  nginxpm redirect get 1

  # Output as JSON
  nginxpm redirect get 1 -o json`,
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

			host, err := c.GetRedirectHost(context.Background(), id)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, host, &output.TableDef{
				Headers: []string{"ID", "DOMAINS", "FORWARD TO", "HTTP CODE", "PRESERVE PATH", "SSL", "ENABLED"},
				RowFunc: func(item interface{}) []string {
					h := item.(*client.RedirectHost)
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
