package cert

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/output"
)

func newCmdCertGet(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a certificate",
		Long: `Get a certificate by ID.

Examples:
  # Get certificate details
  nginxpm cert get 1

  # Output as JSON
  nginxpm cert get 1 -o json`,
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

			cert, err := c.GetCertificate(context.Background(), id)
			if err != nil {
				return err
			}

			return output.Print(f.IOStreams.Out, f.Resolved.Output, cert, &output.TableDef{
				Headers: []string{"ID", "NICE NAME", "PROVIDER", "DOMAINS", "EXPIRES"},
				RowFunc: func(item interface{}) []string {
					ct := item.(*client.Certificate)
					domains := strings.Join(ct.DomainNames, ", ")
					expires := "N/A"
					if ct.ExpiresOn != "" {
						expires = ct.ExpiresOn
					}
					return []string{
						fmt.Sprintf("%d", ct.ID),
						ct.NiceName,
						ct.Provider,
						domains,
						expires,
					}
				},
			})
		},
	}
}
