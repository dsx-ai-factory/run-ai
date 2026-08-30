package version

import (
	"fmt"

	semver "github.com/hashicorp/go-version"
)

func VersionLessThan(version1, version2 string) (bool, error) {
	return compareVersions(version1, version2, LESS)
}

func VersionGreaterThan(version1, version2 string) (bool, error) {
	return compareVersions(version1, version2, GREATER)
}

func VersionGreaterEqualThan(version1, version2 string) (bool, error) {
	isLess, err := VersionLessThan(version1, version2)
	return !isLess, err
}

func compareVersions(version1, version2 string, expected ComparisonResult) (bool, error) {
	result, err := parseAndCompare(version1, version2)
	return result == expected, err
}

func parseAndCompare(version1, version2 string) (ComparisonResult, error) {
	ver1, err := semver.NewVersion(version1)
	if err != nil {
		return UNDEFINED, fmt.Errorf("invalid version1 %s", err)
	}
	ver2, err := semver.NewVersion(version2)
	if err != nil {
		return UNDEFINED, fmt.Errorf("invalid version2 %s", err)
	}
	return ComparisonResult(ver1.Compare(ver2)), nil
}
