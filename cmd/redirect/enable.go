package redirect

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdRedirectEnable(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "enable <id>",
		Short: "Enable a redirection host",
		Long: `Enable a redirection host by ID.

Examples:
  # Enable redirection host 1
  nginxpm redirect enable 1`,
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

			if err := c.EnableRedirectHost(context.Background(), id); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Redirection host %d enabled.\n", id)
			}
			return nil
		},
	}
}
