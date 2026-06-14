package gherkin

import "fmt"

// requiredEnvVars are the build-identity variables the gate-integration task
// injects. newEnvironment errors if any is unset so a misconfigured gate fails
// loudly rather than rendering an empty banner.
var requiredEnvVars = []string{
	"SAURON_APP_NAME",
	"SAURON_APP_VERSION",
	"SAURON_APP_TAG",
	"SAURON_GIT_HASH",
}

// newEnvironment builds the build-identity map from the injected env vars,
// reading through getenv so tests need not mutate the real process environment.
func newEnvironment(getenv func(string) string) (map[string]any, error) {
	for _, key := range requiredEnvVars {
		if getenv(key) == "" {
			return nil, fmt.Errorf("%s is not set; the gate-integration task must inject it", key)
		}
	}

	return map[string]any{
		"App": map[string]any{
			"Name":    getenv("SAURON_APP_NAME"),
			"Version": getenv("SAURON_APP_VERSION"),
			"Tag":     getenv("SAURON_APP_TAG"),
		},
		"GitHash": getenv("SAURON_GIT_HASH"),
	}, nil
}

// newVariables builds the runtime-provided map. HomeDirectory defaults to the
// logical ~/.sauron and is overridden by SAURON_HOME when the gate injects an
// absolute path (which the binary also uses), so both agree. Per-runtime home
// resolution is deferred (see the design doc).
func newVariables(getenv func(string) string) map[string]any {
	home := getenv("SAURON_HOME")
	if home == "" {
		home = "~/.sauron"
	}
	return map[string]any{"HomeDirectory": home}
}
