/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package storage

import (
	"context"
	"testing"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Suite")
}

func storageRunner(newPvcs []corev1.PersistentVolumeClaim) *dromosv1alpha1.WorkloadRunner {
	var storage *dromosv1alpha1.StorageConfig
	if newPvcs != nil {
		storage = &dromosv1alpha1.StorageConfig{NewPvcs: newPvcs}
	}
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{Name: "wl", Namespace: "ns"},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Instructions: &dromosv1alpha1.Instructions{Storage: storage},
		},
	}
}

var _ = Describe("Storage disabled sentinel", func() {
	var fw *framework.Framework

	BeforeEach(func() {
		GinkgoT().Setenv("STORAGE_ENABLED", "false")

		p, err := New(context.Background(), framework.PluginDeps{})
		Expect(err).NotTo(HaveOccurred())
		Expect(p.Name()).To(Equal(PluginName))

		fw, err = framework.NewFrameworkFromPluginBuilders(context.Background(), &framework.Config{},
			[]framework.PluginBuilder{New}, framework.PluginDeps{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("rejects a create asking for PVCs while disabled, with a Forbidden error on the storage path", func() {
		runner := storageRunner([]corev1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "pvc1"}}})
		_, errs := fw.RunValidators(context.Background(), nil, runner)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Type).To(Equal(field.ErrorTypeForbidden))
		Expect(errs[0].Field).To(Equal(storageInstructionPath().String()))
	})

	It("allows updating a running storage workload when the PVCs are unchanged", func() {
		existing := storageRunner([]corev1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "pvc1"}}})
		desired := storageRunner([]corev1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "pvc1"}}})
		_, errs := fw.RunValidators(context.Background(), existing, desired)
		Expect(errs).To(BeEmpty())
	})

	It("rejects changing the PVCs on an update while disabled", func() {
		existing := storageRunner([]corev1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "pvc1"}}})
		desired := storageRunner([]corev1.PersistentVolumeClaim{
			{ObjectMeta: metav1.ObjectMeta{Name: "pvc1"}},
			{ObjectMeta: metav1.ObjectMeta{Name: "pvc2"}},
		})
		_, errs := fw.RunValidators(context.Background(), existing, desired)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Type).To(Equal(field.ErrorTypeForbidden))
	})

	It("does not reject a runner with no storage instruction or no PVCs", func() {
		bare := storageRunner(nil)
		_, errs := fw.RunValidators(context.Background(), nil, bare)
		Expect(errs).To(BeEmpty())

		empty := storageRunner([]corev1.PersistentVolumeClaim{})
		_, errs = fw.RunValidators(context.Background(), nil, empty)
		Expect(errs).To(BeEmpty())
	})
})
