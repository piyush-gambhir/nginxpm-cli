package dead

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

// NewCmdDead returns the dead host parent command.
func NewCmdDead(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dead",
		Short:   "Manage 404 dead hosts",
		Long:    "List, create, update, delete, enable, and disable dead hosts that return 404.",
		Aliases: []string{"dead-host", "dh", "404"},
	}

	cmd.AddCommand(newCmdDeadList(f))
	cmd.AddCommand(newCmdDeadGet(f))
	cmd.AddCommand(newCmdDeadCreate(f))
	cmd.AddCommand(newCmdDeadUpdate(f))
	cmd.AddCommand(newCmdDeadDelete(f))
	cmd.AddCommand(newCmdDeadEnable(f))
	cmd.AddCommand(newCmdDeadDisable(f))

	return cmd
}
