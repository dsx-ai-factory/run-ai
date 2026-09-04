/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	k8sobjectstracker "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"

	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TrueAsString      = "true"
	WorkloadNameLabel = "workloadName"
)

func GetNamespace(namespace string, cachedReader client.Reader) (*corev1.Namespace, error) {
	namespaceObj := &corev1.Namespace{}
	err := cachedReader.Get(context.TODO(), types.NamespacedName{Name: namespace}, namespaceObj)
	if err != nil {
		zap.S().Errorf("Could not get namespace %s, error: %s", namespace, err.Error())
		return nil, err
	}
	return namespaceObj, nil
}

func IsRunaiManagedNamespace(namespace string, cachedReader client.Reader) bool {
	// validate the project itself exists - includes getting the namespace
	// and validation of queue label on the namespace.
	_, err := GetProjectFromNamespaceName(namespace, cachedReader)
	return err == nil
}

// IsUserExposedCredentialsForRunai - in order for user to expose credentials to runai,
// he needs to mark them with label run.ai/resource, with value indicating the type of credentials:
// password, access-key or docker-registry
func IsUserExposedCredentialsForRunai(secret *corev1.Secret) bool {
	if secret.Labels == nil {
		return false
	}

	value, found := secret.Labels[config.Config().ExistingSecretToSyncLabel]
	return found && value != ""
}

func IsUserExposedPvcForRunai(pvc *corev1.PersistentVolumeClaim) bool {
	if pvc.Labels == nil {
		// note, this is not a typo. unlike secret that requires the run.ai/resource label in order
		// for it to be exposed, PVC that does not have any label is candidate for exposure to runai
		return true
	}

	if IsAssetRelatedResource(&pvc.ObjectMeta) {
		return false // asset sync owned resource
	}

	if IsWorkloadControllerResource(&pvc.ObjectMeta) {
		return false // workload controller created resource
	}

	if IsDataVolomeRelatedResource(&pvc.ObjectMeta) {
		return false // data volume created resource
	}

	return true
}

// IsUserExposedConfigMapForRunai - in order for user to expose config map to runai,
// he needs to mark them with label run.ai/resource,
func IsUserExposedConfigMapForRunai(configMap *corev1.ConfigMap) bool {
	if configMap.Labels == nil {
		return false
	}

	_, found := configMap.Labels[config.Config().ConfigMapToSyncLabel]
	return found
}

func IsAssetRelatedResource(om *metav1.ObjectMeta) bool {
	if om == nil || om.Labels == nil {
		return false
	}

	assetsLabelValue, assetsLabelFound := om.Labels[commonconstants.RunaiAssetLabel]
	return assetsLabelFound && assetsLabelValue != ""
}

func IsWorkloadControllerResource(om *metav1.ObjectMeta) bool {
	if om == nil || om.Labels == nil {
		return false
	}

	workloadNameLabel, workloadNameFound := om.Labels[WorkloadNameLabel]
	return workloadNameFound && workloadNameLabel != ""
}

func IsDataVolomeRelatedResource(om *metav1.ObjectMeta) bool {
	if om == nil || om.Labels == nil {
		return false
	}

	datavolumeId, found := om.Labels[v1alpha1.DataVolumeIdLabel]
	return found && datavolumeId != ""
}

// IsClusterOrDeptScopedResource validates if a secret/pvc is labeled with cluster/department scope
func IsClusterOrDeptScopedResource(objectMeta *metav1.ObjectMeta) bool {
	if objectMeta.Labels == nil {
		return false
	}

	tenantScope, tenantLabelFound := objectMeta.Labels[commonconstants.TenantScopeLabel]
	clusterScope, clusterLabelFound := objectMeta.Labels[commonconstants.ClusterWideResourceLabel]
	departmentId, deptLabelFound := objectMeta.Labels[commonconstants.DepartmentScopeLabel]

	return (clusterLabelFound && clusterScope == TrueAsString) ||
		(tenantLabelFound && tenantScope == TrueAsString) ||
		(deptLabelFound && departmentId != "")
}

func GetProjectNameFromNamespace(namespace *corev1.Namespace) string {
	project, found := namespace.Labels[constants.RunaiQueueLabel]
	if !found {
		return ""
	}
	return project
}

// ProjectData is the org-unit projection the shared sync path needs from a project, sourced
// from either the kai.resources project.
// Namespace is taken from status so that it will be the actual namespace and not the requested one.
// Carrying a neutral shape keeps the workload/pod/revision sync path CRD-agnostic.
type ProjectData struct {
	Name             string
	Namespace        string
	ProjectId        string
	DepartmentId     string
	DepartmentName   string
	DefaultNodePools []string
}

func projectDataFromKai(project *kaiorg.Project) ProjectData {
	return ProjectData{
		Name:             project.Name,
		Namespace:        project.Status.Namespace,
		ProjectId:        project.Labels[v2.ProjectIdLabel],
		DepartmentId:     project.Labels[schedulingv2alpha2.DepartmentIdLabel],
		DepartmentName:   project.Spec.Parent,
		DefaultNodePools: project.Spec.DefaultNodePools,
	}
}

