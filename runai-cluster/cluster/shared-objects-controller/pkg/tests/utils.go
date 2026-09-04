/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega" //nolint:staticcheck // gomega dot import is intentional
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/shared-objects-controller/pkg/controller/datavolume"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// gvkRestoringClient wraps a fake client so that objects returned from Get/List
// keep their TypeMeta (apiVersion/kind) populated. The newer controller-runtime
// fake client strips TypeMeta, but production (informer-backed) clients and the
// pre-upgrade fake client preserve it. The datavolume reconciler relies on
// DataVolume.Kind (via SetOwner) when stamping owner references, so without this
// the owner refs would have an empty Kind and IsOwner would never match.
type gvkRestoringClient struct {
	client.WithWatch
	scheme *runtime.Scheme
}

func newGvkRestoringClient(inner client.WithWatch, scheme *runtime.Scheme) client.WithWatch {
	return &gvkRestoringClient{WithWatch: inner, scheme: scheme}
}

func (c *gvkRestoringClient) restoreGVK(obj runtime.Object) {
	if obj == nil {
		return
	}
	if gvks, _, err := c.scheme.ObjectKinds(obj); err == nil && len(gvks) > 0 {
		obj.GetObjectKind().SetGroupVersionKind(gvks[0])
	}
}

func (c *gvkRestoringClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if err := c.WithWatch.Get(ctx, key, obj, opts...); err != nil {
		return err
	}
	c.restoreGVK(obj)
	return nil
}

func (c *gvkRestoringClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if err := c.WithWatch.Patch(ctx, obj, patch, opts...); err != nil {
		return err
	}
	c.restoreGVK(obj)
	return nil
}

func (c *gvkRestoringClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if err := c.WithWatch.Update(ctx, obj, opts...); err != nil {
		return err
	}
	c.restoreGVK(obj)
	return nil
}

type testDvPreset string

const dv1 testDvPreset = "1"
const dv2 testDvPreset = "2"
const dv3 testDvPreset = "3"
const dv4 testDvPreset = "4"

func getDvFromClient(ctx context.Context, name string, k8sClient client.Client) *v1alpha1.DataVolume {
	dvKey := types.NamespacedName{Name: name}
	dvObj := &v1alpha1.DataVolume{}
	err := k8sClient.Get(ctx, dvKey, dvObj)
	if err != nil {
		return nil
	}
	return dvObj
}

//nolint:unused
func _getPvcFromClient(ctx context.Context, name, namespace string, k8sClient client.Client) *v1.PersistentVolumeClaim {
	pvcKey := types.NamespacedName{Name: name, Namespace: namespace}
	pvcObj := &v1.PersistentVolumeClaim{}
	err := k8sClient.Get(ctx, pvcKey, pvcObj)
	if err != nil {
		return nil
	}
	return pvcObj
}

//nolint:unused
func _getPvFromClient(ctx context.Context, name string, k8sClient client.Client) *v1.PersistentVolume {
	pvKey := types.NamespacedName{Name: name}
	pvObj := &v1.PersistentVolume{}
	err := k8sClient.Get(ctx, pvKey, pvObj)
	if err != nil {
		return nil
	}
	return pvObj
}

func convertPvcsToNamespaceToPvcMap(pvcs []v1.PersistentVolumeClaim) map[string]*v1.PersistentVolumeClaim {
	pvcMap := make(map[string]*v1.PersistentVolumeClaim)
	for i := range pvcs {
		pvc := &pvcs[i]
		pvcMap[pvc.Namespace] = pvc
	}
	return pvcMap
}

func convertPvsToPvMap(pvs []v1.PersistentVolume) map[string]*v1.PersistentVolume {
	pvMap := make(map[string]*v1.PersistentVolume)
	for i := range pvs {
		pv := &pvs[i]

		nsName, ok := pv.Labels[v1alpha1.DataVolumePvForNamespaceLabel]
		if !ok {
			continue
		}

		pvMap[nsName] = pv
	}
	return pvMap
}

func getTestScheme() *runtime.Scheme {
	scheme := scheme_init.Scheme()
	Expect(v1.AddToScheme(scheme)).Should(Succeed())
	Expect(v1alpha1.AddToScheme(scheme)).Should(Succeed())
	Expect(kaiorg.AddToScheme(scheme)).Should(Succeed())
	return scheme
}

func buildNamespace(name, projName string) *v1.Namespace {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			UID:  types.UID(name),
		},
	}

	if projName != "" {
		ns.Labels = map[string]string{
			constants.RunaiQueueLabel: projName,
		}
	}

	return ns
}

