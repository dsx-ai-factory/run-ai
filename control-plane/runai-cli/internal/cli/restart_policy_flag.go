package cli

import (
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	typeNullableRestartPolicy = "restart-policy"
)

type NullableRestartPolicyFlag struct {
	value *openapi.NullableRestartPolicy
}

func NewNullableRestartPolicyFlag(value *openapi.NullableRestartPolicy) *NullableRestartPolicyFlag {
	return &NullableRestartPolicyFlag{value: value}
}

func (rp *NullableRestartPolicyFlag) String() string {
	if !rp.IsSet() {
		return ""
	}
	return string(*rp.value.Get())
}

func (rp *NullableRestartPolicyFlag) IsSet() bool {
	return rp.value.IsSet()
}

func (rp *NullableRestartPolicyFlag) Set(value string) error {
	restartPolicy, err := openapi.NewRestartPolicyFromValue(value)
	if err != nil {
		return err
	}
	rp.value.Set(restartPolicy)
	return nil
}

func (rp *NullableRestartPolicyFlag) Value() *openapi.NullableRestartPolicy {
	return rp.value
}

// Type for CLI help value
func (rp *NullableRestartPolicyFlag) Type() string {
	return typeNullableRestartPolicy
}
