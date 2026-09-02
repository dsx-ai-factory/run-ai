/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("networking ownership binding", func() {
	const workloadID = "cp-workload-id"

	It("persists an extensible versioned ownership contract", func() {
		var configuration *models.WorkloadConfiguration

		enableNetworkingOwnershipContract(&configuration)

		Expect(configuration.NetworkingOwnershipContract).To(Equal(&models.NetworkingOwnershipContract{
			Version: networkingOwnershipContractV1,
		}))
		Expect(usesNetworkingOwnershipContract(configuration)).To(BeTrue())

		configuration.NetworkingOwnershipContract.Version = "future"
		Expect(usesNetworkingOwnershipContract(configuration)).To(BeFalse())
	})

	It("overwrites the protected identity and preserves existing selector and pod labels", func(ctx context.Context) {
		selector := map[string]string{
			"app":                     "demo",
			networkingWorkloadIDLabel: "caller-value",
		}
		services := []api.NetworkingService{
			{Spec: &api.ServiceSpec{Selector: &selector, Ports: []api.ServicePort{{Port: 8080}}}},
			{Ref: &api.ServiceRef{Name: "existing-service", Port: 8080}},
		}
		networking := models.NetworkingConfiguration{{Services: &services}}
		record := &models.WorkloadsQueueRecord{
			ID:              workloadID,
			ResourceVersion: "v1",
			EffectiveManifest: dao_utils.JsonBMap{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata":   map[string]interface{}{"name": "demo"},
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"labels": map[string]interface{}{
								"app":                     "demo",
								networkingWorkloadIDLabel: "caller-value",
							},
						},
						"spec": map[string]interface{}{
							"containers": []interface{}{map[string]interface{}{"name": "main", "image": "busybox"}},
						},
					},
				},
			},
			Configuration: &models.WorkloadConfiguration{
				Networking:                  &networking,
				NetworkingOwnershipContract: &models.NetworkingOwnershipContract{Version: networkingOwnershipContractV1},
			},
		}
		workloadType := createWorkloadTypeWithRi(createRiWithGvk(
			&schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			&kartav1alpha1.SpecDefinition{PodTemplateSpecPath: convutils.Ref(".spec.template")},
		))

		Expect(applyNetworkingOwnershipBinding(ctx, record, workloadType)).To(Succeed())

		Expect(selector).To(Equal(map[string]string{
			"app":                     "demo",
			networkingWorkloadIDLabel: workloadID,
		}))
		Expect(services[1].Ref).To(Equal(&api.ServiceRef{Name: "existing-service", Port: 8080}))

		labels, found, err := unstructured.NestedStringMap(
			map[string]interface{}(record.EffectiveManifest),
			"spec", "template", "metadata", "labels",
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(labels).To(Equal(map[string]string{
			"app":                     "demo",
			networkingWorkloadIDLabel: workloadID,
		}))
	})

	It("creates a selector map when the caller omits one", func(ctx context.Context) {
		services := []api.NetworkingService{{
			Spec: &api.ServiceSpec{Ports: []api.ServicePort{{Port: 8080}}},
		}}
		networking := models.NetworkingConfiguration{{Services: &services}}

		Expect(bindCreatedServiceSelectors(&networking, workloadID)).To(Equal(1))
		Expect(services[0].Spec.Selector).NotTo(BeNil())
		Expect(*services[0].Spec.Selector).To(Equal(map[string]string{
			networkingWorkloadIDLabel: workloadID,
		}))
	})

	It("stamps every pod template described by a multi-component Karta", func(ctx context.Context) {
		services := []api.NetworkingService{{
			Spec: &api.ServiceSpec{Ports: []api.ServicePort{{Port: 8080}}},
		}}
		networking := models.NetworkingConfiguration{{Services: &services}}
		record := &models.WorkloadsQueueRecord{
			ID:              workloadID,
			ResourceVersion: "v1",
			EffectiveManifest: dao_utils.JsonBMap{
				"apiVersion": "example.run.ai/v1",
				"kind":       "DistributedWorkload",
				"metadata":   map[string]interface{}{"name": "demo"},
				"spec": map[string]interface{}{
					"leader": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{"labels": map[string]interface{}{"role": "leader"}},
							"spec":     map[string]interface{}{"containers": []interface{}{map[string]interface{}{"name": "main", "image": "busybox"}}},
						},
					},
					"worker": map[string]interface{}{
						"template": map[string]interface{}{
							"metadata": map[string]interface{}{"labels": map[string]interface{}{"role": "worker"}},
							"spec":     map[string]interface{}{"containers": []interface{}{map[string]interface{}{"name": "main", "image": "busybox"}}},
						},
					},
				},
			},
			Configuration: &models.WorkloadConfiguration{
				Networking:                  &networking,
				NetworkingOwnershipContract: &models.NetworkingOwnershipContract{Version: networkingOwnershipContractV1},
			},
		}
		karta := &kartav1alpha1.Karta{Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "root",
					Kind: &kartav1alpha1.GroupVersionKind{Group: "example.run.ai", Version: "v1", Kind: "DistributedWorkload"},
				},
				ChildComponents: []kartav1alpha1.ComponentDefinition{
					{Name: "leader", SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: convutils.Ref(".spec.leader.template")}},
					{Name: "worker", SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: convutils.Ref(".spec.worker.template")}},
				},
			},
		}}

		Expect(applyNetworkingOwnershipBinding(ctx, record, createWorkloadTypeWithRi(karta))).To(Succeed())
		for _, component := range []string{"leader", "worker"} {
			labels, found, err := unstructured.NestedStringMap(
				map[string]interface{}(record.EffectiveManifest),
				"spec", component, "template", "metadata", "labels",
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(labels).To(HaveKeyWithValue(networkingWorkloadIDLabel, workloadID))
			Expect(labels).To(HaveKeyWithValue("role", component))
		}
	})

	DescribeTable("does not materialize absent optional Kubeflow replicas",
		func(
			ctx context.Context,
			karta *kartav1alpha1.Karta,
			kind string,
			replicaSpecsField string,
			presentRoles []string,
			absentRoles []string,
		) {
			replicaSpecs := map[string]interface{}{}
			for _, role := range presentRoles {
				replicaSpecs[role] = validReplicaForNetworkingTest(role)
			}
			record := networkingOwnershipRecordForTest(workloadID, "v1", kubeflowManifestForNetworkingTest(
				kind,
				replicaSpecsField,
				replicaSpecs,
			))

			Expect(applyNetworkingOwnershipBinding(ctx, record, createWorkloadTypeWithRi(karta))).To(Succeed())

			updatedReplicaSpecs := replicaSpecsForNetworkingTest(record, replicaSpecsField)
			Expect(updatedReplicaSpecs).To(HaveLen(len(presentRoles)))
			for _, role := range presentRoles {
				Expect(updatedReplicaSpecs).To(HaveKey(role))
				expectReplicaNetworkingLabel(record, replicaSpecsField, role, workloadID)
			}
			for _, role := range absentRoles {
				Expect(updatedReplicaSpecs).NotTo(HaveKey(role))
			}
			expectCreatedServiceNetworkingSelector(record, workloadID)
		},
		Entry(
			"for a worker-only PyTorchJob",
			kubeflowPyTorchKartaForNetworkingTest(),
			"PyTorchJob",
			"pytorchReplicaSpecs",
			[]string{"Worker"},
			[]string{"Master"},
		),
		Entry(
			"for a worker-only TFJob",
			kubeflowTFKartaForNetworkingTest(),
			"TFJob",
			"tfReplicaSpecs",
			[]string{"Worker"},
			[]string{"Chief", "PS", "Evaluator"},
		),
		Entry(
			"for a Chief-and-Worker TFJob",
			kubeflowTFKartaForNetworkingTest(),
			"TFJob",
			"tfReplicaSpecs",
			[]string{"Chief", "Worker"},
			[]string{"PS", "Evaluator"},
		),
	)

	for _, test := range []struct {
		name                string
		masterValue         interface{}
		templateKeyExpected bool
	}{
		{name: "null replica", masterValue: nil},
		{name: "replica without a pod template", masterValue: map[string]interface{}{}},
		{name: "null pod template", masterValue: map[string]interface{}{"template": nil}, templateKeyExpected: true},
	} {
		test := test
		It("preserves an explicit "+test.name, func(ctx context.Context) {
			replicaSpecs := map[string]interface{}{
				"Master": test.masterValue,
				"Worker": validReplicaForNetworkingTest("Worker"),
			}
			record := networkingOwnershipRecordForTest(workloadID, "v1", kubeflowManifestForNetworkingTest(
				"PyTorchJob",
				"pytorchReplicaSpecs",
				replicaSpecs,
			))

			Expect(applyNetworkingOwnershipBinding(
				ctx,
				record,
				createWorkloadTypeWithRi(kubeflowPyTorchKartaForNetworkingTest()),
			)).To(Succeed())

			updatedReplicaSpecs := replicaSpecsForNetworkingTest(record, "pytorchReplicaSpecs")
			Expect(updatedReplicaSpecs).To(HaveLen(2))
			Expect(updatedReplicaSpecs).To(HaveKey("Master"))
			if test.masterValue == nil {
				Expect(updatedReplicaSpecs["Master"]).To(BeNil())
			} else {
				master, ok := updatedReplicaSpecs["Master"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				if test.templateKeyExpected {
					Expect(master).To(HaveKey("template"))
					Expect(master["template"]).To(BeNil())
				} else {
					Expect(master).NotTo(HaveKey("template"))
				}
			}
			expectReplicaNetworkingLabel(record, "pytorchReplicaSpecs", "Worker", workloadID)
		})
	}

	It("treats an explicitly present empty pod template as a writable target", func(ctx context.Context) {
		replicaSpecs := map[string]interface{}{
			"Master": map[string]interface{}{"template": map[string]interface{}{}},
			"Worker": validReplicaForNetworkingTest("Worker"),
		}
		record := networkingOwnershipRecordForTest(workloadID, "v1", kubeflowManifestForNetworkingTest(
			"PyTorchJob",
			"pytorchReplicaSpecs",
			replicaSpecs,
		))

		Expect(applyNetworkingOwnershipBinding(
			ctx,
			record,
			createWorkloadTypeWithRi(kubeflowPyTorchKartaForNetworkingTest()),
		)).To(Succeed())

		expectReplicaNetworkingLabel(record, "pytorchReplicaSpecs", "Master", workloadID)
		expectReplicaNetworkingLabel(record, "pytorchReplicaSpecs", "Worker", workloadID)
	})

	DescribeTable("fails closed when a created Service has no present writable pod-label target",
		func(ctx context.Context, replicaSpecs map[string]interface{}) {
			record := networkingOwnershipRecordForTest(workloadID, "v1", kubeflowManifestForNetworkingTest(
				"PyTorchJob",
				"pytorchReplicaSpecs",
				replicaSpecs,
			))

			err := applyNetworkingOwnershipBinding(
				ctx,
				record,
				createWorkloadTypeWithRi(kubeflowPyTorchKartaForNetworkingTest()),
			)

			Expect(err).To(MatchError(ContainSubstring("Karta exposes no writable pod-label target")))
			Expect(replicaSpecsForNetworkingTest(record, "pytorchReplicaSpecs")).To(Equal(replicaSpecs))
		},
		Entry("when every replica is missing", map[string]interface{}{}),
		Entry("when every pod template is null", map[string]interface{}{
			"Master": map[string]interface{}{"template": nil},
			"Worker": map[string]interface{}{"template": nil},
		}),
	)

	DescribeTable("skips an absent optional bare PodSpec when another target is writable",
		func(ctx context.Context, includeNullTarget bool) {
			karta := kubeflowPyTorchKartaForNetworkingTest()
			karta.Spec.StructureDefinition.ChildComponents = append(
				karta.Spec.StructureDefinition.ChildComponents,
				kartav1alpha1.ComponentDefinition{
					Name:     "optional-pod-spec",
					Kind:     &kartav1alpha1.GroupVersionKind{Version: "v1", Kind: "Pod"},
					OwnerRef: convutils.Ref("pytorchjob"),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodSpecPath: convutils.Ref(".spec.optionalPodSpec"),
					},
				},
			)
			manifest := kubeflowManifestForNetworkingTest(
				"PyTorchJob",
				"pytorchReplicaSpecs",
				map[string]interface{}{"Worker": validReplicaForNetworkingTest("Worker")},
			)
			if includeNullTarget {
				manifest["spec"].(map[string]interface{})["optionalPodSpec"] = nil
			}
			record := networkingOwnershipRecordForTest(workloadID, "v1", manifest)

			Expect(applyNetworkingOwnershipBinding(ctx, record, createWorkloadTypeWithRi(karta))).To(Succeed())

			expectReplicaNetworkingLabel(record, "pytorchReplicaSpecs", "Worker", workloadID)
			podSpec, found, err := unstructured.NestedFieldNoCopy(
				map[string]interface{}(record.EffectiveManifest),
				"spec", "optionalPodSpec",
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(Equal(includeNullTarget))
			Expect(podSpec).To(BeNil())
		},
		Entry("when the target is missing", false),
		Entry("when the target is null", true),
	)

	It("fails closed for mixed present and absent targets in one Karta component", func(ctx context.Context) {
		record := networkingOwnershipRecordForTest(workloadID, "v1", dao_utils.JsonBMap{
			"apiVersion": "example.run.ai/v1",
			"kind":       "GroupedWorkload",
			"metadata":   map[string]interface{}{"name": "grouped-demo"},
			"spec": map[string]interface{}{
				"groups": []interface{}{
					map[string]interface{}{
						"name":     "present",
						"template": validReplicaForNetworkingTest("present")["template"],
					},
					map[string]interface{}{"name": "absent", "template": nil},
				},
			},
		})
		karta := &kartav1alpha1.Karta{Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "root",
					Kind: &kartav1alpha1.GroupVersionKind{Group: "example.run.ai", Version: "v1", Kind: "GroupedWorkload"},
				},
				ChildComponents: []kartav1alpha1.ComponentDefinition{{
					Name:           "groups",
					Kind:           &kartav1alpha1.GroupVersionKind{Version: "v1", Kind: "Pod"},
					OwnerRef:       convutils.Ref("root"),
					InstanceIdPath: convutils.Ref(".spec.groups[].name"),
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						PodTemplateSpecPath: convutils.Ref(".spec.groups[].template"),
					},
					PodSelector: &kartav1alpha1.PodSelector{
						ComponentInstanceSelector: &kartav1alpha1.ComponentInstanceSelector{
							IdPath: ".metadata.labels[\"group\"]",
						},
					},
				}},
			},
		}}

		err := applyNetworkingOwnershipBinding(ctx, record, createWorkloadTypeWithRi(karta))

		Expect(err).To(MatchError(ContainSubstring("returned both present and absent pod targets")))
		groups, found, nestedErr := unstructured.NestedSlice(
			map[string]interface{}(record.EffectiveManifest), "spec", "groups",
		)
		Expect(nestedErr).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(groups).To(HaveLen(2))
		presentGroup, ok := groups[0].(map[string]interface{})
		Expect(ok).To(BeTrue())
		labels, labelsFound, labelsErr := unstructured.NestedStringMap(
			presentGroup, "template", "metadata", "labels",
		)
		Expect(labelsErr).NotTo(HaveOccurred())
		Expect(labelsFound).To(BeTrue())
		Expect(labels).To(HaveKeyWithValue("role", "present"))
		Expect(labels).NotTo(HaveKey(networkingWorkloadIDLabel))
	})

	It("stamps fragmented pod labels used by operator-backed workloads", func(ctx context.Context) {
		services := []api.NetworkingService{{Spec: &api.ServiceSpec{Ports: []api.ServicePort{{Port: 8080}}}}}
		networking := models.NetworkingConfiguration{{Services: &services}}
		record := &models.WorkloadsQueueRecord{
			ID:              workloadID,
			ResourceVersion: "v1alpha1",
			EffectiveManifest: dao_utils.JsonBMap{
				"apiVersion": "apps.example.com/v1alpha1",
				"kind":       "OperatorWorkload",
				"metadata":   map[string]interface{}{"name": "demo"},
				"spec":       map[string]interface{}{"labels": map[string]interface{}{"app": "demo"}},
			},
			Configuration: &models.WorkloadConfiguration{
				Networking:                  &networking,
				NetworkingOwnershipContract: &models.NetworkingOwnershipContract{Version: networkingOwnershipContractV1},
			},
		}
		karta := &kartav1alpha1.Karta{Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Name: "root",
					Kind: &kartav1alpha1.GroupVersionKind{Group: "apps.example.com", Version: "v1alpha1", Kind: "OperatorWorkload"},
					SpecDefinition: &kartav1alpha1.SpecDefinition{
						FragmentedPodSpecDefinition: &kartav1alpha1.FragmentedPodSpecDefinition{
							LabelsPath: convutils.Ref(".spec.labels"),
						},
					},
				},
			},
		}}

		Expect(applyNetworkingOwnershipBinding(ctx, record, createWorkloadTypeWithRi(karta))).To(Succeed())
		labels, found, err := unstructured.NestedStringMap(map[string]interface{}(record.EffectiveManifest), "spec", "labels")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(labels).To(HaveKeyWithValue("app", "demo"))
		Expect(labels).To(HaveKeyWithValue(networkingWorkloadIDLabel, workloadID))
	})

	It("strips a spoofed identity without requiring Karta for ServiceRef", func(ctx context.Context) {
		services := []api.NetworkingService{{
			Ref: &api.ServiceRef{Name: "existing-service", Port: 8080},
		}}
		networking := models.NetworkingConfiguration{{Services: &services}}
		manifest := manifestWithNetworkingWorkloadID("victim-workload-id")
		record := &models.WorkloadsQueueRecord{
			ID:                workloadID,
			EffectiveManifest: manifest,
			Configuration: &models.WorkloadConfiguration{
				Networking:                  &networking,
				NetworkingOwnershipContract: &models.NetworkingOwnershipContract{Version: networkingOwnershipContractV1},
			},
		}

		Expect(applyNetworkingOwnershipBinding(ctx, record, nil)).To(Succeed())
		Expect(services[0].Ref.Name).To(Equal("existing-service"))
		labels, found, err := unstructured.NestedStringMap(
			map[string]interface{}(record.EffectiveManifest),
			"spec", "template", "metadata", "labels",
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(labels).To(HaveKeyWithValue("app", "demo"))
		Expect(labels).NotTo(HaveKey(networkingWorkloadIDLabel))
	})

	It("strips a spoofed identity from a workload without networking", func(ctx context.Context) {
		record := &models.WorkloadsQueueRecord{
			ID:                workloadID,
			EffectiveManifest: manifestWithNetworkingWorkloadID("victim-workload-id"),
			Configuration: &models.WorkloadConfiguration{
				NetworkingOwnershipContract: &models.NetworkingOwnershipContract{Version: networkingOwnershipContractV1},
			},
		}

		Expect(applyNetworkingOwnershipBinding(ctx, record, nil)).To(Succeed())
		labels, found, err := unstructured.NestedStringMap(
			map[string]interface{}(record.EffectiveManifest),
			"spec", "template", "metadata", "labels",
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(labels).To(HaveKeyWithValue("app", "demo"))
		Expect(labels).NotTo(HaveKey(networkingWorkloadIDLabel))
	})

	It("leaves records created before the ownership contract unchanged", func(ctx context.Context) {
		selector := map[string]string{networkingWorkloadIDLabel: "legacy-value"}
		services := []api.NetworkingService{{
			Spec: &api.ServiceSpec{Selector: &selector, Ports: []api.ServicePort{{Port: 8080}}},
		}}
		networking := models.NetworkingConfiguration{{Services: &services}}
		record := &models.WorkloadsQueueRecord{
			ID:                workloadID,
			EffectiveManifest: dao_utils.JsonBMap{"apiVersion": "v1", "kind": "Pod"},
			Configuration:     &models.WorkloadConfiguration{Networking: &networking},
		}

		Expect(applyNetworkingOwnershipBinding(ctx, record, nil)).To(Succeed())
		Expect(selector[networkingWorkloadIDLabel]).To(Equal("legacy-value"))
	})

	It("fails closed when Karta exposes only a bare PodSpec", func(ctx context.Context) {
		services := []api.NetworkingService{{
			Spec: &api.ServiceSpec{Ports: []api.ServicePort{{Port: 8080}}},
		}}
		networking := models.NetworkingConfiguration{{Services: &services}}
		record := &models.WorkloadsQueueRecord{
			ID:              workloadID,
			ResourceVersion: "v1",
			EffectiveManifest: dao_utils.JsonBMap{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata":   map[string]interface{}{"name": "demo"},
				"spec": map[string]interface{}{
					"containers": []interface{}{map[string]interface{}{"name": "main", "image": "busybox"}},
				},
			},
			Configuration: &models.WorkloadConfiguration{
				Networking:                  &networking,
				NetworkingOwnershipContract: &models.NetworkingOwnershipContract{Version: networkingOwnershipContractV1},
			},
		}
		workloadType := createWorkloadTypeWithRi(createRiWithGvk(
			&schema.GroupVersionKind{Version: "v1", Kind: "Pod"},
			&kartav1alpha1.SpecDefinition{PodSpecPath: convutils.Ref(".spec")},
		))

		err := applyNetworkingOwnershipBinding(ctx, record, workloadType)
		Expect(err).To(MatchError(ContainSubstring("cannot carry the pod label")))
	})
})

