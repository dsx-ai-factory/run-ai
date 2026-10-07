/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	safefakeclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	pvcBlockName       = "pvc-block"
	pvcBlockName2      = "pvc-block2"
	pvcFileSystemName  = "pvc-filesystem"
	pvcFileSystemName2 = "pvc-filesystem2"
	namespace          = "test-namespace"
)

var _ = Describe("Volumes tests", func() {
	Context("DataVolumes", func() {
		var (
			workloadNamespace       string
			testPvcName             string
			testPvcNameTerminating  string
			dataVolumeId            string
			dataVolumeIdTerminating string
			k8sClient               client.Client
			storageFields           field_groups.StorageFields
			namespacedPvc           *corev1.PersistentVolumeClaim
		)

		dataVolumeId = "5e0445ea-6d67-4641-a791-b4a372ef722c"
		workloadNamespace = "test-namespace"
		testPvcName = "test-pvc"
		testPvcNameTerminating = "test-pvc-terminating"
		dataVolumeIdTerminating = "5e0445ea-6d67-4641-a791-b4a372ef722c-terminating"

		BeforeEach(func() {

			// Initialize the k8s client
			scheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(scheme)).To(Succeed())

			// Initialize the PVC that was created by the DataVolume
			namespacedPvc = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testPvcName,
					Namespace: workloadNamespace,
					Labels: map[string]string{
						v1alpha1.DataVolumeIdLabel: dataVolumeId,
					},
				},
			}

			terminatingPvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testPvcNameTerminating,
					Namespace: workloadNamespace,
					Labels: map[string]string{
						v1alpha1.DataVolumeIdLabel: dataVolumeIdTerminating,
					},
					// instead of adding deletionTimestamp to the object,
					// we add a finalizer to stub the terminating state
					Finalizers: []string{"bla bla bla"},
				},
			}

			k8sClient = initFakeClientWithObjectWithStatus(scheme, nil, namespacedPvc, terminatingPvc)

			err := k8sClient.Delete(context.Background(), terminatingPvc)
			if err != nil {
				return
			}

			// Initialize the storage fields
			storageFields = field_groups.StorageFields{
				DataVolumes: &field_types.DataVolumeField{},
			}
		})

		Context("validate data volumes fields", func() {

			It("should return nil for empty data volumes", func() {
				_, err := GetDataVolumes(context.Background(), &storageFields, nil, k8sClient, workloadNamespace)
				Expect(err).To(BeNil())
			})

			It("should return error for dataVolumes with no Name field equals true", func() {
				storageFields.DataVolumes.Value = &map[string]field_types.DataVolumeItemAndRules{
					"dv1": {
						Value: &field_types.DataVolumeItemValue{
							MountPath: ptr.To("/mount-path"),
						},
					},
				}

				_, err := GetDataVolumes(context.Background(), &storageFields, nil, k8sClient, workloadNamespace)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("DataVolume id is missing"))
			})

			It("should return error for dataVolumes with no MountPath field equals true", func() {
				storageFields.DataVolumes.Value = &map[string]field_types.DataVolumeItemAndRules{
					"dv1": {
						Value: &field_types.DataVolumeItemValue{
							Id: ptr.To("dv1"),
						},
					},
				}

				_, err := GetDataVolumes(context.Background(), &storageFields, nil, k8sClient, workloadNamespace)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("DataVolume mountPath is missing"))
			})

			It("should return error for dataVolumes not found", func() {
				storageFields.DataVolumes.Value = &map[string]field_types.DataVolumeItemAndRules{
					"dv-001": {
						Value: &field_types.DataVolumeItemValue{
							Id:        ptr.To("dv1"),
							MountPath: ptr.To("/mount-path"),
						},
					},
				}

				_, err := GetDataVolumes(context.Background(), &storageFields, nil, k8sClient, workloadNamespace)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("DataVolume dv1 not found in namespace test-namespace"))
			})

			It("should return error pvc terminating", func() {
				storageFields.DataVolumes.Value = &map[string]field_types.DataVolumeItemAndRules{
					"dv-0011": {
						Value: &field_types.DataVolumeItemValue{
							Id:        ptr.To(dataVolumeIdTerminating),
							MountPath: ptr.To("/mount-path"),
						},
					},
				}

				_, err := GetDataVolumes(context.Background(), &storageFields, nil, k8sClient, workloadNamespace)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("DataVolume 5e0445ea-6d67-4641-a791-b4a372ef722c-terminating is terminating"))
			})

			It("should succeed for dataVolumes found", func() {
				storageFields.DataVolumes.Value = &map[string]field_types.DataVolumeItemAndRules{
					"dv-0011": {
						Value: &field_types.DataVolumeItemValue{
							Id:        ptr.To(dataVolumeId),
							MountPath: ptr.To("/mount-path"),
						},
					},
				}

				dataVolumes, err := GetDataVolumes(context.Background(), &storageFields, nil, k8sClient, workloadNamespace)
				Expect(err).To(BeNil())
				Expect(len(dataVolumes)).To(Equal(1))
				Expect(dataVolumes[0].VolumeSource.PersistentVolumeClaim.ClaimName).To(Equal(testPvcName))
			})
		})
	})
	Context("CreateEphemeralVolumes", func() {
		var (
			pvcs         *field_types.PvcsField
			volumeMounts *[]corev1.Volume
			workloadName string
		)

		BeforeEach(func() {
			pvcs = &field_types.PvcsField{}
			volumeMounts = &[]corev1.Volume{}
			workloadName = "test-workload"
		})

		Context("validate ephemeral PVCs fields", func() {
			It("should return nil for empty PVCs value", func() {
				err := CreateEphemeralVolumes(pvcs, volumeMounts, workloadName)
				Expect(err).To(BeNil())
			})

			It("should return nil for PVC with Ephemeral field equals false", func() {
				pvcs.Value = &map[string]field_types.PvcsItemAndRules{
					"pvc1": {
						Value: &field_types.PvcsItemValue{
							Ephemeral: ptr.To(false),
						},
					},
				}
				err := CreateEphemeralVolumes(pvcs, volumeMounts, workloadName)
				Expect(err).To(BeNil())
			})

			It("should return error for PVC with ExistingPvc field equals true", func() {
				pvcs.Value = &map[string]field_types.PvcsItemAndRules{
					"pvc1": {
						Value: &field_types.PvcsItemValue{
							Ephemeral:   ptr.To(true),
							ExistingPvc: ptr.To(true),
						},
					},
				}
				err := CreateEphemeralVolumes(pvcs, volumeMounts, workloadName)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("cannot use existing PVC"))
			})

			It("should return error for PVC with Size invalid field", func() {
				sizes := []string{"-10M", "0G", "0Mi", "10s", "0"}
				for _, size := range sizes {
					pvcs.Value = &map[string]field_types.PvcsItemAndRules{
						"pvc1": {
							Value: &field_types.PvcsItemValue{
								Ephemeral: ptr.To(true),
								Size:      ptr.To(size),
							},
						},
					}
					err := CreateEphemeralVolumes(pvcs, volumeMounts, workloadName)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("invalid size for creating a PVC"))
				}
			})

			It("should return nil for a PVC with Size valid field", func() {
				sizes := []string{"10M", "10G", "10Mi"}
				for _, size := range sizes {
					pvcs.Value = &map[string]field_types.PvcsItemAndRules{
						"pvc1": {
							Value: &field_types.PvcsItemValue{
								Ephemeral: ptr.To(true),
								Size:      ptr.To(size),
							},
						},
					}
					err := CreateEphemeralVolumes(pvcs, volumeMounts, workloadName)
					Expect(err).To(BeNil())
				}
			})

			//todo: add more tests on this function
		})
	})
	Context("IdentifyPvcsVolumeMode", func() {
		var (
			ctx            context.Context
			pvcs           *field_types.PvcsField
			k8sClient      client.Client
			blockPvc1      *corev1.PersistentVolumeClaim
			blockPvc2      *corev1.PersistentVolumeClaim
			fileSystemPvc1 *corev1.PersistentVolumeClaim
			fileSystemPvc2 *corev1.PersistentVolumeClaim
		)

		BeforeEach(func() {
			ctx = context.Background()
			scheme := runtime.NewScheme()
			blockPvc1 = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvcBlockName,
					Namespace: namespace,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeMode: ptr.To(corev1.PersistentVolumeBlock),
				},
			}
			blockPvc2 = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvcBlockName2,
					Namespace: namespace,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeMode: ptr.To(corev1.PersistentVolumeBlock),
				},
			}
			// with explicit file system
			fileSystemPvc1 = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvcFileSystemName,
					Namespace: namespace,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeMode: ptr.To(corev1.PersistentVolumeFilesystem),
				},
			}
			// with default access mode (file system)
			fileSystemPvc2 = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pvcFileSystemName2,
					Namespace: namespace,
				},
			}
			Expect(corev1.AddToScheme(scheme)).To(Succeed())
			k8sClient = initFakeClientWithObjectWithStatus(scheme, nil, blockPvc1, blockPvc2, fileSystemPvc1, fileSystemPvc2)
		})

		It("should return nil when pvcs is nil", func() {
			pvcs = nil
			err := IdentifyPvcsVolumeMode(ctx, pvcs, namespace, k8sClient)
			Expect(err).To(BeNil())
		})

		It("should continue when pvc is deleted", func() {
			pvcs = &field_types.PvcsField{
				Value: &map[string]field_types.PvcsItemAndRules{
					"pvc1": {
						Deleted: ptr.To(true),
					},
				},
			}
			err := IdentifyPvcsVolumeMode(ctx, pvcs, namespace, k8sClient)
			Expect(err).To(BeNil())
		})

		It("should return error when pvc does not exist", func() {
			pvcs = &field_types.PvcsField{
				Value: &map[string]field_types.PvcsItemAndRules{
					"pvc1": {
						Value: &field_types.PvcsItemValue{
							ExistingPvc: ptr.To(true),
							ClaimName:   ptr.To("non-existing-pvc"),
						},
					},
				},
			}
			err := IdentifyPvcsVolumeMode(ctx, pvcs, namespace, k8sClient)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("non-existing-pvc"))
		})

		It("should not return error when pvc empty", func() {
			pvcs = &field_types.PvcsField{}
			err := IdentifyPvcsVolumeMode(ctx, pvcs, namespace, k8sClient)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error when claimName is empty", func() {
			pvcs = &field_types.PvcsField{
				Value: &map[string]field_types.PvcsItemAndRules{
					"pvc1": {
						Value: &field_types.PvcsItemValue{
							ExistingPvc: ptr.To(true),
						},
					},
				},
			}
			err := IdentifyPvcsVolumeMode(ctx, pvcs, namespace, k8sClient)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("claimName is mandatory when binding to existing PVC"))
		})

		It("should copy the VolumeMode from the claim to our list", func() {
			pvcs = &field_types.PvcsField{
				Value: &map[string]field_types.PvcsItemAndRules{
					"pvc1": {
						Value: &field_types.PvcsItemValue{
							ExistingPvc: ptr.To(true),
							ClaimName:   ptr.To(pvcBlockName),
						},
					},
					"pvc2": {
						Value: &field_types.PvcsItemValue{
							ExistingPvc: ptr.To(true),
							ClaimName:   ptr.To(pvcFileSystemName),
						},
					},
					"pvc3": {
						Value: &field_types.PvcsItemValue{
							ExistingPvc: ptr.To(false),
						},
					},
					"pvc4": {
						Value: &field_types.PvcsItemValue{
							ExistingPvc: ptr.To(true),
							ClaimName:   ptr.To(pvcFileSystemName2),
						},
					},
					"pvc5": {
						Value: &field_types.PvcsItemValue{
							ExistingPvc: ptr.To(true),
							ClaimName:   ptr.To(pvcBlockName2),
						},
					},
				},
			}
			//before change all volume modes are nil
			Expect((*pvcs.Value)["pvc1"].Value.VolumeMode).To(BeNil())
			Expect((*pvcs.Value)["pvc2"].Value.VolumeMode).To(BeNil())
			Expect((*pvcs.Value)["pvc3"].Value.VolumeMode).To(BeNil())
			Expect((*pvcs.Value)["pvc4"].Value.VolumeMode).To(BeNil())
			Expect((*pvcs.Value)["pvc5"].Value.VolumeMode).To(BeNil())
			err := IdentifyPvcsVolumeMode(ctx, pvcs, namespace, k8sClient)
			Expect(err).To(BeNil())
			Expect((*pvcs.Value)["pvc1"].Value.VolumeMode).To(Equal(ptr.To(string(corev1.PersistentVolumeBlock))))
			Expect((*pvcs.Value)["pvc2"].Value.VolumeMode).To(Equal(ptr.To(string(corev1.PersistentVolumeFilesystem))))
			Expect((*pvcs.Value)["pvc3"].Value.VolumeMode).To(BeNil())
			Expect((*pvcs.Value)["pvc4"].Value.VolumeMode).To(BeNil())
			Expect((*pvcs.Value)["pvc5"].Value.VolumeMode).To(Equal(ptr.To(string(corev1.PersistentVolumeBlock))))
		})
	})

	Context("PVC Validations", func() {
		var (
			storageFields field_groups.StorageFields
		)

		Context("When creating a new PVC", func() {
			It("should require size field", func() {
				storageFields = field_groups.StorageFields{
					Pvcs: &field_types.PvcsField{
						Value: &map[string]field_types.PvcsItemAndRules{
							"pvc-1": {
								Value: &field_types.PvcsItemValue{
									StorageClass: ptr.To("sc-1"),
								},
							},
						},
					},
				}

				_, err := GetPvcVolumes(context.Background(), "missing-size", "", &storageFields)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("size is mandatory for creating a PVC"))
			})

			It("should use the provided suffix", func() {
				storageFields = field_groups.StorageFields{
					Pvcs: &field_types.PvcsField{
						Value: &map[string]field_types.PvcsItemAndRules{
							"pvc-1": {
								Value: &field_types.PvcsItemValue{
									Size: ptr.To("10M"),
								},
							},
						},
					},
				}

				result, err := GetPvcVolumes(context.Background(), "workload-1", "suffix", &storageFields)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(result)).To(Equal(len(storageFields.Pvcs.Deref())))
				Expect(result[0].PersistentVolumeClaim.ClaimName).To(Equal("pvc-workload-1-suffix-0"))
			})
		})

		Context("When binding to an existing PVC", func() {
			It("should reject redundant size field", func() {
				storageFields = field_groups.StorageFields{
					Pvcs: &field_types.PvcsField{
						Value: &map[string]field_types.PvcsItemAndRules{
							"pvc-1": {
								Value: &field_types.PvcsItemValue{
									ClaimName:   ptr.To("claim1"),
									Size:        ptr.To("10M"),
									ExistingPvc: ptr.To(true),
								},
							},
						},
					},
				}

				_, err := GetPvcVolumes(context.Background(), "missing-size", "", &storageFields)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("size is redundant when binding to"))
			})

			It("should verify and use the provided claim name", func() {
				storageFields = field_groups.StorageFields{
					Pvcs: &field_types.PvcsField{
						Value: &map[string]field_types.PvcsItemAndRules{
							"pvc-1": {
								Value: &field_types.PvcsItemValue{
									ClaimName:   ptr.To("claim1"),
									ExistingPvc: ptr.To(true),
								},
							},
						},
					},
				}

				result, err := GetPvcVolumes(context.Background(), "workload-1", "", &storageFields)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(result)).To(Equal(len(storageFields.Pvcs.Deref())))
				Expect(result[0].PersistentVolumeClaim.ClaimName).To(Equal("claim1"))
			})

			It("should require claim name field", func() {
				storageFields = field_groups.StorageFields{
					Pvcs: &field_types.PvcsField{
						Value: &map[string]field_types.PvcsItemAndRules{
							"pvc-1": {
								Value: &field_types.PvcsItemValue{
									ExistingPvc: ptr.To(true),
								},
							},
						},
					},
				}

				_, err := GetPvcVolumes(context.Background(), "workload-1", "", &storageFields)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("is mandatory"))
			})
		})

		Context("When using a mixture of multiple PVCs of different types", func() {
			It("should use correct claim name and index", func() {
				// important notes:
				// * pvc order is important - in order to keep the map sorted we use the prefixes pvc1, pvc2, ...
				// * we intentionally test a case where a new pvc comes AFTER an existing one, in order to check
				//   the indexing, as taken care in skipPvcBuilding function
				storageFields = field_groups.StorageFields{
					Pvcs: &field_types.PvcsField{
						Value: &map[string]field_types.PvcsItemAndRules{
							"pvc1-new": {
								Value: &field_types.PvcsItemValue{
									Size: ptr.To("10M"),
								},
							},
							"pvc2-new": {
								Value: &field_types.PvcsItemValue{
									Size: ptr.To("20M"),
								},
							},
							"pvc3-ephemeral": {
								Value: &field_types.PvcsItemValue{
									Size:      ptr.To("30M"),
									Ephemeral: ptr.To(true),
								},
							},
							"pvc4-existing": {
								Value: &field_types.PvcsItemValue{
									ClaimName:   ptr.To("existing-pvc"),
									ExistingPvc: ptr.To(true),
								},
							},
							"pvc5-new": {
								Value: &field_types.PvcsItemValue{
									Size: ptr.To("40M"),
								},
							},
						},
					},
				}

				result, err := GetPvcVolumes(context.Background(), "workload-1", "", &storageFields)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(result)).To(Equal(4)) // new and existing PVCs, not ephemeral
				Expect(result[0].PersistentVolumeClaim.ClaimName).To(Equal("pvc-workload-1-0"))
				Expect(result[1].PersistentVolumeClaim.ClaimName).To(Equal("pvc-workload-1-1"))
				Expect(result[2].PersistentVolumeClaim.ClaimName).To(Equal("existing-pvc"))
				Expect(result[3].PersistentVolumeClaim.ClaimName).To(Equal("pvc-workload-1-3"))
			})
		})
	})

	Context("GetS3VolumeMounts tests", func() {
		var (
			storageFields *field_groups.StorageFields
		)

		BeforeEach(func() {
			storageFields = &field_groups.StorageFields{
				S3buckets: &field_types.S3bucketField{},
			}
		})

		It("should return empty slice for nil S3buckets", func() {
			storageFields.S3buckets = nil
			mounts := GetS3VolumeMounts(context.Background(), storageFields)
			Expect(mounts).To(BeEmpty())
		})

		It("should return empty slice for empty S3buckets", func() {
			storageFields.S3buckets.Value = &map[string]field_types.S3bucketItemAndRules{}
			mounts := GetS3VolumeMounts(context.Background(), storageFields)
			Expect(mounts).To(BeEmpty())
		})

		It("should skip nil S3bucket values", func() {
			storageFields.S3buckets.Value = &map[string]field_types.S3bucketItemAndRules{
				"s3-1": {
					Value: nil,
				},
			}
			mounts := GetS3VolumeMounts(context.Background(), storageFields)
			Expect(mounts).To(BeEmpty())
		})

		It("should create volume mount for single S3bucket", func() {
			storageFields.S3buckets.Value = &map[string]field_types.S3bucketItemAndRules{
				"s3-1": {
					Value: &field_types.S3bucketItemValue{
						Path: ptr.To("/mnt/s3-bucket"),
					},
				},
			}
			mounts := GetS3VolumeMounts(context.Background(), storageFields)
			Expect(len(mounts)).To(Equal(1))
			Expect(mounts[0].Name).To(Equal("s3-0"))
			Expect(mounts[0].MountPath).To(Equal("/mnt/s3-bucket"))
			mountPropagation := corev1.MountPropagationHostToContainer
			Expect(mounts[0].MountPropagation).To(Equal(&mountPropagation))
		})

		It("should create volume mounts for multiple S3buckets", func() {
			storageFields.S3buckets.Value = &map[string]field_types.S3bucketItemAndRules{
				"s3-1": {
					Value: &field_types.S3bucketItemValue{
						Path: ptr.To("/mnt/s3-bucket-1"),
					},
				},
				"s3-2": {
					Value: &field_types.S3bucketItemValue{
						Path: ptr.To("/mnt/s3-bucket-2"),
					},
				},
			}
			mounts := GetS3VolumeMounts(context.Background(), storageFields)
			Expect(len(mounts)).To(Equal(2))
			Expect(mounts[0].Name).To(Equal("s3-0"))
			Expect(mounts[0].MountPath).To(Equal("/mnt/s3-bucket-1"))
			Expect(mounts[1].Name).To(Equal("s3-1"))
			Expect(mounts[1].MountPath).To(Equal("/mnt/s3-bucket-2"))
		})
	})

	Context("GetS3Volumes tests", func() {
		var (
			storageFields *field_groups.StorageFields
		)

		BeforeEach(func() {
			storageFields = &field_groups.StorageFields{
				S3buckets: &field_types.S3bucketField{},
			}
		})

		It("should return empty slice for nil S3buckets", func() {
			storageFields.S3buckets = nil
			volumes := GetS3Volumes(context.Background(), storageFields)
			Expect(volumes).To(BeEmpty())
		})

		It("should return empty slice for empty S3buckets", func() {
			storageFields.S3buckets.Value = &map[string]field_types.S3bucketItemAndRules{}
			volumes := GetS3Volumes(context.Background(), storageFields)
			Expect(volumes).To(BeEmpty())
		})

		It("should skip nil S3bucket values", func() {
			storageFields.S3buckets.Value = &map[string]field_types.S3bucketItemAndRules{
				"s3-1": {
					Value: nil,
				},
			}
			volumes := GetS3Volumes(context.Background(), storageFields)
			Expect(volumes).To(BeEmpty())
		})

		It("should create volume for single S3bucket", func() {
			storageFields.S3buckets.Value = &map[string]field_types.S3bucketItemAndRules{
				"s3-1": {
					Value: &field_types.S3bucketItemValue{
						Bucket: ptr.To("my-bucket"),
					},
				},
			}
			volumes := GetS3Volumes(context.Background(), storageFields)
			Expect(len(volumes)).To(Equal(1))
			Expect(volumes[0].Name).To(Equal("s3-0"))
			Expect(volumes[0].VolumeSource.EmptyDir).ToNot(BeNil())
		})

		It("should create volumes for multiple S3buckets", func() {
			storageFields.S3buckets.Value = &map[string]field_types.S3bucketItemAndRules{
				"s3-1": {
					Value: &field_types.S3bucketItemValue{
						Bucket: ptr.To("my-bucket-1"),
					},
				},
				"s3-2": {
					Value: &field_types.S3bucketItemValue{
						Bucket: ptr.To("my-bucket-2"),
					},
				},
			}
			volumes := GetS3Volumes(context.Background(), storageFields)
			Expect(len(volumes)).To(Equal(2))
			Expect(volumes[0].Name).To(Equal("s3-0"))
			Expect(volumes[1].Name).To(Equal("s3-1"))
		})
	})

	Context("GetHostPathVolumes tests", func() {
		var (
			storageFields *field_groups.StorageFields
		)

		BeforeEach(func() {
			storageFields = &field_groups.StorageFields{
				Volumes: &field_types.VolumesField{},
			}
		})

		It("should return empty slice for nil volumes", func() {
			storageFields.Volumes = nil
			volumes := GetHostPathVolumes(context.Background(), storageFields)
			Expect(volumes).To(BeEmpty())
		})

		It("should return empty slice for empty volumes", func() {
			storageFields.Volumes.Value = &map[string]field_types.VolumeItemAndRules{}
			volumes := GetHostPathVolumes(context.Background(), storageFields)
			Expect(volumes).To(BeEmpty())
		})

		It("should skip nil volume values", func() {
			storageFields.Volumes.Value = &map[string]field_types.VolumeItemAndRules{
				"vol-1": {
					Value: nil,
				},
			}
			volumes := GetHostPathVolumes(context.Background(), storageFields)
			Expect(volumes).To(BeEmpty())
		})

		It("should create hostPath volume for regular path", func() {
			storageFields.Volumes.Value = &map[string]field_types.VolumeItemAndRules{
				"vol-1": {
					Value: &field_types.VolumeItemValue{
						SourcePath: ptr.To("/host/path"),
					},
				},
			}
			volumes := GetHostPathVolumes(context.Background(), storageFields)
			Expect(len(volumes)).To(Equal(1))
			Expect(volumes[0].Name).To(Equal("host-path-volume-0"))
			Expect(volumes[0].VolumeSource.HostPath).ToNot(BeNil())
			Expect(volumes[0].VolumeSource.HostPath.Path).To(Equal("/host/path"))
			Expect(volumes[0].VolumeSource.NFS).To(BeNil())
		})

		It("should create NFS volume for backward compatibility when NfsServer is set", func() {
			storageFields.Volumes.Value = &map[string]field_types.VolumeItemAndRules{
				"vol-1": {
					Value: &field_types.VolumeItemValue{
						SourcePath: ptr.To("/nfs/path"),
						NfsServer:  ptr.To("nfs.example.com"),
					},
				},
			}
			volumes := GetHostPathVolumes(context.Background(), storageFields)
			Expect(len(volumes)).To(Equal(1))
			Expect(volumes[0].Name).To(Equal("host-path-volume-0"))
			Expect(volumes[0].VolumeSource.NFS).ToNot(BeNil())
			Expect(volumes[0].VolumeSource.NFS.Server).To(Equal("nfs.example.com"))
			Expect(volumes[0].VolumeSource.NFS.Path).To(Equal("/nfs/path"))
			Expect(volumes[0].VolumeSource.HostPath).To(BeNil())
		})

		It("should create multiple volumes with correct indexing", func() {
			storageFields.Volumes.Value = &map[string]field_types.VolumeItemAndRules{
				"vol-1": {
					Value: &field_types.VolumeItemValue{
						SourcePath: ptr.To("/host/path1"),
					},
				},
				"vol-2": {
					Value: &field_types.VolumeItemValue{
						SourcePath: ptr.To("/host/path2"),
						NfsServer:  ptr.To("nfs.example.com"),
					},
				},
			}
			volumes := GetHostPathVolumes(context.Background(), storageFields)
			Expect(len(volumes)).To(Equal(2))
			Expect(volumes[0].Name).To(Equal("host-path-volume-0"))
			Expect(volumes[1].Name).To(Equal("host-path-volume-1"))
		})
	})

	Context("GetNfsVolumes tests", func() {
		var (
			storageFields *field_groups.StorageFields
		)

		BeforeEach(func() {
			storageFields = &field_groups.StorageFields{
				Nfs: &field_types.NfsField{},
			}
		})

		It("should return empty slice for nil NFS volumes", func() {
			storageFields.Nfs = nil
			volumes := GetNfsVolumes(context.Background(), storageFields)
			Expect(volumes).To(BeEmpty())
		})

		It("should return empty slice for empty NFS volumes", func() {
			storageFields.Nfs.Value = &map[string]field_types.NfsItemAndRules{}
			volumes := GetNfsVolumes(context.Background(), storageFields)
			Expect(volumes).To(BeEmpty())
		})

		It("should skip nil NFS volume values", func() {
			storageFields.Nfs.Value = &map[string]field_types.NfsItemAndRules{
				"nfs-1": {
					Value: nil,
				},
			}
			volumes := GetNfsVolumes(context.Background(), storageFields)
			Expect(volumes).To(BeEmpty())
		})

		It("should create NFS volume", func() {
			storageFields.Nfs.Value = &map[string]field_types.NfsItemAndRules{
				"nfs-1": {
					Value: &field_types.NfsItemValue{
						NfsServer:  ptr.To("nfs.example.com"),
						SourcePath: ptr.To("/nfs/share"),
					},
				},
			}
			volumes := GetNfsVolumes(context.Background(), storageFields)
			Expect(len(volumes)).To(Equal(1))
			Expect(volumes[0].Name).To(Equal("nfs-volume-0"))
			Expect(volumes[0].VolumeSource.NFS).ToNot(BeNil())
			Expect(volumes[0].VolumeSource.NFS.Server).To(Equal("nfs.example.com"))
			Expect(volumes[0].VolumeSource.NFS.Path).To(Equal("/nfs/share"))
		})

		It("should create multiple NFS volumes with correct indexing", func() {
			storageFields.Nfs.Value = &map[string]field_types.NfsItemAndRules{
				"nfs-1": {
					Value: &field_types.NfsItemValue{
						NfsServer:  ptr.To("nfs1.example.com"),
						SourcePath: ptr.To("/nfs/share1"),
					},
				},
				"nfs-2": {
					Value: &field_types.NfsItemValue{
						NfsServer:  ptr.To("nfs2.example.com"),
						SourcePath: ptr.To("/nfs/share2"),
					},
				},
			}
			volumes := GetNfsVolumes(context.Background(), storageFields)
			Expect(len(volumes)).To(Equal(2))
			Expect(volumes[0].Name).To(Equal("nfs-volume-0"))
			Expect(volumes[1].Name).To(Equal("nfs-volume-1"))
		})
	})

	Context("GetConfigMapVolumes tests", func() {
		var (
			configMapVolumes *field_types.ConfigMapVolumeField
		)

		BeforeEach(func() {
			configMapVolumes = &field_types.ConfigMapVolumeField{}
		})

		It("should return empty slice for nil configMapVolumes", func() {
			volumes := GetConfigMapVolumes(context.Background(), nil)
			Expect(volumes).To(BeEmpty())
		})

		It("should return empty slice for empty configMapVolumes", func() {
			configMapVolumes.Value = &map[string]field_types.ConfigMapVolumeItemAndRules{}
			volumes := GetConfigMapVolumes(context.Background(), configMapVolumes)
			Expect(volumes).To(BeEmpty())
		})

		It("should create ConfigMap volume", func() {
			configMapVolumes.Value = &map[string]field_types.ConfigMapVolumeItemAndRules{
				"cm-1": {
					Value: &field_types.ConfigMapVolumeItemValue{
						Name:      ptr.To("my-config"),
						MountPath: ptr.To("/etc/config"),
					},
				},
			}
			volumes := GetConfigMapVolumes(context.Background(), configMapVolumes)
			Expect(len(volumes)).To(Equal(1))
			Expect(volumes[0].Name).To(Equal("cm-volume-0"))
			Expect(volumes[0].VolumeSource.ConfigMap).ToNot(BeNil())
			Expect(volumes[0].VolumeSource.ConfigMap.LocalObjectReference.Name).To(Equal("my-config"))
		})

		It("should create multiple ConfigMap volumes with correct indexing", func() {
			configMapVolumes.Value = &map[string]field_types.ConfigMapVolumeItemAndRules{
				"cm-1": {
					Value: &field_types.ConfigMapVolumeItemValue{
						Name:      ptr.To("config-1"),
						MountPath: ptr.To("/etc/config1"),
					},
				},
				"cm-2": {
					Value: &field_types.ConfigMapVolumeItemValue{
						Name:      ptr.To("config-2"),
						MountPath: ptr.To("/etc/config2"),
					},
				},
			}
			volumes := GetConfigMapVolumes(context.Background(), configMapVolumes)
			Expect(len(volumes)).To(Equal(2))
			Expect(volumes[0].Name).To(Equal("cm-volume-0"))
			Expect(volumes[1].Name).To(Equal("cm-volume-1"))
		})
	})

	Context("GetSecretVolumes tests", func() {
		var (
			secretVolumes *field_types.SecretVolumeField
		)

		BeforeEach(func() {
			secretVolumes = &field_types.SecretVolumeField{}
		})

		It("should return empty slice for nil secretVolumes", func() {
			volumes := GetSecretVolumes(context.Background(), nil)
			Expect(volumes).To(BeEmpty())
		})

		It("should return empty slice for empty secretVolumes", func() {
			secretVolumes.Value = &map[string]field_types.SecretVolumeItemAndRules{}
			volumes := GetSecretVolumes(context.Background(), secretVolumes)
			Expect(volumes).To(BeEmpty())
		})

		It("should create Secret volume", func() {
			secretVolumes.Value = &map[string]field_types.SecretVolumeItemAndRules{
				"secret-1": {
					Value: &field_types.SecretVolumeItemValue{
						Name:      ptr.To("my-secret"),
						MountPath: ptr.To("/etc/secret"),
					},
				},
			}
			volumes := GetSecretVolumes(context.Background(), secretVolumes)
			Expect(len(volumes)).To(Equal(1))
			Expect(volumes[0].Name).To(Equal("secret-volume-0"))
			Expect(volumes[0].VolumeSource.Secret).ToNot(BeNil())
			Expect(volumes[0].VolumeSource.Secret.SecretName).To(Equal("my-secret"))
		})

		It("should create multiple Secret volumes with correct indexing", func() {
			secretVolumes.Value = &map[string]field_types.SecretVolumeItemAndRules{
				"secret-1": {
					Value: &field_types.SecretVolumeItemValue{
						Name:      ptr.To("secret-1"),
						MountPath: ptr.To("/etc/secret1"),
					},
				},
				"secret-2": {
					Value: &field_types.SecretVolumeItemValue{
						Name:      ptr.To("secret-2"),
						MountPath: ptr.To("/etc/secret2"),
					},
				},
			}
			volumes := GetSecretVolumes(context.Background(), secretVolumes)
			Expect(len(volumes)).To(Equal(2))
			Expect(volumes[0].Name).To(Equal("secret-volume-0"))
			Expect(volumes[1].Name).To(Equal("secret-volume-1"))
		})
	})

	Context("NeedToSetupRunaiHome tests", func() {
		var (
			containerFields *field_groups.ContainerFields
		)

		BeforeEach(func() {
			containerFields = &field_groups.ContainerFields{
				CreateHomeDir: &field_types.BoolField{},
				RunAsUser:     &field_types.BoolField{},
				Environment:   &field_types.DictionaryField{},
			}
		})

		It("should return false for nil containerFields", func() {
			result := NeedToSetupRunaiHome(nil, false)
			Expect(result).To(BeFalse())
		})

		It("should return false for Jupyter workloads", func() {
			result := NeedToSetupRunaiHome(containerFields, true)
			Expect(result).To(BeFalse())
		})

		It("should return false when neither CreateHomeDir nor RunAsUser is set", func() {
			result := NeedToSetupRunaiHome(containerFields, false)
			Expect(result).To(BeFalse())
		})

		It("should return true when CreateHomeDir is true and no HOME environment", func() {
			containerFields.CreateHomeDir.Value = ptr.To(true)
			result := NeedToSetupRunaiHome(containerFields, false)
			Expect(result).To(BeTrue())
		})

		It("should return false when CreateHomeDir is false and RunAsUser is true", func() {
			containerFields.CreateHomeDir.Value = ptr.To(false)
			containerFields.RunAsUser.Value = ptr.To(true)
			result := NeedToSetupRunaiHome(containerFields, false)
			Expect(result).To(BeFalse())
		})

		It("should return true when CreateHomeDir is unset and RunAsUser is true", func() {
			containerFields.RunAsUser.Value = ptr.To(true)
			result := NeedToSetupRunaiHome(containerFields, false)
			Expect(result).To(BeTrue())
		})

		It("should return true when CreateHomeDir is true and HOME is runai default", func() {
			containerFields.CreateHomeDir.Value = ptr.To(true)
			containerFields.Environment.Value = &map[string]field_types.DictionaryItemAndRules{
				constants.HomeDirEnvName: {
					Value: ptr.To(constants.RunaiDefaultHomeDir),
				},
			}
			result := NeedToSetupRunaiHome(containerFields, false)
			Expect(result).To(BeTrue())
		})

		It("should return false when CreateHomeDir is true but HOME is not runai default", func() {
			containerFields.CreateHomeDir.Value = ptr.To(true)
			containerFields.Environment.Value = &map[string]field_types.DictionaryItemAndRules{
				constants.HomeDirEnvName: {
					Value: ptr.To("/custom/home"),
				},
			}
			result := NeedToSetupRunaiHome(containerFields, false)
			Expect(result).To(BeFalse())
		})
	})

	Context("GetVolumeDevices tests", func() {
		var (
			storageFields *field_groups.StorageFields
		)

		BeforeEach(func() {
			storageFields = &field_groups.StorageFields{
				Pvcs: &field_types.PvcsField{},
			}
		})

		It("should return empty slice for nil PVCs", func() {
			storageFields.Pvcs = nil
			devices := GetVolumeDevices(context.Background(), storageFields)
			Expect(devices).To(BeEmpty())
		})

		It("should return empty slice for empty PVCs", func() {
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{}
			devices := GetVolumeDevices(context.Background(), storageFields)
			Expect(devices).To(BeEmpty())
		})

		It("should skip nil PVC values", func() {
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
				"pvc-1": {
					Value: nil,
				},
			}
			devices := GetVolumeDevices(context.Background(), storageFields)
			Expect(devices).To(BeEmpty())
		})

		It("should skip ephemeral PVCs", func() {
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
				"pvc-1": {
					Value: &field_types.PvcsItemValue{
						Ephemeral: ptr.To(true),
						Path:      ptr.To("/dev/xvda"),
					},
				},
			}
			devices := GetVolumeDevices(context.Background(), storageFields)
			Expect(devices).To(BeEmpty())
		})

		It("should skip filesystem PVCs", func() {
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
				"pvc-1": {
					Value: &field_types.PvcsItemValue{
						VolumeMode: ptr.To(string(corev1.PersistentVolumeFilesystem)),
						Path:       ptr.To("/dev/xvda"),
					},
				},
			}
			devices := GetVolumeDevices(context.Background(), storageFields)
			Expect(devices).To(BeEmpty())
		})

		It("should create volume device for block PVC", func() {
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
				"pvc-1": {
					Value: &field_types.PvcsItemValue{
						VolumeMode: ptr.To(string(corev1.PersistentVolumeBlock)),
						Path:       ptr.To("/dev/xvda"),
					},
				},
			}
			devices := GetVolumeDevices(context.Background(), storageFields)
			Expect(len(devices)).To(Equal(1))
			Expect(devices[0].Name).To(Equal("pvc-volume-0"))
			Expect(devices[0].DevicePath).To(Equal("/dev/xvda"))
		})

		It("should create multiple volume devices with correct indexing", func() {
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
				"pvc-1": {
					Value: &field_types.PvcsItemValue{
						VolumeMode: ptr.To(string(corev1.PersistentVolumeBlock)),
						Path:       ptr.To("/dev/xvda"),
					},
				},
				"pvc-2": {
					Value: &field_types.PvcsItemValue{
						VolumeMode: ptr.To(string(corev1.PersistentVolumeBlock)),
						Path:       ptr.To("/dev/xvdb"),
					},
				},
			}
			devices := GetVolumeDevices(context.Background(), storageFields)
			Expect(len(devices)).To(Equal(2))
			Expect(devices[0].Name).To(Equal("pvc-volume-0"))
			Expect(devices[1].Name).To(Equal("pvc-volume-1"))
		})
	})

	Context("GetVolumes tests", func() {
		var (
			storageFields    *field_groups.StorageFields
			containerFields  *field_groups.ContainerFields
			networkingFields *field_groups.NetworkingFields
			k8sClient        client.Client
			namespace        string
		)

		BeforeEach(func() {
			// Initialize the k8s client
			scheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(scheme)).To(Succeed())
			k8sClient = initFakeClientWithObjectWithStatus(scheme, nil)

			namespace = "test-namespace"
			storageFields = &field_groups.StorageFields{
				Pvcs:             &field_types.PvcsField{},
				DataVolumes:      &field_types.DataVolumeField{},
				Volumes:          &field_types.VolumesField{},
				Nfs:              &field_types.NfsField{},
				S3buckets:        &field_types.S3bucketField{},
				LargeShm:         &field_types.BoolField{},
				ConfigMapVolumes: &field_types.ConfigMapVolumeField{},
				SecretVolumes:    &field_types.SecretVolumeField{},
				GitSync:          &field_types.GitSyncField{},
			}
			containerFields = &field_groups.ContainerFields{
				CreateHomeDir: &field_types.BoolField{},
				RunAsUser:     &field_types.BoolField{},
				Environment:   &field_types.DictionaryField{},
			}
			networkingFields = &field_groups.NetworkingFields{
				ExposedUrls: &field_types.ExposedUrlsField{},
			}
		})

		It("should return empty slice for empty storage fields", func() {
			volumes, err := GetVolumes(context.Background(), storageFields, nil, networkingFields, k8sClient, namespace, "workload", "", false, true, false)
			Expect(err).To(BeNil())
			Expect(volumes).To(BeEmpty())
		})

		It("should return error when S3 is not supported but S3 volumes exist", func() {
			storageFields.S3buckets.Value = &map[string]field_types.S3bucketItemAndRules{
				"s3-1": {
					Value: &field_types.S3bucketItemValue{
						Bucket: ptr.To("my-bucket"),
					},
				},
			}
			_, err := GetVolumes(context.Background(), storageFields, nil, networkingFields, k8sClient, namespace, "workload", "", false, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("this workload type does not support s3 volumes"))
		})

		It("should add large shm volume when requested", func() {
			storageFields.LargeShm.Value = ptr.To(true)
			volumes, err := GetVolumes(context.Background(), storageFields, nil, networkingFields, k8sClient, namespace, "workload", "", false, true, false)
			Expect(err).To(BeNil())
			Expect(len(volumes)).To(Equal(1))
			Expect(volumes[0].Name).To(Equal(dshmVolumeName))
			Expect(volumes[0].VolumeSource.EmptyDir.Medium).To(Equal(corev1.StorageMediumMemory))
		})

		It("should add home directory volume when needed", func() {
			containerFields.CreateHomeDir.Value = ptr.To(true)
			volumes, err := GetVolumes(context.Background(), storageFields, containerFields, networkingFields, k8sClient, namespace, "workload", "", false, true, false)
			Expect(err).To(BeNil())
			Expect(len(volumes)).To(Equal(1))
			Expect(volumes[0].Name).To(Equal(homeVolumeName))
			Expect(volumes[0].VolumeSource.EmptyDir).ToNot(BeNil())
		})

		It("should add git-sync volumes", func() {
			storageFields.GitSync.Value = &map[string]field_types.GitSyncItemAndRules{
				"git-1": {
					Value: &field_types.GitSyncItemValue{
						Repository: ptr.To("https://github.com/example/repo.git"),
						Directory:  ptr.To("/custom/code"),
					},
				},
			}
			volumes, err := GetVolumes(context.Background(), storageFields, nil, networkingFields, k8sClient, namespace, "workload", "", false, true, false)
			Expect(err).To(BeNil())
			Expect(len(volumes)).To(Equal(1))
			Expect(volumes[0].Name).To(Equal("code-sync-0"))
			Expect(volumes[0].VolumeSource.EmptyDir).ToNot(BeNil())
		})

		It("should combine multiple volume types", func() {
			storageFields.LargeShm.Value = ptr.To(true)
			containerFields.CreateHomeDir.Value = ptr.To(true)
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
				"pvc-1": {
					Value: &field_types.PvcsItemValue{
						Size: ptr.To("10Gi"),
					},
				},
			}
			storageFields.ConfigMapVolumes.Value = &map[string]field_types.ConfigMapVolumeItemAndRules{
				"cm-1": {
					Value: &field_types.ConfigMapVolumeItemValue{
						Name:      ptr.To("my-config"),
						MountPath: ptr.To("/etc/config"),
					},
				},
			}

			volumes, err := GetVolumes(context.Background(), storageFields, containerFields, networkingFields, k8sClient, namespace, "workload", "", false, true, false)
			Expect(err).To(BeNil())
			Expect(len(volumes)).To(Equal(4)) // PVC, ConfigMap, SHM, Home
		})

		Context("EmptyDirVolumes", func() {
			It("should return no volumes if EmptyDirVolumes is nil or empty", func() {
				volumes, err := GetVolumes(context.Background(), storageFields, nil, networkingFields, k8sClient, namespace, "workload", "", false, true, false)
				Expect(err).To(BeNil())
				Expect(volumes).To(BeEmpty())
			})

			It("should create a single emptyDir volume with all fields set", func() {
				storageFields.EmptyDirVolumes = &field_types.EmptyDirVolumeField{
					Value: &map[string]field_types.EmptyDirVolumeItemAndRules{
						"ed-1": {
							Value: &field_types.EmptyDirVolumeItemValue{
								Path:      ptr.To("/mnt/ed1"),
								Medium:    ptr.To("Memory"),
								SizeLimit: ptr.To("123Mi"),
							},
						},
					},
				}
				volumes, err := GetVolumes(context.Background(), storageFields, nil, networkingFields, k8sClient, namespace, "workload", "", false, true, false)
				Expect(err).To(BeNil())
				Expect(len(volumes)).To(Equal(1))
				v := volumes[0]
				Expect(v.Name).To(Equal("emptydir-volume-0"))
				Expect(v.VolumeSource.EmptyDir).ToNot(BeNil())
				Expect(v.VolumeSource.EmptyDir.Medium).To(Equal(corev1.StorageMediumMemory))
				Expect(v.VolumeSource.EmptyDir.SizeLimit.String()).To(Equal("123Mi"))
			})

			It("should create multiple emptyDir volumes with correct indexing", func() {
				storageFields.EmptyDirVolumes = &field_types.EmptyDirVolumeField{
					Value: &map[string]field_types.EmptyDirVolumeItemAndRules{
						"ed-1": {Value: &field_types.EmptyDirVolumeItemValue{Path: ptr.To("/mnt/ed1")}},
						"ed-2": {Value: &field_types.EmptyDirVolumeItemValue{Path: ptr.To("/mnt/ed2")}},
					},
				}
				volumes, err := GetVolumes(context.Background(), storageFields, nil, networkingFields, k8sClient, namespace, "workload", "", false, true, false)
				Expect(err).To(BeNil())
				Expect(len(volumes)).To(Equal(2))
				Expect(volumes[0].Name).To(Equal("emptydir-volume-0"))
				Expect(volumes[1].Name).To(Equal("emptydir-volume-1"))
			})

			It("should error if path is missing", func() {
				storageFields.EmptyDirVolumes = &field_types.EmptyDirVolumeField{
					Value: &map[string]field_types.EmptyDirVolumeItemAndRules{
						"ed-1": {
							Value: &field_types.EmptyDirVolumeItemValue{
								// Path is missing
							},
						},
					},
				}
				_, err := GetVolumes(context.Background(), storageFields, nil, networkingFields, k8sClient, namespace, "workload", "", false, true, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ed-1: path is mandatory for emptyDir volume"))
			})

			It("should error if medium is invalid", func() {
				storageFields.EmptyDirVolumes = &field_types.EmptyDirVolumeField{
					Value: &map[string]field_types.EmptyDirVolumeItemAndRules{
						"ed-1": {
							Value: &field_types.EmptyDirVolumeItemValue{
								Path:   ptr.To("/mnt/ed1"),
								Medium: ptr.To("disk"),
							},
						},
					},
				}
				_, err := GetVolumes(context.Background(), storageFields, nil, networkingFields, k8sClient, namespace, "workload", "", false, true, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ed-1: medium must be either empty or '%s' for emptyDir volume", corev1.StorageMediumMemory))
			})

			It("should error if sizeLimit is invalid", func() {
				storageFields.EmptyDirVolumes = &field_types.EmptyDirVolumeField{
					Value: &map[string]field_types.EmptyDirVolumeItemAndRules{
						"ed-1": {
							Value: &field_types.EmptyDirVolumeItemValue{
								Path:      ptr.To("/mnt/ed1"),
								SizeLimit: ptr.To("not-a-size"),
							},
						},
					},
				}
				_, err := GetVolumes(context.Background(), storageFields, nil, networkingFields, k8sClient, namespace, "workload", "", false, true, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ed-1: invalid sizeLimit for emptyDir volume: not-a-size"))
			})
		})
	})

	Context("GetVolumeMounts tests", func() {
		var (
			storageFields   *field_groups.StorageFields
			containerFields *field_groups.ContainerFields
		)

		BeforeEach(func() {
			storageFields = &field_groups.StorageFields{
				Pvcs:             &field_types.PvcsField{},
				DataVolumes:      &field_types.DataVolumeField{},
				Volumes:          &field_types.VolumesField{},
				Nfs:              &field_types.NfsField{},
				S3buckets:        &field_types.S3bucketField{},
				LargeShm:         &field_types.BoolField{},
				ConfigMapVolumes: &field_types.ConfigMapVolumeField{},
				SecretVolumes:    &field_types.SecretVolumeField{},
				GitSync:          &field_types.GitSyncField{},
				MountPropagation: &field_types.BoolField{},
			}
			containerFields = &field_groups.ContainerFields{
				CreateHomeDir: &field_types.BoolField{},
				RunAsUser:     &field_types.BoolField{},
				Environment:   &field_types.DictionaryField{},
			}
		})

		It("should return empty slice for empty storage fields", func() {
			mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
			Expect(err).To(BeNil())
			Expect(mounts).To(BeEmpty())
		})

		It("should return error for ephemeral PVC when not supported", func() {
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
				"pvc-1": {
					Value: &field_types.PvcsItemValue{
						Ephemeral: ptr.To(true),
						ClaimName: ptr.To("ephemeral-pvc"),
						Path:      ptr.To("/data"),
					},
				},
			}
			_, err := GetVolumeMounts(context.Background(), storageFields, nil, false, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("this workload type doesn't support ephemeral PVC"))
		})

		It("should create mount for ephemeral PVC when supported", func() {
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
				"pvc-1": {
					Value: &field_types.PvcsItemValue{
						Ephemeral: ptr.To(true),
						ClaimName: ptr.To("ephemeral-pvc"),
						Path:      ptr.To("/data"),
					},
				},
			}
			mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
			Expect(err).To(BeNil())
			Expect(len(mounts)).To(Equal(1))
			Expect(mounts[0].Name).To(Equal("ephemeral-pvc"))
			Expect(mounts[0].MountPath).To(Equal("/data"))
		})

		It("should create mount for host path volume", func() {
			storageFields.Volumes.Value = &map[string]field_types.VolumeItemAndRules{
				"vol-1": {
					Value: &field_types.VolumeItemValue{
						SourcePath: ptr.To("/host/path"),
						TargetPath: ptr.To("/mount/path"),
						ReadOnly:   ptr.To(true),
					},
				},
			}
			mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
			Expect(err).To(BeNil())
			Expect(len(mounts)).To(Equal(1))
			Expect(mounts[0].Name).To(Equal("host-path-volume-0"))
			Expect(mounts[0].MountPath).To(Equal("/mount/path"))
			Expect(mounts[0].ReadOnly).To(BeTrue())
		})

		It("should add mount propagation for NFS volumes when enabled", func() {
			storageFields.Nfs.Value = &map[string]field_types.NfsItemAndRules{
				"nfs-1": {
					Value: &field_types.NfsItemValue{
						NfsServer:  ptr.To("nfs.example.com"),
						TargetPath: ptr.To("/nfs/mount"),
					},
				},
			}
			storageFields.MountPropagation.Value = ptr.To(true)
			mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
			Expect(err).To(BeNil())
			Expect(len(mounts)).To(Equal(1))
			mountPropagation := corev1.MountPropagationHostToContainer
			Expect(mounts[0].MountPropagation).To(Equal(&mountPropagation))
		})

		It("should create mount for data volume", func() {
			storageFields.DataVolumes.Value = &map[string]field_types.DataVolumeItemAndRules{
				"dv-1": {
					Value: &field_types.DataVolumeItemValue{
						Id:        ptr.To("data-volume-id"),
						MountPath: ptr.To("/data"),
					},
				},
			}
			mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
			Expect(err).To(BeNil())
			Expect(len(mounts)).To(Equal(1))
			Expect(mounts[0].Name).To(Equal("data-volume-0"))
			Expect(mounts[0].MountPath).To(Equal("/data"))
			Expect(mounts[0].ReadOnly).To(BeTrue())
		})

		It("should add large shm mount when requested", func() {
			storageFields.LargeShm.Value = ptr.To(true)
			mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
			Expect(err).To(BeNil())
			Expect(len(mounts)).To(Equal(1))
			Expect(mounts[0].Name).To(Equal(dshmVolumeName))
			Expect(mounts[0].MountPath).To(Equal(dshmPath))
		})

		It("should add home directory mount when needed", func() {
			containerFields.CreateHomeDir.Value = ptr.To(true)
			mounts, err := GetVolumeMounts(context.Background(), storageFields, containerFields, false, true)
			Expect(err).To(BeNil())
			Expect(len(mounts)).To(Equal(1))
			Expect(mounts[0].Name).To(Equal(homeVolumeName))
			Expect(mounts[0].MountPath).To(Equal(constants.RunaiDefaultHomeDir))
		})

		It("should add git-sync mounts", func() {
			storageFields.GitSync.Value = &map[string]field_types.GitSyncItemAndRules{
				"git-1": {
					Value: &field_types.GitSyncItemValue{
						Repository: ptr.To("https://github.com/example/repo.git"),
						Directory:  ptr.To("/custom/code"),
					},
				},
			}
			mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
			Expect(err).To(BeNil())
			Expect(len(mounts)).To(Equal(1))
			Expect(mounts[0].Name).To(Equal("code-sync-0"))
			Expect(mounts[0].MountPath).To(Equal("/custom/code"))
		})

		It("should add git-sync mounts with default path when directory not specified", func() {
			storageFields.GitSync.Value = &map[string]field_types.GitSyncItemAndRules{
				"git-1": {
					Value: &field_types.GitSyncItemValue{
						Repository: ptr.To("https://github.com/example/repo.git"),
					},
				},
			}
			mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
			Expect(err).To(BeNil())
			Expect(len(mounts)).To(Equal(1))
			Expect(mounts[0].Name).To(Equal("code-sync-0"))
			Expect(mounts[0].MountPath).To(Equal("/code-git-1"))
		})

		It("should add ConfigMap and Secret volume mounts", func() {
			storageFields.ConfigMapVolumes.Value = &map[string]field_types.ConfigMapVolumeItemAndRules{
				"cm-1": {
					Value: &field_types.ConfigMapVolumeItemValue{
						Name:      ptr.To("my-config"),
						MountPath: ptr.To("/etc/config"),
					},
				},
			}
			storageFields.SecretVolumes.Value = &map[string]field_types.SecretVolumeItemAndRules{
				"secret-1": {
					Value: &field_types.SecretVolumeItemValue{
						Name:      ptr.To("my-secret"),
						MountPath: ptr.To("/etc/secret"),
					},
				},
			}
			mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
			Expect(err).To(BeNil())
			Expect(len(mounts)).To(Equal(2))
			// Check ConfigMap mount
			cmFound := false
			secretFound := false
			for _, mount := range mounts {
				if mount.Name == "cm-volume-0" && mount.MountPath == "/etc/config" {
					cmFound = true
				}
				if mount.Name == "secret-volume-0" && mount.MountPath == "/etc/secret" {
					secretFound = true
				}
			}
			Expect(cmFound).To(BeTrue())
			Expect(secretFound).To(BeTrue())
		})

		It("should skip mounting block PVCs", func() {
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
				"pvc-1": {
					Value: &field_types.PvcsItemValue{
						VolumeMode: ptr.To(string(corev1.PersistentVolumeBlock)),
						Path:       ptr.To("/dev/xvda"),
						Size:       ptr.To("10Gi"),
					},
				},
			}
			mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
			Expect(err).To(BeNil())
			Expect(mounts).To(BeEmpty())
		})

		It("should mount filesystem PVCs", func() {
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
				"pvc-1": {
					Value: &field_types.PvcsItemValue{
						VolumeMode: ptr.To(string(corev1.PersistentVolumeFilesystem)),
						Path:       ptr.To("/data"),
						Size:       ptr.To("10Gi"),
					},
				},
			}
			mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
			Expect(err).To(BeNil())
			Expect(len(mounts)).To(Equal(1))
			Expect(mounts[0].Name).To(Equal("pvc-volume-0"))
			Expect(mounts[0].MountPath).To(Equal("/data"))
		})

		It("should combine multiple mount types", func() {
			storageFields.LargeShm.Value = ptr.To(true)
			containerFields.CreateHomeDir.Value = ptr.To(true)
			storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
				"pvc-1": {
					Value: &field_types.PvcsItemValue{
						Size: ptr.To("10Gi"),
						Path: ptr.To("/data"),
					},
				},
			}
			storageFields.S3buckets.Value = &map[string]field_types.S3bucketItemAndRules{
				"s3-1": {
					Value: &field_types.S3bucketItemValue{
						Path: ptr.To("/s3data"),
					},
				},
			}

			mounts, err := GetVolumeMounts(context.Background(), storageFields, containerFields, false, true)
			Expect(err).To(BeNil())
			Expect(len(mounts)).To(Equal(4)) // PVC, S3, SHM, Home
		})

		Context("EmptyDirVolumes", func() {
			It("should return no mounts if EmptyDirVolumes is nil or empty", func() {
				mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
				Expect(err).To(BeNil())
				Expect(mounts).To(BeEmpty())
			})

			It("should create a single mount for emptyDir volume", func() {
				storageFields.EmptyDirVolumes = &field_types.EmptyDirVolumeField{
					Value: &map[string]field_types.EmptyDirVolumeItemAndRules{
						"ed-1": {
							Value: &field_types.EmptyDirVolumeItemValue{
								Path: ptr.To("/mnt/ed1"),
							},
						},
					},
				}
				mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
				Expect(err).To(BeNil())
				Expect(len(mounts)).To(Equal(1))
				Expect(mounts[0].Name).To(Equal("emptydir-volume-0"))
				Expect(mounts[0].MountPath).To(Equal("/mnt/ed1"))
			})

			It("should create multiple mounts for multiple emptyDir volumes", func() {
				storageFields.EmptyDirVolumes = &field_types.EmptyDirVolumeField{
					Value: &map[string]field_types.EmptyDirVolumeItemAndRules{
						"ed-1": {Value: &field_types.EmptyDirVolumeItemValue{Path: ptr.To("/mnt/ed1")}},
						"ed-2": {Value: &field_types.EmptyDirVolumeItemValue{Path: ptr.To("/mnt/ed2")}},
					},
				}
				mounts, err := GetVolumeMounts(context.Background(), storageFields, nil, false, true)
				Expect(err).To(BeNil())
				Expect(len(mounts)).To(Equal(2))
				Expect(mounts[0].Name).To(Equal("emptydir-volume-0"))
				Expect(mounts[1].Name).To(Equal("emptydir-volume-1"))
			})
		})
	})
})

func initFakeClientWithObjectWithStatus(scheme *runtime.Scheme, statusObjects []client.Object, objects ...client.Object) client.Client {
	testClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).WithStatusSubresource(statusObjects...).Build()
	return safefakeclient.NewFakeClientConcurrentSafe(testClient)
}
