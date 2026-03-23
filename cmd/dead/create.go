package dead

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdDeadCreate(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a dead host",
		Long: `Create a new 404 dead host from a JSON or YAML definition file.

Examples:
  # Create a dead host from a JSON file
  nginxpm dead create -f dead.json

  # Create from stdin
  cat dead.yaml | nginxpm dead create -f -`,
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

			host, err := c.CreateDeadHost(context.Background(), body)
			if err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Dead host %d created.\n", host.ID)
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, host, nil)
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
