package redirect

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdRedirectUpdate(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a redirection host",
		Long: `Update an existing redirection host from a JSON or YAML file.

Examples:
  # Update redirection host 1 from a JSON file
  nginxpm redirect update 1 -f redirect.json

  # Update from stdin
  nginxpm redirect update 1 -f -`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ID %q: %w", args[0], err)
			}

			if file == "" {
				return cmdutil.FlagErrorf("required flag --file (-f) not set")
			}

			var body interface{}
			if err := cmdutil.UnmarshalInput(file, &body); err != nil {
				return err
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			if _, err := c.UpdateRedirectHost(context.Background(), id, body); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Redirection host %d updated.\n", id)
			}
			return nil
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
