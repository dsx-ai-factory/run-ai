package test

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	schedulingv1 "k8s.io/api/scheduling/v1"

	assets_openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/handlers"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	policy_openapi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/common"
	configmaps "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/config-maps"
	dpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	rbacv1 "k8s.io/api/rbac/v1"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func generateBackendClusterPoliciesInner(name string, scope policy_openapi.ScopeType, updateCounter, scopeId int, toDelete bool,
	effective *policy_openapi.SupersetPolicy, distributed *policy_openapi.DistributedPolicyDefaultsAndRulesV2,
) policy_openapi.PolicySyncResponsePoliciesInner {
	backendClusterPoliciesInner := policy_openapi.PolicySyncResponsePoliciesInner{
		Meta: policy_openapi.PolicyCreationFields{
			Name:  name,
			Scope: scope,
		},
		UpdateCounter: intTo64Pointer(updateCounter),
	}

	if toDelete {
		backendClusterPoliciesInner.IsDeleted = NullableBool(true)
	}

	backendClusterPoliciesInner.Effective = *policy_openapi.NewNullableSupersetPolicy(effective)
	backendClusterPoliciesInner.Distributed = *policy_openapi.NewNullableDistributedPolicyDefaultsAndRulesV2(distributed)

	switch scope {
	case policy_openapi.SCOPETYPE_DEPARTMENT:
		backendClusterPoliciesInner.Meta.DepartmentId = NullableString(strconv.Itoa(scopeId))
	case policy_openapi.SCOPETYPE_PROJECT:
		backendClusterPoliciesInner.Meta.ProjectId = NullableInt32(scopeId)
	}

	return backendClusterPoliciesInner
}

func NullableInt32(number int) policy_openapi.NullableInt32 {
	numberToSet := int32(number)
	NullableInt32 := policy_openapi.NullableInt32{}
	NullableInt32.Set(&numberToSet)
	return NullableInt32
}

func NullableString(string string) policy_openapi.NullableString {
	stringToSet := string
	NullableString := policy_openapi.NullableString{}
	NullableString.Set(&stringToSet)
	return NullableString
}

func NullableBool(boolean bool) policy_openapi.NullableBool {
	booleanToSet := boolean
	NullableBool := policy_openapi.NullableBool{}
	NullableBool.Set(&booleanToSet)
	return NullableBool
}

func intTo64Pointer(num int) *int64 {
	res := int64(num)
	return &res
}

func _getNamespaceFromClient(namespaceName string, k8sClient client.Client) *corev1.Namespace { //nolint:unused
	namespaceKey := types.NamespacedName{Name: namespaceName}
	namespaceObj := &corev1.Namespace{}
	err := k8sClient.Get(context.Background(), namespaceKey, namespaceObj)
	if err != nil {
		return nil
	}
	return namespaceObj
}

func getLegacyProjectFromClient(projectName string, k8sClient client.Client) *runaiv2.Project {
	projectKey := types.NamespacedName{Name: projectName}
	projectObj := &runaiv2.Project{}
	err := k8sClient.Get(context.Background(), projectKey, projectObj)
	if err != nil {
		return nil
	}
	return projectObj
}

func getBasicProjectObject(projectName, projectId, departmentName, departmentId string, statusNamespace string) *runaiv2.Project {
	basicProject := &runaiv2.Project{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Project",
			APIVersion: "run.ai/v2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			Labels: map[string]string{
				runaiv2.ProjectIdLabel:               projectId,
				schedulingv2alpha2.DepartmentIdLabel: departmentId,
			},
		},
		Spec: runaiv2.ProjectSpec{
			Department: departmentName,
			Queues: []runaiv2.QueueProjectSpec{
				{
					Name:     projectName,
					Priority: ptr.To(int32(210)),
				},
			},
		},
		Status: runaiv2.ProjectStatus{},
	}
	if statusNamespace != "" {
		basicProject.Status.Namespace = statusNamespace
	}
	return basicProject
}

func getLegacyDepartmentFromClient(departmentName string, k8sClient client.Client) *schedulingv2alpha2.Department {
	departmentKey := types.NamespacedName{Name: departmentName}
	departmentObj := &schedulingv2alpha2.Department{}
	err := k8sClient.Get(context.Background(), departmentKey, departmentObj)
	if err != nil {
		return nil
	}
	return departmentObj
}

