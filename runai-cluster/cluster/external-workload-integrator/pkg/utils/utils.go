/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"

	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimescheme "sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	randomSuffixLen = 4
	maxNameLen      = 58 // with buffer for "-" and suffix
)

func IsRunaiManagedNamespace(namespace *corev1.Namespace) bool {
	// runai managed namespaces have the runai/queue label
	_, found := namespace.Labels[constants.RunaiQueueLabel]
	return found
}

// GetProjectFromNamespace returns the kai.resources project that owns the namespace.
func GetProjectFromNamespace(ctx context.Context, namespaceName string,
	cachedReader client.Reader) (*kaiorg.Project, error) {
	namespace, err := GetNamespace(ctx, namespaceName, cachedReader)
	if err != nil {
		return nil, err
	}

	// runai managed namespaces have the runai/queue label
	projectName, found := namespace.Labels[constants.RunaiQueueLabel]
	if !found {
		return nil, fmt.Errorf("namespace <%s> doesn't include project label", namespaceName)
	}

	return GetProject(ctx, projectName, cachedReader)
}

func GetGVKKeys[T any](gvks map[schema.GroupVersionKind]T) []schema.GroupVersionKind {
	result := []schema.GroupVersionKind{}

	for gvk := range gvks {
		result = append(result, gvk)
	}
	return result
}

func ConvertGVKMap[T any](gvks map[schema.GroupVersionKind]T) map[metav1.GroupVersionKind]T {
	result := map[metav1.GroupVersionKind]T{}

	for gvk, t := range gvks {
		result[metav1.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind}] = t
	}
	return result
}

func ConvertMapsToBoolValueMap[T, K any](gvkMap1 map[schema.GroupVersionKind]T, gvkMap2 map[schema.GroupVersionKind]K) map[schema.GroupVersionKind]bool {
	result := map[schema.GroupVersionKind]bool{}

	for gvk := range gvkMap1 {
		result[gvk] = true
	}
	for gvk := range gvkMap2 {
		result[gvk] = true
	}
	return result
}

func DerefInt32(num *int32, defaultValue int32) int32 {
	if num == nil {
		return defaultValue
	}
	return *num
}

func Int32PtrOrNil(num int32) *int32 {
	if num == 0 {
		return nil
	}
	return &num
}

func ParseAnnotationInt32(labels map[string]string, labelKey string) (int32, error) {
	return parseInt32FromStringMap(labels, labelKey, "annotation")
}

func ParseLabelInt32(labels map[string]string, labelKey string) (int32, error) {
	return parseInt32FromStringMap(labels, labelKey, "label")
}

