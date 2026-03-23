package stream

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdStreamGet(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Get a stream by ID",
		Long: `Get the full details of a TCP/UDP stream by its numeric ID.

Examples:
  # Get stream details
  nginxpm stream get 5

  # Output as JSON
  nginxpm stream get 5 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return cmdutil.FlagErrorf("invalid stream ID: %s", args[0])
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			stream, err := c.GetStream(context.Background(), id)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, stream, nil)
		},
	}

	return cmd
}
