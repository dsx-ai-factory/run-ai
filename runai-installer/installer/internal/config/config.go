package config

import (
	"fmt"
	"strings"
	"time"
)

type InstallerMode int

const (
	ModeTUI InstallerMode = iota
	ModeCLI
	ModeAPI
)

type RunaiInstallationsOptions string

const (
	CPOnly                RunaiInstallationsOptions = "cp-only"
	SelfHosted            RunaiInstallationsOptions = "self-hosted"
	SHWithResearcherSetup RunaiInstallationsOptions = "sh-with-researcher-setup"
	Airgap                RunaiInstallationsOptions = "airgap"
)

func (r RunaiInstallationsOptions) String() string {
	return string(r)
}

func (r *RunaiInstallationsOptions) Set(s string) error {
	switch strings.ToLower(s) {
	case "cp-only":
		*r = CPOnly
	case "self-hosted":
		*r = SelfHosted
	case "sh-with-researcher-setup":
		*r = SHWithResearcherSetup
	case "airgap":
		*r = Airgap
	default:
		*r = SelfHosted
	}
	return nil
}

func (r RunaiInstallationsOptions) Type() string { return "RunaiInstallationsOptions" }

func (r RunaiInstallationsOptions) AllowedInModes() []InstallerMode {
	switch r {
	case CPOnly:
		return []InstallerMode{ModeAPI}
	case SelfHosted:
		return []InstallerMode{ModeTUI, ModeCLI, ModeAPI}
	case SHWithResearcherSetup:
		return []InstallerMode{ModeAPI}
	case Airgap:
		return []InstallerMode{ModeCLI, ModeAPI}
	default:
		return []InstallerMode{ModeTUI, ModeCLI, ModeAPI}
	}
}

func (i InstallerMode) String() string {
	switch i {
	case ModeTUI:
		return "tui"
	case ModeCLI:
		return "lib"
	case ModeAPI:
		return "api"
	default:
		return "tui"
	}
}

func (i *InstallerMode) Set(s string) error {
	switch strings.ToLower(s) {
	case "tui":
		*i = ModeTUI
	case "lib":
		*i = ModeCLI
	case "api":
		*i = ModeAPI
	default:
		*i = ModeTUI
	}
	return nil
}

func (i InstallerMode) Type() string { return "InstallerMode" }

type GPUType int

const (
	GPUTypeReal GPUType = iota
	GPUTypeFake
)

func (g GPUType) String() string {
	switch g {
	case GPUTypeReal:
		return "real"
	case GPUTypeFake:
		return "fake"
	default:
		return "real"
	}
}

func (g *GPUType) Set(s string) error {
	switch strings.ToLower(s) {
	case "real":
		*g = GPUTypeReal
	case "fake":
		*g = GPUTypeFake
	default:
		*g = GPUTypeReal
	}
	return nil
}

func (g GPUType) Type() string { return "GPUType" }

type RunaiEnvType int

const (
	EnvProd RunaiEnvType = iota
	EnvTest
	EnvLab
	EnvStaging
)

func (e RunaiEnvType) String() string {
	switch e {
	case EnvProd:
		return "prod"
	case EnvTest:
		return "test"
	case EnvLab:
		return "lab"
	case EnvStaging:
		return "staging"
	default:
		return "test"
	}
}

func (e *RunaiEnvType) Set(s string) error {
	switch strings.ToLower(s) {
	case "prod":
		*e = EnvProd
	case "test":
		*e = EnvTest
	case "lab":
		*e = EnvLab
	case "staging":
		*e = EnvStaging
	default:
		*e = EnvTest
	}
	return nil
}

func (e RunaiEnvType) Type() string { return "RunaiEnvType" }

type InstallConfig struct {
	DevMode            bool
	LogsDir            string
	StepTimeout        time.Duration
	ExecutionTimeout   time.Duration
	GPUType            GPUType
	Env                RunaiEnvType
	ClusterEnv         *RunaiEnvType
	CpVersion          string
	ClusterVersion     string
	Mode               InstallerMode
	PreReq             ArrayFlags
	ConfigValues       MapFlags
	// NOTE: two mechanisms exist for controlling which components are installed:
	// RunaiInstallFormat selects a named set of components, while SkipControlPlane
	// and SkipCluster suppress individual components at runtime.  Consider
	// consolidating to one approach for uniformity.
	RunaiInstallFormat RunaiInstallationsOptions
	AirgapExtractedDir string
	// SkipControlPlane skips the Control Plane installation step.
	// Useful when only prerequisites need to be installed.
	SkipControlPlane bool
	// SkipCluster skips the RunAI Cluster installation step.
	// Useful when only prerequisites or the Control Plane need to be installed.
	SkipCluster bool
}

func (c *InstallConfig) EffectiveClusterEnv() RunaiEnvType {
	if c.ClusterEnv != nil {
		return *c.ClusterEnv
	}
	return c.Env
}

func (c *InstallConfig) IsAirgap() bool {
	return c.RunaiInstallFormat == Airgap
}

type UpgradeConfig struct {
	DevMode             bool
	LogsDir             string
	StepTimeout         time.Duration
	ExecutionTimeout    time.Duration
	Env                 RunaiEnvType
	ClusterEnv          *RunaiEnvType
	ControlPlaneVersion string
	ClusterVersion      string
	UpgradeControlPlane bool
	UpgradeCluster      bool
	FQDN                string
	ClusterUUID         string
	CACertPath          string

	AirgapPackagePath  string
	RegistryURL        string
	RegistryUsername   string
	RegistryPassword   string
	AirgapExtractedDir string

	CPCarryForwardValues      []string
	ClusterCarryForwardValues []string
}

func (c *UpgradeConfig) EffectiveClusterEnv() RunaiEnvType {
	if c.ClusterEnv != nil {
		return *c.ClusterEnv
	}
	return c.Env
}

func (c *UpgradeConfig) IsAirgap() bool {
	return c.AirgapPackagePath != ""
}

type UninstallConfig struct {
	DevMode          bool
	LogsDir          string
	StepTimeout      time.Duration
	ExecutionTimeout time.Duration
}

func (c *InstallConfig) ValidateRunaiInstallFormat() error {
	if c.RunaiInstallFormat == "" {
		return nil
	}

	allowedModes := c.RunaiInstallFormat.AllowedInModes()
	for _, m := range allowedModes {
		if m == c.Mode {
			return nil
		}
	}

	var modeNames []string
	for _, m := range allowedModes {
		modeNames = append(modeNames, m.String())
	}
	return fmt.Errorf("runai-install-format '%s' is only allowed in mode(s): %s (current mode: %s)",
		c.RunaiInstallFormat, strings.Join(modeNames, ", "), c.Mode)
}