// buildKaiProject / buildKaiDepartment are the kai.resources twins, created alongside the legacy
// fixtures used by the datavolume suite.
// Spec.Parent replaces the legacy Spec.Department.
func buildKaiProject(projectName, departmentName string) *kaiorg.Project {
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			UID:  types.UID(projectName),
		},
		Spec: kaiorg.ProjectSpec{
			Parent: departmentName,
		},
	}
}

func buildKaiDepartment(departmentName string) *kaiorg.Department {
	return &kaiorg.Department{
		ObjectMeta: metav1.ObjectMeta{
			Name: departmentName,
			UID:  types.UID(departmentName),
		},
	}
}

func buildPvc(whichPv testDvPreset, originalPvcAccessMode *v1.PersistentVolumeAccessMode) *v1.PersistentVolumeClaim {
	switch whichPv {
	case dv1:
		return buildPvcInner2(pvcName1, testNamespace1, pvName1, originalPvcAccessMode)
	case dv2:
		return buildPvcInner2(pvcName2, testNamespace2, pvName2, originalPvcAccessMode)
	case dv3:
		return buildPvcInner2(pvcName3, testNamespace3, pvName3, originalPvcAccessMode)
	case dv4:
		return buildPvcInner2(pvcName4, testNamespace4, pvName4, originalPvcAccessMode)
	}
	return nil
}

func buildPvcOfDvInNs(whichPv testDvPreset, name, namespace, volumeName string, phase v1.PersistentVolumeClaimPhase) *v1.PersistentVolumeClaim {
	dvName := fmt.Sprintf("dv%s", whichPv)
	return buildPvcWithLabels(name, namespace, volumeName, phase, v1.ReadWriteOnce, ptr.To(dvName), map[string]string{v1alpha1.DataVolumeNameLabel: dvName, v1alpha1.DataVolumeIdLabel: dvName})
}

func buildPvcInner2(name, namespace, volumeName string, originalPvcAccessMode *v1.PersistentVolumeAccessMode) *v1.PersistentVolumeClaim {
	return buildPvcInner(name, namespace, volumeName, v1.ClaimBound, originalPvcAccessMode)
}

func buildPvcInner(name, namespace, volumeName string, phase v1.PersistentVolumeClaimPhase, originalPvcAccessModel *v1.PersistentVolumeAccessMode) *v1.PersistentVolumeClaim {
	accessMode := v1.ReadWriteOnce
	if originalPvcAccessModel != nil {
		accessMode = *originalPvcAccessModel
	}
	return buildPvcWithLabels(name, namespace, volumeName, phase, accessMode, nil, map[string]string{})
}

func buildPvcWithLabels(name, namespace, volumeName string, phase v1.PersistentVolumeClaimPhase,
	accessMode v1.PersistentVolumeAccessMode, datavolumeOwner *string, labels map[string]string) *v1.PersistentVolumeClaim {
	if phase == "" {
		phase = v1.ClaimBound
	}

	ownerRefs := []metav1.OwnerReference{}
	if datavolumeOwner != nil {
		ownerRefs = append(ownerRefs, metav1.OwnerReference{
			Kind:       v1alpha1.DataVolumeKind,
			APIVersion: "run/v1alpha1",
			Name:       *datavolumeOwner,
			UID:        types.UID(*datavolumeOwner),
		})
	}

	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			UID:             types.UID(name),
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: phase,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{accessMode},
			VolumeName:  volumeName,

			StorageClassName: ptr.To("blabla"),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"match": "label"},
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      "some-key",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"some-val"},
				}},
			},
			Resources: v1.VolumeResourceRequirements{
				Limits: v1.ResourceList{
					"nvidia.com/gpu": resource.MustParse("500m"),
				},
				Requests: v1.ResourceList{
					"cpu": resource.MustParse("300m"),
				},
			},
			VolumeMode: ptr.To(v1.PersistentVolumeBlock),
			DataSource: &v1.TypedLocalObjectReference{
				APIGroup: ptr.To("v1"),
				Kind:     "PersistentVolumeClaim",
				Name:     "some-pvc",
			},
			DataSourceRef: &v1.TypedObjectReference{
				APIGroup: ptr.To("some-ver"),
				Kind:     "some-kind",
				Name:     "some-pvc-2",
			},
		},
	}
}

