package audit

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdAuditList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List audit log entries",
		Long: `List all audit log entries.

Shows ID, date, user ID, action, object type, and object ID.

Examples:
  # List all audit log entries
  nginxpm audit list

  # Output as JSON
  nginxpm audit list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			entries, err := c.ListAuditLog(context.Background())
			if err != nil {
				return err
			}

			if len(entries) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No audit log entries found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, entries, &output.TableDef{
				Headers: []string{"ID", "DATE", "USER ID", "ACTION", "OBJECT TYPE", "OBJECT ID"},
				RowFunc: func(item interface{}) []string {
					e := item.(client.AuditEntry)
					return []string{
						fmt.Sprintf("%d", e.ID),
						e.CreatedOn,
						fmt.Sprintf("%d", e.UserID),
						e.Action,
						e.ObjectType,
						fmt.Sprintf("%d", e.ObjectID),
					}
				},
			})
		},
	}
}
