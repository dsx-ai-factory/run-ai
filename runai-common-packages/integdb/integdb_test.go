package integdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveConfig_NilConf_TakenFromDefaults(t *testing.T) {
	conf := resolveConfig(nil)
	require.Equal(t, "16", conf.ImageVersion)
	require.Equal(t, "postgres", conf.User)
	require.Equal(t, "postgres", conf.Password)
	require.Equal(t, "backend", conf.Name)
	require.Equal(t, "test", conf.Schema)
	require.Equal(t, "file:///migrations", conf.MigrationDirPath)
}

func TestResolveConfig_NilConf_TakenFromEnv(t *testing.T) {
	t.Setenv("INTEGDB_IMAGE_VERSION", "9")
	t.Setenv("INTEGDB_USER", "custom")
	t.Setenv("INTEGDB_PASSWORD", "secret")
	t.Setenv("INTEGDB_NAME", "mydb")
	t.Setenv("INTEGDB_SCHEMA", "my_schema")
	t.Setenv("INTEGDB_MIGRATION_DIR_PATH", "file:///tmp")

	conf := resolveConfig(nil)
	require.Equal(t, "9", conf.ImageVersion)
	require.Equal(t, "custom", conf.User)
	require.Equal(t, "secret", conf.Password)
	require.Equal(t, "mydb", conf.Name)
	require.Equal(t, "my_schema", conf.Schema)
	require.Equal(t, "file:///tmp", conf.MigrationDirPath)
}

func TestResolveConfig_Partial_DefaultsFillGaps(t *testing.T) {
	conf := resolveConfig(&IntegdbConfig{
		Schema:           "my_schema",
		MigrationDirPath: "file:///tmp",
		ImageVersion:     "9",
	})

	require.Equal(t, "9", conf.ImageVersion)
	require.Equal(t, "my_schema", conf.Schema)
	require.Equal(t, "file:///tmp", conf.MigrationDirPath)
	require.Equal(t, "postgres", conf.User)
	require.Equal(t, "postgres", conf.Password)
	require.Equal(t, "backend", conf.Name)
}

func TestResolveConfig_Full_NoOverride(t *testing.T) {
	conf := resolveConfig(&IntegdbConfig{
		User:             "custom",
		Password:         "secret",
		Name:             "mydb",
		Schema:           "my_schema",
		MigrationDirPath: "file:///tmp",
		ImageVersion:     "9",
	})
	require.Equal(t, "9", conf.ImageVersion)
	require.Equal(t, "custom", conf.User)
	require.Equal(t, "secret", conf.Password)
	require.Equal(t, "mydb", conf.Name)
	require.Equal(t, "my_schema", conf.Schema)
	require.Equal(t, "file:///tmp", conf.MigrationDirPath)
}

func TestGenerateEncryptionKey(t *testing.T) {
	key := generateEncryptionKey()
	require.Equal(t, 32, len(key))
}

func TestMajorVersion(t *testing.T) {
	require.Equal(t, "16", majorVersion("16"))
	require.Equal(t, "16", majorVersion("16.2"))
	require.Equal(t, "16", majorVersion("16-bookworm"))
	require.Equal(t, "17", majorVersion("17.0-alpine"))
}
