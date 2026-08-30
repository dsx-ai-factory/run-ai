package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"

	tenant_lifecycle "github.com/run-ai/runai/control-plane/tenants-manager/pkg/pubsub/tenant-lifecycle"
	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/pevent"
)

// ---------------------------------------------------------------------------
// DB helper
// ---------------------------------------------------------------------------

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSslMode, dbScheme)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func setTenantFailed(t *testing.T, tenantId int32, diagnostics string) {
	t.Helper()
	db := openDB(t)
	_, err := db.Exec(`UPDATE tenants SET status = 'Failed' WHERE id = $1`, tenantId)
	require.NoError(t, err, "failed to set tenant %d to Failed in DB", tenantId)
	now := time.Now().UTC()
	_, err = db.Exec(
		`INSERT INTO provisioning_runs (tenant_id, started_at, ended_at, status, diagnostics)
		 VALUES ($1, $2, $3, 'failed', $4)`,
		tenantId, now.Add(-time.Second), now, diagnostics,
	)
	require.NoError(t, err, "failed to insert provisioning run for tenant %d", tenantId)
}

func getTenantStatusFromDB(t *testing.T, tenantId int32) (status, diagnostics string) {
	t.Helper()
	db := openDB(t)
	row := db.QueryRow(`SELECT status FROM tenants WHERE id = $1`, tenantId)
	require.NoError(t, row.Scan(&status))
	row = db.QueryRow(`
		SELECT COALESCE(diagnostics, '') FROM provisioning_runs
		WHERE tenant_id = $1 ORDER BY ended_at DESC NULLS LAST LIMIT 1`, tenantId)
	_ = row.Scan(&diagnostics)
	return
}

// ---------------------------------------------------------------------------
// NATS helper
// ---------------------------------------------------------------------------

func publishTenantProvision(t *testing.T, event *tenant_lifecycle.TenantProvision) {
	t.Helper()

	nc, err := nats.Connect(natsURL, nats.UserInfo(natsUser, natsPass))
	require.NoError(t, err, "failed to connect to NATS")
	t.Cleanup(func() { nc.Close() })

	js, err := jetstream.New(nc)
	require.NoError(t, err)

	envelope := pevent.New[tenant_lifecycle.TenantProvision]().WithPayload(event)
	payload, err := json.Marshal(envelope)
	require.NoError(t, err)

	topic := pevent.New[tenant_lifecycle.TenantProvision]().GetTopicName()
	_, err = js.Publish(context.Background(), topic, payload)
	require.NoError(t, err, "failed to publish TenantProvision to NATS")
}

// ---------------------------------------------------------------------------
// Poll helper
// ---------------------------------------------------------------------------

func pollTenantStatus(t *testing.T, tenantId int32, wantStatus string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, _ := getTenantStatusFromDB(t, tenantId)
		if status == wantStatus {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	status, diag := getTenantStatusFromDB(t, tenantId)
	t.Fatalf("tenant %d did not reach status %q within %s (current: %q, diagnostics: %q)",
		tenantId, wantStatus, timeout, status, diag)
}

// sampleProvisionEvent builds a minimal TenantProvision from a created tenant's
// metadata (retrieved from the DB) so Path 2 can re-publish it.
func buildProvisionEvent(t *testing.T, tenantId int32, tenantName, mimirId, email string) *tenant_lifecycle.TenantProvision {
	t.Helper()
	contractType := "normal"
	return &tenant_lifecycle.TenantProvision{
		TenantId:     uint(tenantId),
		TenantName:   tenantName,
		MimirId:      mimirId,
		Email:        email,
		ContractType: &contractType,
		Role:         authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR,
	}
}
