/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/thoas/go-funk"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	secretType    ObjectType = "corev1.Secret"
	pvcType       ObjectType = "corev1.PersistentVolumeClaim"
	configmapType ObjectType = "corev1.ConfigMap"

	clusterScope    ReplicationScope = "cluster"
	departmentScope ReplicationScope = "department"

	DockerRegistryAssetKind = "docker-registry"
	PasswordAssetKind       = "password"
	AccessKeyAssetKind      = "access-key"
	GenericSecretAssetKind  = "generic"

	DefaultServiceAccountName = "default"
)

type ObjectType string
type ReplicationScope string

// GVK = group-version-kind (schema.GroupVersionKind)
// a convenient way to create a generic object with a type.

type ClusterWideGVKCommon struct {
	Client                client.Client
	gvkObjectConfigs      map[schema.GroupVersionKind][]ObjectConfig
	typeToObjectInspector map[ObjectType]ObjectInspector
}

type SecretObjectInspector struct {
	Client client.Client
}
type PvcObjectInspector bool
type ConfigMapObjectInspector bool

type ObjectInspector interface {
	DeepCopySpec(from, to client.Object)
	AreEqual(left, right client.Object) bool
}

// immutableSpecInspector is the opt-in extension for kinds whose Spec the
// apiserver treats as immutable (today: ResourceClaimTemplate). The shared
// handler checks for this interface at the Update site and routes those
// kinds through delete-and-recreate instead of calling Client.Update,
// which the apiserver would reject.
type immutableSpecInspector interface {
	IsSpecImmutable() bool
}

type ObjectConfig struct {
	ScopeWideLabel         string
	AdditionalCallback     AdditionalCallbackFn
	AdditionalObjectFilter labels.Selector
	Scope                  ReplicationScope
}

type AdditionalCallbackFn func(
	ctx context.Context, namespace string, objectsReplicatedInNamespaceNames, unreferencedObjectNames []string) error

type ClusterWideReplicatorGVKConfig struct {
	ClusterWideSecrets               bool
	ClusterWidePvcs                  bool
	ClusterWideConfigMaps            bool
	ClusterWideResourceClaimTemplates bool
}

func NewClusterWideGVKCommon(
	ctx context.Context, client client.Client, config *ClusterWideReplicatorGVKConfig) (*ClusterWideGVKCommon, error) {
	clusterWideGVKCommon := &ClusterWideGVKCommon{
		Client:                client,
		gvkObjectConfigs:      map[schema.GroupVersionKind][]ObjectConfig{},
		typeToObjectInspector: map[ObjectType]ObjectInspector{},
	}

	err := clusterWideGVKCommon.buildGVKObjectConfig(config)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("Failed building GVK action handlers for cluster-wide replicated resources")
		return nil, err
	}
	return clusterWideGVKCommon, nil
}

func (handler *ClusterWideGVKCommon) GetGVKObjectConfigs() map[schema.GroupVersionKind][]ObjectConfig {
	return handler.gvkObjectConfigs
}

func (handler *ClusterWideGVKCommon) ListObjectsWithLabel(ctx context.Context, namespace, labelKey,
	labelVal string, gvk schema.GroupVersionKind,
	additionalFilterLabelSelector labels.Selector) ([]client.Object, error) {
	objectList := &unstructured.UnstructuredList{}
	objectList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gvk.Group,
		Kind:    gvk.Kind,
		Version: gvk.Version,
	})

	listOptions := []client.ListOption{client.InNamespace(namespace)}
	mergedLabelSelector, err := handler.CreateMergedObjectFilterSelector(ctx, labelKey, labelVal,
		additionalFilterLabelSelector)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("Failed to create merged object filter selector while reconciling Project")
		return nil, err
	}
	if mergedLabelSelector != nil {
		listOptions = append(listOptions, client.MatchingLabelsSelector{Selector: mergedLabelSelector})
	}

	err = handler.Client.List(ctx, objectList, listOptions...)
	if err != nil {
		if labelKey != "" {
			log.Ctx(ctx).Error().Err(err).Msgf(
				"Failed to list objects in namespace with scope-wide label while reconciling Project. gvk: %s, namespace: %s, scopeWideLabelKey: %s, scopeWideLabelVal: %s",
				gvk.String(), namespace, labelKey, labelVal)
		} else {
			log.Ctx(ctx).Error().Err(err).Msgf("Failed to list objects in namespace while reconciling Project. gvk: %s, namespace: %s",
				gvk.String(), namespace)
		}

		return []client.Object{}, err
	}

	objectItems := []client.Object{}
	err = objectList.EachListItem(func(item runtime.Object) error {
		gvkObject, innerErr := scheme.Scheme.New(gvk)
		if innerErr != nil {
			return innerErr
		}
		innerErr = scheme.Scheme.Convert(item, gvkObject, nil)
		if innerErr != nil {
			return innerErr
		}

		objectItems = append(objectItems, gvkObject.(client.Object))
		return nil
	})
	if err != nil {
		log.Ctx(ctx).Info().Msgf("Failed to convert objects in objectList. namespace: %s, gvk: %s",
			namespace, gvk.String())
		return []client.Object{}, err
	}

	return objectItems, nil
}

