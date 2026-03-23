package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdUserPermissions(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "permissions <id>",
		Short: "Set user permissions",
		Long: `Set permissions for a user from a JSON or YAML file.

Examples:
  # Set permissions for user 1
  nginxpm user permissions 1 -f permissions.json

  # Set from stdin
  nginxpm user permissions 1 -f -`,
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

			var body map[string]interface{}
			if err := cmdutil.UnmarshalInput(file, &body); err != nil {
				return err
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			if err := c.SetUserPermissions(context.Background(), id, body); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Permissions updated for user %d.\n", id)
			}
			return nil
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