func getProjectFromClient(projectName string, k8sClient client.Client) *kaiv1alpha1.Project {
	project := &kaiv1alpha1.Project{}
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: projectName}, project); err != nil {
		return nil
	}
	return project
}

func getDepartmentFromClient(departmentName string, k8sClient client.Client) *kaiv1alpha1.Department {
	department := &kaiv1alpha1.Department{}
	if err := k8sClient.Get(context.Background(), types.NamespacedName{Name: departmentName}, department); err != nil {
		return nil
	}
	return department
}

func getBasicDepartmentObject(departmentName, departmentId string) *schedulingv2alpha2.Department {
	return &schedulingv2alpha2.Department{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Department",
			APIVersion: "scheduling.run.ai/v2alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: departmentName,
			Labels: map[string]string{
				schedulingv2alpha2.DepartmentIdLabel: departmentId,
			},
		},
		Spec: schedulingv2alpha2.DepartmentSpec{
			Queues: []schedulingv2alpha2.QueueDepartmentSpec{{
				Name:     departmentName,
				Nodepool: "default",
				Priority: ptr.To(int32(50)),
			}},
		},
	}
}

func getDatavolumeFromClient(datavolumeName string, k8sClient client.Client) *v1alpha1.DataVolume {
	datavolumeKey := types.NamespacedName{Name: datavolumeName}
	datavolumeObj := &v1alpha1.DataVolume{}
	err := k8sClient.Get(context.Background(), datavolumeKey, datavolumeObj)
	if err != nil {
		return nil
	}
	return datavolumeObj
}

func getBasicNodePoolObject(nodePoolName string) *kaiv1alpha1.NodePool {
	return &kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodePoolName,
		},
	}
}

func getNodePoolFromClient(nodePoolName string, k8sClient client.Client) *kaiv1alpha1.NodePool {
	nodePoolKey := types.NamespacedName{Name: nodePoolName}
	nodePoolObj := &kaiv1alpha1.NodePool{}
	err := k8sClient.Get(context.Background(), nodePoolKey, nodePoolObj)
	if err != nil {
		return nil
	}
	return nodePoolObj
}

func getBasicDatavolumeObject(datavolumeName string) *v1alpha1.DataVolume {
	return &v1alpha1.DataVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DataVolume",
			APIVersion: "run.ai/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: datavolumeName,
		},
	}
}

func getAgentConfigFromClient(k8sClient client.Client) *corev1.ConfigMap {
	p, err := configmaps.GetAgentConfigData(k8sClient)
	if err != nil {
		return nil
	}
	return p
}

func getTrainingPolicyFromClient(name, namespace string, k8sClient client.Client) *tpv1.TrainingPolicy {
	p := getPolicyFromClient(name, namespace, k8sClient, common.PolicyTypeTraining)
	if p == nil {
		return nil
	}
	return p.(*tpv1.TrainingPolicy)
}

func getInteractivePolicyFromClient(name, namespace string, k8sClient client.Client) *ipv1.InteractivePolicy {
	p := getPolicyFromClient(name, namespace, k8sClient, common.PolicyTypeInteractive)
	if p == nil {
		return nil
	}
	return p.(*ipv1.InteractivePolicy)
}

func getDistributedPolicyFromClient(name, namespace string, k8sClient client.Client) *dpv1.DistributedPolicy {
	p := getPolicyFromClient(name, namespace, k8sClient, common.PolicyTypeDistributed)
	if p == nil {
		return nil
	}
	return p.(*dpv1.DistributedPolicy)
}

func getInferencePolicyFromClient(name, namespace string, k8sClient client.Client) *infpv1.InferencePolicy {
	p := getPolicyFromClient(name, namespace, k8sClient, common.PolicyTypeInference)
	if p == nil {
		return nil
	}
	return p.(*infpv1.InferencePolicy)
}

func getPolicyFromClient(name, namespace string, k8sClient client.Client, policyType string) interface{} {
	policyKey := types.NamespacedName{Name: name, Namespace: namespace}

	switch policyType {
	case common.PolicyTypeTraining:
		policyObject := &tpv1.TrainingPolicy{}
		err := k8sClient.Get(context.Background(), policyKey, policyObject)
		if err != nil {
			return nil
		}
		return policyObject
	case common.PolicyTypeInteractive:
		policyObject := &ipv1.InteractivePolicy{}
		err := k8sClient.Get(context.Background(), policyKey, policyObject)
		if err != nil {
			return nil
		}
		return policyObject
	case common.PolicyTypeDistributed:
		policyObject := &dpv1.DistributedPolicy{}
		err := k8sClient.Get(context.Background(), policyKey, policyObject)
		if err != nil {
			return nil
		}
		return policyObject
	case common.PolicyTypeInference:
		policyObject := &infpv1.InferencePolicy{}
		err := k8sClient.Get(context.Background(), policyKey, policyObject)
		if err != nil {
			return nil
		}
		return policyObject
	default:
		return nil
	}
}

