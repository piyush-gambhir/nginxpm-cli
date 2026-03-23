package cert

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
)

func newCmdCertCreate(f *cmdutil.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a certificate",
		Long: `Create a new SSL certificate from a JSON or YAML file.

For Let's Encrypt certificates, this triggers certificate issuance.

Examples:
  # Create a certificate from a JSON file
  nginxpm cert create -f cert.json

  # Create from stdin
  nginxpm cert create -f -`,
		Annotations: map[string]string{"mutates": "true"},
		Args:        cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			cert, err := c.CreateCertificate(context.Background(), body)
			if err != nil {
				return err
			}

			if !f.Quiet {
				fmt.Fprintf(f.IOStreams.Out, "Certificate %d created.\n", cert.ID)
			}
			return nil
		},
	}

	cmdutil.AddFileFlag(cmd, &file)

	return cmd
}
