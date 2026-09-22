/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package personalcreds

import (
	"context"
	"os"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	pcreds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// known-good fixture: encryptionKey decrypts encryptedValue to "private-value-1"
// (same fixture used by workload-common/builders/credentials_test.go).
const encryptionKey = "905c18f0839affcddca117dcac633f79"

var encryptedValue = []byte{
	168, 221, 244, 73, 135, 225, 25, 104, 145, 32, 118, 20, 196, 29, 72, 165,
	224, 40, 97, 180, 83, 157, 205, 140, 158, 186, 60, 73, 173, 193, 146, 219,
	201, 199, 65, 36, 241, 251, 52, 131, 38, 155, 0,
}

var (
	testRunner = &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-runner",
			Namespace: "my-project",
			UID:       "runner-uid-1",
			Annotations: map[string]string{
				pcreds.Annotation: "- sourceSecret: my-token\n",
			},
		},
	}
)

func newReconcileScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
	return scheme
}

func newReconcileClient(scheme *runtime.Scheme, objs ...client.Object) client.Client {
	return testutil.WithSSASupport(
		fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build(),
	)
}

func encryptedSource(name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testRunaiNS},
		Data:       map[string][]byte{"token": encryptedValue},
	}
}

var _ = Describe("HandleAdditionalResources", func() {
	var (
		ctx       context.Context
		scheme    *runtime.Scheme
		k8sClient client.Client
		state     *reconcile.State
		plugin    *Plugin
	)

	BeforeEach(func() {
		ctx = context.Background()
		Expect(os.Setenv(constants.RunaiAdsEncryptionKeyEnvVar, encryptionKey)).To(Succeed())
		scheme = newReconcileScheme()
		state = reconcile.New()
	})

	AfterEach(func() {
		Expect(os.Unsetenv(constants.RunaiAdsEncryptionKeyEnvVar)).To(Succeed())
	})

	It("is a no-op when annotation is absent", func() {
		k8sClient = newReconcileClient(scheme)
		plugin = NewPlugin(k8sClient, testRunaiNS)
		runner := testRunner.DeepCopy()
		runner.Annotations = nil

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())
		Expect(state.PluginIssues(plugin)).To(BeEmpty())
	})

	It("creates the destination secret with decrypted data and ownerRef", func() {
		k8sClient = newReconcileClient(scheme, encryptedSource("my-token"))
		plugin = NewPlugin(k8sClient, testRunaiNS)

		Expect(plugin.HandleAdditionalResources(ctx, testRunner.DeepCopy(), nil, nil, k8sClient, state)).To(Succeed())

		var dest corev1.Secret
		Expect(k8sClient.Get(ctx, types.NamespacedName{
			Name:      pcreds.DestName(testRunner.Name, "my-token"),
			Namespace: testRunner.Namespace,
		}, &dest)).To(Succeed())

		Expect(dest.Data).To(HaveKeyWithValue("token", []byte("private-value-1")))
		Expect(metav1.IsControlledBy(&dest, testRunner)).To(BeTrue())
		Expect(dest.Labels).To(HaveKeyWithValue(pcreds.CopyLabel, "true"))
		Expect(state.PluginIssues(plugin)).To(BeEmpty())
	})

	It("sets destination type to dockerconfigjson when source has the docker-registry asset label", func() {
		source := encryptedSource("my-token")
		source.Labels = map[string]string{constants.RunaiAssetLabel: pcreds.DockerRegistryAssetKind}
		k8sClient = newReconcileClient(scheme, source)
		plugin = NewPlugin(k8sClient, testRunaiNS)

		Expect(plugin.HandleAdditionalResources(ctx, testRunner.DeepCopy(), nil, nil, k8sClient, state)).To(Succeed())

		var dest corev1.Secret
		Expect(k8sClient.Get(ctx, types.NamespacedName{
			Name:      pcreds.DestName(testRunner.Name, "my-token"),
			Namespace: testRunner.Namespace,
		}, &dest)).To(Succeed())

		Expect(dest.Type).To(Equal(corev1.SecretTypeDockerConfigJson))
		Expect(dest.Labels).NotTo(HaveKey(constants.RunaiAssetLabel))
	})

	It("preserves the source secret type when the source is not a docker-registry credential", func() {
		source := encryptedSource("my-token")
		source.Type = corev1.SecretTypeBasicAuth
		k8sClient = newReconcileClient(scheme, source)
		plugin = NewPlugin(k8sClient, testRunaiNS)

		Expect(plugin.HandleAdditionalResources(ctx, testRunner.DeepCopy(), nil, nil, k8sClient, state)).To(Succeed())

		var dest corev1.Secret
		Expect(k8sClient.Get(ctx, types.NamespacedName{
			Name:      pcreds.DestName(testRunner.Name, "my-token"),
			Namespace: testRunner.Namespace,
		}, &dest)).To(Succeed())

		Expect(dest.Type).To(Equal(corev1.SecretTypeBasicAuth))
	})

	It("is idempotent when the destination is already owned by the runner", func() {
		source := encryptedSource("my-token")
		owned := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pcreds.DestName(testRunner.Name, "my-token"),
				Namespace: testRunner.Namespace,
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: dromosv1alpha1.GroupVersion.String(),
					Kind:       "WorkloadRunner",
					Name:       testRunner.Name,
					UID:        testRunner.UID,
					Controller: ptr.To(true),
				}},
			},
			Data: map[string][]byte{"token": []byte("stale-value")},
		}
		k8sClient = newReconcileClient(scheme, source, owned)
		plugin = NewPlugin(k8sClient, testRunaiNS)

		Expect(plugin.HandleAdditionalResources(ctx, testRunner.DeepCopy(), nil, nil, k8sClient, state)).To(Succeed())

		var dest corev1.Secret
		Expect(k8sClient.Get(ctx, types.NamespacedName{
			Name:      owned.Name,
			Namespace: owned.Namespace,
		}, &dest)).To(Succeed())
		Expect(dest.Data).To(HaveKeyWithValue("token", []byte("private-value-1")))
		Expect(state.PluginIssues(plugin)).To(BeEmpty())
	})

	It("records a Storage issue when the destination exists and is not owned by the runner", func() {
		source := encryptedSource("my-token")
		stranger := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pcreds.DestName(testRunner.Name, "my-token"),
				Namespace: testRunner.Namespace,
			},
		}
		k8sClient = newReconcileClient(scheme, source, stranger)
		plugin = NewPlugin(k8sClient, testRunaiNS)

		err := plugin.HandleAdditionalResources(ctx, testRunner.DeepCopy(), nil, nil, k8sClient, state)
		Expect(err).To(HaveOccurred())

		issues := state.PluginIssues(plugin)
		Expect(issues).To(HaveLen(1))
		Expect(issues[0].Group).To(Equal(issue.Storage))
		Expect(issues[0].IsCritical).To(BeTrue())
	})

	It("records a Storage issue when the source secret is missing", func() {
		k8sClient = newReconcileClient(scheme)
		plugin = NewPlugin(k8sClient, testRunaiNS)

		err := plugin.HandleAdditionalResources(ctx, testRunner.DeepCopy(), nil, nil, k8sClient, state)
		Expect(err).To(HaveOccurred())

		issues := state.PluginIssues(plugin)
		Expect(issues).To(HaveLen(1))
		Expect(issues[0].Group).To(Equal(issue.Storage))
	})

	It("records an Internal issue when the encryption key env var is missing", func() {
		Expect(os.Unsetenv(constants.RunaiAdsEncryptionKeyEnvVar)).To(Succeed())
		k8sClient = newReconcileClient(scheme, encryptedSource("my-token"))
		plugin = NewPlugin(k8sClient, testRunaiNS)

		err := plugin.HandleAdditionalResources(ctx, testRunner.DeepCopy(), nil, nil, k8sClient, state)
		Expect(err).To(HaveOccurred())

		issues := state.PluginIssues(plugin)
		Expect(issues).To(HaveLen(1))
		Expect(issues[0].Group).To(Equal(issue.Internal))
	})

	It("processes multiple entries and stops at the first failure", func() {
		k8sClient = newReconcileClient(scheme, encryptedSource("a"))
		plugin = NewPlugin(k8sClient, testRunaiNS)
		runner := testRunner.DeepCopy()
		runner.Annotations[pcreds.Annotation] = "- sourceSecret: a\n- sourceSecret: missing\n"

		err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)
		Expect(err).To(HaveOccurred())

		var destA corev1.Secret
		Expect(k8sClient.Get(ctx, types.NamespacedName{
			Name:      pcreds.DestName(runner.Name, "a"),
			Namespace: runner.Namespace,
		}, &destA)).To(Succeed())
		Expect(state.PluginIssues(plugin)).To(HaveLen(1))
	})

	It("records an Internal issue when the annotation is malformed (defensive)", func() {
		k8sClient = newReconcileClient(scheme)
		plugin = NewPlugin(k8sClient, testRunaiNS)
		runner := testRunner.DeepCopy()
		runner.Annotations[pcreds.Annotation] = "not: a: list"

		err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)
		Expect(err).To(HaveOccurred())

		issues := state.PluginIssues(plugin)
		Expect(issues).To(HaveLen(1))
		Expect(issues[0].Group).To(Equal(issue.Internal))
		Expect(issues[0].IsCritical).To(BeTrue())
	})

	It("records an Internal issue when the source secret has no data", func() {
		empty := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "my-token", Namespace: testRunaiNS},
		}
		k8sClient = newReconcileClient(scheme, empty)
		plugin = NewPlugin(k8sClient, testRunaiNS)

		err := plugin.HandleAdditionalResources(ctx, testRunner.DeepCopy(), nil, nil, k8sClient, state)
		Expect(err).To(HaveOccurred())

		issues := state.PluginIssues(plugin)
		Expect(issues).To(HaveLen(1))
		Expect(issues[0].Group).To(Equal(issue.Internal))
	})

	It("records an Internal issue when source data cannot be decrypted", func() {
		corrupt := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "my-token", Namespace: testRunaiNS},
			Data:       map[string][]byte{"token": []byte("not-valid-aes-gcm-ciphertext")},
		}
		k8sClient = newReconcileClient(scheme, corrupt)
		plugin = NewPlugin(k8sClient, testRunaiNS)

		err := plugin.HandleAdditionalResources(ctx, testRunner.DeepCopy(), nil, nil, k8sClient, state)
		Expect(err).To(HaveOccurred())

		issues := state.PluginIssues(plugin)
		Expect(issues).To(HaveLen(1))
		Expect(issues[0].Group).To(Equal(issue.Internal))
	})
})