func getPvcFromClient(name, namespace string, k8sClient client.Client) *corev1.PersistentVolumeClaim {
	key := types.NamespacedName{Name: name, Namespace: namespace}

	pvcResource := corev1.PersistentVolumeClaim{}
	err := k8sClient.Get(context.Background(), key, &pvcResource)
	if err != nil {
		return nil
	}
	return &pvcResource
}

func getSecretFromClient(name, namespace string, k8sClient client.Client) *corev1.Secret {
	key := types.NamespacedName{Name: name, Namespace: namespace}

	secretResource := corev1.Secret{}
	err := k8sClient.Get(context.Background(), key, &secretResource)
	if err != nil {
		return nil
	}
	return &secretResource
}

func getTrainingWorkloadFromClient(name, namespace string, k8sClient client.Client) *twv2.TrainingWorkload {
	p := getWorkloadFromClient(name, namespace, k8sClient, common.WorkloadTypeTraining)
	if p == nil {
		return nil
	}
	return p.(*twv2.TrainingWorkload)
}

func getInteractiveWorkloadFromClient(name, namespace string, k8sClient client.Client) *iwv2.InteractiveWorkload {
	p := getWorkloadFromClient(name, namespace, k8sClient, common.WorkloadTypeInteractive)
	if p == nil {
		return nil
	}
	return p.(*iwv2.InteractiveWorkload)
}

func getDistributedWorkloadFromClient(name, namespace string, k8sClient client.Client) *dwv2.DistributedWorkload {
	p := getWorkloadFromClient(name, namespace, k8sClient, common.WorkloadTypeDistributed)
	if p == nil {
		return nil
	}
	return p.(*dwv2.DistributedWorkload)
}

func getInferenceWorkloadFromClient(name, namespace string, k8sClient client.Client) *infwv2.InferenceWorkload {
	p := getWorkloadFromClient(name, namespace, k8sClient, common.WorkloadTypeInference)
	if p == nil {
		return nil
	}
	return p.(*infwv2.InferenceWorkload)
}

func getWorkloadFromClient(name, namespace string, k8sClient client.Client, workloadType string) interface{} {
	workloadKey := types.NamespacedName{Name: name, Namespace: namespace}

	switch workloadType {
	case common.WorkloadTypeTraining:
		workload := &twv2.TrainingWorkload{}
		err := k8sClient.Get(context.Background(), workloadKey, workload)
		if err != nil {
			return nil
		}
		return workload
	case common.WorkloadTypeInteractive:
		workload := &iwv2.InteractiveWorkload{}
		err := k8sClient.Get(context.Background(), workloadKey, workload)
		if err != nil {
			return nil
		}
		return workload
	case common.WorkloadTypeDistributed:
		workload := &dwv2.DistributedWorkload{}
		err := k8sClient.Get(context.Background(), workloadKey, workload)
		if err != nil {
			return nil
		}
		return workload
	case common.WorkloadTypeInference:
		workload := &infwv2.InferenceWorkload{}
		err := k8sClient.Get(context.Background(), workloadKey, workload)
		if err != nil {
			return nil
		}
		return workload
	default:
		return nil
	}
}

func getWorkloadV2FromClient(name, namespace string, gvk schema.GroupVersionKind, k8sClient client.Client) *unstructured.Unstructured {
	clientObjectKey := client.ObjectKey{Name: name, Namespace: namespace}

	u := unstructured.Unstructured{}

	u.SetGroupVersionKind(gvk)
	err := k8sClient.Get(context.Background(), clientObjectKey, &u)
	if err == nil {
		return &u
	}
	return nil
}

func getAccessRuleFromClient(accessRuleName string, k8sClient client.Client) *v1alpha1.AccessRule {
	accessRuleKey := types.NamespacedName{Name: accessRuleName}
	accessRuleObj := &v1alpha1.AccessRule{}
	err := k8sClient.Get(context.Background(), accessRuleKey, accessRuleObj)
	if err != nil {
		return nil
	}
	return accessRuleObj
}

