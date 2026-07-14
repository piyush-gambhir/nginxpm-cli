package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/piyush-gambhir/nginxpm-cli/cli-go/internal/config"
)

func TestReadOnlyBlocksMutatingCommand(t *testing.T) {
	cmd := &cobra.Command{
		Use:         "delete",
		Annotations: map[string]string{"mutates": "true"},
	}
	err := checkPermissions(cmd, &config.ResolvedConfig{ReadOnly: true})
	if err == nil || !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("expected read-only error, got %v", err)
	}
}

func TestReadOnlyAllowsReadCommand(t *testing.T) {
	cmd := &cobra.Command{Use: "list"}
	if err := checkPermissions(cmd, &config.ResolvedConfig{ReadOnly: true}); err != nil {
		t.Fatalf("read command should be allowed: %v", err)
	}
}

func TestEnvFlagEnabled(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  bool
	}{{"true", true}, {"TRUE", true}, {"1", true}, {"false", false}, {"0", false}, {"anything", false}} {
		t.Setenv("NGINXPM_TEST_FLAG", tc.value)
		if got := envFlagEnabled("NGINXPM_TEST_FLAG"); got != tc.want {
			t.Errorf("envFlagEnabled(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}

func TestMutatingCommandsAreAnnotated(t *testing.T) {
	root := newRootCmd()
	mutating := map[string]bool{}
	var walk func(*cobra.Command)
	walk = func(parent *cobra.Command) {
		for _, child := range parent.Commands() {
			if child.Annotations != nil && child.Annotations["mutates"] == "true" {
				mutating[child.CommandPath()] = true
			}
			walk(child)
		}
	}
	walk(root)

	want := []string{
		"nginxpm proxy create", "nginxpm proxy update", "nginxpm proxy delete", "nginxpm proxy enable", "nginxpm proxy disable",
		"nginxpm redirect create", "nginxpm redirect update", "nginxpm redirect delete", "nginxpm redirect enable", "nginxpm redirect disable",
		"nginxpm stream create", "nginxpm stream update", "nginxpm stream delete", "nginxpm stream enable", "nginxpm stream disable",
		"nginxpm dead create", "nginxpm dead update", "nginxpm dead delete", "nginxpm dead enable", "nginxpm dead disable",
		"nginxpm cert create", "nginxpm cert delete", "nginxpm cert renew",
		"nginxpm access create", "nginxpm access update", "nginxpm access delete",
		"nginxpm user create", "nginxpm user update", "nginxpm user delete",
		"nginxpm user password", "nginxpm user permissions", "nginxpm setting update",
	}
	for _, path := range want {
		if !mutating[path] {
			t.Errorf("mutating command %q is missing its annotation", path)
		}
	}
}
