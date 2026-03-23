package access

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdAccessGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get an access list",
		Long: `Get an access list by ID.

Examples:
  # Get access list details
  nginxpm access get 1

  # Output as JSON
  nginxpm access get 1 -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid ID %q: %w", args[0], err)
			}

			c, err := f.Client()
			if err != nil {
				return err
			}

			al, err := c.GetAccessList(context.Background(), id)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, al, &output.TableDef{
				Headers: []string{"ID", "NAME", "SATISFY ANY", "PASS AUTH"},
				RowFunc: func(item interface{}) []string {
					a := item.(*client.AccessList)
					satisfyAny := "no"
					if a.SatisfyAny {
						satisfyAny = "yes"
					}
					passAuth := "no"
					if a.PassAuth {
						passAuth = "yes"
					}
					return []string{
						fmt.Sprintf("%d", a.ID),
						a.Name,
						satisfyAny,
						passAuth,
					}
				},
			})
		},
	}
}
