/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"context"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	field_types "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("resolveStorageRefs", func() {
	const ns = "runai-project-a"

	newParser := func(objects ...client.Object) *WorkloadParser {
		builder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(objects...)
		return NewWorkloadParser(fakecachedclient.NewFakeCachedClient(builder))
	}

	It("resolves PVC, Secret, and ConfigMap refs to their UIDs", func() {
		pvc := &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "data", Namespace: ns, UID: "pvc-uid"}}
		secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "regcred", Namespace: ns, UID: "secret-uid"}}
		cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cfg", Namespace: ns, UID: "cm-uid"}}
		parser := newParser(pvc, secret, cm)

		resolved := parser.resolveStorageRefs(context.TODO(), ns, []storageRef{
			{kind: v1alpha1.StorageRefKindPVC, name: "data"},
			{kind: v1alpha1.StorageRefKindSecret, name: "regcred"},
			{kind: v1alpha1.StorageRefKindConfigMap, name: "cfg"},
		})

		Expect(resolved).To(HaveLen(3))
		Expect(resolved[0].Name).To(Equal("data"))
		Expect(resolved[0].Kind).To(Equal(openapi.ResourceRefKind(v1alpha1.StorageRefKindPVC)))
		Expect(resolved[0].Uid).ToNot(BeNil())
		Expect(*resolved[0].Uid).To(Equal("pvc-uid"))
		Expect(*resolved[1].Uid).To(Equal("secret-uid"))
		Expect(resolved[1].Kind).To(Equal(openapi.ResourceRefKind(v1alpha1.StorageRefKindSecret)))
		Expect(*resolved[2].Uid).To(Equal("cm-uid"))
		Expect(resolved[2].Kind).To(Equal(openapi.ResourceRefKind(v1alpha1.StorageRefKindConfigMap)))
	})

	It("skips refs that are not present in cache", func() {
		secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "regcred", Namespace: ns, UID: "secret-uid"}}
		parser := newParser(secret)

		resolved := parser.resolveStorageRefs(context.TODO(), ns, []storageRef{
			{kind: v1alpha1.StorageRefKindPVC, name: "missing"},
			{kind: v1alpha1.StorageRefKindSecret, name: "regcred"},
		})

		Expect(resolved).To(HaveLen(1))
		Expect(resolved[0].Name).To(Equal("regcred"))
	})

	It("skips refs with an unsupported kind instead of resolving them as a ConfigMap", func() {
		cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "bogus", Namespace: ns, UID: "cm-uid"}}
		parser := newParser(cm)

		resolved := parser.resolveStorageRefs(context.TODO(), ns, []storageRef{
			{kind: v1alpha1.StorageRefKind("Deployment"), name: "bogus"},
		})

		Expect(resolved).To(BeEmpty())
	})

	It("returns an empty slice when there are no refs", func() {
		parser := newParser()
		Expect(parser.resolveStorageRefs(context.TODO(), ns, nil)).To(BeEmpty())
	})
})

var _ = Describe("nativeStorageRefs", func() {
	It("collects PVCs, secret and configMap volumes, gitSync, s3, and imagePullSecrets with dedup", func() {
		sf := field_groups.StorageFields{
			Pvcs: field_types.NewPvcsField(map[string]field_types.PvcsItemAndRules{
				"a": {Value: &field_types.PvcsItemValue{ClaimName: ptr.To("existing-pvc"), ExistingPvc: ptr.To(true)}},
				"b": {Value: &field_types.PvcsItemValue{Size: ptr.To("1Gi")}},
				"c": {Value: &field_types.PvcsItemValue{Ephemeral: ptr.To(true)}},
			}),
			SecretVolumes: field_types.NewSecretVolumeField(map[string]field_types.SecretVolumeItemAndRules{
				"s": {Value: &field_types.SecretVolumeItemValue{Name: ptr.To("my-secret")}},
			}),
			ConfigMapVolumes: field_types.NewConfigMapVolumeField(map[string]field_types.ConfigMapVolumeItemAndRules{
				"cm": {Value: &field_types.ConfigMapVolumeItemValue{Name: ptr.To("my-cm")}},
			}),
			GitSync: field_types.NewGitSyncField(map[string]field_types.GitSyncItemAndRules{
				"g": {Value: &field_types.GitSyncItemValue{Repository: ptr.To("repo"), SecretName: ptr.To("git-secret")}},
			}),
			S3buckets: field_types.NewS3bucketField(map[string]field_types.S3bucketItemAndRules{
				"x": {Value: &field_types.S3bucketItemValue{Bucket: ptr.To("b"), Path: ptr.To("/"), SecretName: ptr.To("my-secret")}},
			}),
		}
		imagePullSecrets := &field_types.StringField{Value: ptr.To("regcred my-secret")}

		refs := nativeStorageRefs("tw-1", "runai-project-a", sf, imagePullSecrets)

		Expect(refs).To(ConsistOf(
			storageRef{kind: v1alpha1.StorageRefKindPVC, name: "existing-pvc"},
			storageRef{kind: v1alpha1.StorageRefKindPVC, name: "pvc-tw-1-1"},
			storageRef{kind: v1alpha1.StorageRefKindSecret, name: "my-secret"},
			storageRef{kind: v1alpha1.StorageRefKindConfigMap, name: "my-cm"},
			storageRef{kind: v1alpha1.StorageRefKindSecret, name: "git-secret"},
			storageRef{kind: v1alpha1.StorageRefKindSecret, name: "regcred"},
		))
	})

	It("returns nil for empty storage fields and no imagePullSecrets", func() {
		Expect(nativeStorageRefs("tw-1", "runai-project-a", field_groups.StorageFields{}, nil)).To(BeEmpty())
	})
})
