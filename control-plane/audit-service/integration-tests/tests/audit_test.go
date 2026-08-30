//go:build integration
// +build integration

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/aws/smithy-go/ptr"
	"github.com/gocarina/gocsv"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	events "github.com/run-ai/runai/runai-application-packages/async-events"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/audit-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/audit-service/internal/services"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	openapi "github.com/run-ai/runai/control-plane/audit-service/pkg/client/generated"
)

var (
	db *gorm.DB
	/***** HOSTS *****/
	// host     = "localhost"
	// redisUrl = "localhost:6379"
	// dbHost   = "localhost"
	host     = "audit-service"
	redisUrl = "redis:6379"
	dbHost   = "postgres"

	auditServiceUrl = fmt.Sprintf("http://%s:%d", host, port)

	port       = 8080
	dbPort     = 5432
	dbUser     = "runai"
	dbPassword = "admin"
	dbName     = "runai"
	dbScheme   = "audit_service"
	client     = openapi.NewAPIClient(&openapi.Configuration{
		Servers: openapi.ServerConfigurations{{
			URL: auditServiceUrl,
		}},
	})

	defaultAuditLogMetadata = map[string]string{}
)

func TestMain(m *testing.M) {
	setEnvs()
	createDirSecretTesting()
	defer deleteDirSecretTesting()
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s database=%s sslmode=disable TimeZone=UTC search_path=%s", dbHost, dbPort, dbUser, dbPassword, dbName, dbScheme)

	db, _ = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	db = db.Exec(fmt.Sprintf("SET search_path TO %s", dbScheme))
	db.Exec(fmt.Sprintf("delete from %s.%s", dbScheme, models.AuditLogsTable))
	err := audit.Init(nil)
	if err != nil {
		fmt.Printf("error on audit.Init(&cfg): %v", err)
	}
	err = runaipubsub.Init(nil)
	if err != nil {
		fmt.Printf("error on runaipubsub.Init(nil): %v", err)
	}

	os.Exit(m.Run())
}

func setEnvs() {
	os.Setenv("SERVICE_NAME", "audit-service")
	os.Setenv("SECRETS_MOUNT_PATH", "../secrets")
	os.Setenv("PUBSUB_TYPE", "redis")
	os.Setenv("PUBSUB_URL", redisUrl)
	os.Setenv("PUBSUB_RETRIES", "10")
	os.Setenv("PUBSUB_CONNECTION_TIMEOUT", "30")
	os.Setenv("AUDIT_SERVICE_URL", auditServiceUrl)
}

func TestCreateLogs_ShouldStoreToDB(t *testing.T) {
	tenant := "1"
	l := produceAuditLog(t, "ShouldStoreToDB test: 1", tenant, "test subject", defaultAuditLogMetadata)
	defer deleteTenantLogs(tenant)

	logs := make([]map[string]interface{}, 0)

	err := retry.Retry(
		func(u uint) error {
			res := db.Table("audit_service." + models.AuditLogsTable).Find(&logs)
			if res.Error != nil || len(logs) == 0 {
				return fmt.Errorf("didn't find log")
			}
			return res.Error
		},
		strategy.Wait(1*time.Second),
		strategy.Limit(3),
	)

	require.NoError(t, err)
	require.Equal(t, len(logs), 1)

	log := logs[0]
	require.Equal(t, log["request_id"], *l.RequestId)
	require.Equal(t, log["cluster_id"], *l.ClusterId)
	require.Equal(t, log["subject"], l.Subject)
	require.Equal(t, log["result"], "Succeeded")
}

func TestCreateLogsWithoutRequestId(t *testing.T) {
	tenant := "12345"
	defer deleteTenantLogs(tenant)

	al := audit.NewAuditLog().
		WithClusterId(uuid.NewString()).
		WithTenantId(tenant).
		WithSubject("subject").
		WithSubjectType("subject_type").
		WithAction("POST").
		WithStatusCode(200)

	err := al.Write()
	require.NoError(t, err)

	logs := make([]map[string]interface{}, 0)

	err = retry.Retry(
		func(u uint) error {
			res := db.Table("audit_service." + models.AuditLogsTable).Find(&logs)
			if res.Error != nil || len(logs) == 0 {
				return fmt.Errorf("didn't find log")
			}
			return res.Error
		},
		strategy.Wait(1*time.Second),
		strategy.Limit(3),
	)

	require.NoError(t, err)
	require.Equal(t, len(logs), 1)

	log := logs[0]
	require.NotEmpty(t, log["request_id"])
	require.Equal(t, *al.Log.ClusterId, log["cluster_id"])
	require.Equal(t, al.Log.Subject, log["subject"])
	require.Equal(t, "Succeeded", log["result"])
}

