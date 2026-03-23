package stream

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdStreamEnable(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <id>",
		Short: "Enable a stream",
		Long: `Enable a TCP/UDP stream by ID, activating its forwarding rules.

Examples:
  # Enable stream 5
  nginxpm stream enable 5`,
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

			if err := c.EnableStream(context.Background(), id); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Stream %d enabled.\n", id)
			}
			return nil
		},
	}

	return cmd
}
