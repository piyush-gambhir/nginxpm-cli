package setting

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdSettingGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a setting",
		Long: `Get a specific setting by ID.

Examples:
  # Get the default-site setting
  nginxpm setting get default-site

  # Output as JSON
  nginxpm setting get default-site -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			setting, err := c.GetSetting(context.Background(), args[0])
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, setting, &output.TableDef{
				Headers: []string{"ID", "VALUE"},
				RowFunc: func(item interface{}) []string {
					s := item.(*client.Setting)
					return []string{
						s.ID,
						fmt.Sprintf("%v", s.Value),
					}
				},
			})
		},
	}
}
