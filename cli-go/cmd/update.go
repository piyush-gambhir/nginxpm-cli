package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/build"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/config"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/output"
	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/update"
)

// Seams for tests: the target platform, the executable being replaced, and
// whether stdin can answer a prompt.
var (
	goos            = runtime.GOOS
	goarch          = runtime.GOARCH
	executablePath  = currentExecutable
	stdinIsTerminal = func() bool { return term.IsTerminal(int(os.Stdin.Fd())) }
)

// updateCheckResult is the `update --check -o json|yaml` payload.
type updateCheckResult struct {
	CurrentVersion  string `json:"current_version" yaml:"current_version"`
	LatestVersion   string `json:"latest_version" yaml:"latest_version"`
	UpdateAvailable bool   `json:"update_available" yaml:"update_available"`
	ReleaseURL      string `json:"release_url" yaml:"release_url"`
	InstallMethod   string `json:"install_method" yaml:"install_method"`
}

func newUpdateCmd() *cobra.Command {
	var checkOnly, yes bool

	cmd := &cobra.Command{
		Use:         "update",
		Annotations: map[string]string{"mutates": "true"},
		Short:       "Update nginxpm to the latest version",
		Long: `Check for and install the latest nginxpm release from GitHub Releases.

nginxpm update downloads the release archive for this OS and architecture
(.tar.gz on macOS and Linux, .zip on Windows), verifies it against the
release's checksums.txt (SHA-256), and replaces the running executable. On
Windows the running nginxpm.exe is moved aside to nginxpm.exe.old, which the
next run deletes. If the executable's directory is not writable, re-run with
sudo (Administrator on Windows) or reinstall into a writable directory. A build
in a Go bin directory ($GOBIN, $GOPATH/bin, ~/go/bin) is not replaced: update
it from source with git pull && make install.

--check always asks GitHub and reports the current and latest versions; it
works with -o json and with --read-only. Installing is blocked by --read-only.
Without --yes, update asks for confirmation, and fails under --no-input or when
stdin is not a terminal.

In an interactive terminal, other commands check GitHub for a new release at
most once a day and print a notice on stderr. The check is skipped when stderr
is not a terminal, when CI is set, with --quiet, and when
NGINXPM_NO_UPDATE_NOTIFIER or NO_UPDATE_NOTIFIER is set to any value.

Examples:
  nginxpm update --check
  nginxpm update --check -o json
  nginxpm update
  nginxpm update --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, checkOnly, yes)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only report whether a newer release exists (always queries GitHub)")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Install without asking for confirmation")

	return cmd
}

func runUpdate(cmd *cobra.Command, checkOnly, yes bool) error {
	out := cmd.OutOrStdout()
	current := build.Version
	if !update.IsReleaseVersion(current) {
		fmt.Fprintf(out, "Update checking is not available for development builds (version %q).\n", current)
		fmt.Fprintln(out, "Install a release to enable updates.")
		return nil
	}
	format := flagOutput
	switch format {
	case "", "table", "json", "yaml":
	default:
		return fmt.Errorf("unsupported output format: %s (use table, json, or yaml)", format)
	}
	if !checkOnly {
		if err := checkUpdateAllowed(cmd); err != nil {
			return err
		}
	}

	configDir := config.ConfigDir()
	info, err := update.CheckForUpdateFresh(current, configDir)
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}
	execPath, execErr := executablePath()
	goInstall := execErr == nil && isGoInstall(execPath)

	if checkOnly {
		return printUpdateCheck(out, format, info, goInstall)
	}
	if !info.Available {
		fmt.Fprintf(out, "nginxpm v%s is already the latest version.\n", trimV(current))
		return nil
	}
	fmt.Fprintf(out, "Update available: v%s -> v%s\n", trimV(current), info.LatestVersion)
	if goInstall {
		fmt.Fprintf(out, "nginxpm in %s was built from source, so it does not replace itself.\n", filepath.Dir(execPath))
		fmt.Fprintf(out, "Update with: %s\n", update.SourceUpdateCommand)
		fmt.Fprintf(out, "Release notes: %s\n", info.ReleaseURL)
		return nil
	}
	if execErr != nil {
		return execErr
	}

	if !yes {
		if flagNoInput || !stdinIsTerminal() {
			return fmt.Errorf("update needs confirmation: pass --yes to install v%s without a prompt", info.LatestVersion)
		}
		fmt.Fprint(out, "Update now? [Y/n] ")
		line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
		answer := strings.ToLower(strings.TrimSpace(line))
		if (err != nil && answer == "") || (answer != "" && answer != "y" && answer != "yes") {
			fmt.Fprintln(out, "Update cancelled.")
			return nil
		}
	}

	installer := &update.Installer{GOOS: goos, GOARCH: goarch, ExecPath: execPath, Out: out}
	if err := installer.Install(cmd.Context(), info.LatestVersion); err != nil {
		return err
	}
	update.ClearCache(configDir)
	fmt.Fprintf(out, "Updated nginxpm v%s -> v%s\n", trimV(current), info.LatestVersion)
	fmt.Fprintf(out, "Release notes: %s\n", info.ReleaseURL)
	return nil
}

// checkUpdateAllowed applies read-only mode (flag, NGINXPM_READ_ONLY, or the
// profile's read_only) to installing; `update --check` is always allowed.
func checkUpdateAllowed(cmd *cobra.Command) error {
	readOnly := flagReadOnly || envFlagEnabled("NGINXPM_READ_ONLY")
	if resolved, _, err := loadAndResolveConfig(cmd); err == nil && resolved.ReadOnly {
		readOnly = true
	}
	return checkPermissions(cmd, &config.ResolvedConfig{ReadOnly: readOnly})
}

func printUpdateCheck(w io.Writer, format string, info *update.UpdateInfo, goInstall bool) error {
	method := "self"
	if goInstall {
		method = "go"
	}
	if format == "json" || format == "yaml" {
		return output.Print(w, format, updateCheckResult{
			CurrentVersion:  trimV(info.CurrentVersion),
			LatestVersion:   info.LatestVersion,
			UpdateAvailable: info.Available,
			ReleaseURL:      info.ReleaseURL,
			InstallMethod:   method,
		}, nil)
	}
	available := "no"
	if info.Available {
		available = "yes"
	}
	fmt.Fprintf(w, "Current version:  v%s\n", trimV(info.CurrentVersion))
	fmt.Fprintf(w, "Latest version:   v%s\n", info.LatestVersion)
	fmt.Fprintf(w, "Update available: %s\n", available)
	if info.Available {
		fmt.Fprintf(w, "Update with: %s\n", update.UpdateCommand(goInstall))
	}
	fmt.Fprintf(w, "Release notes: %s\n", info.ReleaseURL)
	return nil
}

// currentExecutable is the running binary with symlinks resolved, so the
// update replaces the real file rather than a link to it.
func currentExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("finding current executable: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return "", fmt.Errorf("resolving executable path: %w", err)
	}
	return resolved, nil
}

func isGoInstall(execPath string) bool {
	home, _ := os.UserHomeDir()
	return update.IsGoInstall(execPath, os.Getenv, home)
}

func trimV(v string) string {
	return strings.TrimPrefix(v, "v")
}
