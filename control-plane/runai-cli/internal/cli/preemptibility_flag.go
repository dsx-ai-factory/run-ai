package cli

import (
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	typeNullablePreemptibility = "preemptibility"
)

type NullablePreemptibilityFlag struct {
	value *openapi.NullablePreemptibility
}

func NewNullablePreemptibilityFlag(value *openapi.NullablePreemptibility) *NullablePreemptibilityFlag {
	return &NullablePreemptibilityFlag{value: value}
}

func (pf *NullablePreemptibilityFlag) String() string {
	if !pf.IsSet() {
		return ""
	}
	return string(*pf.value.Get())
}

func (pf *NullablePreemptibilityFlag) IsSet() bool {
	return pf.value.IsSet()
}

func (pf *NullablePreemptibilityFlag) Set(value string) error {
	preemptibility, err := openapi.NewPreemptibilityFromValue(value)
	if err != nil {
		return err
	}
	pf.value.Set(preemptibility)
	return nil
}

func (pf *NullablePreemptibilityFlag) Value() *openapi.NullablePreemptibility {
	return pf.value
}

// Type for CLI help value
func (pf *NullablePreemptibilityFlag) Type() string {
	return typeNullablePreemptibility
}
