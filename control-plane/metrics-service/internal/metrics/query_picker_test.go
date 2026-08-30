package metrics

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
)

func initProcessedMetricConfigsForTesting() map[string]*ProcessedMetricConfiguration {
	version290, _ := version.NewSemver("2.9.0")
	version295, _ := version.NewSemver("2.9.5")
	version2130, _ := version.NewSemver("2.13.0")
	version310, _ := version.NewSemver("3.1.0")
	processedMetricConfigs := make(map[string]*ProcessedMetricConfiguration)
	processedMetricConfigs["metric1"] = &ProcessedMetricConfiguration{
		CrossVersionQuery: "query1",
		VersionedQueries: []*VersionedQuery{
			{
				Version: version2130,
				Query:   "query2",
			},
			{
				Version: version310,
				Query:   "query5",
			},
			{
				Version: version290,
				Query:   "query3",
			},
			{
				Version: version295,
				Query:   "query4",
			},
		},
	}
	return processedMetricConfigs
}

func TestSortQueriesByVersion(t *testing.T) {
	processedMetricConfigs := initProcessedMetricConfigsForTesting()
	sortQueriesByVersion(processedMetricConfigs["metric1"])
	assert.Equal(t, "query3", processedMetricConfigs["metric1"].VersionedQueries[0].Query)
	assert.Equal(t, "query4", processedMetricConfigs["metric1"].VersionedQueries[1].Query)
	assert.Equal(t, "query2", processedMetricConfigs["metric1"].VersionedQueries[2].Query)
	assert.Equal(t, "query5", processedMetricConfigs["metric1"].VersionedQueries[3].Query)
}

func TestGetQueryByVersion_Exact(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", "2.9.0")
	assert.NoError(t, err)
	assert.Equal(t, "query3", query)
}

func TestGetQueryByVersion_BetweenHotfixes(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", "2.9.1")
	assert.NoError(t, err)
	assert.Equal(t, "query3", query)
}

func TestGetQueryByVersion_BetweenMinors(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", "2.12.1")
	assert.NoError(t, err)
	assert.Equal(t, "query4", query)
}

func TestGetQueryByVersion_BetweenMajors(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", "3.0.5")
	assert.NoError(t, err)
	assert.Equal(t, "query2", query)
}

func TestGetQueryByVersion_Latest(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", "3.5")
	assert.NoError(t, err)
	assert.Equal(t, "query5", query)
}

func TestGetQueryByVersion_CrossCluster(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", CROSS_CLUSTER_VERSION)
	assert.NoError(t, err)
	assert.Equal(t, "query1", query)
}

func TestGetQueryByVersion_Invalid(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", "bad")
	assert.Error(t, err)
	assert.Equal(t, "", query)
	assert.Contains(t, err.Error(), "failed to parse version")
}

func TestGetQueryByVersion_MetricNotFound(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric2", "2.9.0")
	assert.Error(t, err)
	assert.Equal(t, "", query)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetQueryByVersion_EmptyVersion(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", "")
	assert.Error(t, err)
	assert.Equal(t, "", query)
	assert.Contains(t, err.Error(), "empty")
}

func TestGetQueryByVersion_NoQueryForVersion(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", "1.0.0")
	assert.Error(t, err)
	assert.Equal(t, "", query)
	assert.Contains(t, err.Error(), "no query found")
}

func TestGetQueryByVersion_DevVersion(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", "0.0.0-1015653.d36-master-7c63e3")
	assert.NoError(t, err)
	assert.Equal(t, "query5", query)
}

func TestTrimVersionSuffix_StagingVersion(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	assert.Equal(t, "2.9.5", queryPicker.trimVersionSuffix("2.9.5-staging.0"))
}

func TestGetQueryByVersion_StagingVersion(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", "2.9.5-staging.0")
	assert.NoError(t, err)
	assert.Equal(t, "query4", query)
}

func TestTrimVersionSuffix_SaasVersion(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	assert.Equal(t, "2.13.0", queryPicker.trimVersionSuffix("2.13.0-saas.18"))
}

func TestGetQueryByVersion_SaasVersion(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	query, err := queryPicker.GetQueryByVersion("metric1", "2.13.0-saas.15")
	assert.NoError(t, err)
	assert.Equal(t, "query2", query)
}

func TestTrimVersionSuffix_MultipleDashes(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	assert.Equal(t, "2.13.0", queryPicker.trimVersionSuffix("2.13.0-saas.18-aa-bb-cc"))
}

func TestTrimVersionSuffix_NoDashes(t *testing.T) {
	queryPicker := NewMetricQueryPicker(initProcessedMetricConfigsForTesting())
	assert.Equal(t, "2.13.0", queryPicker.trimVersionSuffix("2.13.0"))
}
