package stream

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdStreamUpdate(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a stream",
		Long: `Update an existing TCP/UDP stream by ID from a JSON or YAML definition file.

Examples:
  # Update stream 5 from a file
  nginxpm stream update 5 -f stream.json

  # Update from stdin
  cat stream.yaml | nginxpm stream update 5 -f -`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return cmdutil.FlagErrorf("invalid stream ID: %s", args[0])
			}

			if file == "" {
				return cmdutil.FlagErrorf("file is required (-f)")
			}

			var body interface{}
			if err := cmdutil.UnmarshalInput(file, &body); err != nil {
				return err
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			stream, err := c.UpdateStream(context.Background(), id, body)
			if err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Stream %d updated.\n", id)
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, stream, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
