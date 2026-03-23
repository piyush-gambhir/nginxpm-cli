package redirect

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

// NewCmdRedirect returns the redirect parent command.
func NewCmdRedirect(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "redirect",
		Short:   "Manage redirection hosts",
		Long:    "List, create, update, delete, enable, and disable redirection hosts.",
		Aliases: []string{"redirection", "redir", "rh"},
	}

	cmd.AddCommand(newCmdRedirectList(f))
	cmd.AddCommand(newCmdRedirectGet(f))
	cmd.AddCommand(newCmdRedirectCreate(f))
	cmd.AddCommand(newCmdRedirectUpdate(f))
	cmd.AddCommand(newCmdRedirectDelete(f))
	cmd.AddCommand(newCmdRedirectEnable(f))
	cmd.AddCommand(newCmdRedirectDisable(f))

	return cmd
}
