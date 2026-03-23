package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/internal/config"
)

func newLoginCmd(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Interactively log in to an Nginx Proxy Manager instance and save a profile",
		Long: `Interactively configure and test a connection to an Nginx Proxy Manager instance.

Prompts for the server URL, email, password, and TLS settings. Tests the
connection by verifying the server is reachable, then authenticates with the
provided credentials. Saves the configuration as a named profile.

Examples:
  # Start interactive login
  nginxpm login`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if f.NoInput {
				return fmt.Errorf("interactive input required but --no-input is set. Use environment variables (NGINXPM_URL, NGINXPM_EMAIL, NGINXPM_PASSWORD) instead of 'nginxpm login'.")
			}

			reader := bufio.NewReader(os.Stdin)
			out := f.IOStreams.Out

			// Prompt for URL.
			fmt.Fprint(out, "Nginx Proxy Manager URL: ")
			urlStr, _ := reader.ReadString('\n')
			urlStr = strings.TrimSpace(urlStr)
			if urlStr == "" {
				return fmt.Errorf("URL is required")
			}

			// Prompt for Email.
			fmt.Fprint(out, "Email: ")
			email, _ := reader.ReadString('\n')
			email = strings.TrimSpace(email)
			if email == "" {
				return fmt.Errorf("email is required")
			}

			// Prompt for Password.
			fmt.Fprint(out, "Password: ")
			password, _ := reader.ReadString('\n')
			password = strings.TrimSpace(password)
			if password == "" {
				return fmt.Errorf("password is required")
			}

			// Prompt for TLS skip verification.
			fmt.Fprint(out, "Skip TLS verification? (y/N) [N]: ")
			insecureStr, _ := reader.ReadString('\n')
			insecureStr = strings.TrimSpace(strings.ToLower(insecureStr))
			insecure := insecureStr == "y" || insecureStr == "yes"

			profile := config.Profile{
				URL:      urlStr,
				Email:    email,
				Password: password,
				Insecure: insecure,
			}

			// Test connection by verifying the server is reachable (no auth needed).
			fmt.Fprintln(out, "Testing connection...")
			c := client.NewClientWithToken(urlStr, insecure)
			if f.Verbose {
				c.EnableVerboseLogging(f.IOStreams.ErrOut)
			}

			_, err := c.GetStatus(context.Background())
			if err != nil {
				return fmt.Errorf("connection test failed: %w", err)
			}
			fmt.Fprintln(out, "Server is reachable.")

			// Authenticate with the provided credentials.
			fmt.Fprintln(out, "Authenticating...")
			resolved := &config.ResolvedConfig{
				URL:      profile.URL,
				Email:    profile.Email,
				Password: profile.Password,
				Insecure: profile.Insecure,
			}
			authClient, err := client.NewClient(resolved)
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}
			_ = authClient // authentication succeeded if no error

			fmt.Fprintln(out, "Authentication successful!")

			// Prompt for profile name.
			fmt.Fprint(out, "Profile name [default]: ")
			profileName, _ := reader.ReadString('\n')
			profileName = strings.TrimSpace(profileName)
			if profileName == "" {
				profileName = "default"
			}

			// Save to config.
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			// Overwrite if exists.
			cfg.Profiles[profileName] = profile
			cfg.CurrentProfile = profileName

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Fprintf(out, "Profile %q saved and set as current.\n", profileName)
			return nil
		},
	}
}
