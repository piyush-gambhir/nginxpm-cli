package stream

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdStreamList(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all streams",
		Long: `List all TCP/UDP stream forwarding entries.

Shows incoming port, forwarding destination, protocol flags, and enabled status.

Examples:
  # List all streams
  nginxpm stream list

  # Output as JSON
  nginxpm stream list -o json`,
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := f.Client()
			if err != nil {
				return err
			}

			streams, err := c.ListStreams(context.Background())
			if err != nil {
				return err
			}

			if len(streams) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No streams found.")
				return nil
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, streams, &output.TableDef{
				Headers: []string{"ID", "INCOMING PORT", "FORWARD TO", "TCP", "UDP", "ENABLED"},
				RowFunc: func(item interface{}) []string {
					s := item.(client.Stream)
					tcp := "no"
					if s.TCPForwarding {
						tcp = "yes"
					}
					udp := "no"
					if s.UDPForwarding {
						udp = "yes"
					}
					enabled := "no"
					if s.Enabled {
						enabled = "yes"
					}
					return []string{
						fmt.Sprintf("%d", s.ID),
						fmt.Sprintf("%d", s.IncomingPort),
						fmt.Sprintf("%s:%d", s.ForwardingHost, s.ForwardingPort),
						tcp,
						udp,
						enabled,
					}
				},
			})
		},
	}

	return cmd
}
