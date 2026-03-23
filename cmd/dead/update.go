package dead

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdDeadUpdate(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a dead host",
		Long: `Update an existing 404 dead host by ID from a JSON or YAML definition file.

Examples:
  # Update dead host 3 from a file
  nginxpm dead update 3 -f dead.json

  # Update from stdin
  cat dead.yaml | nginxpm dead update 3 -f -`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return cmdutil.FlagErrorf("invalid dead host ID: %s", args[0])
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

			host, err := c.UpdateDeadHost(context.Background(), id, body)
			if err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Dead host %d updated.\n", id)
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, host, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
