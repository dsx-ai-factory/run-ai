package cli

import (
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	typeDistributedInferenceStartupPolicy = "distributed-inference-startup-policy"
)

type DistributedInferenceStartupPolicyFlag struct {
	value *openapi.NullableDistributedInferenceStartupPolicy
}

func NewDistributedInferenceStartupPolicyFlag(value *openapi.NullableDistributedInferenceStartupPolicy) *DistributedInferenceStartupPolicyFlag {
	return &DistributedInferenceStartupPolicyFlag{value: value}
}

func (sp *DistributedInferenceStartupPolicyFlag) String() string {
	if !sp.IsSet() {
		return ""
	}
	return string(*sp.value.Get())
}

func (sp *DistributedInferenceStartupPolicyFlag) IsSet() bool {
	return sp.value.IsSet()
}

func (sp *DistributedInferenceStartupPolicyFlag) Set(value string) error {
	startupPolicy, err := openapi.NewDistributedInferenceStartupPolicyFromValue(value)
	if err != nil {
		return err
	}
	sp.value.Set(startupPolicy)
	return nil
}

func (sp *DistributedInferenceStartupPolicyFlag) Value() *openapi.NullableDistributedInferenceStartupPolicy {
	return sp.value
}

func (sp *DistributedInferenceStartupPolicyFlag) Type() string {
	return typeDistributedInferenceStartupPolicy
}