func parseInt32FromStringMap(strMap map[string]string, labelKey string, mapName string) (int32, error) {
	valueStr, found := strMap[labelKey]
	if !found {
		return 0, fmt.Errorf("%s <%s> doesn't exist", mapName, labelKey)
	}

	value, err := strconv.ParseInt(valueStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(value), nil
}

func GenerateExternalWorkloadName(gvk schema.GroupVersionKind, originalWorkloadName string) string {
	// kind + name should be unique per namespace
	externalWorkloadName := fmt.Sprintf("ew-%s-%s", gvk.Kind, originalWorkloadName)
	externalWorkloadK8sName := convertToK8sResourceName(externalWorkloadName)

	if len(externalWorkloadK8sName) < len(externalWorkloadName) {
		// if name was truncated, add a random suffix to make it unique
		externalWorkloadK8sName = fmt.Sprintf("%s-%s", externalWorkloadK8sName, rand.String(randomSuffixLen))
	}

	return externalWorkloadK8sName
}

// AddChildIdsToStatusObjectIfNeeded - adds the childIds to the ObservedStatus.ChildIds, if not already there.
// returns whether some childIds was added or not.
func AddChildIdsToStatusObjectIfNeeded(externalWorkload *v1alpha1.ExternalWorkload,
	childIdsToInsert []v1alpha1.ChildId) bool {
	if len(childIdsToInsert) == 0 {
		return false
	}

	if externalWorkload.Status.ObservedStatus.ChildIds == nil {
		externalWorkload.Status.ObservedStatus.ChildIds = []v1alpha1.ChildId{}
	}

	someAdded := false
	for _, childIdToInsert := range childIdsToInsert {
		isChildIdInStatus, _ := IsChildIdInStatus(externalWorkload, childIdToInsert)
		if !isChildIdInStatus {
			externalWorkload.Status.ObservedStatus.ChildIds = append(externalWorkload.Status.ObservedStatus.ChildIds,
				childIdToInsert)
			someAdded = true
		}
	}

	return someAdded
}

func IsChildIdInStatus(externalWorkload *v1alpha1.ExternalWorkload, childId v1alpha1.ChildId) (bool, int) {
	for index, statusChildId := range externalWorkload.Status.ObservedStatus.ChildIds {
		if childId.UID == statusChildId.UID {
			return true, index
		}
	}
	return false, -1
}

func RemoveOldChildIdsFromStatus(externalWorkload *v1alpha1.ExternalWorkload) bool {
	someRemoved := false

	updatedList := []v1alpha1.ChildId{}
	for _, existing := range externalWorkload.Status.ObservedStatus.ChildIds {
		if isOldChildIdToRemove(existing) {
			someRemoved = true
		} else {
			updatedList = append(updatedList, existing)
		}
	}

	if someRemoved {
		externalWorkload.Status.ObservedStatus.ChildIds = updatedList
	}
	return someRemoved
}

func isOldChildIdToRemove(childId v1alpha1.ChildId) bool {
	return childId.Type == "PodGroup"
}

func DeleteChildIdsFromStatus(externalWorkload *v1alpha1.ExternalWorkload,
	childIdsToDelete []v1alpha1.ChildId) bool {
	if len(childIdsToDelete) == 0 {
		return false
	}

	if externalWorkload.Status.ObservedStatus.ChildIds == nil {
		return false
	}

	someDeleted := false
	for _, childIdToDelete := range childIdsToDelete {
		isChildIdInStatus, index := IsChildIdInStatus(externalWorkload, childIdToDelete)
		if isChildIdInStatus {
			childIds := externalWorkload.Status.ObservedStatus.ChildIds
			externalWorkload.Status.ObservedStatus.ChildIds = append(childIds[:index], childIds[index+1:]...)
			someDeleted = true
		}
	}

	return someDeleted
}

func AddUnstructuredToSchemeIfNeeded(gvk schema.GroupVersionKind, scheme *apiruntime.Scheme) {
	if _, found := scheme.AllKnownTypes()[gvk]; found {
		return
	}

	groupVersion := schema.GroupVersion{Group: gvk.Group, Version: gvk.Version}
	listGvk := schema.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind + "List",
	}
	unstructuredObject := &unstructured.Unstructured{}
	unstructuredObject.SetGroupVersionKind(gvk)
	unstructuredList := &unstructured.UnstructuredList{}
	unstructuredList.SetGroupVersionKind(listGvk)

	schemeBuilder := &runtimescheme.Builder{GroupVersion: groupVersion}
	schemeBuilder.Register(unstructuredObject, unstructuredList)
	scheme.AddKnownTypeWithName(gvk, unstructuredObject)
	scheme.AddKnownTypeWithName(listGvk, unstructuredList)
	utilruntime.Must(schemeBuilder.AddToScheme(scheme))
}

func GetGvkFromCrd(crd *apiextensionsv1.CustomResourceDefinition) (*schema.GroupVersionKind, error) {
	version := ""
	for _, crdVersion := range crd.Spec.Versions {
		if crdVersion.Storage {
			version = crdVersion.Name
			break
		}
	}
	if version == "" {
		// shouldn't happen, but of course validating
		return nil, fmt.Errorf("crd <%s/%s> has no storage version", crd.Spec.Group, crd.Spec.Names.Kind)
	}

	return &schema.GroupVersionKind{
		Group:   crd.Spec.Group,
		Version: version,
		Kind:    crd.Spec.Names.Kind,
	}, nil
}

func MergeAnnotations(annotations ...map[string]string) map[string]string {
	result := map[string]string{}

	for _, annotation := range annotations {
		for key, value := range annotation {
			result[key] = value
		}
	}

	return result
}

func convertToK8sResourceName(name string) string {
	result := name
	if len(name) > maxNameLen {
		result = name[:maxNameLen]
	}

	result = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			return r
		}
		return '-'
	}, result)

	result = strings.ToLower(result)
	return result
}

func GetNamespace(ctx context.Context, namespaceName string, cachedReader client.Reader) (*corev1.Namespace, error) {
	namespace := &corev1.Namespace{}
	err := cachedReader.Get(ctx, types.NamespacedName{Name: namespaceName}, namespace)
	if err != nil {
		log.Error().Msgf("Failed to get namespace %s, error: %s", namespaceName, err.Error())
		return nil, err
	}
	return namespace, nil
}

