/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	controller_utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	twv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	resourceUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	workloadUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/run-ai/runai/runai-cluster/common/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("PvcBuilder", func() {
	var (
		pvcBuilder  PvcBuilder
		scheme      *runtime.Scheme
		k8sClient   client.Client
		ctx         context.Context
		owner       *twv1.TrainingWorkload
		JobWorkload *jobs.JobWorkload
	)
	const testWorkloadName = "test"
	BeforeEach(func() {
		scheme = runtime.NewScheme()
		register.InitSchema(scheme)
		ctx = context.Background()
		k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)
		pvcBuilder = PvcBuilder{
			builder: workloadbuilders.NewPvcBuilder(k8sClient),
		}
		JobWorkload = workloadUtils.NewJobWorkloadBuilder().WithTraining().WithNamespace(defaultNamespace).WithName(testWorkloadName).WithUid("test-uid").Build()
		owner = twv1.NewFromJobWorkload(JobWorkload)
		owner.UID = "test-uid"
	})

	const pvcSize = "1Gi"
	const pvcTestName = "pvcClaim"
	Context("isValidPvc", func() {
		Context("No pvc in cluster", func() {
			It("should return true", func() {
				err := pvcBuilder.builder.IsValidPvc(ctx, defaultNamespace, pvcTestName, owner, "")
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("Pvc already exists in cluster with the same name", func() {
			It("should return true when the pvc owner same", func() {
				pvc := resourceUtils.CreatePvc(pvcTestName, defaultNamespace, resource.MustParse(pvcSize), owner.UID, constants.KindTrainingWorkload, "", nil, nil, nil)

				Expect(controller_utils.SetControllerReference(owner, pvc, scheme)).To(Succeed())
				Expect(k8sClient.Create(ctx, pvc)).To(Succeed())
				err := pvcBuilder.builder.IsValidPvc(ctx, defaultNamespace, pvcTestName, owner, "")
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return false when pvc owner different", func() {
				pvc := resourceUtils.CreatePvc(pvcTestName, defaultNamespace, resource.MustParse(pvcSize), "different-uid", constants.KindTrainingWorkload, "", nil, nil, nil)
				Expect(k8sClient.Create(ctx, pvc)).To(Succeed())
				err := pvcBuilder.builder.IsValidPvc(ctx, defaultNamespace, pvcTestName, owner, "")
				Expect(err).To(HaveOccurred())
			})

		})
		Context("Backward compatibility, no ownership exists", func() {
			var (
				oldPvc *v1.PersistentVolumeClaim
			)

			BeforeEach(func() {
				oldPvc = resourceUtils.CreatePvc(pvcTestName, defaultNamespace, resource.MustParse(pvcSize), "", "", "", nil, nil, nil)
				oldPvc.Labels = map[string]string{
					constants.LabelRelease: testWorkloadName,
				}
				Expect(k8sClient.Create(ctx, oldPvc)).To(Succeed())
			})
			It("should return true, when pvc has label of the Workload", func() {
				err := pvcBuilder.builder.IsValidPvc(ctx, defaultNamespace, pvcTestName, owner, JobWorkload.Name)
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return false, when pvc has no label of the Workload", func() {
				oldPvc.Labels = map[string]string{}
				Expect(k8sClient.Update(ctx, oldPvc)).To(Succeed())
				err := pvcBuilder.builder.IsValidPvc(ctx, defaultNamespace, pvcTestName, owner, JobWorkload.Name)
				Expect(err).To(HaveOccurred())
			})
			It("should return false, when pvc has  label of different Workload", func() {
				oldPvc.Labels = map[string]string{
					constants.LabelRelease: "different-Workload",
				}
				Expect(k8sClient.Update(ctx, oldPvc)).To(Succeed())
				err := pvcBuilder.builder.IsValidPvc(ctx, defaultNamespace, pvcTestName, owner, JobWorkload.Name)
				Expect(err).To(HaveOccurred())
			})
		})
	})
	Context("Build Pvc", func() {
		var (
			pvcList  []*v1.PersistentVolumeClaim
			workload *jobs.JobWorkload
		)
		const storageName = "storage"
		const workloadOwnershipError = "is owned by another workload"
		Context("With 3 new PVC", func() {

			BeforeEach(func() {
				pvcList = []*v1.PersistentVolumeClaim{
					resourceUtils.CreatePvc(pvcTestName, defaultNamespace, resource.MustParse(pvcSize), owner.UID, constants.KindTrainingWorkload, storageName, nil, nil, nil),
					resourceUtils.CreatePvc("pvcClaim2", defaultNamespace, resource.MustParse(pvcSize), owner.UID, constants.KindTrainingWorkload, storageName, nil, nil, nil),
					resourceUtils.CreatePvc("pvcClaim3", defaultNamespace, resource.MustParse(pvcSize), owner.UID, constants.KindTrainingWorkload, storageName, nil, nil, nil),
				}
				builder := workloadUtils.NewJobWorkloadBuilder().WithTraining().WithNamespace(defaultNamespace).WithName(testWorkloadName).WithUid("test-uid")
				for _, pvc := range pvcList {
					builder.WithPVC(pvc, false, false)
				}
				workload = builder.Build()
				owner = twv1.NewFromJobWorkload(workload)
			})
			It("should return all pvc", func() {
				objectList, err := pvcBuilder.Build(ctx, workload, ResourceBuilderOptions{Owner: owner})
				Expect(err).ToNot(HaveOccurred())
				Expect(objectList).To(HaveLen(3))
				for _, obj := range objectList {
					pvc := obj.(*v1.PersistentVolumeClaim)
					Expect(pvc.OwnerReferences).To(HaveLen(1))
					Expect(pvc.OwnerReferences[0].UID).To(Equal(owner.UID))
					Expect(pvc.Name).To(Or(Equal(pvcTestName), Equal("pvcClaim2"), Equal("pvcClaim3")))
					Expect(*pvc.Spec.StorageClassName).To(Equal(storageName))
				}
			})
			It("should return err when pvc already exists with different owner", func() {
				pvc := pvcList[0]
				otherOwner := twv1.NewFromJobWorkload(workload)
				otherOwner.UID = "different-uid"
				Expect(controller_utils.SetControllerReference(otherOwner, pvc, scheme)).To(Succeed())
				Expect(k8sClient.Create(ctx, pvc)).To(Succeed())
				_, err := pvcBuilder.Build(ctx, workload, ResourceBuilderOptions{Owner: owner})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(workloadOwnershipError))
			})
			It("should match preview pvc", func() {
				previewPvcList, err := pvcBuilder.GetPreview(ctx, workload)
				Expect(err).ToNot(HaveOccurred())
				Expect(previewPvcList).To(HaveLen(3))
				objectList, err := pvcBuilder.Build(ctx, workload, ResourceBuilderOptions{Owner: owner})
				Expect(err).ToNot(HaveOccurred())
				Expect(objectList).To(HaveLen(3))
				pvcNameNamespace := []types.NamespacedName{}
				for _, obj := range objectList {
					pvcNameNamespace = append(pvcNameNamespace, types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()})
				}
				for _, pvc := range previewPvcList {
					Expect(pvcNameNamespace).To(ContainElement(types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}))
				}
			})
			It("should fail if pvc size is not set", func() {
				workload.Spec.Pvcs.Deref()[pvcTestName].Value.Size = utils.Ref("")
				_, err := pvcBuilder.Build(ctx, workload, ResourceBuilderOptions{Owner: owner})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid size for creating a PVC"))
			})
		})
		Context("With 1 exiting PVC 1 ephemeral and 1 regular PCV", func() {
			var (
				existingPvc *v1.PersistentVolumeClaim
				regPvc      *v1.PersistentVolumeClaim
				ephPvc      *v1.PersistentVolumeClaim
			)
			BeforeEach(func() {
				existingPvc = resourceUtils.CreatePvc("ExistingPVC", defaultNamespace, resource.MustParse(pvcSize), owner.UID, constants.KindTrainingWorkload, storageName, nil, nil, nil)
				regPvc = resourceUtils.CreatePvc("regular", defaultNamespace, resource.MustParse(pvcSize), owner.UID, constants.KindTrainingWorkload, storageName, nil, nil, nil)
				ephPvc = resourceUtils.CreatePvc("ephemeral", defaultNamespace, resource.MustParse(pvcSize), owner.UID, constants.KindTrainingWorkload, storageName, nil, nil, nil)
				builder := workloadUtils.NewJobWorkloadBuilder().WithTraining().WithNamespace(defaultNamespace).WithName(testWorkloadName).WithUid("test-uid")
				builder.WithPVC(existingPvc, true, false)
				builder.WithPVC(regPvc, false, false)
				builder.WithPVC(ephPvc, false, true)
				workload = builder.Build()
				owner = twv1.NewFromJobWorkload(workload)
			})
			It("should return only regular pvc", func() {
				objectList, err := pvcBuilder.Build(ctx, workload, ResourceBuilderOptions{Owner: owner})
				Expect(err).ToNot(HaveOccurred())
				Expect(objectList).To(HaveLen(1))
				pvc := objectList[0].(*v1.PersistentVolumeClaim)
				Expect(pvc.OwnerReferences).To(HaveLen(1))
				Expect(pvc.OwnerReferences[0].UID).To(Equal(owner.UID))
				Expect(pvc.Name).To(Equal(regPvc.Name))
			})
			It("should return err when pvc already exists with different owner", func() {
				pvc := regPvc
				otherOwner := twv1.NewFromJobWorkload(workload)
				otherOwner.UID = "different-uid"
				Expect(controller_utils.SetControllerReference(otherOwner, pvc, scheme)).To(Succeed())
				Expect(k8sClient.Create(ctx, pvc)).To(Succeed())
				_, err := pvcBuilder.Build(ctx, workload, ResourceBuilderOptions{Owner: owner})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(workloadOwnershipError))
			})
			It("should match preview pvc", func() {
				previewPvcList, err := pvcBuilder.GetPreview(ctx, workload)
				Expect(err).ToNot(HaveOccurred())
				Expect(previewPvcList).To(HaveLen(1))
				objectList, err := pvcBuilder.Build(ctx, workload, ResourceBuilderOptions{Owner: owner})
				Expect(err).ToNot(HaveOccurred())
				Expect(objectList).To(HaveLen(1))
				pvcNameNamespace := []types.NamespacedName{}
				for _, obj := range objectList {
					pvcNameNamespace = append(pvcNameNamespace, types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()})
				}
				for _, pvc := range previewPvcList {
					Expect(pvcNameNamespace).To(ContainElement(types.NamespacedName{Name: pvc.Name, Namespace: pvc.Namespace}))
				}
			})
		})
	})
})
