package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/build"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/update"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long: `Print the nginxpm-cli version, commit hash, and build date.

When an earlier update check is cached, also print the latest release and
whether an update is available. version never contacts GitHub; run
nginxpm update --check for a fresh answer.

Examples:
  nginxpm version`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "nginxpm-cli version %s\n", build.Version)
			fmt.Fprintf(cmd.OutOrStdout(), "  commit: %s\n", build.Commit)
			fmt.Fprintf(cmd.OutOrStdout(), "  built:  %s\n", build.Date)
			if info := update.CachedInfo(build.Version, config.ConfigDir()); info != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "  latest: %s\n", info.LatestVersion)
				if update.IsReleaseVersion(build.Version) {
					fmt.Fprintf(cmd.OutOrStdout(), "  update_available: %t\n", info.Available)
				}
			}
		},
	}
}
