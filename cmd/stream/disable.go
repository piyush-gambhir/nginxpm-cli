package stream

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdStreamDisable(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable <id>",
		Short: "Disable a stream",
		Long: `Disable a TCP/UDP stream by ID, deactivating its forwarding rules.

Examples:
  # Disable stream 5
  nginxpm stream disable 5`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return cmdutil.FlagErrorf("invalid stream ID: %s", args[0])
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			if err := c.DisableStream(context.Background(), id); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Stream %d disabled.\n", id)
			}
			return nil
		},
	}

	return cmd
}