func TestGetLogs_ShouldGetStoredLogsInDB(t *testing.T) {
	tenant := "2"
	l := produceAuditLog(t, "ShouldGetStoredLogsInDB test: 2", tenant, "test subject", defaultAuditLogMetadata)
	defer deleteTenantLogs(tenant)
	response, err := getAuditLogsFromApi(tenant, nil, []string{}, 1, 5, 1)
	require.NoError(t, err)
	require.Equal(t, *response.Total, int32(1))
	fetchedLogs := response.GetAuditLogs()
	require.Equal(t, 1, len(fetchedLogs))
	log := fetchedLogs[0]
	require.Equal(t, log.TenantId, l.TenantId)
	require.Equal(t, *log.RequestId, *l.RequestId)
	require.Equal(t, *log.ClusterId, *l.ClusterId)
	require.Equal(t, log.Subject, l.Subject)
	require.Equal(t, log.Result, "Succeeded")
}

func TestGetLogs_NoLogsForTenant(t *testing.T) {
	tenant := "3"
	_ = produceAuditLog(t, "NoLogsForTenant test: 3", tenant, "test subject", defaultAuditLogMetadata)
	defer deleteTenantLogs(tenant)
	response, err := getAuditLogsFromApi("1001", nil, []string{}, 1, 2, 1)
	require.NoError(t, err)

	fetchedLogs := response.GetAuditLogs()
	require.Empty(t, fetchedLogs)
}

func TestGetLogs_ShouldGetStoredLogsInDBWithTimeFilter(t *testing.T) {
	tenant := "4"
	l := produceAuditLog(t, "TimeFilter test: 4", tenant, "test subject", defaultAuditLogMetadata)
	defer deleteTenantLogs(tenant)
	start := time.Now().Add(-24 * time.Hour)
	response, err := getAuditLogsFromApi(tenant, &start, []string{}, 1, 5, 1)
	require.NoError(t, err)
	require.NotNil(t, response)

	fetchedLogs := response.GetAuditLogs()
	require.Equal(t, 1, len(fetchedLogs))
	log := fetchedLogs[0]
	require.Equal(t, log.TenantId, l.TenantId)
	require.Equal(t, *log.RequestId, *l.RequestId)
	require.Equal(t, *log.ClusterId, *l.ClusterId)
	require.Equal(t, log.Subject, l.Subject)
	require.Equal(t, log.Result, "Succeeded")
}

func TestGetLogs_ShouldGetStoredLogsInDBWithFieldFilters(t *testing.T) {
	tenant := "5"
	requiredSubject := uuid.New().String()
	_ = produceAuditLog(t, fmt.Sprintf("FieldsFilters test: 5, Subject %s ", requiredSubject), tenant, requiredSubject, defaultAuditLogMetadata)
	insertAuditLogs(t, tenant, 10)
	defer deleteTenantLogs(tenant)
	response, err := getAuditLogsFromApi(tenant, nil, []string{"subject==" + requiredSubject}, 20, 5, 0)
	require.NoError(t, err)

	fetchedLogs := response.GetAuditLogs()
	require.Equal(t, 1, len(fetchedLogs))
	log := fetchedLogs[0]
	require.Equal(t, log.Subject, requiredSubject)
}

func TestGetLogs_ShouldGetStoredLogsInDBWithMultipleFilters(t *testing.T) {
	tenant := "6"
	requiredSubject := uuid.New().String()
	firstLog := produceAuditLog(t, fmt.Sprintf("FieldsFilters test: 6, Subject %s ", requiredSubject), tenant, requiredSubject, defaultAuditLogMetadata)
	_ = produceAuditLog(t, fmt.Sprintf("FieldsFilters test: 6, Subject %s ", requiredSubject), tenant, requiredSubject, defaultAuditLogMetadata)
	defer deleteTenantLogs(tenant)
	response, err := getAuditLogsFromApi(tenant, nil, []string{"subject==" + requiredSubject, "cluster_name==" + *firstLog.ClusterName}, 20, 5, 0)
	require.NoError(t, err)

	fetchedLogs := response.GetAuditLogs()
	require.Equal(t, 1, len(fetchedLogs))
	log := fetchedLogs[0]
	require.Equal(t, log.Subject, requiredSubject)
	require.Equal(t, *log.ClusterName, *firstLog.ClusterName)
}

