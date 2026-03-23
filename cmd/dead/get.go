package dead

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdDeadGet(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a dead host by ID",
		Long: `Get the full details of a 404 dead host by its numeric ID.

Examples:
  # Get dead host details
  nginxpm dead get 3

  # Output as JSON
  nginxpm dead get 3 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return cmdutil.FlagErrorf("invalid dead host ID: %s", args[0])
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			host, err := c.GetDeadHost(context.Background(), id)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, host, nil)
		},
	}

	return cmd
}
