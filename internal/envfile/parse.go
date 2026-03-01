package envfile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Parse reads KEY=VALUE pairs from the given reader.
// Lines starting with # are comments. Empty lines are ignored.
// Values may optionally be wrapped in double quotes (quotes are stripped).
// No shell variable expansion is performed.
func Parse(r io.Reader) (map[string]string, error) {
	env := make(map[string]string)
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: invalid format (expected KEY=VALUE): %q", lineNum, line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNum)
		}

		// Strip optional double quotes
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		env[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading env file: %w", err)
	}

	return env, nil
}

// ParseFile reads KEY=VALUE pairs from the given file path.
// If the file does not exist, returns an empty map (not an error).
func ParseFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("opening env file: %w", err)
	}
	defer func() { _ = f.Close() }()

	return Parse(f)
}
