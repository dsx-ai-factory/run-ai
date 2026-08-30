package config

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/clusters"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/database"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/datasources"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/metrics"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/metricstores"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/providers"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/reports"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/tenants"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempConfigFile(content string) *os.File {
	file, err := os.CreateTemp("", "service-config*.yaml")
	if err != nil {
		log.Fatal(err)
	}
	file.WriteString(content)
	file.Close()
	return file
}

func initConfigForTesting(t *testing.T, configContent string) error {
	return initConfigForTestingWithSecrets(t, configContent, nil)
}

func createTempSecretsDir(t *testing.T, secrets map[string]string) string {
	secretsDir, err := os.MkdirTemp("", "secrets")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(secretsDir)
	})

	for name, value := range secrets {
		err := os.WriteFile(filepath.Join(secretsDir, name), []byte(value), 0600)
		require.NoError(t, err)
	}
	return secretsDir
}

func initConfigForTestingWithSecrets(t *testing.T, configContent string, secrets map[string]string) error {
	config.Init(&config.Config{
		KeyDelimiter: ";",
	})

	file := createTempConfigFile(configContent)
	t.Cleanup(func() {
		os.Remove(file.Name())
	})

	secretsDir := createTempSecretsDir(t, secrets)
	return Init(file.Name(), secretsDir)
}

func TestInitConfig_Defaults(t *testing.T) {
	defer config.FlushAll()
	configContent := ``
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)

	serviceConf := config.Get[*ServiceConfiguration]()
	assert.Equal(t, "info", serviceConf.LogLevel)
	assert.Equal(t, false, serviceConf.UseMocks)

	reportsConf := config.Get[*reports.ReportsConfiguration]()
	assert.Equal(t, 5, reportsConf.FetchLimit)
	assert.Equal(t, 10, reportsConf.ConcurrencyLimit)
	assert.Equal(t, 60, reportsConf.RunPeriodSec)
	assert.Equal(t, "metric_service_reports", reportsConf.S3Config.Bucket)

	clustersConf := config.Get[*clusters.ClustersConfiguration]()
	assert.Equal(t, "http://runai-backend-cluster-service:8080", clustersConf.ServiceUrl)
	assert.Equal(t, 1, clustersConf.Cache.VersionHistoriesTtlMinutes)
	assert.Equal(t, 60, clustersConf.Cache.ClusterNamesTtlMinutes)

	tenantsConf := config.Get[*tenants.TenantsConfiguration]()
	assert.Equal(t, "http://runai-backend-tenants-manager:8080", tenantsConf.ServiceUrl)

	datasourcesConf := config.Get[*datasources.DatasourcesConfiguration]()
	assert.Equal(t, "http://runai-backend-thanos-query:9090", datasourcesConf.DatasourceUrl)

	cleanerConf := config.Get[*database.CleanerConfiguration]()
	assert.Equal(t, false, cleanerConf.IsCleanupEnabled)
	assert.Equal(t, 0, cleanerConf.CleanerIntervalHours)
	assert.Equal(t, 0, cleanerConf.DeleteReportsTTLDays)

	dbCronJobsConf := config.Get[*database.CronJobsConfiguration]()
	assert.Equal(t, 60, dbCronJobsConf.Reports.IntervalSeconds)
	assert.Equal(t, 1, dbCronJobsConf.Reports.MaxRetryCount)
	assert.Equal(t, 1, dbCronJobsConf.Reports.RetryFailedReportsMinAgeMinutes)
	assert.Equal(t, 30, dbCronJobsConf.Reports.RetryFailedReportsMaxAgeMinutes)
	assert.Equal(t, 30, dbCronJobsConf.Reports.FailStuckProcessingReportsMinAgeMinutes)

	metricsConf := config.Get[*metrics.MetricsConfiguration]()
	assert.Equal(t, make(map[string]*metrics.MetricConfiguration), metricsConf.Metrics)

	datasourceProviderConf := config.Get[*providers.DatasourceProviderConfig]()
	assert.Equal(t, providers.DPTStatic, datasourceProviderConf.Type)
	assert.Equal(t, metricstores.Datasource{
		Type: metricstores.DatasourceTypeThanos,
		Read: metricstores.DatasourceEndpoint{
			Url: "http://runai-backend-thanos-query:9090",
		},
		Write: metricstores.DatasourceEndpoint{},
	}, datasourceProviderConf.StaticDatasource)
}

func TestInitConfig_CustomServiceConfiguration(t *testing.T) {
	defer config.FlushAll()
	configContent := `
log_level: debug
use_mocks: true
`
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)
	serviceConf := config.Get[*ServiceConfiguration]()
	assert.Equal(t, "debug", serviceConf.LogLevel)
	assert.Equal(t, true, serviceConf.UseMocks)
}

