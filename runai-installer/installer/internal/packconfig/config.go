// Package packconfig defines the airgap pack config file and maps it onto the
// orchestrator spec, mirroring installconfig's conventions.
package packconfig

type Config struct {
	Settings       Settings         `yaml:"settings,omitempty"`
	RegistryAuth   []RegistryAuth   `yaml:"registryAuth,omitempty"`
	ImageOverrides []ArtifactSource `yaml:"imageOverrides,omitempty"`
}

type Settings struct {
	RunaiVersion string `yaml:"runaiVersion,omitempty"`
	Architecture string `yaml:"architecture,omitempty"`
	Output       string `yaml:"output,omitempty"`
}

// RegistryAuth holds the credentials for one registry, using exactly one of
// three schemes: token (token, tokenPath or tokenEnv), basic (username with
// password, passwordPath or passwordEnv), or none for an anonymous registry.
type RegistryAuth struct {
	Registry           string `yaml:"registry"`
	Username           string `yaml:"username,omitempty"`
	Password           string `yaml:"password,omitempty"`
	PasswordPath       string `yaml:"passwordPath,omitempty"`
	PasswordEnv        string `yaml:"passwordEnv,omitempty"`
	Token              string `yaml:"token,omitempty"`
	TokenPath          string `yaml:"tokenPath,omitempty"`
	TokenEnv           string `yaml:"tokenEnv,omitempty"`
	BearerToken        string `yaml:"bearerToken,omitempty"`
	BearerTokenPath    string `yaml:"bearerTokenPath,omitempty"`
	BearerTokenEnv     string `yaml:"bearerTokenEnv,omitempty"`
	CAChainPath        string `yaml:"caChainPath,omitempty"`
	Insecure           bool   `yaml:"insecure,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify,omitempty"`
}

type ArtifactSource struct {
	Group  string `yaml:"group,omitempty"`
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
}
