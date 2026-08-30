package dao_utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

const (
	DefaultPostgresPort            = "5432"
	postgresImage                  = "postgres:16.3-alpine"
	postgresReadyLogMsg            = "database system is ready to accept connections"
	postgresReadReplicaReadyLogMsg = "database system is ready to accept read-only connections"
	waitingPollInterval            = 100 * time.Millisecond
	waitingOccurrence              = 2
	postReadySleepDuration         = 200 * time.Millisecond
	All                            = "1=1"
	PrimaryContainerAlias          = "primary_container"
)

// OpenPostgresForTesting runs a postgres docker and sets env vars so the tests
// will utilize this docker as persistent storage instead of "real" postgres.
func OpenPostgresForTesting(t *testing.T, dbConfig *DataBaseConfig) {
	dbConfig.Host, dbConfig.Port, dbConfig.ReadHost, dbConfig.ReadPort = spinPostgresWithReadReplica(t, dbConfig)
	dbConfig.MigrationFiles = os.Getenv(EnvMigrationFiles)
	Init(dbConfig)
}

func MountSecretsToDir(dbUser, dbPassword string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "secrets")
	if err != nil {
		return "", err
	}
	dbUserFile := filepath.Join(tmpDir, EnvDbUser)
	dbPasswordFile := filepath.Join(tmpDir, EnvDbPassword)
	if err := os.WriteFile(dbUserFile, []byte(dbUser), 0600); err != nil {
		return "", err
	}
	if err := os.WriteFile(dbPasswordFile, []byte(dbPassword), 0600); err != nil {
		return "", err
	}
	return tmpDir, nil
}

// SetPostgresEnv sets environment variables related to Postgres configuration.
func SetPostgresEnv(dbSchemaName, migrationFiles string) error {
	secretsDir, err := MountSecretsToDir("postgres", "admin")
	if err != nil {
		return err
	}
	_ = os.Setenv(EnvSecretsMountPath, secretsDir)
	_ = os.Setenv(EnvDbName, "postgres")
	_ = os.Setenv(EnvDbHost, "temporary-value")
	_ = os.Setenv(EnvDbSchema, dbSchemaName)
	_ = os.Setenv(EnvMigrationFiles, migrationFiles)
	return nil
}

// spinSharedNetwork creates a new Docker network (unique for each test).
func spinSharedNetwork(ctx context.Context) (*testcontainers.DockerNetwork, string, error) {
	net, err := network.New(ctx, network.WithAttachable())
	if err != nil {
		return nil, "", err
	}
	return net, net.Name, nil
}

// spinPostgresWithReadReplica spins up a primary and replica postgres on a new network.
func spinPostgresWithReadReplica(t *testing.T, dbConfig *DataBaseConfig) (string, string, string, string) {
	t.Helper()
	ctx := context.Background()

	pgNetwork, networkName, err := spinSharedNetwork(ctx)
	require.Nil(t, err, "failed to create network")

	primaryHost, primaryPort, err := spinPostgres(t, dbConfig, networkName)
	require.Nil(t, err, "failed to start primary postgres")

	replicaContainer, err := spinReplicaContainer(ctx, networkName)
	require.Nil(t, err, "failed to start replica postgres")

	replicaHost, err := replicaContainer.Host(ctx)
	require.Nil(t, err, "failed to get replica host")

	replicaPort, err := replicaContainer.MappedPort(ctx, DefaultPostgresPort)
	require.Nil(t, err, "failed to get replica port")

	// Cleanup containers and network after test
	t.Cleanup(func() {
		_ = replicaContainer.Terminate(ctx)
		_ = pgNetwork.Remove(ctx)
	})

	return primaryHost, primaryPort, replicaHost, replicaPort.Port()
}

// spinPostgres spins up the PostgreSQL primary container.
func spinPostgres(t *testing.T, dbConfig *DataBaseConfig, networkName string) (string, string, error) {
	t.Helper()
	ctx := context.Background()

	primaryContainer, err := spinPrimaryContainer(ctx, dbConfig, networkName)
	if err != nil {
		return "", "", err
	}

	primaryHost, err := primaryContainer.Host(ctx)
	if err != nil {
		return "", "", err
	}

	primaryPort, err := primaryContainer.MappedPort(ctx, DefaultPostgresPort)
	if err != nil {
		return "", "", err
	}

	// Terminate container after test
	t.Cleanup(func() {
		_ = primaryContainer.Terminate(ctx)
	})

	return primaryHost, primaryPort.Port(), nil
}

