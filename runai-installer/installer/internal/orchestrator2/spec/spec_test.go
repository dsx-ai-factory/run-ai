package spec

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validInstallSpec() *Spec {
	return &Spec{
		Command:      CommandInstall,
		Registry:     &Registry{Type: RegistryTypeJFrog, Token: "tok"},
		ControlPlane: &ControlPlane{Domain: "runai.example.com"},
		Cluster:      &Cluster{Name: "cluster01"},
	}
}

func TestValidate_Install_OK(t *testing.T) {
	require.NoError(t, validInstallSpec().Validate())
}

func TestValidate_Install_ControlPlaneOnly(t *testing.T) {
	s := validInstallSpec()
	s.Cluster = nil
	require.NoError(t, s.Validate())
}

func TestValidate_Install_ClusterOnly(t *testing.T) {
	s := validInstallSpec()
	s.ControlPlane = nil
	require.NoError(t, s.Validate())
}

func TestValidate_Install_SelfSignedTLS(t *testing.T) {
	s := validInstallSpec()
	s.SelfSignedTLS = &SelfSignedTLS{Enabled: true}
	require.NoError(t, s.Validate())
}

func TestSelfSignedTLSDNSNames(t *testing.T) {
	cases := []struct {
		name string
		spec *Spec
		want []string
	}{
		{
			name: "control plane only",
			spec: &Spec{ControlPlane: &ControlPlane{Domain: "cp.example.com"}},
			want: []string{"*.cp.example.com", "cp.example.com"},
		},
		{
			name: "cluster only",
			spec: &Spec{Cluster: &Cluster{Name: "cluster", Domain: "cluster.example.com"}},
			want: []string{"*.cluster.example.com", "cluster.example.com"},
		},
		{
			name: "shared domain with inference",
			spec: &Spec{
				ControlPlane: &ControlPlane{Domain: "runai.example.com"},
				Cluster: &Cluster{
					Name:      "cluster",
					Inference: &Inference{WildcardDomain: "*.inference.runai.example.com"},
				},
			},
			want: []string{"*.inference.runai.example.com", "*.runai.example.com", "runai.example.com"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.spec.SelfSignedTLSDNSNames())
		})
	}
}

func TestValidate_Install_SelfSignedTLSHardFailures(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Spec)
		match  string
	}{
		{
			name: "certificate conflict",
			mutate: func(s *Spec) {
				s.SelfSignedTLS = &SelfSignedTLS{Enabled: true, ConflictingFields: []string{"controlPlane.tls"}}
			},
			match: "cannot be combined",
		},
		{
			name: "gateway API",
			mutate: func(s *Spec) {
				s.SelfSignedTLS = &SelfSignedTLS{Enabled: true}
				s.ExternalConnection = &ExternalConnection{Type: ExternalConnectionGatewayAPI}
			},
			match: "Gateway",
		},
		{
			name: "OpenShift",
			mutate: func(s *Spec) {
				s.SelfSignedTLS = &SelfSignedTLS{Enabled: true}
				s.KubernetesDistribution = "openshift"
			},
			match: "OpenShift",
		},
		{
			name: "missing domain",
			mutate: func(s *Spec) {
				s.SelfSignedTLS = &SelfSignedTLS{Enabled: true}
				s.ControlPlane = nil
				s.Cluster.Domain = ""
			},
			match: "domain",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := validInstallSpec()
			tc.mutate(s)
			require.ErrorContains(t, s.Validate(), tc.match)
		})
	}
}

func TestValidate_Install_NeitherEnabled(t *testing.T) {
	s := validInstallSpec()
	s.ControlPlane = nil
	s.Cluster = nil
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one")
}

func TestValidate_Install_RegistryRequired(t *testing.T) {
	s := validInstallSpec()
	s.Registry = nil
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registry")
}

func TestValidate_Install_RegistryUnknownType(t *testing.T) {
	s := validInstallSpec()
	s.Registry = &Registry{Type: "dockerhub", Token: "tok"}
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registry type")
}

func TestValidate_Install_JFrogTokenRequired(t *testing.T) {
	s := validInstallSpec()
	s.Registry = &Registry{Type: RegistryTypeJFrog}
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token")
}

func TestValidate_NGCRegistry_TokenFree_OK(t *testing.T) {
	s := validInstallSpec()
	s.Registry = &Registry{Type: RegistryTypeNGC}
	require.NoError(t, s.Validate(), "NGC supports a credential-free offering")
}

func TestValidate_Install_ControlPlaneDomainRequired(t *testing.T) {
	s := validInstallSpec()
	s.ControlPlane = &ControlPlane{}
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "domain")
}

func TestValidate_Install_ClusterNameRequired(t *testing.T) {
	s := validInstallSpec()
	s.Cluster = &Cluster{}
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster name")
}

func TestValidate_NGCRegistry_OK(t *testing.T) {
	s := validInstallSpec()
	s.Registry = &Registry{Type: RegistryTypeNGC, Token: "ngc-key"}
	require.NoError(t, s.Validate())
}

