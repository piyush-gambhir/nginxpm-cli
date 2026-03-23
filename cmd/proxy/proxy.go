package proxy

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

// NewCmdProxy returns the proxy parent command.
func NewCmdProxy(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "proxy",
		Short:   "Manage proxy hosts",
		Long:    "List, create, update, delete, enable, and disable proxy hosts.",
		Aliases: []string{"proxy-host", "ph"},
	}

	cmd.AddCommand(newCmdProxyList(f))
	cmd.AddCommand(newCmdProxyGet(f))
	cmd.AddCommand(newCmdProxyCreate(f))
	cmd.AddCommand(newCmdProxyUpdate(f))
	cmd.AddCommand(newCmdProxyDelete(f))
	cmd.AddCommand(newCmdProxyEnable(f))
	cmd.AddCommand(newCmdProxyDisable(f))

	return cmd
}