func (handler *ClusterWideGVKCommon) CreateMergedObjectFilterSelector(ctx context.Context, labelKey, labelVal string,
	additionalFilterLabelSelector labels.Selector) (labels.Selector, error) {
	if labelKey == "" {
		return additionalFilterLabelSelector, nil
	}
	var err error
	if labelVal != "" {
		additionalFilterLabelSelector, err = AddRequirementToLabelSelector(additionalFilterLabelSelector,
			labelKey, selection.Equals, []string{labelVal})
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("Failed to add requirement to label selector while reconciling Project")
			return nil, err
		}
	} else {
		additionalFilterLabelSelector, err = AddRequirementToLabelSelector(additionalFilterLabelSelector,
			labelKey, selection.Exists, nil)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msgf("Failed to add requirement to label selector while reconciling Project")
			return nil, err
		}
	}

	return additionalFilterLabelSelector, nil
}

// unused
func (handler *ClusterWideGVKCommon) ListObjects(
	ctx context.Context, namespace string, gvk schema.GroupVersionKind) ([]client.Object, error) {
	return handler.ListObjectsWithLabel(ctx, namespace, "", "", gvk, nil)
}

func (handler *ClusterWideGVKCommon) ShouldHandleExisting(ctx context.Context, object client.Object) bool {
	switch object.(type) {
	case *corev1.Secret:
		return true
	case *corev1.ConfigMap:
		return true
	case *corev1.PersistentVolumeClaim:
		return true
	case *resourcev1.ResourceClaimTemplate:
		return true
	}
	log.Ctx(ctx).Error().Msgf("unsupported type")
	return false
}

func (handler *ClusterWideGVKCommon) DeepCopySpec(ctx context.Context, from, to client.Object) {
	switch from.(type) {
	case *corev1.Secret:
		handler.typeToObjectInspector[secretType].DeepCopySpec(from, to)
	case *corev1.ConfigMap:
		handler.typeToObjectInspector[configmapType].DeepCopySpec(from, to)
	case *corev1.PersistentVolumeClaim:
		handler.typeToObjectInspector[pvcType].DeepCopySpec(from, to)
	case *resourcev1.ResourceClaimTemplate:
		handler.typeToObjectInspector[rctType].DeepCopySpec(from, to)
	default:
		log.Ctx(ctx).Error().Msgf("unsupported type")
	}
}

func (handler *ClusterWideGVKCommon) AreEqual(ctx context.Context, left, right client.Object) bool {
	switch left.(type) {
	case *corev1.Secret:
		return handler.typeToObjectInspector[secretType].AreEqual(left, right)
	case *corev1.ConfigMap:
		return handler.typeToObjectInspector[configmapType].AreEqual(left, right)
	case *corev1.PersistentVolumeClaim:
		return handler.typeToObjectInspector[pvcType].AreEqual(left, right)
	case *resourcev1.ResourceClaimTemplate:
		return handler.typeToObjectInspector[rctType].AreEqual(left, right)
	}
	log.Ctx(ctx).Error().Msgf("unsupported type")
	return false
}

