package stream

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

// NewCmdStream returns the stream parent command.
func NewCmdStream(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stream",
		Short:   "Manage TCP/UDP streams",
		Long:    "List, create, update, delete, enable, and disable TCP/UDP stream forwarding.",
		Aliases: []string{"streams", "st"},
	}

	cmd.AddCommand(newCmdStreamList(f))
	cmd.AddCommand(newCmdStreamGet(f))
	cmd.AddCommand(newCmdStreamCreate(f))
	cmd.AddCommand(newCmdStreamUpdate(f))
	cmd.AddCommand(newCmdStreamDelete(f))
	cmd.AddCommand(newCmdStreamEnable(f))
	cmd.AddCommand(newCmdStreamDisable(f))

	return cmd
}
