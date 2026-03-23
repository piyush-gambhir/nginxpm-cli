package setting

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdSettingList(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all settings",
		Long: `List all server settings.

Shows ID and value for each setting.

Examples:
  # List all settings
  nginxpm setting list

  # Output as JSON
  nginxpm setting list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			settings, err := c.ListSettings(context.Background())
			if err != nil {
				return err
			}

			if len(settings) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No settings found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, settings, &output.TableDef{
				Headers: []string{"ID", "VALUE"},
				RowFunc: func(item interface{}) []string {
					s := item.(client.Setting)
					return []string{
						s.ID,
						fmt.Sprintf("%v", s.Value),
					}
				},
			})
		},
	}
}