// spinReplicaContainer spins up the PostgreSQL read-replica container.
func spinReplicaContainer(ctx context.Context, networkName string) (testcontainers.Container, error) {
	replicaReq := testcontainers.ContainerRequest{
		Image:        postgresImage,
		ExposedPorts: []string{DefaultPostgresPort},
		Networks:     []string{networkName},
		Env: map[string]string{
			"PGPASSWORD": "replicator_password",
			"PGUSER":     "replicator",
		},
		Cmd: []string{
			"bash", "-c",
			fmt.Sprintf(`
            until pg_basebackup --pgdata=/var/lib/postgresql/data -R --slot=replication_slot --host=%s --port=%s
            do
                echo 'Waiting for primary to connect...'
                sleep 1s
            done
            echo 'Backup done, starting replica...'
            chmod 0700 /var/lib/postgresql/data
            exec postgres
            `, PrimaryContainerAlias, DefaultPostgresPort),
		},
		User: "postgres",
		WaitingFor: wait.ForLog(postgresReadReplicaReadyLogMsg).
			WithPollInterval(waitingPollInterval).
			WithStartupTimeout(2 * time.Minute),
	}

	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: replicaReq,
		Started:          true,
	})
}

// spinPrimaryContainer spins up the PostgreSQL primary container.
func spinPrimaryContainer(ctx context.Context, dbConfig *DataBaseConfig, networkName string) (testcontainers.Container, error) {
	primaryReq := testcontainers.ContainerRequest{
		Image:          postgresImage,
		ExposedPorts:   []string{DefaultPostgresPort},
		Networks:       []string{networkName},
		NetworkAliases: map[string][]string{networkName: {PrimaryContainerAlias}},
		Env: map[string]string{
			"POSTGRES_PASSWORD":         dbConfig.Password,
			"POSTGRES_USER":             dbConfig.User,
			"POSTGRES_DATABASE":         dbConfig.Name,
			"POSTGRES_DB":               "postgres",
			"POSTGRES_HOST_AUTH_METHOD": "scram-sha-256\nhost replication all 0.0.0.0/0 md5",
			"POSTGRES_INITDB_ARGS":      "--auth-host=scram-sha-256",
		},
		Cmd: []string{
			"postgres",
			"-c", "wal_level=replica",
			"-c", "hot_standby=on",
			"-c", "max_wal_senders=10",
			"-c", "max_replication_slots=10",
			"-c", "hot_standby_feedback=on",
		},
		WaitingFor: wait.ForLog(postgresReadyLogMsg).
			WithPollInterval(waitingPollInterval).
			WithOccurrence(waitingOccurrence).
			WithStartupTimeout(2 * time.Minute),
	}

	primaryContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: primaryReq,
			Started:          true,
		},
	)
	if err != nil {
		return nil, err
	}

	// Postgres sometimes needs a small sleep after appearing ready.
	time.Sleep(postReadySleepDuration)

	commands := []string{
		"CREATE USER replicator WITH REPLICATION ENCRYPTED PASSWORD 'replicator_password';",
		"SELECT pg_create_physical_replication_slot('replication_slot');",
	}

	for _, cmd := range commands {
		execCmd := []string{"psql", "-U", dbConfig.User, "-d", dbConfig.Name, "-c", cmd}
		exitCode, stdout, err := primaryContainer.Exec(ctx, execCmd)
		stdoutBytes, _ := io.ReadAll(stdout)
		if exitCode != 0 {
			fmt.Printf("Error executing command: %s\nstdout: %s\nerr: %v\n", cmd, string(stdoutBytes), err)
			os.Exit(1)
		}
	}

	return primaryContainer, nil
}

// CleanTables cleans all the given tables.
func CleanTables(db *gorm.DB, objects ...interface{}) {
	for _, o := range objects {
		db.Delete(&o, All)
	}
}