func TestInitConfig_CustomClusters(t *testing.T) {
	defer config.FlushAll()
	configContent := `
clusters:
  service_url: http://clusters.com
`
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)
	clustersConf := config.Get[*clusters.ClustersConfiguration]()
	assert.Equal(t, "http://clusters.com", clustersConf.ServiceUrl)
}

func TestInitConfig_CustomTenants(t *testing.T) {
	defer config.FlushAll()
	configContent := `
tenants:
  service_url: http://tenants.com
`
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)

	tenantsConf := config.Get[*tenants.TenantsConfiguration]()
	assert.Equal(t, "http://tenants.com", tenantsConf.ServiceUrl)
}

func TestInitConfig_CustomTenants_InvalidConfig(t *testing.T) {
	defer config.FlushAll()
	configContent := `
tenants:
  service_url: not a url
  legacy_service_url: not a url
`
	err := initConfigForTesting(t, configContent)
	assert.Contains(t, err.Error(), "failed to validate config struct TenantsConfiguration")
}

func TestInitConfig_CustomDatasources(t *testing.T) {
	defer config.FlushAll()
	configContent := `
datasources:
  datasource_url: http://datasource.com
`
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)

	datasourcesConf := config.Get[*datasources.DatasourcesConfiguration]()
	assert.Equal(t, "http://datasource.com", datasourcesConf.DatasourceUrl)
}

func TestInitConfig_CustomDatasources_InvalidConfig(t *testing.T) {
	defer config.FlushAll()
	configContent := `
datasources:
  datasource_url: not a url
`
	err := initConfigForTesting(t, configContent)
	assert.Contains(t, err.Error(), "failed to validate config struct DatasourcesConfiguration")
}

func TestInitConfig_CustomCleaner(t *testing.T) {
	defer config.FlushAll()
	configContent := `
cleaner:
  is_cleanup_enabled: true
  cleaner_interval_hours: 10
  delete_report_ttl_days: 7
`
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)

	cleanerConf := config.Get[*database.CleanerConfiguration]()
	assert.Equal(t, true, cleanerConf.IsCleanupEnabled)
	assert.Equal(t, 10, cleanerConf.CleanerIntervalHours)
	assert.Equal(t, 7, cleanerConf.DeleteReportsTTLDays)
}

func TestInitConfig_CustomCleaner_InvalidConfig(t *testing.T) {
	defer config.FlushAll()
	configContent := `
cleaner:
  is_cleanup_enabled: true
  cleaner_interval_hours: -10
  delete_report_ttl_days: -7
`
	err := initConfigForTesting(t, configContent)
	assert.Contains(t, err.Error(), "failed to validate config struct CleanerConfiguration")
}

func TestInitConfig_CustomCronJobs(t *testing.T) {
	defer config.FlushAll()
	configContent := `
cron_jobs:
  reports:
    interval_seconds: 10
    max_retry_count: 10
    retry_failed_reports_min_age_minutes: 11
    retry_failed_reports_max_age_minutes: 12
    fail_stuck_processing_reports_min_age_minutes: 13
`
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)

	dbCronJobsConf := config.Get[*database.CronJobsConfiguration]()
	assert.Equal(t, 10, dbCronJobsConf.Reports.IntervalSeconds)
	assert.Equal(t, 10, dbCronJobsConf.Reports.MaxRetryCount)
	assert.Equal(t, 11, dbCronJobsConf.Reports.RetryFailedReportsMinAgeMinutes)
	assert.Equal(t, 12, dbCronJobsConf.Reports.RetryFailedReportsMaxAgeMinutes)
	assert.Equal(t, 13, dbCronJobsConf.Reports.FailStuckProcessingReportsMinAgeMinutes)
}

func TestInitConfig_CustomMetrics(t *testing.T) {
	defer config.FlushAll()
	configContent := `
metrics:
  metric1:
    cross_version_query: query2
    queries:
      2.9.0: query1
  metric2:
    cross_version_query: query4
    queries:
      2.13.2: query3
`
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)

	metricsConf := config.Get[*metrics.MetricsConfiguration]()
	assert.Equal(t, 2, len(metricsConf.Metrics))
	assert.Equal(t, 1, len(metricsConf.Metrics["metric1"].Queries))
	assert.Equal(t, 1, len(metricsConf.Metrics["metric2"].Queries))
	assert.Equal(t, "query2", metricsConf.Metrics["metric1"].CrossVersionQuery)
	assert.Equal(t, "query4", metricsConf.Metrics["metric2"].CrossVersionQuery)
	assert.Equal(t, "query1", metricsConf.Metrics["metric1"].Queries["2.9.0"])
	assert.Equal(t, "query3", metricsConf.Metrics["metric2"].Queries["2.13.2"])
}