func getBasicAccessRuleObject(accessRuleName, accessRuleId string) *v1alpha1.AccessRule {
	return &v1alpha1.AccessRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AccessRule",
			APIVersion: "run.ai/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: accessRuleName,
			Labels: map[string]string{
				v1alpha1.AccessRuleIdLabel: accessRuleId,
			},
		},
	}
}

func getBasicTrainingPolicyObject(name, namespace string, policyCounter int, scope policy_openapi.ScopeType, labelId string, trainingSpec tpv1.TrainingSpec) *tpv1.TrainingPolicy {
	tp := &tpv1.TrainingPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "TrainingPolicy",
			APIVersion: "run.ai/v2alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				constants.PolicyLabel: "true",
			},
			Namespace: namespace,
			Annotations: map[string]string{
				constants.AnnotationsBackendPolicyCounter: fmt.Sprint(policyCounter),
			},
		},
		Spec: trainingSpec,
	}

	switch scope {
	case policy_openapi.SCOPETYPE_TENANT:
		tp.Labels[constants.TenantScopeLabel] = "true"
	case policy_openapi.SCOPETYPE_PROJECT:
		tp.Labels[constants.ProjectScopeLabel] = labelId
	case policy_openapi.SCOPETYPE_DEPARTMENT:
		tp.Labels[constants.DepartmentScopeLabel] = labelId
	}

	return tp
}

func getBasicDistributedPolicyObject(name, namespace string, policyCounter int, scope policy_openapi.ScopeType, labelId string,
	workerSpec dpv1.DistributedSpec, masterSpec *dpv1.DistributedMasterSpec,
) *dpv1.DistributedPolicy {
	dp := &dpv1.DistributedPolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DistributedPolicy",
			APIVersion: "run.ai/v2alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				constants.PolicyLabel: "true",
			},
			Namespace: namespace,
			Annotations: map[string]string{
				constants.AnnotationsBackendPolicyCounter: fmt.Sprint(policyCounter),
			},
		},
		Spec:       workerSpec,
		MasterSpec: masterSpec,
	}

	switch scope {
	case policy_openapi.SCOPETYPE_TENANT:
		dp.Labels[constants.TenantScopeLabel] = "true"
	case policy_openapi.SCOPETYPE_PROJECT:
		dp.Labels[constants.ProjectScopeLabel] = labelId
	case policy_openapi.SCOPETYPE_DEPARTMENT:
		dp.Labels[constants.DepartmentScopeLabel] = labelId
	}

	return dp
}

func getBasicInteractivePolicyObject(name, namespace string, policyCounter int, scope policy_openapi.ScopeType, scopeId int) *ipv1.InteractivePolicy {
	tp := &ipv1.InteractivePolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "InteractivePolicy",
			APIVersion: "run.ai/v2alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				constants.PolicyLabel: "true",
			},
			Namespace: namespace,
			Annotations: map[string]string{
				constants.AnnotationsBackendPolicyCounter: strconv.Itoa(policyCounter),
			},
		},
	}

	switch scope {
	case policy_openapi.SCOPETYPE_TENANT:
		tp.Labels[constants.TenantScopeLabel] = "true"
	case policy_openapi.SCOPETYPE_PROJECT:
		tp.Labels[constants.ProjectScopeLabel] = strconv.Itoa(scopeId)
	case policy_openapi.SCOPETYPE_DEPARTMENT:
		tp.Labels[constants.DepartmentScopeLabel] = strconv.Itoa(scopeId)
	}

	return tp
}

func getBasicSecretObject(name, namespace string, updateCounter *int64, scope assets_openapi.ScopeType,
	scopeId string, secretType string, data map[string][]byte,
) *corev1.Secret {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Type: corev1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				constants.RunaiAssetLabel: secretType,
			},
			Namespace: namespace,
		},
	}

	if data != nil {
		secret.Data = data
	}

	if updateCounter != nil {
		secret.Annotations = map[string]string{
			handlers.UpdateCountAnnotation: fmt.Sprintf("%d", *updateCounter),
		}
	}

	switch scope {
	case assets_openapi.SCOPETYPE_TENANT:
		secret.Labels[constants.TenantScopeLabel] = "true"
	case assets_openapi.SCOPETYPE_PROJECT:
		secret.Labels[constants.ProjectScopeLabel] = scopeId
	case assets_openapi.SCOPETYPE_DEPARTMENT:
		secret.Labels[constants.DepartmentScopeLabel] = scopeId
	}

	return secret
}

