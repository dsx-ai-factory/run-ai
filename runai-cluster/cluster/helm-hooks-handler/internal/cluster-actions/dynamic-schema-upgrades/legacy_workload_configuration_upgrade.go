package dynamic_schema_upgrades

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	openShiftRouteConfigurationField     = "openshiftRouteConfiguration"
	openShiftRouteTerminationPolicyField = "terminationPolicy"
)

var legacyWorkloadControllerFields = []string{
	openShiftRouteConfigurationField,
	"gitSyncImage",
	"s3FileSystemImage",
	"runaiOauth2ProxyImage",
	"runaiOauth2ProxyRedisFQDN",
	"runaiOauth2ProxyWorkloadAuthSecretName",
	"runaiOauth2ProxySessionStoreType",
	"runaiOauth2ProxyUsernameLoaderImage",
	"externalAuthUrlEnabled",
}

var legacyWorkloadControllerOperands = []string{
	workloadControllerOperandName,
	inferenceWorkloadControllerOperandName,
}

type legacyWorkloadConfigurationUpgrade struct{}

func newLegacyWorkloadConfigurationUpgrader() *legacyWorkloadConfigurationUpgrade {
	return &legacyWorkloadConfigurationUpgrade{}
}

func (u *legacyWorkloadConfigurationUpgrade) shouldPerformUpgrade(resource *unstructured.Unstructured) bool {
	for _, operand := range legacyWorkloadControllerOperands {
		for _, field := range legacyWorkloadControllerFields {
			_, legacyFound, legacyErr := unstructured.NestedFieldNoCopy(
				resource.Object, "spec", operand, field,
			)
			_, currentFound, currentErr := unstructured.NestedFieldNoCopy(
				resource.Object, "spec", anyworkloadControllerOperandName, field,
			)
			if legacyErr == nil && currentErr == nil && legacyFound && !currentFound {
				return true
			}
		}
	}
	return false
}

func (u *legacyWorkloadConfigurationUpgrade) updateRunaiConfig(
	resource *unstructured.Unstructured,
) (*unstructured.Unstructured, error) {
	for _, operand := range legacyWorkloadControllerOperands {
		for _, field := range legacyWorkloadControllerFields {
			value, legacyFound, err := unstructured.NestedFieldNoCopy(
				resource.Object, "spec", operand, field,
			)
			if err != nil {
				return nil, err
			}
			_, currentFound, err := unstructured.NestedFieldNoCopy(
				resource.Object, "spec", anyworkloadControllerOperandName, field,
			)
			if err != nil {
				return nil, err
			}
			if !legacyFound || currentFound {
				continue
			}
			if err := unstructured.SetNestedField(
				resource.Object, runtime.DeepCopyJSONValue(value), "spec", anyworkloadControllerOperandName, field,
			); err != nil {
				return nil, err
			}
		}
	}
	return resource, nil
}

func (u *legacyWorkloadConfigurationUpgrade) upgradeHandlerIdentifier() string {
	return "anyworkload-controller-legacy-workload-configuration"
}
