package spec_checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	infw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"sort"
	"strconv"
)

const (
	SpecChecksumAnnotation         = "run.ai/workload-spec-checksum"
	SpecChecksumMigratedAnnotation = "run.ai/workload-spec-checksum-migrated"
)

// canonicalJSON creates a canonical JSON representation of the input,
// by sorting the keys of all maps in the object
func canonicalJSON(v interface{}) ([]byte, error) {
	// marshal the data into a map[string]interface{}
	var m map[string]interface{}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return json.Marshal(sortedMap(m))
}

// sortedMap recursively sorts all maps in the object
func sortedMap(m map[string]interface{}) map[string]interface{} {
	sorted := make(map[string]interface{})
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		value := m[k]
		switch v := value.(type) {
		case map[string]interface{}:
			sorted[k] = sortedMap(v)
		case []interface{}:
			sorted[k] = sortedSlice(v)
		default:
			sorted[k] = v
		}
	}
	return sorted
}

// sortedSlice recursively sorts all maps in slices
func sortedSlice(s []interface{}) []interface{} {
	for i, v := range s {
		switch v := v.(type) {
		case map[string]interface{}:
			s[i] = sortedMap(v)
		case []interface{}:
			s[i] = sortedSlice(v)
		}
	}
	return s
}

// CalculateSpecChecksum calculates the SHA-256 checksum of the InferenceWorkload's spec.
func CalculateSpecChecksum(workload *infw.InferenceWorkload) (string, error) {
	specJSON, err := canonicalJSON(workload.Spec)
	if err != nil {
		return "", fmt.Errorf("failed to create canonical JSON of spec: %w", err)
	}

	checksum := sha256.Sum256(specJSON)
	return fmt.Sprintf("%x", checksum), nil
}

// IsLegacyMD5Checksum reports whether s is a legacy MD5-format checksum:
// exactly 32 characters, all valid lowercase hex digits.
func IsLegacyMD5Checksum(s string) bool {
	if len(s) != 32 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// SetSpecChecksumAnnotation calculates and sets the spec checksum as an annotation on the InferenceWorkload
func SetSpecChecksumAnnotation(workload *infw.InferenceWorkload, checksum string) (bool, error) {

	if checksum == "" {
		var err error
		checksum, err = CalculateSpecChecksum(workload)
		if err != nil {
			return false, err
		}
	}

	if workload.Annotations == nil {
		workload.Annotations = make(map[string]string)
	}

	if workload.Annotations[SpecChecksumAnnotation] == checksum {
		return false, nil
	}

	workload.Annotations[SpecChecksumAnnotation] = checksum

	log.Debug().Msgf("Updated annotation %s for workload %s/%s: %s\n",
		SpecChecksumAnnotation, workload.Name, workload.Namespace, checksum)
	return true, nil
}

func SetSpecChecksumMigratedAnnotation(workload *infw.InferenceWorkload, migrated bool) bool {
	if workload.Annotations == nil {
		workload.Annotations = make(map[string]string)
	}

	value := strconv.FormatBool(migrated)
	if workload.Annotations[SpecChecksumMigratedAnnotation] == value {
		return false
	}

	workload.Annotations[SpecChecksumMigratedAnnotation] = value
	return true
}

func GetSpecChecksumAnnotation(workload *infw.InferenceWorkload) string {
	return workload.Annotations[SpecChecksumAnnotation]
}

func GetSpecChecksumMigratedAnnotation(workload *infw.InferenceWorkload) bool {
	return workload.Annotations[SpecChecksumMigratedAnnotation] == "true"
}
