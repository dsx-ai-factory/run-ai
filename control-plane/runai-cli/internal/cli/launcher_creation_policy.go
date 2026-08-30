package cli

import openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"

const (
	typeNullableLauncherCreationPolicy = "launcher-creation-policy"

	MPILauncherCreationPolicyAtStartup          = "AtStartup"
	MPILauncherCreationPolicyWaitForWorkerReady = "WaitForWorkersReady"
)

type NullableLauncherCreationPolicyFlag struct {
	value *openapi.NullableMpiLauncherCreationPolicy
}

func NewNullableLauncherCreationPolicyFlag(value *openapi.NullableMpiLauncherCreationPolicy) *NullableLauncherCreationPolicyFlag {
	return &NullableLauncherCreationPolicyFlag{value: value}
}

func (lcp *NullableLauncherCreationPolicyFlag) String() string {
	if !lcp.IsSet() {
		return ""
	}
	return string(*lcp.value.Get())
}

func (lcp *NullableLauncherCreationPolicyFlag) IsSet() bool {
	return lcp.value.IsSet()
}

func (lcp *NullableLauncherCreationPolicyFlag) Set(value string) error {
	launcherCreationPolicy, err := openapi.NewMpiLauncherCreationPolicyFromValue(value)
	if err != nil {
		return err
	}
	lcp.value.Set(launcherCreationPolicy)
	return nil
}

func (lcp *NullableLauncherCreationPolicyFlag) Value() *openapi.NullableMpiLauncherCreationPolicy {
	return lcp.value
}

// Type for CLI help value
func (lcp *NullableLauncherCreationPolicyFlag) Type() string {
	return typeNullableLauncherCreationPolicy
}
