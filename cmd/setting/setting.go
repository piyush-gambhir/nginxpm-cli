package setting

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

// NewCmdSetting returns the setting parent command.
func NewCmdSetting(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "setting",
		Short:   "Manage server settings",
		Long:    "View and update Nginx Proxy Manager server settings (e.g., default site behavior).",
		Aliases: []string{"settings"},
	}

	cmd.AddCommand(newCmdSettingList(f))
	cmd.AddCommand(newCmdSettingGet(f))
	cmd.AddCommand(newCmdSettingUpdate(f))

	return cmd
}
