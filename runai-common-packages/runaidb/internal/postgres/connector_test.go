package postgres

import (
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/stretchr/testify/require"
)

func defaultTestConfig() *dbconfig.Config {
	return &dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{
			Host:             "localhost",
			Port:             5432,
			Name:             "backend",
			Schema:           "public",
			User:             "admin",
			Password:         "secret",
			SslMode:          dbconfig.SslModeDisable,
			StatementTimeout: "2min",
		},
		Pool: dbconfig.PoolConfig{
			MaxConnections:                20,
			MaxIdleConnections:            5,
			MaxIdleConnectionTimeDuration: 30 * time.Minute,
		},
	}
}

func TestNewConnector(t *testing.T) {
	c := NewConnector()
	require.NotNil(t, c)
}

func TestBuildDSN(t *testing.T) {
	c := NewConnector()
	conf := defaultTestConfig()

	dsn := c.buildDSN(&conf.Connection)
	require.Equal(
		t,
		"postgres://admin:secret@localhost:5432/backend?search_path=public&sslmode=disable&statement_timeout=2min&timeZone=UTC",
		dsn.String(),
	)
}

func TestBuildDSN_MultiHost(t *testing.T) {
	c := NewConnector()
	conf := defaultTestConfig()
	conf.Connection.Host = "host1,host2,host3:5433"
	conf.Connection.TargetSessionAttrs = "read-write"

	dsn := c.buildDSN(&conf.Connection)
	require.Equal(
		t,
		"postgres://admin:secret@host1:5432,host2:5432,host3:5433/backend?search_path=public&sslmode=disable&statement_timeout=2min&target_session_attrs=read-write&timeZone=UTC",
		dsn.String(),
	)
}

func TestBuildReadDSN(t *testing.T) {
	c := NewConnector()
	conf := defaultTestConfig()
	conf.ReadConnection = dbconfig.ReadConnectionConfig{
		Host:     "read-host",
		Port:     5433,
		User:     "read-user",
		Password: "read-password",
		SslMode:  dbconfig.SslModeRequire,
	}

	dsn := c.buildReadDSN(conf)
	require.Equal(
		t,
		"postgres://read-user:read-password@read-host:5433/backend?default_transaction_read_only=true&search_path=public&sslmode=require&statement_timeout=2min&timeZone=UTC",
		dsn.String(),
	)
}

func TestBuildReadDSN_DoesNotInheritPrimaryTargetSessionAttrs(t *testing.T) {
	c := NewConnector()
	conf := defaultTestConfig()
	conf.Connection.TargetSessionAttrs = "read-write" // primary wants the writer
	conf.ReadConnection = dbconfig.ReadConnectionConfig{
		Host:     "read-host",
		Port:     5433,
		User:     "read-user",
		Password: "read-password",
	}

	dsn := c.buildReadDSN(conf)
	require.NotContains(t, dsn.String(), "target_session_attrs")
}

func TestBuildReadDSN_TargetSessionAttrs(t *testing.T) {
	c := NewConnector()
	conf := defaultTestConfig()
	conf.Connection.TargetSessionAttrs = "read-write"
	conf.ReadConnection = dbconfig.ReadConnectionConfig{
		Host:               "r1,r2",
		Port:               5433,
		User:               "read-user",
		Password:           "read-password",
		TargetSessionAttrs: "prefer-standby",
	}

	dsn := c.buildReadDSN(conf)
	require.Equal(
		t,
		"postgres://read-user:read-password@r1:5433,r2:5433/backend?default_transaction_read_only=true&search_path=public&sslmode=disable&statement_timeout=2min&target_session_attrs=prefer-standby&timeZone=UTC",
		dsn.String(),
	)
}

func TestRegisterSerializers_NoEncryptionKey(t *testing.T) {
	c := NewConnector()
	conf := defaultTestConfig()

	err := c.registerSerializers(conf)
	require.NoError(t, err)
}

func TestRegisterSerializers_WithEncryptionKey(t *testing.T) {
	c := NewConnector()
	conf := defaultTestConfig()
	conf.EncryptionKey = "12345678901234567890123456789012"

	require.NoError(t, config.Set(conf))
	defer config.FlushAll()

	err := c.registerSerializers(conf)
	require.NoError(t, err)
}

func TestRegisterSerializers_WithInvalidEncryptionKey(t *testing.T) {
	c := NewConnector()
	conf := defaultTestConfig()
	conf.EncryptionKey = "invalid-key"

	require.NoError(t, config.Set(conf))
	defer config.FlushAll()

	err := c.registerSerializers(conf)
	require.EqualError(t, err, "failed to create symmetric encryption serializer: encryption key must be 16, 24, or 32 bytes, got: 11")
}
