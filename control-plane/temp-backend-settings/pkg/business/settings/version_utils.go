package settings

import (
	"strconv"
	"strings"
)

func IsVersionSupported(version string, minimumVersion string) bool {
	if version == "" {
		return false
	}
	version = extractVersion(version)
	return strings.HasPrefix(version, "0.0.0") ||
		strings.Contains(version, "master") ||
		compareVersions(version, minimumVersion) >= 0
}

func compareVersions(a, b string) int {
	aParts := parseVersionParts(a)
	bParts := parseVersionParts(b)

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var av, bv int
		if i < len(aParts) {
			av = aParts[i]
		}
		if i < len(bParts) {
			bv = bParts[i]
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
	}
	return 0
}

func parseVersionParts(v string) []int {
	parts := strings.Split(v, ".")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			n = 0
		}
		result = append(result, n)
	}
	return result
}

func extractVersion(version string) string {
	if idx := strings.Index(version, "-rc"); idx > -1 {
		version = version[:idx]
	}
	version = strings.TrimPrefix(version, "v")
	return version
}
