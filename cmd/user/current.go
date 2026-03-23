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

func newCmdUserCurrent(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current authenticated user",
		Long: `Show the currently authenticated user.

Examples:
  # Show current user
  nginxpm user current

  # Output as JSON
  nginxpm user current -o json`,
		Aliases: []string{"me", "whoami"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			u, err := c.GetUser(context.Background(), "me")
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