func getBasicPvcObject(name, namespace string, updateCounter *int64, scope assets_openapi.ScopeType, scopeId string, spec *corev1.PersistentVolumeClaimSpec) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				constants.RunaiAssetLabel: "pvc",
			},
			Namespace: namespace,
		},
	}

	if spec != nil {
		pvc.Spec = *spec
	}

	if updateCounter != nil {
		pvc.Annotations = map[string]string{
			handlers.UpdateCountAnnotation: fmt.Sprintf("%d", *updateCounter),
		}
	}

	switch scope {
	case assets_openapi.SCOPETYPE_TENANT:
		pvc.Labels[constants.TenantScopeLabel] = "true"
	case assets_openapi.SCOPETYPE_PROJECT:
		pvc.Labels[constants.ProjectScopeLabel] = scopeId
	case assets_openapi.SCOPETYPE_DEPARTMENT:
		pvc.Labels[constants.DepartmentScopeLabel] = scopeId
	}

	return pvc
}

func getBasicPvcObjectWithAnnotations(name, namespace string, updateCounter *int64, scope assets_openapi.ScopeType, scopeId string,
	annotations map[string]string, spec *corev1.PersistentVolumeClaimSpec,
) *corev1.PersistentVolumeClaim {
	pvc := getBasicPvcObject(name, namespace, updateCounter, scope, scopeId, spec)
	pvc.Annotations = annotations
	return pvc
}

func getBasicInferencePolicyObject(name, namespace string, policyCounter int, scope policy_openapi.ScopeType, labelId string, inferenceSpec infpv1.InferenceSpec) *infpv1.InferencePolicy {
	infp := &infpv1.InferencePolicy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "InferencePolicy",
			APIVersion: "run.ai/v2alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				constants.PolicyLabel: "true",
			},
			Namespace: namespace,
			Annotations: map[string]string{
				constants.AnnotationsBackendPolicyCounter: fmt.Sprint(policyCounter),
			},
		},
		Spec: inferenceSpec,
	}

	switch scope {
	case policy_openapi.SCOPETYPE_TENANT:
		infp.Labels[constants.TenantScopeLabel] = "true"
	case policy_openapi.SCOPETYPE_PROJECT:
		infp.Labels[constants.ProjectScopeLabel] = labelId
	case policy_openapi.SCOPETYPE_DEPARTMENT:
		infp.Labels[constants.DepartmentScopeLabel] = labelId
	}

	return infp
}

func getBasicWorkloadV2(name, namespace string, updateCounter int64, spec map[string]interface{}) client.Object {
	u := unstructured.Unstructured{
		Object: spec,
	}

	u.SetName(name)
	u.SetNamespace(namespace)

	annotations := u.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[handlers.UpdateCountAnnotation] = fmt.Sprintf("%d", updateCounter)

	u.SetAnnotations(annotations)

	var obj client.Object = &u
	obj = wrapWithWorkloadRunner(u)
	obj.SetName(name)
	obj.SetNamespace(namespace)

	return obj
}

func wrapWithWorkloadRunner(manifest unstructured.Unstructured) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		TypeMeta: metav1.TypeMeta{
			Kind:       handlers.KindWorkloadRunner,
			APIVersion: handlers.ApiVersionWorkloadRunner,
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template: runtime.RawExtension{Object: &manifest},
		},
	}
}
func getPriorityClass(name string, value int32, id string, source string) *schedulingv1.PriorityClass {
	labels := map[string]string{}
	if id != "" {
		labels[constants.PriorityClassIdLabel] = id
	}
	if source != "" {
		labels[constants.PriorityClassSourceLabel] = source
	}
	pc := schedulingv1.PriorityClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Value: value,
	}
	return &pc
}

func _getRoleBindingFromClient(name, namespace string, k8sClient client.Client) *rbacv1.RoleBinding { //nolint:unused
	rbKey := types.NamespacedName{Name: name, Namespace: namespace}
	rbObj := &rbacv1.RoleBinding{}
	err := k8sClient.Get(context.Background(), rbKey, rbObj)
	if err != nil {
		return nil
	}
	return rbObj
}