func buildPv(whichPv testDvPreset, originalPvAccessMode *v1.PersistentVolumeAccessMode) *v1.PersistentVolume {
	switch whichPv {
	case dv1:
		return buildPvInner(pvName1, pvcName1, testNamespace1, originalPvAccessMode)
	case dv2:
		return buildPvInner(pvName2, pvcName2, testNamespace2, originalPvAccessMode)
	case dv3:
		return buildPvInner(pvName3, pvcName3, testNamespace2, originalPvAccessMode)
	case dv4:
		return buildPvInner(pvName4, pvcName4, testNamespace3, originalPvAccessMode)
	}
	return nil
}

func buildPvOfDvForNs(whichDv testDvPreset, name, claimName, claimNamespace string, originalPvAccessMode *v1.PersistentVolumeAccessMode) *v1.PersistentVolume {
	accessMode := v1.ReadWriteOncePod
	if originalPvAccessMode != nil {
		accessMode = *originalPvAccessMode
	}

	dvName := fmt.Sprintf("dv%s", whichDv)

	return buildPvWithLabels(name, claimName, claimNamespace, accessMode, v1.PersistentVolumeReclaimRetain,
		ptr.To(dvName), map[string]string{
			v1alpha1.DataVolumePvForNamespaceLabel: claimNamespace,
			v1alpha1.DataVolumeNameLabel:           dvName,
			v1alpha1.DataVolumeIdLabel:             dvName,
		})
}

func buildPvInner(name, claimName, claimNamespace string, originalPvAccessMode *v1.PersistentVolumeAccessMode) *v1.PersistentVolume {
	persistentVolumeAccessMode := v1.ReadWriteOncePod
	if originalPvAccessMode != nil {
		persistentVolumeAccessMode = *originalPvAccessMode
	}
	return buildPvWithLabels(name, claimName, claimNamespace, persistentVolumeAccessMode, v1.PersistentVolumeReclaimRecycle, nil, map[string]string{})
}

func buildPvWithStatus(name, claimName, claimNamespace string, phase v1.PersistentVolumePhase, originPvAccessMode *v1.PersistentVolumeAccessMode) *v1.PersistentVolume {
	pv := buildPvInner(name, claimName, claimNamespace, originPvAccessMode)
	pv.Status = v1.PersistentVolumeStatus{Phase: phase}
	return pv
}

func buildPvWithLabels(name, claimName, claimNamespace string, accessMode v1.PersistentVolumeAccessMode,
	persistentVolumeReclaimPolicy v1.PersistentVolumeReclaimPolicy, datavolumeOwner *string, labels map[string]string) *v1.PersistentVolume {
	return buildPvWithLabelsInner(name, claimName, claimNamespace, accessMode, persistentVolumeReclaimPolicy, datavolumeOwner, nil, labels)
}

func buildPvWithLabelsInner(name, claimName, claimNamespace string, accessMode v1.PersistentVolumeAccessMode,
	persistentVolumeReclaimPolicy v1.PersistentVolumeReclaimPolicy, datavolumeOwner *string, phase *v1.PersistentVolumePhase, labels map[string]string) *v1.PersistentVolume {
	var claimRef *v1.ObjectReference = nil
	if claimName != "" {
		claimRef = &v1.ObjectReference{
			Name:      claimName,
			Namespace: claimNamespace,
		}
	}

	ownerRefs := []metav1.OwnerReference{}
	if datavolumeOwner != nil {
		ownerRefs = append(ownerRefs, metav1.OwnerReference{
			Kind:       v1alpha1.DataVolumeKind,
			APIVersion: "run/v1alpha1",
			Name:       *datavolumeOwner,
			UID:        types.UID(*datavolumeOwner),
		})
	}

	phasee := v1.VolumeBound
	if phase != nil {
		phasee = *phase
	}

	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			UID:             types.UID(name),
			Labels:          labels,
			OwnerReferences: ownerRefs,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "v1",
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes:                   []v1.PersistentVolumeAccessMode{accessMode},
			ClaimRef:                      claimRef,
			PersistentVolumeReclaimPolicy: persistentVolumeReclaimPolicy,

			StorageClassName: "blabla",
			Capacity: v1.ResourceList{
				"storage": resource.MustParse("1Gi"),
			},
			MountOptions: []string{"some-option"},
			VolumeMode:   ptr.To(v1.PersistentVolumeFilesystem),
			NodeAffinity: &v1.VolumeNodeAffinity{Required: &v1.NodeSelector{NodeSelectorTerms: []v1.NodeSelectorTerm{{MatchExpressions: []v1.NodeSelectorRequirement{
				{Key: "1", Operator: v1.NodeSelectorOpIn, Values: []string{"2"}},
			}}}}},
			VolumeAttributesClassName: ptr.To("some-class"),
		},
		Status: v1.PersistentVolumeStatus{
			Phase: phasee,
		},
	}
}

