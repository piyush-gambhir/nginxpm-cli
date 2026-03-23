package dead

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdDeadEnable(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable <id>",
		Short: "Enable a dead host",
		Long: `Enable a 404 dead host by ID, activating its routing rules.

Examples:
  # Enable dead host 3
  nginxpm dead enable 3`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return cmdutil.FlagErrorf("invalid dead host ID: %s", args[0])
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			if err := c.EnableDeadHost(context.Background(), id); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Dead host %d enabled.\n", id)
			}
			return nil
		},
	}

	return cmd
}