func projectDataFromLegacy(project *v2.Project) ProjectData {
	return ProjectData{
		Name:             project.Name,
		Namespace:        project.Status.Namespace,
		ProjectId:        project.Labels[v2.ProjectIdLabel],
		DepartmentId:     project.Labels[schedulingv2alpha2.DepartmentIdLabel],
		DepartmentName:   project.Spec.Department,
		DefaultNodePools: project.Spec.DefaultNodePools,
	}
}

func GetProjectId(project *ProjectData) (string, error) {
	if project.ProjectId == "" {
		return "", fmt.Errorf("project <%s> doesn't include project id label", project.Name)
	}
	return project.ProjectId, nil
}

func GetDepartmentId(project *ProjectData) (string, error) {
	if project.DepartmentId == "" {
		return "", fmt.Errorf("project <%s> doesn't include department id label", project.Name)
	}
	return project.DepartmentId, nil
}

func GetProjectIdFromNamespaceName(namespaceName string, cachedReader client.Reader) (string, error) {
	project, err := GetProjectFromNamespaceName(namespaceName, cachedReader)
	if err != nil {
		return "", err
	}
	return GetProjectId(project)
}

func GetAllRunAiManagedNamespaces(cachedReader client.Reader) ([]corev1.Namespace, error) {
	namespaces := &corev1.NamespaceList{}
	err := cachedReader.List(context.TODO(), namespaces, client.HasLabels{constants.RunaiQueueLabel})
	if err != nil {
		zap.S().Errorf("Failed listing namespaces managed by runai, error: %s", err.Error())
		return []corev1.Namespace{}, err
	}
	return namespaces.Items, nil
}

func GetAllRunAiManagedNamespacesMap(cachedReader client.Reader) map[string]bool {
	namespacesList, err := GetAllRunAiManagedNamespaces(cachedReader)
	if err != nil {
		return map[string]bool{}
	}

	// create a map of namespaces (the best way to search for namespace in a list)
	namespacesMap := make(map[string]bool)
	for _, namespace := range namespacesList {
		namespacesMap[namespace.Name] = true
	}
	return namespacesMap
}

// GetAllProjects lists the kai.resources org-unit projects as neutral ProjectData.
func GetAllProjects(cachedReader client.Reader) ([]ProjectData, error) {
	projects := &kaiorg.ProjectList{}
	if err := cachedReader.List(context.Background(), projects); err != nil {
		zap.S().Errorf("Failed listing projects, error: %s", err.Error())
		return nil, err
	}
	result := make([]ProjectData, 0, len(projects.Items))
	for i := range projects.Items {
		result = append(result, projectDataFromKai(&projects.Items[i]))
	}
	return result, nil
}

// GetProject fetches a single kai.resources org-unit project by name as neutral ProjectData.
func GetProject(projectName string, cachedReader client.Reader) (*ProjectData, error) {
	project := &kaiorg.Project{}
	if err := cachedReader.Get(context.Background(), types.NamespacedName{Name: projectName}, project); err != nil {
		zap.S().Errorf("Failed getting project <%s>, error: %s", projectName, err.Error())
		return nil, err
	}
	data := projectDataFromKai(project)
	return &data, nil
}

func GetProjectFromNamespaceName(namespaceName string, cachedReader client.Reader) (*ProjectData, error) {
	namespace, err := GetNamespace(namespaceName, cachedReader)
	if err != nil {
		return nil, err
	}

	projectName := GetProjectNameFromNamespace(namespace)
	if projectName == "" {
		return nil, fmt.Errorf("namespace <%s> doesn't include project label", namespaceName)
	}

	return GetProject(projectName, cachedReader)
}

func ParseLabelInt32(labels map[string]string, labelKey string) (int32, error) {
	valueStr, found := labels[labelKey]
	if !found {
		return -1, fmt.Errorf("label <%s> doesn't exist", labelKey)
	}

	value, err := strconv.ParseInt(valueStr, 10, 32)
	if err != nil {
		return -1, err
	}
	return int32(value), nil
}

// ExtractIssuesFromObject extracts list of issues from the object annotations
func ExtractIssuesFromObject(annotations map[string]string) []k8sobjectstracker.Issue {
	var issues []k8sobjectstracker.Issue
	for k, v := range annotations {
		if strings.HasPrefix(k, commonconstants.CloneAnnotation) {
			issues = append(issues, k8sobjectstracker.Issue{
				Namespace: strings.TrimPrefix(k, fmt.Sprintf("%s/", commonconstants.CloneAnnotation)),
				Issue:     v,
			})
		}
	}
	return issues
}

type NullableValue[T comparable] interface {
	Get() *T
	IsSet() bool
}

func CompareNullables[T comparable](a NullableValue[T], b NullableValue[T]) bool {
	aValue := a.Get()
	bValue := b.Get()
	if a.IsSet() != b.IsSet() {
		return aValue != bValue
	}

	if !a.IsSet() && !b.IsSet() {
		return false
	}

	if aValue == nil || bValue == nil {
		return aValue != bValue
	}

	return *aValue != *bValue
}