func buildDatavolumeWithShouldUseOriginalVolume(dv *v1alpha1.DataVolume, projectNames []string, departmentNames []string, clusterScope string, shouldUseOriginalVolume bool) *v1alpha1.DataVolume {
	dv.Spec.SharedScopes = buildSharedScopes(projectNames, departmentNames, clusterScope)
	dv.Spec.ShouldUseOriginalVolume = shouldUseOriginalVolume
	return dv
}

func buildDatavolume(whichDv testDvPreset, sharedProjects, sharedDepartments []string, sharedCluster string) *v1alpha1.DataVolume {
	return buildDatavolume2(whichDv, false, sharedProjects, sharedDepartments, sharedCluster)
}

func buildDatavolume2(whichDv testDvPreset, shouldDeleteOriginalVolume bool, sharedProjects, sharedDepartments []string, sharedCluster string) *v1alpha1.DataVolume {
	var name, originPvcName, projectName string
	switch whichDv {
	case dv1:
		name = "dv1"
		originPvcName = pvcName1
		projectName = testProject1
	case dv2:
		name = "dv2"
		originPvcName = pvcName2
		projectName = testProject2
	case dv3:
		name = "dv3"
		originPvcName = pvcName3
		projectName = testProject3
	case dv4:
		name = "dv4"
		originPvcName = pvcName4
		projectName = testProject4
	}

	return buildDatavolumeInner(name, originPvcName, projectName, shouldDeleteOriginalVolume, sharedProjects, sharedDepartments, sharedCluster)
}

func buildDatavolumeWithStatus(whichDv testDvPreset, status *v1alpha1.DataVolumeStatus, sharedProjects, sharedDepartments []string, sharedCluster string) *v1alpha1.DataVolume {
	dv := buildDatavolume(whichDv, sharedProjects, sharedDepartments, sharedCluster)
	dv.Status = *status
	return dv
}

func buildDatavolumeInner(name, originPvcName, projectName string, shouldDeleteOriginalVolume bool, sharedProjects, sharedDepartments []string, sharedCluster string) *v1alpha1.DataVolume {
	sharedScopes := buildSharedScopes(sharedProjects, sharedDepartments, sharedCluster)

	return &v1alpha1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			UID:  types.UID(name),
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.DataVolumeKind,
			APIVersion: "run/v1alpha1",
		},
		Spec: v1alpha1.DataVolumeSpec{
			DatavolumeId:               name,
			Description:                "test description",
			OriginPvcName:              originPvcName,
			ProjectName:                projectName,
			ShouldDeleteOriginalVolume: shouldDeleteOriginalVolume,
			SharedScopes:               sharedScopes,
		},
	}
}

func buildSharedScopes(sharedProjects, sharedDepartments []string, sharedCluster string) []v1alpha1.DataVolumeSharedScope {
	sharedScopes := append(buildSharedScopesInner(sharedProjects, v1alpha1.ProjectScope), buildSharedScopesInner(sharedDepartments, v1alpha1.DepartmentScope)...)
	if sharedCluster != "" {
		sharedScopes = append(sharedScopes, buildSharedScopesInner([]string{sharedCluster}, v1alpha1.ClusterScope)...)
	}
	return sharedScopes
}

func buildSharedScopesInner(scopes []string, scopeType v1alpha1.ScopeType) []v1alpha1.DataVolumeSharedScope {
	sharedScopes := make([]v1alpha1.DataVolumeSharedScope, len(scopes))
	for i, scopeName := range scopes {
		sharedScopes[i] = v1alpha1.DataVolumeSharedScope{ScopeType: scopeType, ScopeName: scopeName}
	}
	return sharedScopes
}

func buildDataVolumeStatus(phase v1alpha1.DataVolumePhase, message datavolume.DatavolumePhaseMessage, datavolumePvcName, datavolumePvName string) *v1alpha1.DataVolumeStatus {
	return &v1alpha1.DataVolumeStatus{
		Phase:             phase,
		Message:           string(message),
		DatavolumePvcName: datavolumePvcName,
		DatavolumePvName:  datavolumePvName,
	}
}
