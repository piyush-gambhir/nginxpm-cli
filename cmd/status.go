package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newStatusCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Nginx Proxy Manager server status",
		Long: `Show the current status of the Nginx Proxy Manager server.

This command does not require authentication. It reports the server status,
version, and whether initial setup has been completed.

Examples:
  nginxpm status
  nginxpm status -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Status doesn't need auth, just the URL.
			url := ""
			outputFmt := "table"
			if f.Resolved != nil {
				url = f.Resolved.URL
				outputFmt = f.Resolved.Output
			}
			if url == "" {
				return fmt.Errorf("URL is required (use --url, NGINXPM_URL, or configure a profile)")
			}

			insecure := false
			if f.Resolved != nil {
				insecure = f.Resolved.Insecure
			}

			c := client.NewClientWithToken(url, insecure)
			if f.Verbose {
				c.EnableVerboseLogging(f.IOStreams.ErrOut)
			}

			status, err := c.GetStatus(context.Background())
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, outputFmt, status, &output.TableDef{
				Headers: []string{"Status", "Version", "Setup"},
				RowFunc: func(item interface{}) []string {
					s := item.(*client.Status)
					return []string{
						s.Status,
						fmt.Sprintf("%d.%d.%d", s.Version.Major, s.Version.Minor, s.Version.Revision),
						fmt.Sprintf("%t", s.Setup),
					}
				},
			})
		},
	}
}
