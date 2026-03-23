package audit

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

// NewCmdAudit returns the audit parent command.
func NewCmdAudit(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "audit",
		Short:   "View audit log",
		Long:    "View the audit log of actions performed in Nginx Proxy Manager.",
		Aliases: []string{"audit-log", "log"},
	}

	cmd.AddCommand(newCmdAuditList(f))
	cmd.AddCommand(newCmdAuditGet(f))

	return cmd
}