// GetProject returns the kai.resources project by name.
func GetProject(ctx context.Context, projectName string, cachedReader client.Reader) (*kaiorg.Project, error) {
	project := &kaiorg.Project{}
	err := cachedReader.Get(ctx, types.NamespacedName{Name: projectName}, project)
	if err != nil {
		log.Error().Msgf("Failed to get kai project %s, error: %s", projectName, err.Error())
		return nil, err
	}
	return project, nil
}

func IsSchedulerEnforcedForNamespace(namespace *corev1.Namespace) bool {
	enforce, found := namespace.Annotations[clustercommonconstants.EnforceSchedulerAnnotationName]
	if !found {
		return false
	}

	val, err := strconv.ParseBool(enforce)
	if err != nil {
		log.Error().Msgf("Could not parse <%s> annotation value <%s> for namespace <%s>, error: %s",
			clustercommonconstants.EnforceSchedulerAnnotationName, enforce, namespace.Name, err.Error())
		return false
	}
	return val
}

func IsObjectNilOrEmpty(object client.Object) bool {
	return object == nil || object.GetName() == ""
}

func GetObject(ctx context.Context, cachedClient client.Client, name, namespace string, getEmptyObject func() client.Object) (client.Object, error) {
	obj := getEmptyObject()
	err := cachedClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, obj)
	if err != nil {
		return nil, err
	}
	if obj.GetObjectKind().GroupVersionKind().Kind == "" {
		gvk, err := apiutil.GVKForObject(obj, cachedClient.Scheme())
		if err != nil {
			return nil, err
		}
		obj.GetObjectKind().SetGroupVersionKind(gvk)
	}

	return obj, nil
}

func AddLabelsPatchBytes(labelsToAdd map[string]string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"labels": labelsToAdd}})
}

func IsItemInList[T any](list []T, item T) bool {
	return indexOfFirstItemInList(list, item) > -1
}

func DeleteFirstFromList[T any](list []T, itemToRemove T) []T {
	indexToRemove := indexOfFirstItemInList(list, itemToRemove)

	if indexToRemove >= len(list) || indexToRemove == -1 {
		return list
	}
	if indexToRemove < len(list)-1 {
		copy(list[indexToRemove:], list[indexToRemove+1:]) // Shift a[i+1:] left one index.
	}

	list = list[:len(list)-1] // Truncate slice.
	return list
}

func indexOfFirstItemInList[T any](list []T, item T) int {
	for i, element := range list {
		if reflect.DeepEqual(element, item) {
			return i
		}
	}
	return -1
}

// GetExternalWorkloadOfWorkloadRunner finds the ExternalWorkload of a runner.
func GetExternalWorkloadOfWorkloadRunner(ctx context.Context, runnerUID string, namespace string,
	cachedClient client.Client,
) (*v1alpha1.ExternalWorkload, error) {
	externalWorkloads := &v1alpha1.ExternalWorkloadList{}
	err := cachedClient.List(ctx, externalWorkloads,
		client.InNamespace(namespace),
		client.MatchingLabels{clustercommonconstants.WorkloadRunnerUidLabel: runnerUID})
	if err != nil {
		return nil, err
	}

	if len(externalWorkloads.Items) == 0 {
		return nil, errors.NewNotFound(v1alpha1.Resource("externalworkload"), "")
	}

	return &externalWorkloads.Items[0], nil
}

func GetExternalWorkloadOfObject(ctx context.Context, objectId string, namespace string, cachedClient client.Client) (*v1alpha1.ExternalWorkload, error) {
	externalWorkloadByOriginalFieldSelector :=
		fields.OneTermEqualSelector(constants.ExternalWorkloadByOriginalUIDIndexerName, objectId)

	externalWorkloads := &v1alpha1.ExternalWorkloadList{}
	err := cachedClient.List(ctx, externalWorkloads,
		&client.ListOptions{FieldSelector: externalWorkloadByOriginalFieldSelector},
		client.InNamespace(namespace))
	if err != nil {
		return nil, err
	}

	if len(externalWorkloads.Items) == 0 {
		return nil, errors.NewNotFound(v1alpha1.Resource("externalworkload"), "")
	}

	return &externalWorkloads.Items[0], nil
}

// AddToResourceList adds or merges a resource quantity into a ResourceList.
func AddToResourceList(list corev1.ResourceList, name corev1.ResourceName, qty resource.Quantity) {
	if existing, found := list[name]; found {
		existing.Add(qty)
		list[name] = existing
	} else {
		list[name] = qty.DeepCopy()
	}
}
