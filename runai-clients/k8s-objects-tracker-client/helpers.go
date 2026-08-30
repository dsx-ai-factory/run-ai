package k8s_objects_tracker_client

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	k8sObjectsApi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
)

// IsStorageClassActionPermitted check if a given action flag indicates that action is allowed. note that
// BoolIsTrue/BoolIsFalse cannot be used for this purpose, because they assume no value
// indicates false (not allowed), while in this case no value indicates true (allowed)
func IsStorageClassActionPermitted(permissionFlag k8sObjectsApi.NullableBool) bool {

	// flag not specified - assumption is no restriction
	if permissionFlag.Get() == nil {
		return true
	}

	return *permissionFlag.Get() == true
}

// ValidatePvcAttValueVsConfig validates a value of pvc added attribute VS the configuration of this attr in the storage class confiuguration
func ValidatePvcAttValueVsConfig(config *k8sObjectsApi.PvcAddedAttribute, keyPrefix string, value string) *k8sObjectsApi.Error {
	if config == nil {
		return nil
	}

	switch *config.Format.Get() {
	case k8sObjectsApi.PVCADDEDATTRIBUTEFORMAT_NUMBER:
		asFloat, convErr := strconv.ParseFloat(value, 64)
		if convErr != nil {
			return k8sObjectsApi.NewError(http.StatusBadRequest, chooseError(config, keyPrefix, "invalid value for numeric attribute"))
		}

		var minValue float64

		if config.MinValue.Get() != nil {
			minValue = float64(*config.MinValue.Get())
			if asFloat < float64(*config.MinValue.Get()) {
				return k8sObjectsApi.NewError(http.StatusBadRequest, chooseError(config, keyPrefix, fmt.Sprintf("value must be no less then %v", minValue)))
			}
		}

		if config.MaxValue.Get() != nil {
			maxValue := float64(*config.MaxValue.Get())
			if asFloat > maxValue {
				return k8sObjectsApi.NewError(http.StatusBadRequest, chooseError(config, keyPrefix, fmt.Sprintf("value must be no greater then %v", maxValue)))
			}
		}

		if config.Step.Get() != nil {
			step := float64(*config.Step.Get())
			diff := (asFloat - minValue) / step
			if math.Abs(diff-math.Round(diff)) > 1e-7 {
				return k8sObjectsApi.NewError(http.StatusBadRequest, chooseError(config, keyPrefix, fmt.Sprintf("value does not comply step of %v from %v", step, minValue)))
			}
		}

	case k8sObjectsApi.PVCADDEDATTRIBUTEFORMAT_TEXT:
		if convutils.IsStringHasValue(config.ValidationRegexp.Get()) {
			match, _ := regexp.MatchString(*config.ValidationRegexp.Get(), value)
			if !match {
				return k8sObjectsApi.NewError(http.StatusBadRequest, chooseError(config, keyPrefix, "malformed value"))
			}
		}

	case k8sObjectsApi.PVCADDEDATTRIBUTEFORMAT_BOOLEAN:
		if value != "true" && value != "false" {
			return k8sObjectsApi.NewError(http.StatusBadRequest, chooseError(config, keyPrefix, "value can be either 'true' or 'false'"))
		}
	}

	return nil
}

func chooseError(config *k8sObjectsApi.PvcAddedAttribute, keyPrefix string, customError string) string {
	if convutils.IsStringHasValue(config.ValidationError.Get()) {
		return *config.ValidationError.Get()
	}

	return keyPrefix + customError
}
