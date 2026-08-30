package parsers

import (
	"fmt"
	"github.com/run-ai/runai/runai-application-packages/convutils"

	auth_client "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	k8sobjectstracker "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/util"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var originAssets = k8sobjectstracker.ORIGIN_ASSETS
var originExisting = k8sobjectstracker.ORIGIN_EXISTING

type ScopeTypeAndId struct {
	Type auth_client.ScopeType
	Id   string
}

type ParsedSecret struct {
	Scope           ScopeTypeAndId
	CreationRequest k8sobjectstracker.SecretCreationRequest
}

func ParseSecretToServerScheme(secret *corev1.Secret) (*ParsedSecret, error) {
	var secretTypeLabel string

	var origin k8sobjectstracker.NullableOrigin

	if util.IsAssetRelatedResource(&secret.ObjectMeta) {
		origin.Set(&originAssets)
		secretTypeLabel = config.Config().AssetsSecretToSyncLabel
	} else if util.IsUserExposedCredentialsForRunai(secret) {
		origin.Set(&originExisting)
		secretTypeLabel = config.Config().ExistingSecretToSyncLabel
	} else {
		return nil, fmt.Errorf("could not determine the origin of the seret")
	}

	secretType, found := secret.Labels[secretTypeLabel]
	if !found {
		return nil, fmt.Errorf("label %s is not provided", secretTypeLabel)
	}

	var private k8sobjectstracker.NullableBool
	privateLabel, found := secret.Labels[commonconstants.UserCredentialsLabel]
	if found && privateLabel == util.TrueAsString {
		private.Set(convutils.Ref(true))
	}

	result := ParsedSecret{
		CreationRequest: k8sobjectstracker.SecretCreationRequest{
			Type:    k8sobjectstracker.SecretType(secretType),
			Origin:  origin,
			Issues:  util.ExtractIssuesFromObject(secret.GetAnnotations()),
			Private: private,
		},
	}

	err := parseScope(&result.Scope, &secret.ObjectMeta)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func parseScope(scope *ScopeTypeAndId, objectMeta *metav1.ObjectMeta) error {

	//
	//   for existing secrets:
	//   expecting to have one of the two labels: run.ai/cluster-wide OR run.ai/department
	//   if none of them is provided, need to return error
	//   for assets originated secret, it is possible that the scope will be run.ai/tenant-wide,
	//   and they are sent as cluster scope, cause from control plane point of view they are
	//   similar to cluster scope.
	//
	tenantScopeValue, tenantScopeGiven := objectMeta.Labels[commonconstants.TenantScopeLabel]
	clusterScopeValue, clusterScopeGiven := objectMeta.Labels[commonconstants.ClusterWideResourceLabel]
	departmentId, departmentIdGiven := objectMeta.Labels[commonconstants.DepartmentScopeLabel]
	// if the secret is a user credential, there are no scope labels, we need to set the scope to cluster
	userCredential, found := objectMeta.Labels[commonconstants.UserCredentialsLabel]
	if found && userCredential == util.TrueAsString {
		*scope = ScopeTypeAndId{
			Type: auth_client.SCOPETYPE_CLUSTER,
			Id:   config.Config().ClusterUuid,
		}
		return nil
	}

	tenantScope := tenantScopeGiven && (tenantScopeValue == util.TrueAsString)
	clusterScope := clusterScopeGiven && (clusterScopeValue == util.TrueAsString)

	if departmentIdGiven {
		if tenantScope {
			return fmt.Errorf("labels %s (value: %s) and %s (value: %s) are mutually excluded",
				commonconstants.TenantScopeLabel, tenantScopeValue, commonconstants.DepartmentScopeLabel, departmentId)
		}

		if clusterScope {
			return fmt.Errorf("labels %s (value: %s) and %s (value: %s) are mutually excluded",
				commonconstants.ClusterWideResourceLabel, clusterScopeValue, commonconstants.DepartmentScopeLabel, departmentId)
		}
	}

	if tenantScope || clusterScope {
		*scope = ScopeTypeAndId{
			Type: auth_client.SCOPETYPE_CLUSTER,
			Id:   config.Config().ClusterUuid,
		}
	} else if departmentIdGiven {
		*scope = ScopeTypeAndId{
			Type: auth_client.SCOPETYPE_DEPARTMENT,
			Id:   departmentId,
		}
	} else {
		*scope = ScopeTypeAndId{
			Type: auth_client.SCOPETYPE_PROJECT,
			Id:   objectMeta.Namespace,
		}
	}

	return nil
}