func manifestWithNetworkingWorkloadID(value string) dao_utils.JsonBMap {
	return dao_utils.JsonBMap{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app":                     "demo",
						networkingWorkloadIDLabel: value,
					},
				},
			},
		},
	}
}

type kubeflowReplicaComponentForNetworkingTest struct {
	name string
	key  string
}

// These fixtures mirror the complete pod-template paths in the production
// PyTorchJob and TFJob Kartas without importing the separate workloads module.
func kubeflowPyTorchKartaForNetworkingTest() *kartav1alpha1.Karta {
	return kubeflowReplicaKartaForNetworkingTest(
		"pytorchjob",
		"PyTorchJob",
		"pytorchReplicaSpecs",
		[]kubeflowReplicaComponentForNetworkingTest{
			{name: "master", key: "Master"},
			{name: "worker", key: "Worker"},
		},
	)
}

func kubeflowTFKartaForNetworkingTest() *kartav1alpha1.Karta {
	return kubeflowReplicaKartaForNetworkingTest(
		"tfjob",
		"TFJob",
		"tfReplicaSpecs",
		[]kubeflowReplicaComponentForNetworkingTest{
			{name: "Chief", key: "Chief"},
			{name: "Worker", key: "Worker"},
			{name: "PS", key: "PS"},
			{name: "Evaluator", key: "Evaluator"},
		},
	)
}