func TestDownloadLogs_ShouldDownloadJSON(t *testing.T) {
	tenant := "7"
	requiredSubject := uuid.New().String()
	logs, err := produceAndDownloadLogs(t, requiredSubject, tenant, "JSON", defaultAuditLogMetadata)
	require.NoError(t, err)
	defer deleteTenantLogs(tenant)
	body, err := io.ReadAll(logs)
	require.NoError(t, err)

	fetchedLogs := api.GetAuditLogs200Response{}
	err = json.Unmarshal(body, &fetchedLogs.AuditLogs)
	require.NoError(t, err)
	require.Equal(t, 1, len(fetchedLogs.AuditLogs))
	require.Equal(t, fetchedLogs.AuditLogs[0].Subject, requiredSubject)
}

func TestDownloadLogs_ShouldDownloadCSV(t *testing.T) {
	tenant := "8"
	requiredSubject := uuid.New().String()
	logs, err := produceAndDownloadLogs(t, requiredSubject, tenant, "CSV", defaultAuditLogMetadata)
	defer deleteTenantLogs(tenant)
	fetchedLogs := []services.SerializedAuditLog{}
	err = gocsv.Unmarshal(logs, &fetchedLogs)
	require.NoError(t, err)
	require.Equal(t, 1, len(fetchedLogs))
	require.Equal(t, fetchedLogs[0].Subject, requiredSubject)
}

func TestGetLogs_TotalSamples(t *testing.T) {
	tenant := "9"
	insertAuditLogs(t, tenant, 10)
	defer deleteTenantLogs(tenant)
	time.Sleep(3 * time.Second)
	response, err := getAuditLogsFromApi(tenant, nil, []string{}, 5, 5, 10)

	require.NoError(t, err)
	require.Equal(t, 5, len(response.AuditLogs))
	require.Equal(t, int32(10), *response.Total)
}

func TestDownloadLogs_FilesAreValid_DifferentMetadataScenarios(t *testing.T) {
	tenantId := "10"

	tests := []struct {
		name                string
		metadata            map[string]string
		expectedCsvMetadata string
	}{
		{"No Metadata", nil, ""},
		{"Single Key Metadata", map[string]string{"key1": "value1"}, `key1=value1`},
		{"Multiple Keys Metadata", map[string]string{"name": "x", "appId": "val1"}, `appId=val1,name=x`},
		{"Special Characters Metadata", map[string]string{"key@!": "val#%"}, `key@!=val#%`},
		{"With Inner Quotes", map[string]string{`key"1""`: `val"1""`}, `key""1""""=val""1""""`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer deleteTenantLogs(tenantId)
			assertCSVTestsWithDifferentMetadata(t, uuid.New().String(), tenantId, tt.metadata, tt.expectedCsvMetadata)
			assertJSONTestsWithDifferentMetadata(t, uuid.New().String(), tenantId, tt.metadata)
		})
	}
}

func TestDeleteLogs_OnTenantDelete(t *testing.T) {
	tenantId := "11"

	insertAuditLogs(t, tenantId, 10)
	defer deleteTenantLogs(tenantId)
	response, err := getAuditLogsFromApi(tenantId, nil, []string{}, 10, 15, 10)

	require.NoError(t, err)
	require.Equal(t, 10, len(response.AuditLogs))
	require.Equal(t, int32(10), *response.Total)

	runaipubsub.Publish(context.Background(), events.NewTenantDelete(tenantId))
	time.Sleep(3 * time.Second)
	response, err = getAuditLogsFromApi(tenantId, nil, []string{}, 0, 0, 0)

	require.NoError(t, err)
	require.Equal(t, 0, len(response.AuditLogs))
	require.Equal(t, (*int32)(nil), response.Total)
}

