package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdUserPassword(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "password <id>",
		Short: "Change user password",
		Long: `Change a user's password from a JSON or YAML file.

The file should contain the password change payload with type, current, and secret fields.

Examples:
  # Change password for user 1
  nginxpm user password 1 -f password.json

  # Change from stdin
  nginxpm user password 1 -f -`,
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

			if err := c.ChangePassword(context.Background(), id, body); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Password changed for user %d.\n", id)
			}
			return nil
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
