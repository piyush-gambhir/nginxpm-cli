package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for nginxpm CLI.

To load completions:

Bash:
  $ source <(nginxpm completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ nginxpm completion bash > /etc/bash_completion.d/nginxpm
  # macOS:
  $ nginxpm completion bash > $(brew --prefix)/etc/bash_completion.d/nginxpm

Zsh:
  $ source <(nginxpm completion zsh)
  # To load completions for each session, execute once:
  $ nginxpm completion zsh > "${fpath[1]}/_nginxpm"

Fish:
  $ nginxpm completion fish | source
  # To load completions for each session, execute once:
  $ nginxpm completion fish > ~/.config/fish/completions/nginxpm.fish

PowerShell:
  PS> nginxpm completion powershell | Out-String | Invoke-Expression
  # To load completions for each session, execute once:
  PS> nginxpm completion powershell > nginxpm.ps1
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
	return cmd
}