// IsSpecImmutable reports whether the handler for the given object's kind
// has opted into the delete-and-recreate flow. Used by the shared
// ClusterWideReplicatedResourceHandler to pick the right code path for
// kinds whose Spec the apiserver treats as immutable.
func (handler *ClusterWideGVKCommon) IsSpecImmutable(object client.Object) bool {
	var inspector ObjectInspector
	switch object.(type) {
	case *corev1.Secret:
		inspector = handler.typeToObjectInspector[secretType]
	case *corev1.ConfigMap:
		inspector = handler.typeToObjectInspector[configmapType]
	case *corev1.PersistentVolumeClaim:
		inspector = handler.typeToObjectInspector[pvcType]
	case *resourcev1.ResourceClaimTemplate:
		inspector = handler.typeToObjectInspector[rctType]
	}
	if im, ok := inspector.(immutableSpecInspector); ok {
		return im.IsSpecImmutable()
	}
	return false
}

// PreserveImmutableFields preserves immutable fields from the existing object to the new object
// to avoid breaking Kubernetes constraints (e.g., PVC-PV bindings)
func (handler *ClusterWideGVKCommon) PreserveImmutableFields(existing, new client.Object) {
	// Only PVCs have immutable fields that need preservation
	existingPvc, existingIsPvc := existing.(*corev1.PersistentVolumeClaim)
	newPvc, newIsPvc := new.(*corev1.PersistentVolumeClaim)

	if existingIsPvc && newIsPvc && existingPvc.Spec.VolumeName != "" {
		// Preserve the VolumeName to maintain PV binding
		newPvc.Spec.VolumeName = existingPvc.Spec.VolumeName
	}
}

// buildGVKObjectConfig
// when adding a new object type you want to replicate
// from runai namespace to all runai-managed namespaces (projects namespaces),
// add it here :)
// also make sure to implement the ObjectInspector interface,
// and add the appropriate handlers - in DeepCopySpec and AreEqual.
func (handler *ClusterWideGVKCommon) buildGVKObjectConfig(config *ClusterWideReplicatorGVKConfig) error {
	if config.ClusterWideSecrets {
		err := handler.buildSecretsConfig()
		if err != nil {
			return err
		}
	}

	if config.ClusterWidePvcs {
		err := handler.buildPvcsConfig()
		if err != nil {
			return err
		}
	}

	if config.ClusterWideConfigMaps {
		if err := handler.buildConfigMapConfig(); err != nil {
			return err
		}
	}

	if config.ClusterWideResourceClaimTemplates {
		if err := handler.buildRctConfig(); err != nil {
			return err
		}
	}

	return nil
}

