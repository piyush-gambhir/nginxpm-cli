package cert

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdCertRenew(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "renew <id>",
		Short: "Renew a Let's Encrypt certificate",
		Long: `Renew a Let's Encrypt certificate by ID.

Examples:
  # Renew certificate
  nginxpm cert renew 1`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ID %q: %w", args[0], err)
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			if err := c.RenewCertificate(context.Background(), id); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Certificate %d renewed.\n", id)
			}
			return nil
		},
	}
}
