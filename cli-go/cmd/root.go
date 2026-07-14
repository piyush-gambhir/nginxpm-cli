package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/cmd/access"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/cmd/audit"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/cmd/cert"
	cmdconfig "github.com/piyush-gambhir/nginxpm-cli/cli-go/cmd/config"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/cmd/dead"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/cmd/proxy"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/cmd/redirect"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/cmd/setting"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/cmd/stream"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/cmd/user"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/build"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/client"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/cmdutil"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/update"
)

var (
	flagOutput   string
	flagProfile  string
	flagURL      string
	flagEmail    string
	flagPassword string
	flagInsecure bool
	flagReadOnly bool
	flagNoInput  bool
	flagQuiet    bool
	flagVerbose  bool
)

// OutputFormat is set during PersistentPreRunE and exported for use by main.go.
var OutputFormat string

// Execute is the main entry point for the CLI.
func Execute() error {
	return newRootCmd().Execute()
}

// loadAndResolveConfig loads the config file and resolves auth from flags/env/config.
func loadAndResolveConfig(cmd *cobra.Command) (*config.ResolvedConfig, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("loading config: %w", err)
	}

	// Determine which profile to use.
	profileName := flagProfile
	if profileName == "" {
		profileName = cfg.CurrentProfile
	}
	var profile *config.Profile
	if profileName != "" {
		p, ok := cfg.Profiles[profileName]
		if !ok && cmd.Name() != "login" {
			return nil, nil, fmt.Errorf("profile %q not found", profileName)
		}
		if ok {
			profile = &p
		}
	}

	// Determine output format.
	output := flagOutput
	if output == "" {
		output = cfg.Defaults.Output
	}

	// Resolve configuration.
	resolved := config.Resolve(flagURL, flagEmail, flagPassword, flagInsecure, cmd.Flags().Changed("insecure"), profile, cfg.Defaults)
	if output != "" {
		resolved.Output = output
	}

	return resolved, cfg, nil
}

func checkPermissions(cmd *cobra.Command, resolved *config.ResolvedConfig) error {
	effectiveReadOnly := resolved.ReadOnly
	if flagReadOnly {
		effectiveReadOnly = true
	}
	if effectiveReadOnly && cmd.Annotations != nil && cmd.Annotations["mutates"] == "true" {
		return fmt.Errorf("command '%s' is blocked in read-only mode; remove read_only from the profile or disable the read-only environment setting to permit writes", cmd.CommandPath())
	}
	return nil
}

func envFlagEnabled(name string) bool {
	v := strings.TrimSpace(os.Getenv(name))
	return strings.EqualFold(v, "true") || v == "1"
}

// createClient sets up the HTTP client factory on the factory.
func createClient(ctx context.Context, f *cmdutil.Factory, resolved *config.ResolvedConfig) {
	f.Client = func() (*client.Client, error) {
		c, err := client.NewClientContext(ctx, resolved)
		if err != nil {
			return nil, err
		}
		if flagVerbose {
			c.EnableVerboseLogging(f.IOStreams.ErrOut)
		}
		return c, nil
	}
}

const updateRepo = "piyush-gambhir/nginxpm-cli"

