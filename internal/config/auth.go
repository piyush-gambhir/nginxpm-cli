package config

import (
	"os"
	"strings"
)

// ResolvedConfig holds the final resolved configuration after layering
// flags > env vars > profile values.
type ResolvedConfig struct {
	URL      string
	Email    string
	Password string
	Insecure bool
	Output   string
}

// Resolve merges flag values, environment variables, and profile values.
// Priority: flags > env vars > profile values.
func Resolve(flagURL, flagEmail, flagPassword string, flagInsecure bool, profile *Profile, defaults Defaults) *ResolvedConfig {
	rc := &ResolvedConfig{}

	// URL
	rc.URL = firstNonEmpty(flagURL, os.Getenv("NGINXPM_URL"))
	if rc.URL == "" && profile != nil {
		rc.URL = profile.URL
	}

	// Email
	rc.Email = firstNonEmpty(flagEmail, os.Getenv("NGINXPM_EMAIL"))
	if rc.Email == "" && profile != nil {
		rc.Email = profile.Email
	}

	// Password
	rc.Password = firstNonEmpty(flagPassword, os.Getenv("NGINXPM_PASSWORD"))
	if rc.Password == "" && profile != nil {
		rc.Password = profile.Password
	}

	// Insecure
	if flagInsecure {
		rc.Insecure = true
	} else if envInsecure := os.Getenv("NGINXPM_INSECURE"); envInsecure != "" {
		rc.Insecure = strings.EqualFold(envInsecure, "true") || envInsecure == "1"
	} else if profile != nil {
		rc.Insecure = profile.Insecure
	}

	// Output
	rc.Output = defaults.Output
	if rc.Output == "" {
		rc.Output = "table"
	}

	return rc
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
