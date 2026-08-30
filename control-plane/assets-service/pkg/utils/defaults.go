package utils

import "github.com/run-ai/runai/runai-application-packages/convutils"

func ApplyBoolDefault(optionalValue *bool, defaultValue bool) *bool {
	if optionalValue == nil {
		return convutils.Ref(defaultValue)
	}
	return optionalValue
}
