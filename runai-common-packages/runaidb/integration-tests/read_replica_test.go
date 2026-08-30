package integrationtests

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"github.com/stretchr/testify/require"
)

func requireReadDB(t *testing.T) orm.ReadDB {
	readDB, err := runaidb.ReadDB()
	require.NoError(t, err)
	return readDB
}

func TestReadReplica_UsesDedicatedConnection(t *testing.T) {
	require.NotSame(t, runaidb.DB().Actual(), requireReadDB(t).Actual())
}

func TestReadReplica_UsesReadUser(t *testing.T) {
	var currentUser string
	err := requireReadDB(t).Raw("SELECT current_user").Row().Scan(&currentUser)
	require.NoError(t, err)
	require.Equal(t, readUser, currentUser)
}

func TestReadReplica_TransactionIsReadOnly(t *testing.T) {
	var readOnly string
	err := requireReadDB(t).Raw("SHOW transaction_read_only").Row().Scan(&readOnly)
	require.NoError(t, err)
	require.Equal(t, "on", readOnly)
}

func TestReadReplica_CanSelect(t *testing.T) {
	record := TestSecret{Name: "read-replica-select", Secret: "value"}
	_, err := runaidb.DB().Create(&record)
	require.NoError(t, err)
	require.NotZero(t, record.ID)

	var result TestSecret
	err = requireReadDB(t).First(&result, record.ID)
	require.NoError(t, err)
	require.Equal(t, record.Name, result.Name)
	require.Equal(t, record.Secret, result.Secret)
}

func TestReadReplica_CannotInsert(t *testing.T) {
	err := requireReadDB(t).Actual().Exec(
		"INSERT INTO test_schema.test_secrets (name, secret) VALUES (?, ?)",
		"read-replica-insert", "value",
	).Error
	require.ErrorIs(t, err, orm.ErrReadOnlyConnection)
}

func TestReadReplica_CannotUpdate(t *testing.T) {
	record := TestSecret{Name: "read-replica-update", Secret: "value"}
	_, err := runaidb.DB().Create(&record)
	require.NoError(t, err)

	err = requireReadDB(t).Actual().Exec(
		"UPDATE test_schema.test_secrets SET name = ? WHERE id = ?",
		"updated", record.ID,
	).Error
	require.ErrorIs(t, err, orm.ErrReadOnlyConnection)
}

func TestReadReplica_CannotDelete(t *testing.T) {
	record := TestSecret{Name: "read-replica-delete", Secret: "value"}
	_, err := runaidb.DB().Create(&record)
	require.NoError(t, err)

	err = requireReadDB(t).Actual().Exec(
		"DELETE FROM test_schema.test_secrets WHERE id = ?",
		record.ID,
	).Error
	require.ErrorIs(t, err, orm.ErrReadOnlyConnection)
}
