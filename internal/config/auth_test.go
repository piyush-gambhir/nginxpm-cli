package config

import "testing"

func TestResolveBooleanOverrides(t *testing.T) {
	t.Setenv("NGINXPM_INSECURE", "true")
	t.Setenv("NGINXPM_READ_ONLY", "true")

	rc := Resolve("https://npm.example", "", "", false, true, &Profile{Insecure: true}, Defaults{})
	if rc.Insecure {
		t.Fatal("explicit --insecure=false should override environment and profile")
	}
	if !rc.ReadOnly {
		t.Fatal("NGINXPM_READ_ONLY=true should enable read-only mode")
	}
}
