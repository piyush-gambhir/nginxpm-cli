package access

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

// NewCmdAccess returns the access list parent command.
func NewCmdAccess(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "access",
		Short:   "Manage access lists",
		Long:    "List, create, update, and delete access lists.",
		Aliases: []string{"access-list", "acl"},
	}

	cmd.AddCommand(newCmdAccessList(f))
	cmd.AddCommand(newCmdAccessGet(f))
	cmd.AddCommand(newCmdAccessCreate(f))
	cmd.AddCommand(newCmdAccessUpdate(f))
	cmd.AddCommand(newCmdAccessDelete(f))

	return cmd
}