func (handler *ClusterWideGVKCommon) buildSecretsConfig() error {
	secretGvk, err := getGvk(&corev1.Secret{}, secretType)
	if err != nil {
		return err
	}

	secretObjectInspector := SecretObjectInspector{
		Client: handler.Client,
	}

	handler.typeToObjectInspector[secretType] = secretObjectInspector

	dockerRegistryLabelSelectorAsset, err := CreateLabelSelector(commonconstants.RunaiAssetLabel, selection.Equals,
		[]string{DockerRegistryAssetKind})
	if err != nil {
		return err
	}
	nonDockerRegistryLabelSelectorAsset, err := CreateLabelSelector(commonconstants.RunaiAssetLabel, selection.In,
		[]string{PasswordAssetKind, AccessKeyAssetKind, GenericSecretAssetKind})
	if err != nil {
		return err
	}

	dockerRegistryLabelSelectorUser, err := CreateLabelSelector(commonconstants.RunaiResourceLabel, selection.Equals,
		[]string{DockerRegistryAssetKind})
	if err != nil {
		return err
	}
	nonDockerRegistryLabelSelectorUser, err := CreateLabelSelector(commonconstants.RunaiResourceLabel, selection.In,
		[]string{PasswordAssetKind, AccessKeyAssetKind, GenericSecretAssetKind})
	if err != nil {
		return err
	}

	workloadAuthSecretLabelSelector, err := CreateLabelSelector(commonconstants.LabelName, selection.Equals,
		[]string{commonconstants.WorkloadAuthSecretName})
	if err != nil {
		return err
	}

	runaiCertCASecretLabelSelector, err := CreateLabelSelector(commonconstants.LabelName, selection.Equals,
		[]string{commonconstants.RunaiCASecretName})
	if err != nil {
		return err
	}

	handler.gvkObjectConfigs[secretGvk] = []ObjectConfig{
		//  Asset Sync -> Tenant Scope
		{
			commonconstants.TenantScopeLabel,
			nil,
			nonDockerRegistryLabelSelectorAsset,
			clusterScope,
		},
		{
			commonconstants.TenantScopeLabel,
			secretObjectInspector.updateServiceAccountWithNamespacedSecrets,
			dockerRegistryLabelSelectorAsset,
			clusterScope,
		},

		//  Asset Sync -> Cluster Scope
		{
			commonconstants.ClusterWideResourceLabel,
			nil,
			nonDockerRegistryLabelSelectorAsset,
			clusterScope,
		},
		{
			commonconstants.ClusterWideResourceLabel,
			secretObjectInspector.updateServiceAccountWithNamespacedSecrets,
			dockerRegistryLabelSelectorAsset,
			clusterScope,
		},

		//  Asset Sync -> Department Scope
		{
			commonconstants.DepartmentScopeLabel,
			nil,
			nonDockerRegistryLabelSelectorAsset,
			departmentScope,
		},
		{
			commonconstants.DepartmentScopeLabel,
			secretObjectInspector.updateServiceAccountWithNamespacedSecrets,
			dockerRegistryLabelSelectorAsset,
			departmentScope,
		},

		// Workload Controller -> Cluster Scope
		{
			commonconstants.ClusterWideResourceLabel,
			nil,
			workloadAuthSecretLabelSelector,
			clusterScope,
		},

		//  User Owned -> Cluster Scope
		{
			commonconstants.ClusterWideResourceLabel,
			nil,
			nonDockerRegistryLabelSelectorUser,
			clusterScope,
		},
		{
			commonconstants.ClusterWideResourceLabel,
			secretObjectInspector.updateServiceAccountWithNamespacedSecrets,
			dockerRegistryLabelSelectorUser,
			clusterScope,
		},

		//  User Owned -> Department Scope
		{
			commonconstants.DepartmentScopeLabel,
			nil,
			nonDockerRegistryLabelSelectorUser,
			departmentScope,
		},
		{
			commonconstants.DepartmentScopeLabel,
			secretObjectInspector.updateServiceAccountWithNamespacedSecrets,
			dockerRegistryLabelSelectorUser,
			departmentScope,
		},
		{
			commonconstants.ClusterWideResourceLabel,
			nil,
			runaiCertCASecretLabelSelector,
			clusterScope,
		},
	}

	return nil
}

func (handler *ClusterWideGVKCommon) buildPvcsConfig() error {
	pvcGvk, err := getGvk(&corev1.PersistentVolumeClaim{}, pvcType)
	if err != nil {
		return err
	}

	nonDataSharingPvcLabelSelector, err := CreateLabelSelector(commonconstants.RunaiDataSharingLabel, selection.NotEquals, []string{"true"})
	if err != nil {
		return err
	}

	handler.typeToObjectInspector[pvcType] = PvcObjectInspector(true)
	handler.gvkObjectConfigs[pvcGvk] = []ObjectConfig{
		{
			commonconstants.ClusterWideResourceLabel,
			nil,
			nonDataSharingPvcLabelSelector,
			clusterScope,
		},
		{
			commonconstants.TenantScopeLabel,
			nil,
			nonDataSharingPvcLabelSelector,
			clusterScope,
		},
		{
			commonconstants.DepartmentScopeLabel,
			nil,
			nonDataSharingPvcLabelSelector,
			departmentScope,
		},
	}

	return nil
}

func (handler *ClusterWideGVKCommon) buildConfigMapConfig() error {
	gvk, err := getGvk(&corev1.ConfigMap{}, configmapType)
	if err != nil {
		return err
	}

	handler.typeToObjectInspector[configmapType] = ConfigMapObjectInspector(true)
	handler.gvkObjectConfigs[gvk] = []ObjectConfig{
		{
			commonconstants.ClusterWideResourceLabel,
			nil,
			nil,
			clusterScope,
		},
		{
			commonconstants.DepartmentScopeLabel,
			nil,
			nil,
			departmentScope,
		},
	}

	return nil
}

