package redirect

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdRedirectDisable(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <id>",
		Short: "Disable a redirection host",
		Long: `Disable a redirection host by ID.

Examples:
  # Disable redirection host 1
  nginxpm redirect disable 1`,
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

			if err := c.DisableRedirectHost(context.Background(), id); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Redirection host %d disabled.\n", id)
			}
			return nil
		},
	}
}
