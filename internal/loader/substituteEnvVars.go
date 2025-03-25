package loader

import (
	"os"
	"regexp"
	"strings"
)

var envVarRegexp = regexp.MustCompile(`\$\{([^}]+)\}`)

// substituteEnvVars replaces ${VAR} with the corresponding environment variable.
// Escaped form ${{VAR}} is preserved as literal ${VAR}.
func substituteEnvVars(yamlContent string) string {
	// Step 1: Escape ${{VAR}} to a temporary placeholder
	yamlContent = strings.ReplaceAll(yamlContent, "${{", "__ESCAPED_VAR__START__")
	yamlContent = strings.ReplaceAll(yamlContent, "}}", "__ESCAPED_VAR__END__")

	// Step 2: Substitute all ${VAR}
	yamlContent = envVarRegexp.ReplaceAllStringFunc(yamlContent, func(m string) string {
		key := envVarRegexp.FindStringSubmatch(m)[1]
		return os.Getenv(key)
	})

	// Step 3: Restore escaped ${VAR}
	yamlContent = strings.ReplaceAll(yamlContent, "__ESCAPED_VAR__START__", "${")
	yamlContent = strings.ReplaceAll(yamlContent, "__ESCAPED_VAR__END__", "}")

	return yamlContent
}
