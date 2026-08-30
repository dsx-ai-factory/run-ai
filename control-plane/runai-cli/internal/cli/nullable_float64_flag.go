package cli

import (
	"strconv"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"k8s.io/utils/ptr"
)

const (
	typeNullableFloat64 = "float64"
)

type NullableFloat64Flag struct {
	value *openapi.NullableFloat64
}

func NewNullableFloat64Flag(value *openapi.NullableFloat64) *NullableFloat64Flag {
	return &NullableFloat64Flag{
		value: value,
	}
}

func (nf *NullableFloat64Flag) IsSet() bool {
	return nf.value.IsSet()
}

func (nf *NullableFloat64Flag) String() string {
	if !nf.value.IsSet() {
		return ""
	}
	return strconv.FormatFloat(*nf.value.Get(), 'f', -1, 64)
}

func (nf *NullableFloat64Flag) Value() *float64 {
	return nf.value.Get()
}

func (nf *NullableFloat64Flag) Set(value string) error {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}

	nf.value.Set(ptr.To(float64(f)))
	return nil
}

func (nf *NullableFloat64Flag) Type() string {
	return typeNullableFloat64
}