func kubeflowReplicaKartaForNetworkingTest(
	rootName string,
	kind string,
	replicaSpecsField string,
	components []kubeflowReplicaComponentForNetworkingTest,
) *kartav1alpha1.Karta {
	children := make([]kartav1alpha1.ComponentDefinition, 0, len(components))
	for _, component := range components {
		children = append(children, kartav1alpha1.ComponentDefinition{
			Name:     component.name,
			Kind:     &kartav1alpha1.GroupVersionKind{Version: "v1", Kind: "Pod"},
			OwnerRef: convutils.Ref(rootName),
			SpecDefinition: &kartav1alpha1.SpecDefinition{
				PodTemplateSpecPath: convutils.Ref(
					".spec." + replicaSpecsField + "." + component.key + ".template",
				),
			},
		})
	}

	return &kartav1alpha1.Karta{Spec: kartav1alpha1.KartaSpec{
		StructureDefinition: kartav1alpha1.StructureDefinition{
			RootComponent: kartav1alpha1.ComponentDefinition{
				Name: rootName,
				Kind: &kartav1alpha1.GroupVersionKind{Group: "kubeflow.org", Version: "v1", Kind: kind},
			},
			ChildComponents: children,
		},
	}}
}

func kubeflowManifestForNetworkingTest(
	kind string,
	replicaSpecsField string,
	replicaSpecs map[string]interface{},
) dao_utils.JsonBMap {
	return dao_utils.JsonBMap{
		"apiVersion": "kubeflow.org/v1",
		"kind":       kind,
		"metadata":   map[string]interface{}{"name": "distributed-demo"},
		"spec": map[string]interface{}{
			replicaSpecsField: replicaSpecs,
		},
	}
}

