package metrics

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
)

func TestValidateVersions_Valid(t *testing.T) {
	metricsConfig := &MetricConfiguration{
		CrossVersionQuery: "query1",
		Queries: map[string]string{
			"2.13":  "query2",
			"2.9":   "query3",
			"2.9.5": "query4",
		},
	}
	err := validateNeededVersionsExist(metricsConfig)
	assert.NoError(t, err)
}

func TestValidateVersions_IllegalNoCrossVersion(t *testing.T) {
	metricsConfig := &MetricConfiguration{
		Queries: map[string]string{
			"2.13":  "query2",
			"2.9":   "query3",
			"2.9.5": "query4",
		},
	}
	err := validateNeededVersionsExist(metricsConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing cross cluster version")
}

func TestValidateVersions_IllegalNoSpecificVersion(t *testing.T) {
	metricsConfig := &MetricConfiguration{
		CrossVersionQuery: "query1",
		Queries:           map[string]string{},
	}
	err := validateNeededVersionsExist(metricsConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "need to have at least one version other than cross cluster version")
}

func TestValidateVersions_IllegalNoSpecificVersionNil(t *testing.T) {
	metricsConfig := &MetricConfiguration{
		CrossVersionQuery: "query1",
		Queries:           nil,
	}
	err := validateNeededVersionsExist(metricsConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "need to have at least one version other than cross cluster version")
}

func TestTransformKeysIntoValidSemvers_Valid(t *testing.T) {
	queries := map[string]string{
		"2.13":  "query2",
		"2.9":   "query3",
		"2.9.5": "query4",
	}
	validSemverKeysQueries, err := transformKeysIntoValidSemvers(queries)
	assert.NoError(t, err)
	assert.NotNil(t, validSemverKeysQueries)
	assert.Equal(t, len(queries), len(validSemverKeysQueries))
	assert.Contains(t, validSemverKeysQueries, "2.9.0")
	assert.Contains(t, validSemverKeysQueries, "2.9.5")
	assert.Contains(t, validSemverKeysQueries, "2.13.0")
}

func TestTransformKeysIntoValidSemvers_BadFormat(t *testing.T) {
	queries := map[string]string{
		"2.13": "query",
		"2.9":  "query",
		"bad":  "query",
	}
	validSemverKeysQueries, err := transformKeysIntoValidSemvers(queries)
	assert.Error(t, err)
	assert.Nil(t, validSemverKeysQueries)
	assert.Contains(t, err.Error(), "failed to parse version")
}

func TestTransformKeysIntoValidSemvers_Duplicate(t *testing.T) {
	queries := map[string]string{
		"2.13":  "query",
		"2.9":   "query",
		"2.9.0": "query",
	}
	validSemverKeysQueries, err := transformKeysIntoValidSemvers(queries)
	assert.Error(t, err)
	assert.Nil(t, validSemverKeysQueries)
	assert.Contains(t, err.Error(), "is not unique")
}

func TestCreateVersionedQueries_Created(t *testing.T) {
	queries := map[string]string{
		"2.13":  "query2",
		"2.9":   "query3",
		"2.9.5": "query4",
	}
	versionedQueries, err := createVersionedQueries(queries)
	assert.NoError(t, err)
	assert.NotNil(t, versionedQueries)
	assert.Equal(t, len(queries), len(versionedQueries))
	assert.Contains(t, versionedQueries, &VersionedQuery{version.Must(version.NewSemver("2.13")), "query2"})
	assert.Contains(t, versionedQueries, &VersionedQuery{version.Must(version.NewSemver("2.9")), "query3"})
	assert.Contains(t, versionedQueries, &VersionedQuery{version.Must(version.NewSemver("2.9.5")), "query4"})
}

func TestCreateSortedVersions_VersionBadFormat(t *testing.T) {
	queries := map[string]string{
		"2.13": "quer ",
		"2.9":  "query",
		"bad":  "query",
	}
	sortedVersions, err := createVersionedQueries(queries)
	assert.Error(t, err)
	assert.Nil(t, sortedVersions)
}