func TestInitConfig_CustomMetrics_InvalidSemver(t *testing.T) {
	defer config.FlushAll()
	configContent := `
metrics:
  metric1:
    cross_version_query: query2
    queries:
      2.9: query1
`

	err := initConfigForTesting(t, configContent)
	assert.Error(t, err)
}

func TestInitConfig_CustomMetrics_NoCrossVersion(t *testing.T) {
	defer config.FlushAll()
	configContent := `
metrics:
  metric1:
    queries:
      2.9.0: query1
`

	err := initConfigForTesting(t, configContent)
	assert.Error(t, err)
}

func TestInitConfig_CustomMetrics_NoQueries(t *testing.T) {
	defer config.FlushAll()
	configContent := `
metrics:
  metric1:
    cross_version_query: query2
`

	err := initConfigForTesting(t, configContent)
	assert.Error(t, err)
}

func TestInitConfig_CustomClusters_CacheDefaults(t *testing.T) {
	defer config.FlushAll()
	configContent := `
clusters:
  service_url: tenants.com
`
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)

	clustersConf := config.Get[*clusters.ClustersConfiguration]()
	assert.Equal(t, 1, clustersConf.Cache.VersionHistoriesTtlMinutes)
}

func TestInitConfig_CustomClusters_CacheNotAllDefaults(t *testing.T) {
	defer config.FlushAll()
	configContent := `
clusters:
  service_url: tenants.com
  cache:
    version_histories_ttl_minutes: 60
    get_cluster_names_ttl_minutes: 100
`
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)

	clustersConf := config.Get[*clusters.ClustersConfiguration]()
	assert.Equal(t, 60, clustersConf.Cache.VersionHistoriesTtlMinutes)
	assert.Equal(t, 100, clustersConf.Cache.ClusterNamesTtlMinutes)
}

func TestInitConfig_CustomMetricStores(t *testing.T) {
	defer config.FlushAll()
	configContent := `
metric_stores:
  cache:
    ttl_minutes: 1
`
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)

	metricStoresConf := config.Get[*metricstores.MetricStoresConfiguration]()
	assert.Equal(t, 1, metricStoresConf.Cache.TTLMinutes)
}

func TestInitConfig_CustomDatasourceIdConfig(t *testing.T) {
	defer config.FlushAll()
	configContent := `
release: test_release
env: test_env
`
	err := initConfigForTesting(t, configContent)
	assert.NoError(t, err)
	dsIdConf := config.Get[*metricstores.DatasourceIdConfig]()
	assert.Equal(t, "test_release", dsIdConf.Release)
	assert.Equal(t, "test_env", dsIdConf.Env)
}

func TestInitConfig_CustomDatasourceProvider(t *testing.T) {
	defer config.FlushAll()
	configContent := `
datasource_provider:
  datasource_type: grafanalabs
  static_datasource:
    type: mimir
    read:
      url: http://mimir-read:9090
      auth:
        type: basic
    write:
      url: http://mimir-write:9090
      auth:
        type: basic
`
	secrets := map[string]string{
		"DATASOURCE_PROVIDER.DATASOURCE_AUTH.TOKEN":                       "my-secret-token",
		"DATASOURCE_PROVIDER.STATIC_DATASOURCE.READ.AUTH.BASIC.USER":      "reader",
		"DATASOURCE_PROVIDER.STATIC_DATASOURCE.READ.AUTH.BASIC.PASSWORD":  "reader-pass",
		"DATASOURCE_PROVIDER.STATIC_DATASOURCE.WRITE.AUTH.BASIC.USER":     "writer",
		"DATASOURCE_PROVIDER.STATIC_DATASOURCE.WRITE.AUTH.BASIC.PASSWORD": "writer-pass",
	}

	err := initConfigForTestingWithSecrets(t, configContent, secrets)
	assert.NoError(t, err)

	datasourceProviderConf := config.Get[*providers.DatasourceProviderConfig]()
	assert.Equal(t, &providers.DatasourceProviderConfig{
		Type: providers.DPTGrafanaLabs,
		Auth: providers.DatasourceProviderAuth{
			Token: "my-secret-token",
		},
		StaticDatasource: metricstores.Datasource{
			Type: metricstores.DatasourceTypeMimir,
			Read: metricstores.DatasourceEndpoint{
				Url: "http://mimir-read:9090",
				Auth: &metricstores.DatasourceAuth{
					Type: metricstores.AuthTypeBasic,
					Basic: &metricstores.AuthBasic{
						User:     "reader",
						Password: "reader-pass",
					},
				},
			},
			Write: metricstores.DatasourceEndpoint{
				Url: "http://mimir-write:9090",
				Auth: &metricstores.DatasourceAuth{
					Type: metricstores.AuthTypeBasic,
					Basic: &metricstores.AuthBasic{
						User:     "writer",
						Password: "writer-pass",
					},
				},
			},
		},
	}, datasourceProviderConf)
}
