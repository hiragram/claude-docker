package update

import (
	"fmt"
	"strconv"
	"strings"
)

// parseVersion strips a leading "v" and splits "X.Y.Z" into major, minor, patch.
func parseVersion(s string) (major, minor, patch int, err error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format %q: expected X.Y.Z", s)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version %q: %w", parts[0], err)
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version %q: %w", parts[1], err)
	}
	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version %q: %w", parts[2], err)
	}

	return major, minor, patch, nil
}

// isNewer returns true if latest is strictly newer than current.
func isNewer(latest, current string) (bool, error) {
	lMaj, lMin, lPat, err := parseVersion(latest)
	if err != nil {
		return false, fmt.Errorf("parsing latest version: %w", err)
	}
	cMaj, cMin, cPat, err := parseVersion(current)
	if err != nil {
		return false, fmt.Errorf("parsing current version: %w", err)
	}

	if lMaj != cMaj {
		return lMaj > cMaj, nil
	}
	if lMin != cMin {
		return lMin > cMin, nil
	}
	return lPat > cPat, nil
}
