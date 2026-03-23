package dead

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdDeadList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all dead hosts",
		Long: `List all 404 dead host entries.

Shows domains, SSL status, and enabled status.

Examples:
  # List all dead hosts
  nginxpm dead list

  # Output as JSON
  nginxpm dead list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			hosts, err := c.ListDeadHosts(context.Background())
			if err != nil {
				return err
			}

			if len(hosts) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No dead hosts found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, hosts, &output.TableDef{
				Headers: []string{"ID", "DOMAINS", "SSL", "ENABLED"},
				RowFunc: func(item interface{}) []string {
					h := item.(client.DeadHost)
					ssl := "no"
					if h.SSLForced {
						ssl = "yes"
					}
					enabled := "no"
					if h.Enabled {
						enabled = "yes"
					}
					return []string{
						fmt.Sprintf("%d", h.ID),
						strings.Join(h.DomainNames, ", "),
						ssl,
						enabled,
					}
				},
			})
		},
	}

	return cmd
}