func validReplicaForNetworkingTest(role string) map[string]interface{} {
	return map[string]interface{}{
		"template": map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{"role": role},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{"name": "main", "image": "busybox"},
				},
			},
		},
	}
}

func networkingOwnershipRecordForTest(
	workloadID string,
	resourceVersion string,
	manifest dao_utils.JsonBMap,
) *models.WorkloadsQueueRecord {
	services := []api.NetworkingService{{
		Spec: &api.ServiceSpec{Ports: []api.ServicePort{{Port: 8080}}},
	}}
	networking := models.NetworkingConfiguration{{Services: &services}}
	return &models.WorkloadsQueueRecord{
		ID:                workloadID,
		ResourceVersion:   resourceVersion,
		EffectiveManifest: manifest,
		Configuration: &models.WorkloadConfiguration{
			Networking:                  &networking,
			NetworkingOwnershipContract: &models.NetworkingOwnershipContract{Version: networkingOwnershipContractV1},
		},
	}
}

func replicaSpecsForNetworkingTest(
	record *models.WorkloadsQueueRecord,
	replicaSpecsField string,
) map[string]interface{} {
	GinkgoHelper()
	replicaSpecs, found, err := unstructured.NestedMap(
		map[string]interface{}(record.EffectiveManifest), "spec", replicaSpecsField,
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(found).To(BeTrue())
	return replicaSpecs
}

func expectReplicaNetworkingLabel(
	record *models.WorkloadsQueueRecord,
	replicaSpecsField string,
	role string,
	workloadID string,
) {
	GinkgoHelper()
	labels, found, err := unstructured.NestedStringMap(
		map[string]interface{}(record.EffectiveManifest),
		"spec", replicaSpecsField, role, "template", "metadata", "labels",
	)
	Expect(err).NotTo(HaveOccurred())
	Expect(found).To(BeTrue())
	Expect(labels).To(HaveKeyWithValue(networkingWorkloadIDLabel, workloadID))
}

func expectCreatedServiceNetworkingSelector(record *models.WorkloadsQueueRecord, workloadID string) {
	GinkgoHelper()
	Expect(record.Configuration).NotTo(BeNil())
	Expect(record.Configuration.Networking).NotTo(BeNil())
	Expect(*record.Configuration.Networking).NotTo(BeEmpty())
	services := (*record.Configuration.Networking)[0].Services
	Expect(services).NotTo(BeNil())
	Expect(*services).NotTo(BeEmpty())
	Expect((*services)[0].Spec).NotTo(BeNil())
	Expect((*services)[0].Spec.Selector).NotTo(BeNil())
	Expect(*(*services)[0].Spec.Selector).To(HaveKeyWithValue(networkingWorkloadIDLabel, workloadID))
}
