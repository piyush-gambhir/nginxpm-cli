package audit

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdAuditGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an audit log entry",
		Long: `Get a specific audit log entry by ID.

Examples:
  # Get audit log entry details
  nginxpm audit get 1

  # Output as JSON
  nginxpm audit get 1 -o json`,
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

			entry, err := c.GetAuditEntry(context.Background(), id)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, entry, &output.TableDef{
				Headers: []string{"ID", "DATE", "USER ID", "ACTION", "OBJECT TYPE", "OBJECT ID"},
				RowFunc: func(item interface{}) []string {
					e := item.(*client.AuditEntry)
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