func TestDeleteLogs_OnTenantDelete_MultipleTenants(t *testing.T) {
	tenantId1 := "12"
	tenantId2 := "13"
	insertAuditLogs(t, tenantId1, 10)
	insertAuditLogs(t, tenantId2, 10)
	defer deleteTenantLogs(tenantId1)
	defer deleteTenantLogs(tenantId2)
	response, err := getAuditLogsFromApi(tenantId1, nil, []string{}, 10, 10, 10)

	require.NoError(t, err)
	require.Equal(t, 10, len(response.AuditLogs))
	require.Equal(t, int32(10), *response.Total)

	response, err = getAuditLogsFromApi(tenantId2, nil, []string{}, 10, 10, 10)

	require.NoError(t, err)
	require.Equal(t, 10, len(response.AuditLogs))
	require.Equal(t, int32(10), *response.Total)

	runaipubsub.Publish(context.Background(), events.NewTenantDelete(tenantId1))
	time.Sleep(3 * time.Second)
	response, err = getAuditLogsFromApi(tenantId1, nil, []string{}, 10, 0, 0)

	require.NoError(t, err)
	require.Equal(t, 0, len(response.AuditLogs))
	require.Equal(t, (*int32)(nil), response.Total)

	response, err = getAuditLogsFromApi(tenantId2, nil, []string{}, 10, 10, 10)

	require.NoError(t, err)
	require.Equal(t, 10, len(response.AuditLogs))
	require.Equal(t, int32(10), *response.Total)
}

func TestPostLogs(t *testing.T) {
	tenant := "15"
	requiredSubject := uuid.New().String()
	_ = produceAuditLogUsingApi(t, tenant, requiredSubject, defaultAuditLogMetadata)
	insertAuditLogs(t, tenant, 10)
	defer deleteTenantLogs(tenant)
	response, err := getAuditLogsFromApi(tenant, nil, []string{"subject==" + requiredSubject}, 20, 5, 0)
	require.NoError(t, err)

	fetchedLogs := response.GetAuditLogs()
	require.Equal(t, 1, len(fetchedLogs))
	log := fetchedLogs[0]
	require.Equal(t, log.Subject, requiredSubject)
}

func assertCSVTestsWithDifferentMetadata(t *testing.T, subjectId, tenantId string, metadata map[string]string, expectedMetadata string) {
	logs, err := produceAndDownloadLogs(t, subjectId, tenantId, "CSV", metadata)
	require.NoError(t, err)

	fetchedLogs := []services.SerializedAuditLog{}
	err = gocsv.Unmarshal(logs, &fetchedLogs)
	require.NoError(t, err)

	require.Equal(t, 1, len(fetchedLogs))
	require.Equal(t, subjectId, fetchedLogs[0].Subject)
	require.Equal(t, expectedMetadata, fetchedLogs[0].Metadata)
}

func assertJSONTestsWithDifferentMetadata(t *testing.T, subjectId, tenantId string, metadata map[string]string) {
	logs, err := produceAndDownloadLogs(t, subjectId, tenantId, "JSON", metadata)
	require.NoError(t, err)

	body, err := io.ReadAll(logs)
	require.NoError(t, err)

	fetchedLogs := []api.AuditLog{}
	err = json.Unmarshal(body, &fetchedLogs)

	require.NoError(t, err)
	require.Equal(t, 1, len(fetchedLogs))
	require.Equal(t, subjectId, fetchedLogs[0].Subject)
	require.Equal(t, metadata, fetchedLogs[0].Metadata)
}

func produceAndDownloadLogs(t *testing.T, requiredSubject, tenant, fileType string, metadata map[string]string) (io.ReadCloser, error) {
	_ = produceAuditLog(t, fmt.Sprintf("FieldsFilters test: %s, Subject %s ", tenant, requiredSubject), tenant, requiredSubject, metadata)
	insertAuditLogs(t, tenant, 10)

	res, err := getAuditLogsFromApi(tenant, nil, []string{"subject==" + requiredSubject}, 20, 5, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(res.AuditLogs))

	response, err := downloadAuditLogsFromApi(tenant, fileType, []string{"subject==" + requiredSubject})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	return response.Body, err
}

