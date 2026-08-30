package services

import (
	"math"

	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
)

const (
	MibToMbMultiplyFactor = 1.048576
)

type MemConvertFunc func(float64) float64

func ApplyMemoryConversion(value *float64, f MemConvertFunc) *float64 {
	if value == nil || *value == -1 {
		return value
	}
	converted := f(*value)
	return &converted
}

func MbToMib(x float64) float64 {
	return math.Round(float64(x)/MibToMbMultiplyFactor*1000) / 1000
}

func MibToMb(x float64) float64 {
	return math.Round(float64(x)*MibToMbMultiplyFactor*1000) / 1000
}

func GbToMib(x float64) float64 {
	return MbToMib(GbToMb(x))
}

func GbToMb(x float64) float64 {
	return float64(int64(x)) * 1000
}

func MibToGb(x float64) float64 {
	return MbToGb(MibToMb(x))
}

func MbToGb(x float64) float64 {
	return float64(int64(x)) / 1000
}

func ConvertNone(x float64) float64 {
	return x * 1
}

func GetConversionFn(units api.Units, memoryUnitMb bool) MemConvertFunc {
	conversionFn := ConvertNone
	if units == api.UNITS_MIB && memoryUnitMb {
		conversionFn = MibToMb
	}
	if units == api.UNITS_GB && memoryUnitMb {
		conversionFn = GbToMb
	}
	if units == api.UNITS_GB && !memoryUnitMb {
		conversionFn = GbToMib
	}
	if units == api.UNITS_MB && !memoryUnitMb {
		conversionFn = MbToMib
	}
	return conversionFn
}

func GetMemoryConversionFn(from, to api.Units) MemConvertFunc {
	if from == to {
		return ConvertNone
	}

	if from == api.UNITS_MIB {
		if to == api.UNITS_MB {
			return MibToMb
		}
		if to == api.UNITS_GB {
			return MibToGb
		}
	}

	if from == api.UNITS_MB {
		if to == api.UNITS_MIB {
			return MbToMib
		}
		if to == api.UNITS_GB {
			return MbToGb
		}
	}

	if from == api.UNITS_GB {
		if to == api.UNITS_MIB {
			return GbToMib
		}
		if to == api.UNITS_MB {
			return GbToMb
		}
	}

	return ConvertNone
}
