package envfile

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// WriteFile writes KEY=VALUE pairs to the given file path.
// Keys are sorted for deterministic output.
// If env is empty or nil, no file is created.
func WriteFile(path string, env map[string]string) error {
	if len(env) == 0 {
		return nil
	}

	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%s\n", k, env[k])
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}
