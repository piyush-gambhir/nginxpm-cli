package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/output"
)

// resolution describes arguments using the parser without running the command.
type resolution struct {
	Command     string            `json:"command"`
	Effect      string            `json:"effect"`
	Conditional bool              `json:"conditional"`
	Interactive bool              `json:"interactive"`
	Secret      bool              `json:"secret_output"`
	Flags       map[string]string `json:"flags"`
	Args        []string          `json:"args"`
}

func newCommandsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commands",
		Short: "Offline command discovery for agents",
		Long:  "Inspect commands without running them or loading configuration.",
	}
	cmd.AddCommand(newResolveCmd())
	return cmd
}

func newResolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "resolve -- <arguments>",
		Short:              "Show the command, effect and flags that arguments would run",
		Long:               "Resolve arguments with the CLI parser without executing the target, reading configuration, or contacting the network. The answer is always JSON.",
		Example:            "  nginxpm commands resolve -- proxy list\n  nginxpm commands resolve -- -o json proxy delete 1 --confirm",
		DisableFlagParsing: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Override config/auth setup and the background update check, including on errors.
			OutputFormat = "json"
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			for i, arg := range args {
				if arg == "--" {
					args = args[i+1:]
					break
				}
			}
			probe := newRootCmd()
			probe.SetIn(strings.NewReader(""))
			probe.SetOut(io.Discard)
			probe.SetErr(io.Discard)
			target, rest, err := probe.Find(args)
			if err != nil || target == probe || !target.Runnable() {
				return fmt.Errorf("arguments do not name a runnable nginxpm command")
			}
			if err := target.ParseFlags(rest); err != nil {
				return err
			}
			return output.Print(cmd.OutOrStdout(), OutputFormat, describeResolution(target), nil)
		},
	}
}

func describeResolution(target *cobra.Command) resolution {
	res := resolution{
		Command: target.CommandPath(), Effect: "read", Flags: map[string]string{},
		Args: target.Flags().Args(),
	}
	if res.Args == nil {
		res.Args = []string{}
	}
	ann := target.Annotations
	switch {
	case ann["mutates"] == "true" || ann["mutates"] == "conditional":
		res.Effect, res.Conditional = "remote_write", ann["mutates"] == "conditional"
	case ann["mutates-when-flag"] != "" && target.Flags().Changed(ann["mutates-when-flag"]):
		res.Effect = "remote_write"
	case ann["writes-local"] == "true" || ann["writes-local"] == "conditional" || ann["self-update"] != "" ||
		(ann["writes-local-when-flag"] != "" && target.Flags().Changed(ann["writes-local-when-flag"])):
		res.Effect, res.Conditional = "local_write", ann["writes-local"] == "conditional"
	}
	res.Interactive = ann["interactive"] == "true"
	res.Secret = ann["secret-output"] == "true"
	target.Flags().Visit(func(f *pflag.Flag) { res.Flags[f.Name] = f.Value.String() })
	return res
}
