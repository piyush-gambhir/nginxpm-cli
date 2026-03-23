package user

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

// NewCmdUser returns the user parent command.
func NewCmdUser(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "user",
		Short:   "Manage users",
		Long:    "List, create, update, delete, and manage users.",
		Aliases: []string{"users"},
	}

	cmd.AddCommand(newCmdUserList(f))
	cmd.AddCommand(newCmdUserGet(f))
	cmd.AddCommand(newCmdUserCreate(f))
	cmd.AddCommand(newCmdUserUpdate(f))
	cmd.AddCommand(newCmdUserDelete(f))
	cmd.AddCommand(newCmdUserCurrent(f))
	cmd.AddCommand(newCmdUserPermissions(f))
	cmd.AddCommand(newCmdUserPassword(f))

	return cmd
}
