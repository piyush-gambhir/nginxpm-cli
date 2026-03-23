package cert

import (
	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

// NewCmdCert returns the cert parent command.
func NewCmdCert(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cert",
		Short:   "Manage SSL certificates",
		Long:    "List, create, delete, and renew SSL certificates.",
		Aliases: []string{"certificate", "certificates", "ssl"},
	}

	cmd.AddCommand(newCmdCertList(f))
	cmd.AddCommand(newCmdCertGet(f))
	cmd.AddCommand(newCmdCertCreate(f))
	cmd.AddCommand(newCmdCertDelete(f))
	cmd.AddCommand(newCmdCertRenew(f))
	cmd.AddCommand(newCmdCertDNSProviders(f))
	cmd.AddCommand(newCmdCertTestHTTP(f))

	return cmd
}
