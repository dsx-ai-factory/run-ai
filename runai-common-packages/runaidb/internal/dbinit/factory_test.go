package dbinit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostgresFactory_CreateConnector(t *testing.T) {
	factory := &postgresFactory{}
	require.NotNil(t, factory.CreateConnector())
}

func TestPostgresFactory_CreateMigrator(t *testing.T) {
	factory := &postgresFactory{}
	require.NotNil(t, factory.CreateMigrator(nil))
}

func TestPostgresFactory_CreateSessionCreator(t *testing.T) {
	factory := &postgresFactory{}
	require.NotNil(t, factory.CreateSessionCreator(nil))
}