func TestRegistry_ChartCredentials(t *testing.T) {
	u, p := (&Registry{Type: RegistryTypeJFrog, Token: "tok"}).ChartCredentials()
	assert.Empty(t, u)
	assert.Empty(t, p)

	u, p = (&Registry{Type: RegistryTypeNGC, Token: "ngc-key"}).ChartCredentials()
	assert.Equal(t, "$oauthtoken", u)
	assert.Equal(t, "ngc-key", p)

	u, p = (&Registry{Type: RegistryTypeNGC, Username: "svc", Password: "pw"}).ChartCredentials()
	assert.Equal(t, "svc", u)
	assert.Equal(t, "pw", p)

	u, p = (&Registry{Type: RegistryTypeNGC}).ChartCredentials()
	assert.Empty(t, u)
	assert.Empty(t, p)
}

func TestRegistry_ImagePullConfig(t *testing.T) {
	cfg, ok := (&Registry{Type: RegistryTypeJFrog, Token: "encoded"}).ImagePullConfig()
	assert.True(t, ok)
	assert.Equal(t, "encoded", cfg)

	cfg, ok = (&Registry{Type: RegistryTypeNGC, Token: "ngc-key"}).ImagePullConfig()
	require.True(t, ok)
	raw, err := base64.StdEncoding.DecodeString(cfg)
	require.NoError(t, err)
	var parsed struct {
		Auths map[string]struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"auths"`
	}
	require.NoError(t, json.Unmarshal(raw, &parsed))
	entry, present := parsed.Auths["nvcr.io"]
	require.True(t, present)
	assert.Equal(t, "$oauthtoken", entry.Username)
	assert.Equal(t, "ngc-key", entry.Password)

	_, ok = (&Registry{Type: RegistryTypeNGC}).ImagePullConfig()
	assert.False(t, ok)
}

func TestValidate_Install_Airgap_OK(t *testing.T) {
	s := validInstallSpec()
	s.Connectivity = ConnectivityAirGapped
	s.Registry = nil
	s.Airgap = &Airgap{BundlePath: "/tmp/bundle.tar.gz", WorkingDir: "/data/runai-airgap"}
	require.NoError(t, s.Validate())
}

func TestValidate_Install_Airgap_BundleRequired(t *testing.T) {
	s := validInstallSpec()
	s.Connectivity = ConnectivityAirGapped
	s.Registry = nil
	s.Airgap = nil
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "airgap bundle")
}

func TestValidate_Install_Airgap_WorkingDirRequired(t *testing.T) {
	s := validInstallSpec()
	s.Connectivity = ConnectivityAirGapped
	s.Registry = nil
	s.Airgap = &Airgap{BundlePath: "/tmp/bundle.tar.gz"}
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "working directory")
}

func TestValidate_Install_Airgap_SkipPrepare_BundleNotRequired(t *testing.T) {
	s := validInstallSpec()
	s.Connectivity = ConnectivityAirGapped
	s.Registry = nil
	s.Airgap = &Airgap{WorkingDir: "/data/runai-airgap", SkipPrepare: true}
	require.NoError(t, s.Validate())
}

func TestValidate_Install_Airgap_SkipPrepare_WorkingDirRequired(t *testing.T) {
	s := validInstallSpec()
	s.Connectivity = ConnectivityAirGapped
	s.Registry = nil
	s.Airgap = &Airgap{SkipPrepare: true}
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "working directory")
}

func TestValidate_Install_Connected_RegistryStillRequired(t *testing.T) {
	s := validInstallSpec()
	s.Connectivity = ConnectivityConnected
	s.Registry = nil
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "registry")
}

func TestValidate_UnknownConnectivity(t *testing.T) {
	s := validInstallSpec()
	s.Connectivity = "intranet"
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connectivity")
}

func TestValidate_UnknownEnv(t *testing.T) {
	s := validInstallSpec()
	s.Env = "moon"
	err := s.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "env")
}

func TestCluster_HostRouting(t *testing.T) {
	cases := []struct {
		name    string
		routing string
		helm    map[string]string
		want    bool
	}{
		{name: "default is host", want: true},
		{name: "explicit host", routing: "host", want: true},
		{name: "explicit path", routing: "path", want: false},
		{name: "override path", helm: map[string]string{SubdomainSupportHelmKey: "false"}, want: false},
		{name: "override host", helm: map[string]string{SubdomainSupportHelmKey: "true"}, want: true},
		{name: "explicit setting beats override", routing: "path", helm: map[string]string{SubdomainSupportHelmKey: "true"}, want: false},
		{name: "unparseable override falls back to default", helm: map[string]string{SubdomainSupportHelmKey: "yes"}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Cluster{WorkloadRouting: tc.routing, CustomHelmValues: tc.helm}
			assert.Equal(t, tc.want, c.HostRouting())
		})
	}
	var nilCluster *Cluster
	assert.True(t, nilCluster.HostRouting(), "nil cluster defaults to host")
}

func TestCluster_SubdomainOverride(t *testing.T) {
	host, present, valid := (&Cluster{}).SubdomainOverride()
	assert.False(t, present)
	assert.False(t, valid)
	assert.False(t, host)

	host, present, valid = (&Cluster{CustomHelmValues: map[string]string{SubdomainSupportHelmKey: "false"}}).SubdomainOverride()
	assert.True(t, present)
	assert.True(t, valid)
	assert.False(t, host)

	_, present, valid = (&Cluster{CustomHelmValues: map[string]string{SubdomainSupportHelmKey: "maybe"}}).SubdomainOverride()
	assert.True(t, present)
	assert.False(t, valid)
}
