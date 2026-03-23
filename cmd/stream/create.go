package stream

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdStreamCreate(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a stream",
		Long: `Create a new TCP/UDP stream from a JSON or YAML definition file.

Examples:
  # Create a stream from a JSON file
  nginxpm stream create -f stream.json

  # Create from stdin
  cat stream.yaml | nginxpm stream create -f -`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			stream, err := c.CreateStream(context.Background(), body)
			if err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Stream %d created.\n", stream.ID)
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, stream, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
