package common

import (
	orgunit "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

// Unit conversion divisors used to normalize CPU/memory values from upstream
// APIs into LLM-friendly forms (cores, GiB).
const (
	MibPerGiB         = 1024.0
	MbPerGiB          = 1073.741824 // 1 GiB = 2^30 bytes ≈ 1073.741824 decimal MB
	BytesPerGiB       = 1024.0 * 1024.0 * 1024.0
	MillicoresPerCore = 1000.0
)

func CoresToMillicores(cores float64) float64 { return cores * MillicoresPerCore }
func MillicoresToCores(milli float64) float64 { return milli / MillicoresPerCore }
func GiBToMiB(giB float64) float64            { return giB * MibPerGiB }

// MemoryToGiB normalizes a memory value in the upstream unit to GiB. Centralized so the
// read and write tools agree on the conversion (and a written value round-trips back).
func MemoryToGiB(value float64, units orgunit.Units) float64 {
	switch units {
	case orgunit.UNITS_GB:
		return value / (MbPerGiB / 1000.0)
	case orgunit.UNITS_MB:
		return value / MbPerGiB
	default: // Mib, and the API's most common default
		return value / MibPerGiB
	}
}
