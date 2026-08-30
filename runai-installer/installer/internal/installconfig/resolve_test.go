package installconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeToken(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	return path
}

func TestResolveRegistry_JFrogFlag(t *testing.T) {
	path := writeToken(t, "  jfrog-token\n")

	reg, err := ResolveRegistry(&Config{}, path, "")
	require.NoError(t, err)
	assert.Equal(t, spec.RegistryTypeJFrog, reg.Type)
	assert.Equal(t, "jfrog-token", reg.Token)
}

func TestResolveRegistry_NGCFlag(t *testing.T) {
	path := writeToken(t, "ngc-key")

	reg, err := ResolveRegistry(&Config{}, "", path)
	require.NoError(t, err)
	assert.Equal(t, spec.RegistryTypeNGC, reg.Type)
	assert.Equal(t, "ngc-key", reg.Token)
}

func TestResolveRegistry_BothFlagsRejected(t *testing.T) {
	p1 := writeToken(t, "a")
	p2 := writeToken(t, "b")

	_, err := ResolveRegistry(&Config{}, p1, p2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestResolveRegistry_FromConfig(t *testing.T) {
	path := writeToken(t, "cfg-token")
	cfg := &Config{Registry: &Registry{Type: "ngc", TokenPath: path}}

	reg, err := ResolveRegistry(cfg, "", "")
	require.NoError(t, err)
	assert.Equal(t, spec.RegistryTypeNGC, reg.Type)
	assert.Equal(t, "cfg-token", reg.Token)
}

func TestResolveRegistry_FlagOverridesConfig(t *testing.T) {
	cfgPath := writeToken(t, "cfg-token")
	flagPath := writeToken(t, "flag-token")
	cfg := &Config{Registry: &Registry{Type: "ngc", TokenPath: cfgPath}}

	reg, err := ResolveRegistry(cfg, flagPath, "")
	require.NoError(t, err)
	assert.Equal(t, spec.RegistryTypeJFrog, reg.Type)
	assert.Equal(t, "flag-token", reg.Token)
}

func TestResolveRegistry_FlagCarriesConfigHostAndCA(t *testing.T) {
	flagPath := writeToken(t, "flag-token")
	// Only Host and CAChainPath carry over from the config; the type and token
	// come from the --jfrog-token-path flag, which takes precedence.
	cfg := &Config{Registry: &Registry{
		Host:        "mirror.local:5000",
		CAChainPath: "/etc/runai/registry-ca.crt",
	}}

	reg, err := ResolveRegistry(cfg, flagPath, "")
	require.NoError(t, err)
	assert.Equal(t, spec.RegistryTypeJFrog, reg.Type)
	assert.Equal(t, "flag-token", reg.Token)
	assert.Equal(t, "mirror.local:5000", reg.Host, "host carries over from config when a token flag is used")
	assert.Equal(t, "/etc/runai/registry-ca.crt", reg.CAChainPath, "CA chain carries over from config when a token flag is used")
}

func TestResolveRegistry_NoSource(t *testing.T) {
	_, err := ResolveRegistry(&Config{}, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no registry credentials")
}

func TestResolveRegistry_UnknownConfigType(t *testing.T) {
	path := writeToken(t, "tok")
	cfg := &Config{Registry: &Registry{Type: "dockerhub", TokenPath: path}}

	_, err := ResolveRegistry(cfg, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown registry type")
}

func TestResolveRegistry_MissingFile(t *testing.T) {
	_, err := ResolveRegistry(&Config{}, filepath.Join(t.TempDir(), "nope"), "")
	require.Error(t, err)
}

func TestResolveRegistry_EmptyToken(t *testing.T) {
	path := writeToken(t, "   \n")

	_, err := ResolveRegistry(&Config{}, path, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestValidateRegistryAuth(t *testing.T) {
	cases := []struct {
		name    string
		reg     Registry
		want    registryAuthMode
		wantErr string
	}{
		{name: "host only is unauthenticated", reg: Registry{Host: "h"}, want: authNone},
		{name: "token inline", reg: Registry{Type: "jfrog", Token: "t"}, want: authToken},
		{name: "token path", reg: Registry{Type: "ngc", TokenPath: "/p"}, want: authToken},
		{name: "basic inline", reg: Registry{Username: "u", Password: "p"}, want: authBasic},
		{name: "basic path", reg: Registry{Username: "u", PasswordPath: "/p"}, want: authBasic},
		{name: "token inline and path", reg: Registry{Token: "t", TokenPath: "/p"}, wantErr: "token or tokenPath"},
		{name: "password inline and path", reg: Registry{Username: "u", Password: "p", PasswordPath: "/p"}, wantErr: "password or passwordPath"},
		{name: "basic and token", reg: Registry{Username: "u", Password: "p", Token: "t"}, wantErr: "mutually exclusive"},
		{name: "password without username", reg: Registry{Password: "p"}, wantErr: "requires a username"},
		{name: "username without password", reg: Registry{Username: "u"}, wantErr: "requires a password"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mode, err := validateRegistryAuth(&tc.reg)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, mode)
		})
	}
}

func TestResolveRegistry_BasicInline(t *testing.T) {
	cfg := &Config{Registry: &Registry{Host: "harbor.local", Username: "admin", Password: "Harbor12345"}}

	reg, err := ResolveRegistry(cfg, "", "")
	require.NoError(t, err)
	assert.Equal(t, "harbor.local", reg.Host)
	assert.Equal(t, "admin", reg.Username)
	assert.Equal(t, "Harbor12345", reg.Password)
	assert.Empty(t, reg.Token)
}

func TestResolveRegistry_BasicPasswordPath(t *testing.T) {
	path := writeToken(t, "  Harbor12345\n")
	cfg := &Config{Registry: &Registry{Host: "harbor.local", Username: "admin", PasswordPath: path}}

	reg, err := ResolveRegistry(cfg, "", "")
	require.NoError(t, err)
	assert.Equal(t, "admin", reg.Username)
	assert.Equal(t, "Harbor12345", reg.Password)
}

func TestResolveRegistry_TokenInline(t *testing.T) {
	cfg := &Config{Registry: &Registry{Type: "jfrog", Token: "inline-token"}}

	reg, err := ResolveRegistry(cfg, "", "")
	require.NoError(t, err)
	assert.Equal(t, spec.RegistryTypeJFrog, reg.Type)
	assert.Equal(t, "inline-token", reg.Token)
}

func TestResolveRegistry_HostOnly(t *testing.T) {
	cfg := &Config{Registry: &Registry{Host: "harbor.local"}}

	reg, err := ResolveRegistry(cfg, "", "")
	require.NoError(t, err)
	assert.Equal(t, "harbor.local", reg.Host)
	assert.Empty(t, reg.Token)
	assert.Empty(t, reg.Username)
}

func TestResolveRegistry_InvalidCombo(t *testing.T) {
	cfg := &Config{Registry: &Registry{Username: "admin", Password: "p", Token: "t"}}

	_, err := ResolveRegistry(cfg, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestResolveRegistry_BasicMissingPasswordFile(t *testing.T) {
	cfg := &Config{Registry: &Registry{Username: "admin", PasswordPath: filepath.Join(t.TempDir(), "nope")}}

	_, err := ResolveRegistry(cfg, "", "")
	require.Error(t, err)
}
