package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdUserList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List users",
		Long: `List all users.

Shows ID, name, nickname, email, roles, and disabled status.

Examples:
  # List all users
  nginxpm user list

  # Output as JSON
  nginxpm user list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			users, err := c.ListUsers(context.Background())
			if err != nil {
				return err
			}

			if len(users) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No users found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, users, &output.TableDef{
				Headers: []string{"ID", "NAME", "NICKNAME", "EMAIL", "ROLES", "DISABLED"},
				RowFunc: func(item interface{}) []string {
					u := item.(client.User)
					roles := strings.Join(u.Roles, ", ")
					disabled := "no"
					if u.IsDisabled {
						disabled = "yes"
					}
					return []string{
						fmt.Sprintf("%d", u.ID),
						u.Name,
						u.Nickname,
						u.Email,
						roles,
						disabled,
					}
				},
			})
		},
	}
}
