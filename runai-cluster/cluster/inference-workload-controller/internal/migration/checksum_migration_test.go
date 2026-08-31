/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	safeFake "github.com/run-ai/runai/runai-cluster/cluster/common/utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/utils/spec_checksum"
	register "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/utils/testing_utils"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	infw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// patchFuncClient wraps a client.Client and delegates Patch to a custom function,
// allowing tests to inject failures without relying on object name matching.
type patchFuncClient struct {
	client.Client
	patchFn func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error
}

func (c *patchFuncClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return c.patchFn(ctx, obj, patch, opts...)
}

var _ = Describe("ChecksumSHA256Migration", func() {
	var (
		migration *ChecksumSHA256Migration
		k8sClient client.Client
		scheme    *runtime.Scheme
	)

	scheme = runtime.NewScheme()
	register.InitSchema(scheme)

	newClient := func() client.Client {
		return safeFake.NewFakeClientConcurrentSafe(
			fake.NewClientBuilder().WithScheme(scheme).Build(),
		)
	}

	workloadWithMD5 := func(name, md5hex string) *infw.InferenceWorkload {
		workload := getInferenceWorkload(namespace1, name)
		workload.Annotations = map[string]string{
			spec_checksum.SpecChecksumAnnotation: md5hex,
		}
		return workload
	}

	BeforeEach(func() {
		k8sClient = newClient()
		migration = NewChecksumSHA256Migration(k8sClient)
		for _, ns := range createNamespaces() {
			Expect(k8sClient.Create(context.Background(), &ns)).To(Succeed())
		}
	})

	Context("IsLegacyMD5Checksum", func() {
		It("returns true for a valid 32-char lowercase hex string", func() {
			Expect(spec_checksum.IsLegacyMD5Checksum("d41d8cd98f00b204e9800998ecf8427e")).To(BeTrue())
		})

		It("returns false for a valid 64-char SHA-256 hex string", func() {
			Expect(spec_checksum.IsLegacyMD5Checksum("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")).To(BeFalse())
		})

		It("returns false for a 32-char string containing non-hex characters", func() {
			Expect(spec_checksum.IsLegacyMD5Checksum("d41d8cd98f00b204e9800998ecf8427z")).To(BeFalse())
		})

		It("returns false for an empty string", func() {
			Expect(spec_checksum.IsLegacyMD5Checksum("")).To(BeFalse())
		})
	})

	Context("Migrate", func() {
		It("skips when the InferenceWorkload API is not served", func() {
			missingAPIClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
				List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
					return &meta.NoKindMatchError{GroupKind: infw.GroupVersion.WithKind("InferenceWorkload").GroupKind(), SearchedVersions: []string{infw.GroupVersion.Version}}
				},
			}).Build()

			Expect(NewChecksumSHA256Migration(missingAPIClient).Migrate(context.Background())).To(Succeed())
		})

		It("returns authorization errors while listing inference workloads", func() {
			forbiddenErr := apierrors.NewForbidden(schema.GroupResource{Group: infw.GroupVersion.Group, Resource: "inferenceworkloads"}, "", errors.New("denied"))
			forbiddenClient := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
				List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
					return forbiddenErr
				},
			}).Build()

			Expect(NewChecksumSHA256Migration(forbiddenClient).Migrate(context.Background())).To(MatchError(forbiddenErr))
		})

		It("migrates a workload with a valid MD5 checksum to SHA-256", func() {
			workload := workloadWithMD5("wl-md5", "d41d8cd98f00b204e9800998ecf8427e")
			createObject(k8sClient, workload)

			Expect(migration.Migrate(context.Background())).To(Succeed())

			updated := readInferenceWorkload(k8sClient, workload)
			got := spec_checksum.GetSpecChecksumAnnotation(updated)
			Expect(got).To(HaveLen(64))

			expected, err := spec_checksum.CalculateSpecChecksum(updated)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(expected))
		})

		It("skips a workload that already has a SHA-256 checksum", func() {
			workload := getInferenceWorkload(namespace1, "wl-sha256")
			createObject(k8sClient, workload)
			sha256sum, err := spec_checksum.CalculateSpecChecksum(workload)
			Expect(err).NotTo(HaveOccurred())
			workload.Annotations = map[string]string{spec_checksum.SpecChecksumAnnotation: sha256sum}
			Expect(k8sClient.Update(context.Background(), workload)).To(Succeed())

			Expect(migration.Migrate(context.Background())).To(Succeed())

			Expect(spec_checksum.GetSpecChecksumAnnotation(readInferenceWorkload(k8sClient, workload))).To(Equal(sha256sum))
		})

		It("skips a workload with a 32-char non-hex annotation value", func() {
			nonHex := strings.Repeat("z", 32)
			workload := workloadWithMD5("wl-nonhex", nonHex)
			createObject(k8sClient, workload)

			Expect(migration.Migrate(context.Background())).To(Succeed())

			Expect(spec_checksum.GetSpecChecksumAnnotation(readInferenceWorkload(k8sClient, workload))).To(Equal(nonHex))
		})

		It("returns an error when a patch fails and still processes remaining workloads", func() {
			workload1 := workloadWithMD5("wl-a", "d41d8cd98f00b204e9800998ecf8427e")
			workload2 := workloadWithMD5("wl-b", "d41d8cd98f00b204e9800998ecf8427e")
			createObject(k8sClient, workload1)
			createObject(k8sClient, workload2)

			// Fail the first Patch call regardless of which workload is processed first.
			patchCount := 0
			m := NewChecksumSHA256Migration(&patchFuncClient{
				Client: k8sClient,
				patchFn: func(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
					patchCount++
					if patchCount == 1 {
						return errors.New("simulated patch failure")
					}
					return k8sClient.Patch(ctx, obj, patch, opts...)
				},
			})

			err := m.Migrate(context.Background())
			Expect(err).To(HaveOccurred())

			// Exactly one workload must have been migrated (the second Patch succeeded).
			migratedCount := 0
			for _, wl := range []*infw.InferenceWorkload{workload1, workload2} {
				updated := &infw.InferenceWorkload{}
				Expect(k8sClient.Get(context.Background(), client.ObjectKeyFromObject(wl), updated)).To(Succeed())
				if len(spec_checksum.GetSpecChecksumAnnotation(updated)) == 64 {
					migratedCount++
				}
			}
			Expect(migratedCount).To(Equal(1))
		})

		It("is idempotent — second run skips all workloads already on SHA-256", func() {
			workload := workloadWithMD5("wl-idempotent", "d41d8cd98f00b204e9800998ecf8427e")
			createObject(k8sClient, workload)

			Expect(migration.Migrate(context.Background())).To(Succeed())
			firstChecksum := spec_checksum.GetSpecChecksumAnnotation(readInferenceWorkload(k8sClient, workload))
			Expect(firstChecksum).To(HaveLen(64))

			Expect(migration.Migrate(context.Background())).To(Succeed())
			Expect(spec_checksum.GetSpecChecksumAnnotation(readInferenceWorkload(k8sClient, workload))).To(Equal(firstChecksum))
		})

		It("recomputes the checksum from the current spec when the spec changed while the controller was stopped", func() {
			workload := getInferenceWorkload(namespace1, "wl-spec-change")
			createObject(k8sClient, workload)
			workload.Annotations = map[string]string{
				spec_checksum.SpecChecksumAnnotation: "d41d8cd98f00b204e9800998ecf8427e",
			}
			Expect(k8sClient.Update(context.Background(), workload)).To(Succeed())

			workload.Spec.Image = &flds.StringField{Value: ptr.To("ubuntu:22.04")}
			Expect(k8sClient.Update(context.Background(), workload)).To(Succeed())

			Expect(migration.Migrate(context.Background())).To(Succeed())

			updated := readInferenceWorkload(k8sClient, workload)
			got := spec_checksum.GetSpecChecksumAnnotation(updated)
			Expect(got).To(HaveLen(64))

			expected, err := spec_checksum.CalculateSpecChecksum(updated)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(expected))
		})
	})
})
