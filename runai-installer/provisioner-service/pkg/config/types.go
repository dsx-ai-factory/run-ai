package config

type Server struct {
	Port int `mapstructure:"port"`
}

func (s *Server) SetDefaults() {
	s.Port = 8080
}

type TLS struct {
	CertPath         string `mapstructure:"runai_tls_cert_path"`
	KeyPath          string `mapstructure:"runai_tls_key_path"`
	CAChainPath      string `mapstructure:"runai_ca_chain_path"`
	WildcardCertPath string `mapstructure:"runai_wildcard_tls_cert_path"`
	WildcardKeyPath  string `mapstructure:"runai_wildcard_tls_key_path"`
}

func (t *TLS) SetDefaults() {
	t.CertPath = "/var/runai/tls/tls.crt"
	t.KeyPath = "/var/runai/tls/tls.key"
	t.CAChainPath = ""
	t.WildcardCertPath = "/var/runai/wildcard/tls.crt"
	t.WildcardKeyPath = "/var/runai/wildcard/tls.key"
}

type JFrog struct {
	Token string `mapstructure:"runai_jfrog_token" config:"secret"`
}

func (j *JFrog) SetDefaults() {
}

type Installation struct {
	ClusterName             string `mapstructure:"runai_cluster_name"`
	CpVersion               string `mapstructure:"runai_cp_version"`
	ClusterVersion          string `mapstructure:"runai_cluster_version"`
	StepTimeoutMinutes      int    `mapstructure:"runai_step_timeout_minutes"`
	ExecutionTimeoutMinutes int    `mapstructure:"runai_execution_timeout_minutes"`
}

// DefaultRunaiVersion is the default Run:AI version installed by the provisioner.
// To install the latest available version instead, set RUNAI_CP_VERSION and
// RUNAI_CLUSTER_VERSION to ">0.0.0-0" via environment variables.
const DefaultRunaiVersion = "2.25.9"

func (i *Installation) SetDefaults() {
	i.ClusterName = "runai-cluster"
	i.CpVersion = DefaultRunaiVersion
	i.ClusterVersion = DefaultRunaiVersion
	i.StepTimeoutMinutes = 15
	i.ExecutionTimeoutMinutes = 50
}
