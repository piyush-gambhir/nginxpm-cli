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

func newCmdUserGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a user",
		Long: `Get a user by ID.

The ID can be a numeric user ID or "me" for the current user.

Examples:
  # Get user details
  nginxpm user get 1

  # Get current user
  nginxpm user get me

  # Output as JSON
  nginxpm user get 1 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			u, err := c.GetUser(context.Background(), args[0])
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, u, &output.TableDef{
				Headers: []string{"ID", "NAME", "NICKNAME", "EMAIL", "ROLES", "DISABLED"},
				RowFunc: func(item interface{}) []string {
					usr := item.(*client.User)
					roles := strings.Join(usr.Roles, ", ")
					disabled := "no"
					if usr.IsDisabled {
						disabled = "yes"
					}
					return []string{
						fmt.Sprintf("%d", usr.ID),
						usr.Name,
						usr.Nickname,
						usr.Email,
						roles,
						disabled,
					}
				},
			})
		},
	}
}
