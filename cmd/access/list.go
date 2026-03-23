package access

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdAccessList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List access lists",
		Long: `List all access lists.

Shows ID, name, satisfy any, and pass auth status.

Examples:
  # List all access lists
  nginxpm access list

  # Output as JSON
  nginxpm access list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			lists, err := c.ListAccessLists(context.Background())
			if err != nil {
				return err
			}

			if len(lists) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No access lists found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, lists, &output.TableDef{
				Headers: []string{"ID", "NAME", "SATISFY ANY", "PASS AUTH"},
				RowFunc: func(item interface{}) []string {
					al := item.(client.AccessList)
					satisfyAny := "no"
					if al.SatisfyAny {
						satisfyAny = "yes"
					}
					passAuth := "no"
					if al.PassAuth {
						passAuth = "yes"
					}
					return []string{
						fmt.Sprintf("%d", al.ID),
						al.Name,
						satisfyAny,
						passAuth,
					}
				},
			})
		},
	}
}
