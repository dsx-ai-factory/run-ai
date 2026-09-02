/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/rand"
)

func newDummyAPIPvc() *workloadsManagerApi.WorkloadV2UpdateConfigurationPvc {
	suffix := rand.String(4)
	storageClass := "my-storageClass-" + suffix
	volumeMode := workloadsManagerApi.WORKLOADV2UPDATECONFIGURATIONPVCVOLUMEMODE_FILESYSTEM

	return &workloadsManagerApi.WorkloadV2UpdateConfigurationPvc{
		Metadata: workloadsManagerApi.WorkloadV2UpdateConfigurationPvcMetadata{
			Name: "pvc-" + suffix,
		},
		Spec: workloadsManagerApi.WorkloadV2UpdateConfigurationPvcSpec{
			AccessModes: []workloadsManagerApi.WorkloadV2UpdateConfigurationPvcAccessMode{
				workloadsManagerApi.WORKLOADV2UPDATECONFIGURATIONPVCACCESSMODE_READ_WRITE_ONCE,
				workloadsManagerApi.WORKLOADV2UPDATECONFIGURATIONPVCACCESSMODE_READ_ONLY_MANY,
			},
			Resources: workloadsManagerApi.WorkloadV2UpdateConfigurationPvcResources{
				Requests: workloadsManagerApi.WorkloadV2UpdateConfigurationPvcResourcesRequests{
					Storage: "10Gi",
				},
			},
			StorageClassName: *workloadsManagerApi.NewNullableString(&storageClass),
			VolumeMode:       *workloadsManagerApi.NewNullableWorkloadV2UpdateConfigurationPvcVolumeMode(&volumeMode),
		},
	}
}

var _ = Describe("Storage Configuration Conversion", func() {
	Describe("NewPersistentVolumeClaimFromApiPVC", func() {
		It("converts a fully populated PVC end-to-end", func() {
			labels := map[string]string{"team": "ml"}
			annotations := map[string]string{"note": "primary"}

			apiPvc := newDummyAPIPvc()
			apiPvc.Metadata.Labels = &labels
			apiPvc.Metadata.Annotations = &annotations
			pvc, err := NewPersistentVolumeClaimFromApiPVC(apiPvc)
			Expect(err).NotTo(HaveOccurred())

			Expect(pvc.Name).To(Equal(apiPvc.Metadata.Name))
			Expect(pvc.Labels).To(Equal(labels))
			Expect(pvc.Annotations).To(Equal(annotations))

			Expect(pvc.Spec.AccessModes).To(ConsistOf(
				corev1.ReadWriteOnce,
				corev1.ReadOnlyMany,
			))
			Expect(pvc.Spec.Resources.Requests).To(HaveKey(corev1.ResourceStorage))
			Expect(pvc.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse(apiPvc.Spec.Resources.Requests.Storage)))
			Expect(pvc.Spec.StorageClassName).NotTo(BeNil())
			Expect(*pvc.Spec.StorageClassName).To(Equal(*apiPvc.Spec.StorageClassName.Get()))
			Expect(pvc.Spec.VolumeMode).NotTo(BeNil())
			Expect(*pvc.Spec.VolumeMode).To(Equal(corev1.PersistentVolumeFilesystem))
		})

		It("returns an error when storage quantity is malformed", func() {
			apiPvc := newDummyAPIPvc()
			apiPvc.Spec.Resources.Requests.Storage = "not-a-quantity"
			_, err := NewPersistentVolumeClaimFromApiPVC(apiPvc)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unmarshal as PersistentVolumeClaimSpec"))
		})

		// Drift detector: the API spec is intentionally a structural subset of
		// corev1.PersistentVolumeClaimSpec. If a new API field is introduced
		// whose JSON tag has no counterpart in corev1, this strict-decode test fails
		It("API spec JSON is a strict subset of corev1.PersistentVolumeClaimSpec", func() {
			apiPvc := newDummyAPIPvc()
			data, err := json.Marshal(&apiPvc)
			Expect(err).NotTo(HaveOccurred())

			dec := json.NewDecoder(bytes.NewReader(data))
			dec.DisallowUnknownFields()
			var k8sSpec corev1.PersistentVolumeClaim
			Expect(dec.Decode(&k8sSpec)).To(Succeed())
		})
	})

	Describe("convertStorageConfig", func() {
		It("converts multiple PVCs in order", func() {
			apiPvcs := []workloadsManagerApi.WorkloadV2UpdateConfigurationPvc{*newDummyAPIPvc(), *newDummyAPIPvc()}
			cfg, err := convertStorageConfig(apiPvcs, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.NewPvcs).To(HaveLen(len(apiPvcs)))
			for i, testPvc := range cfg.NewPvcs {
				Expect(testPvc.Name).To(Equal(apiPvcs[i].Metadata.Name))
			}
		})

		It("returns an empty StorageConfig when given no PVCs", func() {
			cfg, err := convertStorageConfig(nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg).NotTo(BeNil())
			Expect(cfg.NewPvcs).To(BeEmpty())
		})

	})
})
