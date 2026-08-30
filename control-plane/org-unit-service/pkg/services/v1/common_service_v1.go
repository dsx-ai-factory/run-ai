package v1

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	"github.com/run-ai/runai/runai-application-packages/versions"
)

func ValidateMetricsSupported(clusterVersion string) error {
	supported, err := versions.IsVersionSupported(clusterVersion, MetricsClusterMinVersion)
	if err != nil {
		return err
	}
	if !supported {
		return fmt.Errorf("getting project metrics is not supported for cluster version %s", clusterVersion)
	}
	return nil
}

func ValidateMetricsRequest(ctx context.Context, clusterId string, start time.Time, end time.Time, samples *int32, clustersClient cluster_service_client.IClusterExternalClient) error {
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return err
	}
	if start.After(end) {
		return errors.New("start cannot be after end")
	}
	if samples == nil {
		return errors.New("number of samples can not be nil")
	}
	if *samples == 0 {
		*samples = DefaultNumberOfSamples
	} else if samples != nil && *samples > MaxNumberOfSamples {
		*samples = MaxNumberOfSamples
	}
	clusterVersion, err := services.GetClusterVersion(ctx, clusterUuid, clustersClient)
	if err != nil {
		return err
	}
	return ValidateMetricsSupported(clusterVersion)
}

func GetTenantAndClusterUuid(ctx context.Context, clusterId string) (*string, uuid.UUID, error) {
	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return nil, uuid.UUID{}, err
	}

	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return nil, uuid.UUID{}, err
	}

	return tenantId, clusterUuid, nil
}

func ShouldFetchUtilization(permittedScopes *openapi.PermittedScopes) bool {
	return permittedScopes == nil ||
		len(permittedScopes.Clusters) > 0 ||
		len(permittedScopes.Tenants) > 0 ||
		(permittedScopes.System != nil && *(permittedScopes.System))
}

// StringOrFloatHookFunc is a decode hook function to convert strings and floats to float32.
func StringOrFloatHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf(float32(0)) {
		return data, nil
	}

	switch f.Kind() {
	case reflect.String:
		strVal, ok := data.(string)
		if !ok {
			return data, nil
		}
		floatVal, err := strconv.ParseFloat(strVal, 32)
		if err != nil {
			return nil, fmt.Errorf("could not parse %v to float32: %v", data, err)
		}
		return float32(floatVal), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal := reflect.ValueOf(data).Int()
		return float32(intVal), nil
	case reflect.Float32, reflect.Float64:
		return float32(reflect.ValueOf(data).Float()), nil
	}

	return data, nil
}

// StringOrIntHookFunc is a decode hook function to convert strings and int to int32.
func StringOrIntHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf(int32(0)) {
		return data, nil
	}

	switch f.Kind() {
	case reflect.String:
		strVal, ok := data.(string)
		if !ok {
			return data, nil
		}
		intVal, err := strconv.ParseInt(strVal, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("could not parse %v to int32: %v", data, err)
		}
		return int32(intVal), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal := reflect.ValueOf(data).Int()
		return int32(intVal), nil
	case reflect.Float32, reflect.Float64:
		floatVal := reflect.ValueOf(data).Float()
		return int32(floatVal), nil
	}

	return data, nil
}
