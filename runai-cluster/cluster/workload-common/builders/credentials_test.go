/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"os"

	controllerutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// encryptedValue contains a encrypted value for the secret ("private-value-1").
var encryptedValue = []byte{
	168, 221, 244, 73, 135, 225, 25, 104, 145, 32, 118, 20, 196, 29, 72, 165,
	224, 40, 97, 180, 83, 157, 205, 140, 158, 186, 60, 73, 173, 193, 146, 219,
	201, 199, 65, 36, 241, 251, 52, 131, 38, 155, 0}

var _ = Describe("Credentials builder tests", func() {
	var (
		scheme     *runtime.Scheme
		fakeClient client.Client

		b   *CredentialsBuilder
		ctx context.Context

		workloadBaseName string
		owner            *corev1.Pod
		workloadInfo     WorkloadInfo
	)

	BeforeEach(func() {
		// Setup common test variables
		scheme = setupScheme()
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		b = NewCredentialsBuilder(fakeClient, constants.NamespaceRunAi)
		ctx = workloadproject.SetProjectInContext(context.Background(), defaultNamespace)

		workloadBaseName = "test-workload"
		owner = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: defaultNamespace,
				UID:       "test-uid",
			},
		}

		workloadInfo = WorkloadInfo{
			Name:         workloadBaseName,
			Namespace:    defaultNamespace,
			Initials:     "iw",
			Owner:        owner,
			WorkloadKind: constants.KindInteractiveWorkload,
			JobType:      constants.KindRunaiJob,
		}
		err := os.Setenv(constants.RunaiAdsEncryptionKeyEnvVar, "905c18f0839affcddca117dcac633f79")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err := os.Unsetenv(constants.RunaiAdsEncryptionKeyEnvVar)
		Expect(err).To(BeNil())
	})

	Context("GetPreview and Build", func() {
		It("should return empty result for empty credentials", func() {

			preview, err := b.Preview(ctx, &workloadInfo, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(BeEmpty())

			result, err := b.Build(ctx, &workloadInfo, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeEmpty())
		})

		It("build all credential", func() {
			mockScheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(mockScheme)).To(Succeed())

			existsCredentials := []*corev1.Secret{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-1",
						Namespace: constants.NamespaceRunAi,
					},
					Data: map[string][]byte{
						"key1": encryptedValue,
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-2",
						Namespace: constants.NamespaceRunAi,
					},
					Data: map[string][]byte{
						"key1": encryptedValue,
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-3",
						Namespace: constants.NamespaceRunAi,
					},
					Data: map[string][]byte{
						"key1": encryptedValue,
					},
				},
			}

			mockClient := &MockClient{
				Client: fake.NewClientBuilder().WithScheme(mockScheme).WithObjects(existsCredentials[0]).WithObjects(existsCredentials[1]).Build(),
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					for _, s := range existsCredentials {
						if key.Name == s.Name && key.Namespace == s.Namespace {
							secret := obj.(*corev1.Secret)
							*secret = *s
							return nil
						}
					}
					return apierrors.NewNotFound(schema.GroupResource{Resource: "secrets"}, key.Name)
				},
				scheme: mockScheme,
			}

			b = NewCredentialsBuilder(mockClient, constants.NamespaceRunAi)

			envsValue := map[string]flds.DictionaryItemAndRules{
				"test-1": {
					Value: convutils.Ref("USERCREDENTIAL:test-1,key-1"),
				},
				"test-2": {
					Value: convutils.Ref("USERCREDENTIAL:test-2,key-2"),
				},
			}
			envs := []*flds.DictionaryField{
				{
					Value: &envsValue,
				},
			}

			imagePullSecrets := []*flds.StringField{
				{
					Value: convutils.Ref("USERCREDENTIAL:test-3"),
				},
			}

			preview, err := b.Preview(ctx, &workloadInfo, envs, imagePullSecrets)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(ConsistOf([]Resource{
				{Kind: secretGroupVersion, Name: "iw-test-workload-test-1", Namespace: defaultNamespace},
				{Kind: secretGroupVersion, Name: "iw-test-workload-test-2", Namespace: defaultNamespace},
				{Kind: secretGroupVersion, Name: "iw-test-workload-test-3", Namespace: defaultNamespace},
			}))

			result, err := b.Build(ctx, &workloadInfo, envs, imagePullSecrets)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result[0].Name).To(MatchRegexp("iw-test-workload-test-[1-3]"))
			Expect(result[0].Namespace).To(Equal(defaultNamespace))
			Expect(result[0].Data).To(HaveKeyWithValue("key1", []byte("private-value-1")))
			Expect(result[1].Name).To(MatchRegexp("iw-test-workload-test-[1-3]"))
			Expect(result[1].Namespace).To(Equal(defaultNamespace))
			Expect(result[1].Data).To(HaveKeyWithValue("key1", []byte("private-value-1")))
			Expect(result[2].Name).To(MatchRegexp("iw-test-workload-test-[1-3]"))
			Expect(result[2].Namespace).To(Equal(defaultNamespace))
			Expect(result[2].Data).To(HaveKeyWithValue("key1", []byte("private-value-1")))
		})

		It("build one credential, when there is an one already exists", func() {
			mockScheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(mockScheme)).To(Succeed())

			existsCredentials := []*corev1.Secret{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-1",
						Namespace: constants.NamespaceRunAi,
					},
					Data: map[string][]byte{
						"key1": encryptedValue,
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-2",
						Namespace: constants.NamespaceRunAi,
					},
					Data: map[string][]byte{
						"key1": encryptedValue,
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "iw-test-workload-test-1",
						Namespace: defaultNamespace,
					},
					Data: map[string][]byte{
						"key1": []byte("private-value-1"),
					},
				},
			}

			controllerutils.SetControllerReference(owner, existsCredentials[2], b.k8sClient.Scheme())

			mockClient := &MockClient{
				Client: fake.NewClientBuilder().WithScheme(mockScheme).WithObjects(existsCredentials[0]).WithObjects(existsCredentials[1]).Build(),
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					for _, s := range existsCredentials {
						if key.Name == s.Name && key.Namespace == s.Namespace {
							secret := obj.(*corev1.Secret)
							*secret = *s
							return nil
						}
					}
					return apierrors.NewNotFound(schema.GroupResource{Resource: "secrets"}, key.Name)
				},
				scheme: mockScheme,
			}

			b = NewCredentialsBuilder(mockClient, constants.NamespaceRunAi)

			envsValue := map[string]flds.DictionaryItemAndRules{
				"test-1": {
					Value: convutils.Ref("USERCREDENTIAL:test-1,key-1"),
				},
				"test-2": {
					Value: convutils.Ref("USERCREDENTIAL:test-2,key-2"),
				},
			}
			envs := []*flds.DictionaryField{
				{
					Value: &envsValue,
				},
			}

			preview, err := b.Preview(ctx, &workloadInfo, envs, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(preview).To(ConsistOf([]Resource{
				{Kind: secretGroupVersion, Name: "iw-test-workload-test-1", Namespace: defaultNamespace},
				{Kind: secretGroupVersion, Name: "iw-test-workload-test-2", Namespace: defaultNamespace},
			}))

			result, err := b.Build(ctx, &workloadInfo, envs, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(2))
			Expect(result[0].Name).To(MatchRegexp("iw-test-workload-test-[12]"))
			Expect(result[0].Namespace).To(Equal(defaultNamespace))
			Expect(result[0].Data).To(HaveKeyWithValue("key1", []byte("private-value-1")))
			Expect(result[1].Name).To(MatchRegexp("iw-test-workload-test-[12]"))
			Expect(result[1].Namespace).To(Equal(defaultNamespace))
			Expect(result[1].Data).To(HaveKeyWithValue("key1", []byte("private-value-1")))
		})

		It("missing one source credential", func() {
			mockScheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(mockScheme)).To(Succeed())

			existsCredentials := []*corev1.Secret{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-1",
						Namespace: constants.NamespaceRunAi,
					},
					Data: map[string][]byte{
						"key1": encryptedValue,
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-2",
						Namespace: constants.NamespaceRunAi,
					},
					Data: map[string][]byte{
						"key1": encryptedValue,
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "iw-test-workload-test-1",
						Namespace: defaultNamespace,
					},
					Data: map[string][]byte{
						"key1": []byte("private-value-1"),
					},
				},
			}

			controllerutils.SetControllerReference(owner, existsCredentials[2], b.k8sClient.Scheme())

			mockClient := &MockClient{
				Client: fake.NewClientBuilder().WithScheme(mockScheme).WithObjects(existsCredentials[0]).WithObjects(existsCredentials[1]).Build(),
				MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
					for _, s := range existsCredentials {
						if key.Name == s.Name && key.Namespace == s.Namespace {
							secret := obj.(*corev1.Secret)
							*secret = *s
							return nil
						}
					}
					return apierrors.NewNotFound(schema.GroupResource{Resource: "secrets"}, key.Name)
				},
				scheme: mockScheme,
			}

			b = NewCredentialsBuilder(mockClient, constants.NamespaceRunAi)

			envsValue1 := map[string]flds.DictionaryItemAndRules{
				"test-1": {
					Value: convutils.Ref("USERCREDENTIAL:test-1,key-1"),
				},
				"test-2": {
					Value: convutils.Ref("USERCREDENTIAL:test-2,key-2"),
				},
			}
			envsValue2 := map[string]flds.DictionaryItemAndRules{
				"test-3": {
					Value: convutils.Ref("USERCREDENTIAL:test-3,key-3"),
				},
			}

			envs := []*flds.DictionaryField{
				{
					Value: &envsValue1,
				},
				{
					Value: &envsValue2,
				},
			}

			preview, err := b.Preview(ctx, &workloadInfo, envs, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(preview)).To(Equal(3))
			Expect(preview).To(ConsistOf([]Resource{
				{Kind: secretGroupVersion, Name: "iw-test-workload-test-1", Namespace: defaultNamespace},
				{Kind: secretGroupVersion, Name: "iw-test-workload-test-2", Namespace: defaultNamespace},
				{Kind: secretGroupVersion, Name: "iw-test-workload-test-3", Namespace: defaultNamespace},
			}))

			result, err := b.Build(ctx, &workloadInfo, envs, nil)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

	})
})