func _getClusterRoleBindingFromClient(name string, k8sClient client.Client) *rbacv1.ClusterRoleBinding { //nolint:unused
	crbKey := types.NamespacedName{Name: name}
	crbObj := &rbacv1.ClusterRoleBinding{}
	err := k8sClient.Get(context.Background(), crbKey, crbObj)
	if err != nil {
		return nil
	}
	return crbObj
}

func _listAccessRules(k8sClient client.Client) *v1alpha1.AccessRuleList { //nolint:unused
	accessRules := &v1alpha1.AccessRuleList{}
	err := k8sClient.List(context.Background(), accessRules)
	if err != nil {
		return nil
	}
	return accessRules
}

func _createNamespaceIfNeeded(namespace, project string, k8sClient client.Client) { //nolint:unused
	namespaceObj := _getNamespaceFromClient(namespace, k8sClient)
	if namespaceObj != nil {
		return
	}

	namespaceObj = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        namespace,
			Annotations: map[string]string{},
		},
	}
	if namespace != runaiv1.DefaultNamespace && project != "" {
		namespaceObj.Labels = map[string]string{
			"runai/queue": project,
		}
	}
	utils.ExpectCreateResource(k8sClient, namespaceObj)
}

func _createDepartment(departmentName, departmentId string, k8sClient client.Client) { //nolint:unused
	utils.ExpectCreateResource(k8sClient, getBasicDepartmentObject(departmentName, departmentId))
}

func _createAccessRule(accessRuleName, accessRuleId string, accessRuleSpec *v1alpha1.AccessRuleSpec, k8sClient client.Client) *v1alpha1.AccessRule { //nolint:unused
	accessRuleObj := &v1alpha1.AccessRule{
		ObjectMeta: metav1.ObjectMeta{
			Name: accessRuleName,
			Labels: map[string]string{
				v1alpha1.AccessRuleIdLabel: accessRuleId,
			},
		},
		Spec: *accessRuleSpec,
	}

	utils.ExpectCreateResource(k8sClient, accessRuleObj)
	return accessRuleObj
}

func _deleteNamespace(namespace string, k8sClient client.Client) { //nolint:unused
	By(fmt.Sprintf("Deleting Namespace <%s>", namespace))

	namespaceObj := _getNamespaceFromClient(namespace, k8sClient)
	if namespaceObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, namespaceObj)
	}
}

func _deleteProject(project string, k8sClient client.Client) { //nolint:unused
	By(fmt.Sprintf("Deleting project <%s>", project))

	projectObj := getLegacyProjectFromClient(project, k8sClient)
	if projectObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, projectObj)
	}
}

func _deleteDepartment(department string, k8sClient client.Client) { //nolint:unused
	By(fmt.Sprintf("Deleting department <%s>", department))

	departmentObj := getLegacyDepartmentFromClient(department, k8sClient)
	if departmentObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, departmentObj)
	}
}

func _deleteAccessRule(accessRule string, k8sClient client.Client) { //nolint:unused
	By(fmt.Sprintf("Deleting accessRule <%s>", accessRule))

	accessRuleObj := getAccessRuleFromClient(accessRule, k8sClient)
	if accessRuleObj != nil {
		utils.DeleteAndPollUntilDeleted(k8sClient, accessRuleObj)
	}
}

func _expectUpdateResourceStatus(k8sClient client.Client, resource client.Object) { //nolint:unused
	By(fmt.Sprintf("updating status of resource %s/%s...", resource.GetNamespace(), resource.GetName()))

	ExpectWithOffset(1, k8sClient.Status().Update(context.Background(), resource)).To(Succeed())

	By(fmt.Sprintf("succesfully updated status of resource %s/%s", resource.GetNamespace(), resource.GetName()))
}

var sampleDeployment = `{
		"apiVersion": "apps/v1",
		"kind": "Deployment",
		"metadata": {
			"name": "{name}"
		},
		"spec": {
			"replicas": 1,
			"selector": {
				"matchLabels": {
					"app": "{name}"
				}
			},
			"template": {
				"metadata": {
					"labels": {
						"app": "{name}"
					}
				},
				"spec": {
					"containers": [
						{
							"name": "container-{name}",
							"image": "nginx:latest"
						}
					]
				}
			}
		}
	}`

func generateSampleDeployment(name string) map[string]interface{} {
	withName := strings.ReplaceAll(sampleDeployment, "{name}", name)
	asJson, _ := gojsonq.LoadJsonBytes([]byte(withName))
	return asJson
}