func getGvk(object runtime.Object, objectType ObjectType) (schema.GroupVersionKind, error) {
	gvks, _, err := scheme.Scheme.ObjectKinds(object)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	if len(gvks) == 0 {
		err = fmt.Errorf("no kinds registered for object <%v> in scheme", objectType)
		return schema.GroupVersionKind{}, err
	}
	if len(gvks) > 1 {
		err = fmt.Errorf("too many kinds registered for object <%v> in scheme; make sure only one is registered", objectType)
		return schema.GroupVersionKind{}, err
	}

	return gvks[0], nil
}

// PVCS

func (pvc PvcObjectInspector) DeepCopySpec(from, to client.Object) {
	fromPvc, ok := from.(*corev1.PersistentVolumeClaim)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return
	}
	toPvc, ok := to.(*corev1.PersistentVolumeClaim)
	if !ok {
		return
	}

	fromPvc.Spec.DeepCopyInto(&toPvc.Spec)
	// don't copy VolumeName field - a (pvc <-> pc) binding is a one-to-one mapping,
	// so if an existing pvc is bound to a pv, the replicated pvc in a different namespace will
	// be bound to a different pv -> hence we don't want this field to be copied to the new object.
	toPvc.Spec.VolumeName = ""
}

func (pvc PvcObjectInspector) AreEqual(left, right client.Object) bool {
	leftPvc, ok := left.(*corev1.PersistentVolumeClaim)
	if !ok {
		return false
	}
	rightPvc, ok := right.(*corev1.PersistentVolumeClaim)
	if !ok {
		return false
	}

	leftRunaiAnnotations := filterRunaiAnnotations(leftPvc.Annotations)
	rightRunaiAnnotations := filterRunaiAnnotations(rightPvc.Annotations)

	return reflect.DeepEqual(leftRunaiAnnotations, rightRunaiAnnotations)
}

// ConfigMap

func (cm ConfigMapObjectInspector) DeepCopySpec(from, to client.Object) {
	f, ok := from.(*corev1.ConfigMap)
	if !ok {
		// shouldn't happen because we convert objects to the correct type
		return
	}
	t, ok := to.(*corev1.ConfigMap)
	if !ok {
		return
	}

	fdc := f.DeepCopy()
	t.Data = fdc.Data
	t.BinaryData = fdc.BinaryData
	t.Immutable = fdc.Immutable
}

func (cm ConfigMapObjectInspector) AreEqual(left, right client.Object) bool {
	l, ok := left.(*corev1.ConfigMap)
	if !ok {
		return false
	}
	r, ok := right.(*corev1.ConfigMap)
	if !ok {
		return false
	}

	return reflect.DeepEqual(l.Name, r.Name) &&
		reflect.DeepEqual(l.Namespace, r.Namespace) &&
		reflect.DeepEqual(l.Data, r.Data) &&
		reflect.DeepEqual(l.BinaryData, r.BinaryData) &&
		reflect.DeepEqual(l.OwnerReferences, r.OwnerReferences) &&
		reflect.DeepEqual(l.Labels[replicatedProjectNameLabelKey], r.Labels[replicatedProjectNameLabelKey]) &&
		reflect.DeepEqual(l.Labels[replicatedProjectIdLabelKey], r.Labels[replicatedProjectIdLabelKey])
}

// Secrets

func (secret SecretObjectInspector) DeepCopySpec(from, to client.Object) {
	fromSecret, ok := from.(*corev1.Secret)
	if !ok {
		return
	}
	toSecret, ok := to.(*corev1.Secret)
	if !ok {
		return
	}

	toSecret.Data = fromSecret.Data
	toSecret.StringData = fromSecret.StringData
	toSecret.Type = fromSecret.Type
}

func (secret SecretObjectInspector) AreEqual(left, right client.Object) bool {
	leftSecret, ok := left.(*corev1.Secret)
	if !ok {
		return false
	}
	rightSecret, ok := right.(*corev1.Secret)
	if !ok {
		return false
	}

	return reflect.DeepEqual(leftSecret.Name, rightSecret.Name) &&
		reflect.DeepEqual(leftSecret.Namespace, rightSecret.Namespace) &&
		reflect.DeepEqual(leftSecret.Data, rightSecret.Data) &&
		reflect.DeepEqual(leftSecret.StringData, rightSecret.StringData) &&
		reflect.DeepEqual(leftSecret.Type, rightSecret.Type) &&
		reflect.DeepEqual(leftSecret.OwnerReferences, rightSecret.OwnerReferences)
}