func getAuditLogsFromApi(tenantId string, start *time.Time, filters []string, samples int, maxRetries int, expectedTotalCount int) (*openapi.GetAuditLogs200Response, error) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@run.ai")
	logs := client.AuditLogsAPI.GetAuditLogs(ctx).
		Start(time.Now().UTC().Add(-24 * time.Hour)).
		End(time.Now().UTC().Add(24 * time.Hour)).
		SortBy("timestamp").
		SortOrder("asc").
		FilterBy(filters).
		NumberOfSamples(int32(samples))
	if start != nil {
		logs.Start((*start).UTC())
	}

	var res *openapi.GetAuditLogs200Response
	var err error
	retryErr := retry.Retry(
		func(u uint) error {
			res, _, err = logs.Execute()
			if err != nil || len(res.AuditLogs) < expectedTotalCount || res.Total == nil || *res.Total < int32(expectedTotalCount) {
				fmt.Printf("didn't find any messages for tenant %s\n", tenantId)
				return fmt.Errorf("didn't find any messages for tenant %s\n", tenantId)
			}
			return nil
		},
		strategy.Limit(uint(maxRetries+1)),
		strategy.Wait(2*time.Second),
	)
	if retryErr != nil {
		fmt.Printf("get logs retry was exhausted: %v\n", retryErr)
	}

	return res, err
}

func downloadAuditLogsFromApi(tenantId string, fileType string, filters []string) (*http.Response, error) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@run.ai")
	return client.AuditLogsAPI.DownloadAuditLogs(ctx).
		FileType(fileType).
		Start(time.Now().Add(-24 * time.Hour).UTC()).
		End(time.Now().Add(24 * time.Hour).UTC()).
		FilterBy(filters).Execute()
}

func produceAuditLog(t *testing.T, id, tenantId, subject string, metadata map[string]string) openapi.AuditLog {
	l := getAuditLog(id, tenantId, subject, metadata)
	al := audit.NewAuditLog().
		WithClusterId(*l.ClusterId).
		WithClusterName(*l.ClusterName).
		WithRequestId(*l.RequestId).
		WithTenantId(tenantId).
		WithSubject(subject).
		WithSubjectType(subject).
		WithEntityType(*l.EntityType).
		WithEntityId(*l.EntityId).
		WithEntityName(*l.EntityName).
		WithAction(l.Action).
		WithStatusCode(int(*l.HttpStatusCode))

	for k, v := range metadata {
		al.WithMetadata(k, v)
	}

	al.Log.SetId(id)
	err := al.Write()
	require.NoError(t, err)
	return l
}

func produceAuditLogUsingApi(t *testing.T, tenantId, subject string, metadata map[string]string) openapi.AuditLog {
	l := getAuditLog("", tenantId, subject, metadata)
	_, err := client.InternalAuditLogsAPI.SaveAuditLog(context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)).AuditLog(l).Execute()
	require.NoError(t, err)
	return l
}

func getAuditLog(id, tenantId, subject string, metadata map[string]string) openapi.AuditLog {
	return openapi.AuditLog{
		Id:             ptr.String(id),
		RequestId:      ptr.String(uuid.NewString()),
		ClusterId:      ptr.String(uuid.NewString()),
		ClusterName:    ptr.String(uuid.NewString()),
		Timestamp:      ptr.Time(time.Now()),
		TenantId:       tenantId,
		Subject:        subject,
		SubjectType:    subject,
		Metadata:       &metadata,
		EntityType:     ptr.String("type"),
		EntityId:       ptr.String("id"),
		EntityName:     ptr.String("name"),
		Action:         "GET",
		HttpStatusCode: ptr.Int32(200),
		Result:         "Ok",
	}
}

func createDirSecretTesting() error {
	if err := os.MkdirAll("secrets", 0o755); err != nil {
		return err
	}
	return nil
}

func deleteDirSecretTesting() error {
	if err := os.RemoveAll("secrets"); err != nil {
		return err
	}
	return nil
}

func insertAuditLogs(t *testing.T, tenantId string, count int) {
	for i := 0; i < count; i++ {
		subject := fmt.Sprintf("%d: %s", i, uuid.New().String())
		_ = produceAuditLog(t, fmt.Sprintf("FieldsFilters test: %s, Subject %s ", tenantId, subject), tenantId, subject, defaultAuditLogMetadata)
	}
}

func deleteTenantLogs(tenantId string) {
	db.Exec(fmt.Sprintf("delete from %s.%s where tenant_id = %s", dbScheme, models.AuditLogsTable, tenantId))
}