func newRootCmd() *cobra.Command {
	f := &cmdutil.Factory{
		IOStreams: cmdutil.DefaultIOStreams(),
	}

	// Channel-based update check result passing from PersistentPreRun to PersistentPostRun.
	var updateResult chan *update.UpdateInfo

	rootCmd := &cobra.Command{
		Use:   "nginxpm",
		Short: "Nginx Proxy Manager CLI - manage Nginx Proxy Manager from the command line",
		Long: `A command-line interface for managing Nginx Proxy Manager proxy hosts, redirections, streams, certificates, and more.

Full command reference (for agents/LLMs): https://nginxpm-cli.pages.dev/llms.txt
Claude Code skill: https://github.com/piyush-gambhir/nginxpm-cli/blob/main/nginxpm/SKILL.md`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Check env vars for --no-input, --quiet, --verbose.
			if envFlagEnabled("NGINXPM_NO_INPUT") {
				flagNoInput = true
			}
			if !cmd.Flags().Changed("quiet") {
				flagQuiet = envFlagEnabled("NGINXPM_QUIET")
			}
			if !cmd.Flags().Changed("verbose") {
				flagVerbose = envFlagEnabled("NGINXPM_VERBOSE")
			}
			f.NoInput = flagNoInput
			f.Quiet = flagQuiet
			f.Verbose = flagVerbose

			// Start background update check for most commands.
			cmdName := cmd.Name()
			skipUpdateCheck := cmdName == "update" || cmdName == "version" || cmdName == "completion" || cmdName == "help"
			if !skipUpdateCheck && build.Version != "dev" && build.Version != "" {
				updateResult = make(chan *update.UpdateInfo, 1)
				go func() {
					info, _ := update.CheckForUpdate(build.Version, updateRepo, config.ConfigDir())
					updateResult <- info
				}()
			}

			// Skip auth setup for commands that don't need it.
			if cmdName == "version" || cmdName == "completion" || cmdName == "help" || cmdName == "update" {
				return nil
			}
			// Also skip for config subcommands.
			if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
				return nil
			}

			resolved, cfg, err := loadAndResolveConfig(cmd)
			if err != nil {
				return err
			}

			// Set exported OutputFormat for use by main.go error handler.
			OutputFormat = resolved.Output

			f.Resolved = resolved
			if err := checkPermissions(cmd, resolved); err != nil {
				return err
			}

			f.Config = func() (*config.Config, error) {
				return cfg, nil
			}

			// Skip client creation for commands that handle auth themselves.
			if cmdName == "login" || cmdName == "status" {
				return nil
			}

			createClient(cmd.Context(), f, resolved)

			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if updateResult == nil {
				return
			}
			select {
			case info := <-updateResult:
				if info != nil && info.Available {
					update.PrintUpdateNotice(os.Stderr, info)
				}
			case <-time.After(2 * time.Second):
				// Don't block command output waiting for update check.
			}
		},
	}

	// Global persistent flags.
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().StringVar(&flagProfile, "profile", "", "Configuration profile to use")
	rootCmd.PersistentFlags().StringVar(&flagURL, "url", "", "Nginx Proxy Manager URL")
	rootCmd.PersistentFlags().StringVar(&flagEmail, "email", "", "Email for authentication")
	rootCmd.PersistentFlags().StringVarP(&flagPassword, "password", "p", "", "Password for authentication")
	rootCmd.PersistentFlags().BoolVarP(&flagInsecure, "insecure", "k", false, "Skip TLS certificate verification")
	rootCmd.PersistentFlags().BoolVar(&flagReadOnly, "read-only", false, "Block write operations (safety mode for agents)")
	rootCmd.PersistentFlags().BoolVar(&flagNoInput, "no-input", false, "Disable all interactive prompts (for CI/agent use)")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress informational output")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose HTTP logging")

	// Register subcommands.
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newUpdateCmd())
	rootCmd.AddCommand(newLoginCmd(f))
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newStatusCmd(f))
	rootCmd.AddCommand(cmdconfig.NewCmdConfig(f))
	rootCmd.AddCommand(proxy.NewCmdProxy(f))
	rootCmd.AddCommand(redirect.NewCmdRedirect(f))
	rootCmd.AddCommand(stream.NewCmdStream(f))
	rootCmd.AddCommand(dead.NewCmdDead(f))
	rootCmd.AddCommand(cert.NewCmdCert(f))
	rootCmd.AddCommand(access.NewCmdAccess(f))
	rootCmd.AddCommand(user.NewCmdUser(f))
	rootCmd.AddCommand(audit.NewCmdAudit(f))
	rootCmd.AddCommand(setting.NewCmdSetting(f))

	return rootCmd
}
