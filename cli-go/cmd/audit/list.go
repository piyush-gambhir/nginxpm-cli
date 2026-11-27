package audit

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/output"
)

const auditListCap = 100

func newCmdAuditList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List audit log entries",
		Long: `List the most recent audit log entries.

Nginx Proxy Manager returns at most the 100 newest entries; use "audit get <id>"
for older ones. Shows ID, date, user ID, action, object type, and object ID.

Examples:
  # List the most recent audit log entries
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

			entries, err := c.ListAuditLog(cmd.Context())
			if err != nil {
				return err
			}

			if !f.Quiet && len(entries) >= auditListCap {
				fmt.Fprintf(f.IOStreams.ErrOut, "Note: Nginx Proxy Manager returns at most the %d newest audit entries; older entries may exist (use audit get <id>).\n", auditListCap)
			}

			if len(entries) == 0 && f.Resolved.Output == "table" {
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