func (secret SecretObjectInspector) updateServiceAccountWithNamespacedSecrets(
	ctx context.Context, namespace string, secretsReplicatedInNamespace, unreferencedSecrets []string) error {
	serviceAccount, err := secret.getDefaultServiceAccountForNamespace(ctx, namespace)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("Failed to retrieve the default Service Account for Namespace, can't update secrets. namespace: %s",
			namespace)
		return nil
	}

	return secret.updateServiceAccountWithSecrets(ctx, serviceAccount, secretsReplicatedInNamespace, unreferencedSecrets)
}

func (secret SecretObjectInspector) getDefaultServiceAccountForNamespace(ctx context.Context, namespace string,
) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{}
	err := secret.Client.Get(ctx,
		client.ObjectKey{
			Namespace: namespace,
			Name:      DefaultServiceAccountName,
		}, serviceAccount)
	return serviceAccount, err
}

// Updates the Namespace' default Service Account with all current namespaced Secrets
func (secret SecretObjectInspector) updateServiceAccountWithSecrets(
	ctx context.Context, serviceAccount *corev1.ServiceAccount, secretsReplicatedInNamespace, unreferencedSecrets []string) error {
	serviceAccountOriginalImagePullSecrets := append([]corev1.LocalObjectReference{}, serviceAccount.ImagePullSecrets...)
	updateServiceAccountWithCurrentSecrets(serviceAccount, secretsReplicatedInNamespace, unreferencedSecrets)

	if reflect.DeepEqual(serviceAccount.ImagePullSecrets, serviceAccountOriginalImagePullSecrets) {
		log.Ctx(ctx).Debug().Msgf("Default Service Account is updated with ImagePullSecrets, skipping update. namespace: %s, serviceAccount: %s, imagePullSecrets: %s",
			serviceAccount.Namespace, serviceAccount.Name,
			fmt.Sprintf("%v", serviceAccount.ImagePullSecrets))
		return nil
	}

	log.Ctx(ctx).Info().Msgf("Updating Namespace's default Service Account with ImagePullSecrets. namespace: %s, serviceAccount: %s, imagePullSecrets: %s",
		serviceAccount.Namespace, serviceAccount.Name,
		fmt.Sprintf("%v", serviceAccount.ImagePullSecrets))

	err := secret.Client.Update(ctx, serviceAccount)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("Error updating Namespace's default Service Account with namespaced secrets. namespace: %s, serviceAccount: %s",
			serviceAccount.Namespace, serviceAccount.Name)
		return err
	}
	return nil
}

func updateServiceAccountWithCurrentSecrets(
	serviceAccount *corev1.ServiceAccount, secretsReplicatedInNamespace, unreferencedSecrets []string) {
	var serviceAccountSecrets []string
	for _, serviceAccountSecret := range serviceAccount.ImagePullSecrets {
		if !funk.ContainsString(unreferencedSecrets, serviceAccountSecret.Name) {
			// Remove unreferenced secrets from the service account
			serviceAccountSecrets = append(serviceAccountSecrets, serviceAccountSecret.Name)
		}
	}

	serviceAccountSecrets = append(serviceAccountSecrets, secretsReplicatedInNamespace...)
	serviceAccountSecrets = funk.UniqString(serviceAccountSecrets)
	serviceAccount.ImagePullSecrets = make([]corev1.LocalObjectReference, 0)
	for _, secret := range serviceAccountSecrets {
		serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, corev1.LocalObjectReference{Name: secret})
	}
}

func filterRunaiAnnotations(annotations map[string]string) map[string]string {
	filtered := make(map[string]string)
	for k, v := range annotations {
		if strings.Contains(k, "run.ai") {
			filtered[k] = v
		}
	}
	return filtered
}

func CreateLabelSelector(key string, operator selection.Operator, values []string) (labels.Selector, error) {
	return AddRequirementToLabelSelector(nil, key, operator, values)
}

func AddRequirementToLabelSelector(selector labels.Selector, key string, operator selection.Operator,
	values []string) (labels.Selector, error) {
	req, err := labels.NewRequirement(key, operator, values)
	if err != nil {
		return nil, err
	}
	if selector == nil {
		selector = labels.NewSelector()
	}
	selector = selector.Add(*req)
	return selector, nil
}
