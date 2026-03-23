package setting

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdSettingUpdate(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a setting",
		Long: `Update a server setting from a JSON or YAML file.

Examples:
  # Update the default-site setting from a JSON file
  nginxpm setting update default-site -f setting.json

  # Update from stdin
  nginxpm setting update default-site -f -`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if _, err := c.UpdateSetting(context.Background(), args[0], body); err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Setting %q updated.\n", args[0])
			}
			return nil
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
