package handlers

// calling the import here to ensure that the helper init takes place before anything else
// see inside the helper package for more info
import (
	"time"

	_ "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/tests/helper"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	workloads_openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("WorkloadTypesConfigTests", func() {
	var (
		inspector WtcConfigMapObjectInspector
	)

	BeforeEach(func() {
		inspector = WtcConfigMapObjectInspector{}
	})

	// Helper function to create sample WorkloadTypeConfig items
	createSampleWorkloadTypeConfigs := func() []workloads_openapi.WorkloadTypeConfig {
		strPtr := func(s string) *string { return &s }
		timePtr := func(t time.Time) *time.Time { return &t }
		configs := make([]workloads_openapi.WorkloadTypeConfig, 2)
		configs[0] = workloads_openapi.WorkloadTypeConfig{
			CategoryId:   "b3e1c1e2-1234-4a5b-8c6d-abcdef123456",
			PriorityId:   "c4d2e3f4-5678-4b9c-8d7e-fedcba654321",
			Name:         "Deployment",
			Id:           strPtr("a1b2c3d4-1111-2222-3333-444455556666"),
			CategoryName: strPtr("CategoryA"),
			PriorityName: strPtr("PriorityA"),
			UpdatedAt:    timePtr(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
			UpdatedBy:    strPtr("userA"),
		}
		configs[1] = workloads_openapi.WorkloadTypeConfig{
			CategoryId:   "e5f6a7b8-9876-4c3b-8e9f-123456789abc",
			PriorityId:   "f7e8d9c0-4321-4d5e-8f0a-abcdefabcdef",
			Name:         "RayCluster",
			Id:           strPtr("d4c3b2a1-7777-8888-9999-000011112222"),
			CategoryName: strPtr("CategoryB"),
			PriorityName: strPtr("PriorityB"),
			UpdatedAt:    timePtr(time.Date(2024, 2, 2, 15, 30, 0, 0, time.UTC)),
			UpdatedBy:    strPtr("userB"),
		}
		return configs
	}

	// Helper function to create sample WorkloadCategory items
	createSampleWorkloadCategories := func() []workloads_openapi.WorkloadCategory {
		strPtr := func(s string) *string { return &s }
		timePtr := func(t time.Time) *time.Time { return &t }
		categories := make([]workloads_openapi.WorkloadCategory, 2)
		categories[0] = workloads_openapi.WorkloadCategory{
			Name:      "CategoryA",
			Id:        "cat1",
			CreatedAt: timePtr(time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)),
			CreatedBy: strPtr("userA"),
		}
		categories[1] = workloads_openapi.WorkloadCategory{
			Name:      "CategoryB",
			Id:        "cat2",
			CreatedAt: timePtr(time.Date(2024, 2, 2, 15, 30, 0, 0, time.UTC)),
			CreatedBy: strPtr("userB"),
		}
		return categories
	}

	// Helper function to create a ConfigMap with workload type configs and categories
	createConfigMapWithWorkloadConfigs := func(configs []workloads_openapi.WorkloadTypeConfig, categories []workloads_openapi.WorkloadCategory) *corev1.ConfigMap {
		cm := &corev1.ConfigMap{}
		cm.Data = make(map[string]string)
		err := setWorkloadsTypeConfig(cm, configs)
		Expect(err).To(BeNil())
		err = setWorkloadsCategories(cm, categories)
		Expect(err).To(BeNil())
		return cm
	}

	Context("WorkloadsTypesConfigObjectInspector", func() {
		Describe("IsUpToDate", func() {
			It("should return false when either object is nil", func() {
				cm := &corev1.ConfigMap{}
				Expect(inspector.IsUpToDate(nil, cm)).To(BeFalse())
				Expect(inspector.IsUpToDate(cm, nil)).To(BeFalse())
				Expect(inspector.IsUpToDate(nil, nil)).To(BeFalse())
			})

			It("should return false when objects are not ConfigMaps", func() {
				pod := &corev1.Pod{}
				cm := &corev1.ConfigMap{}
				Expect(inspector.IsUpToDate(pod, cm)).To(BeFalse())
				Expect(inspector.IsUpToDate(cm, pod)).To(BeFalse())
			})

			It("should return false when ConfigMaps lack the required keys", func() {
				cm1 := &corev1.ConfigMap{Data: map[string]string{}}
				cm2 := &corev1.ConfigMap{Data: map[string]string{}}
				Expect(inspector.IsUpToDate(cm1, cm2)).To(BeFalse())
			})

			It("should return true when ConfigMaps have identical workload configurations", func() {
				configs := createSampleWorkloadTypeConfigs()
				categories := createSampleWorkloadCategories()
				cm1 := createConfigMapWithWorkloadConfigs(configs, categories)
				cm2 := createConfigMapWithWorkloadConfigs(configs, categories)
				Expect(inspector.IsUpToDate(cm1, cm2)).To(BeTrue())
			})

			It("should return false when ConfigMaps have different workload type configurations", func() {
				configs1 := createSampleWorkloadTypeConfigs()
				configs2 := createSampleWorkloadTypeConfigs()
				categories := createSampleWorkloadCategories()
				// Modify the second config to make it different
				configs2[0].Name = "DifferentName"
				cm1 := createConfigMapWithWorkloadConfigs(configs1, categories)
				cm2 := createConfigMapWithWorkloadConfigs(configs2, categories)
				Expect(inspector.IsUpToDate(cm1, cm2)).To(BeFalse())
			})

			It("should return false when ConfigMaps have different workload category configurations", func() {
				configs := createSampleWorkloadTypeConfigs()
				categories1 := createSampleWorkloadCategories()
				categories2 := createSampleWorkloadCategories()
				// Modify the second categories to make it different
				categories2[0].Name = "DifferentCategory"
				cm1 := createConfigMapWithWorkloadConfigs(configs, categories1)
				cm2 := createConfigMapWithWorkloadConfigs(configs, categories2)
				Expect(inspector.IsUpToDate(cm1, cm2)).To(BeFalse())
			})
		})

		Describe("DeepCopySpec", func() {
			It("should not modify target when source is not a ConfigMap", func() {
				pod := &corev1.Pod{}
				configs := createSampleWorkloadTypeConfigs()
				categories := createSampleWorkloadCategories()
				target := createConfigMapWithWorkloadConfigs(configs, categories)
				originalData := target.Data
				inspector.DeepCopySpec(pod, target)
				Expect(target.Data).To(Equal(originalData))
			})

			It("should copy all required keys from source to target", func() {
				sourceConfigs := createSampleWorkloadTypeConfigs()
				sourceCategories := createSampleWorkloadCategories()
				targetConfigs := createSampleWorkloadTypeConfigs()
				targetCategories := createSampleWorkloadCategories()
				// Modify target configs to be different from source
				targetConfigs[0].Name = "DifferentName"
				targetCategories[0].Name = "DifferentCategory"

				source := createConfigMapWithWorkloadConfigs(sourceConfigs, sourceCategories)
				target := createConfigMapWithWorkloadConfigs(targetConfigs, targetCategories)

				inspector.DeepCopySpec(source, target)
				Expect(target.Data[constants.WorkloadsConfigMapTypesKey]).To(Equal(source.Data[constants.WorkloadsConfigMapTypesKey]))
				Expect(target.Data[constants.WorkloadsConfigMapCategoriesKey]).To(Equal(source.Data[constants.WorkloadsConfigMapCategoriesKey]))
			})

			It("should initialize target Data map if nil", func() {
				sourceConfigs := createSampleWorkloadTypeConfigs()
				sourceCategories := createSampleWorkloadCategories()
				source := createConfigMapWithWorkloadConfigs(sourceConfigs, sourceCategories)
				target := &corev1.ConfigMap{}
				inspector.DeepCopySpec(source, target)
				Expect(target.Data).NotTo(BeNil())
				Expect(target.Data[constants.WorkloadsConfigMapTypesKey]).To(Equal(source.Data[constants.WorkloadsConfigMapTypesKey]))
				Expect(target.Data[constants.WorkloadsConfigMapCategoriesKey]).To(Equal(source.Data[constants.WorkloadsConfigMapCategoriesKey]))
			})
		})
	})
})

var _ = Describe("WorkloadTypes Karta", Ordered, func() {
	var (
		wth *WorkloadTypeKartasHandler
	)

	BeforeAll(func() {
		wth = &WorkloadTypeKartasHandler{backendWorkloadTypes: map[string]kartaForAgent{}}
	})

	It("extractKarta should convert backend resourceDefinitions into Karta objects", func() {
		// Build a minimal Karta map as returned by backend
		ridMap := map[string]interface{}{
			"apiVersion": "optimization.nvidia.com/v1alpha1",
			"kind":       "Karta",
			"metadata":   map[string]interface{}{},
			"spec": map[string]interface{}{
				"structureDefinition": map[string]interface{}{
					"rootComponent": map[string]interface{}{
						"name": "Pod",
						"kind": map[string]interface{}{
							"group":   "",
							"version": "v1alpha1",
							"kind":    "Pod",
						},
						"specDefinition": map[string]interface{}{
							"podSpecPath":  "spec",
							"metadataPath": "metadata",
						},
					},
				},
			},
		}

		version := "v1alpha1"
		wt := workloads_openapi.WorkloadTypeConfig{
			CategoryId: "cat",
			PriorityId: "prio",
			Name:       "Pod",
			Group:      "",
			Kartas: []workloads_openapi.WorkloadTypeKartasWithVersionInner{
				{
					Version: &version,
					Karta:   ridMap,
				},
			},
		}

		resp := &workloads_openapi.ListWorkloadTypes200Response{Types: []workloads_openapi.WorkloadTypeConfig{wt}}

		objs, err := wth.extractKartas(resp)
		Expect(err).To(BeNil())
		Expect(objs).To(HaveLen(1))

		karta, ok := objs[0].(*kartav1alpha1.Karta)
		Expect(ok).To(BeTrue())
		Expect(karta.Spec.StructureDefinition.RootComponent.Kind).To(BeNil())
		Expect(karta.Spec.StructureDefinition.RootComponent.SpecDefinition).To(BeNil())

		_, found := wth.backendWorkloadTypes[karta.Name]
		Expect(found).To(BeTrue())

		fullObject, err := wth.GetExpectedObject(karta.Name, "")
		Expect(err).To(BeNil())
		Expect(fullObject).ToNot(BeNil())
		karta, ok = fullObject.(*kartav1alpha1.Karta)
		Expect(ok).To(BeTrue())
		Expect(karta.Spec.StructureDefinition.RootComponent.Kind).NotTo(BeNil())
		Expect(karta.Spec.StructureDefinition.RootComponent.Kind.Kind).To(Equal("Pod"))
		Expect(karta.Spec.StructureDefinition.RootComponent.Kind.Version).To(Equal("v1alpha1"))
		Expect(karta.Spec.StructureDefinition.RootComponent.Kind.Group).To(Equal(""))
		Expect(karta.Spec.StructureDefinition.RootComponent.SpecDefinition).ToNot(BeNil())
		Expect(karta.Spec.StructureDefinition.RootComponent.SpecDefinition.PodSpecPath).ToNot(BeNil())
		Expect(*karta.Spec.StructureDefinition.RootComponent.SpecDefinition.PodSpecPath).To(Equal("spec"))
		Expect(karta.Spec.StructureDefinition.RootComponent.SpecDefinition.MetadataPath).ToNot(BeNil())
		Expect(*karta.Spec.StructureDefinition.RootComponent.SpecDefinition.MetadataPath).To(Equal("metadata"))

		expectedName, genErr := generateKartaName(karta)
		Expect(genErr).To(BeNil())
		Expect(karta.GetName()).To(Equal(expectedName))
	})

	It("generateKartaName should produce deterministic name from GVK", func() {
		karta := &kartav1alpha1.Karta{}
		karta.Spec.StructureDefinition.RootComponent.Kind = &kartav1alpha1.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		}

		name, err := generateKartaName(karta)
		Expect(err).To(BeNil())
		Expect(name).To(Equal("krt-v1-apps-deployment"))
	})

	It("extractKartas should handle multiple workload types built with buildResourceDefinitionForGvk", func() {
		ver := "v1"
		wtJob := workloads_openapi.WorkloadTypeConfig{
			CategoryId: "cat",
			PriorityId: "prio",
			Name:       "Job",
			Group:      "batch",
			Kartas: []workloads_openapi.WorkloadTypeKartasWithVersionInner{
				{Version: &ver, Karta: buildResourceDefinitionForGvk("batch", "v1", "Job")},
			},
		}
		wtDeploy := workloads_openapi.WorkloadTypeConfig{
			CategoryId: "cat",
			PriorityId: "prio",
			Name:       "Deployment",
			Group:      "apps",
			Kartas: []workloads_openapi.WorkloadTypeKartasWithVersionInner{
				{Version: &ver, Karta: buildResourceDefinitionForGvk("apps", "v1", "Deployment")},
			},
		}

		resp := &workloads_openapi.ListWorkloadTypes200Response{Types: []workloads_openapi.WorkloadTypeConfig{wtJob, wtDeploy}}
		objs, err := wth.extractKartas(resp)
		Expect(err).To(BeNil())
		Expect(objs).To(HaveLen(2))

		kinds := map[string]bool{}
		groups := map[string]bool{}
		versions := map[string]bool{}
		for _, obj := range objs {
			_, ok := wth.backendWorkloadTypes[obj.GetName()]
			Expect(ok).To(BeTrue())
			fullObj, err := wth.GetExpectedObject(obj.GetName(), "")
			Expect(err).To(BeNil())
			Expect(fullObj).ToNot(BeNil())

			r, ok := fullObj.(*kartav1alpha1.Karta)
			Expect(ok).To(BeTrue())
			Expect(r.Spec.StructureDefinition.RootComponent.Kind).NotTo(BeNil())
			kinds[r.Spec.StructureDefinition.RootComponent.Kind.Kind] = true
			groups[r.Spec.StructureDefinition.RootComponent.Kind.Group] = true
			versions[r.Spec.StructureDefinition.RootComponent.Kind.Version] = true
		}
		Expect(kinds).To(HaveKey("Job"))
		Expect(kinds).To(HaveKey("Deployment"))
		Expect(groups).To(HaveKey("batch"))
		Expect(groups).To(HaveKey("apps"))
		Expect(versions).To(HaveKey("v1"))
	})

	It("generateKartaName should normalize dots in group (ray.io)", func() {
		karta := &kartav1alpha1.Karta{}
		karta.Spec.StructureDefinition.RootComponent.Kind = &kartav1alpha1.GroupVersionKind{
			Group:   "ray.io",
			Version: "v1",
			Kind:    "RayCluster",
		}
		name, err := generateKartaName(karta)
		Expect(err).To(BeNil())
		// dots converted to dashes
		Expect(name).To(Equal("krt-v1-ray-io-raycluster"))
	})

	It("should handle very long kind name with same group and two versions (v2, v2alpha2)", func() {
		longKind := "ThisIsAnExtremelyLongCustomKindNameThatExceedsTheKubernetesMaxResourceNameLengthByALot"
		group := "custom.run.ai"
		v2 := "v2"
		v2a2 := "v2alpha2"

		wt := workloads_openapi.WorkloadTypeConfig{
			CategoryId: "cat",
			PriorityId: "prio",
			Name:       longKind,
			Group:      group,
			Kartas: []workloads_openapi.WorkloadTypeKartasWithVersionInner{
				{Version: &v2, Karta: buildResourceDefinitionForGvk(group, v2, longKind)},
				{Version: &v2a2, Karta: buildResourceDefinitionForGvk(group, v2a2, longKind)},
			},
		}

		resp := &workloads_openapi.ListWorkloadTypes200Response{Types: []workloads_openapi.WorkloadTypeConfig{wt}}
		objs, err := wth.extractKartas(resp)
		Expect(err).To(BeNil())
		Expect(objs).To(HaveLen(2))
		Expect(objs[0]).ToNot(Equal(objs[1]))

		// names already generated inside extractKartas
		for _, obj := range objs {
			r := obj.(*kartav1alpha1.Karta)
			Expect(r.GetName()).ToNot(BeEmpty())
			Expect(r.GetName()).To(ContainSubstring("krt-"))
			Expect(r.GetName()).To(ContainSubstring("-custom-run-ai-"))
		}
	})
})

// buildResourceDefinitionForGvk constructs a minimal Karta resourceDefinition map
// similar to the control-plane helper, for use in agent tests.
func buildResourceDefinitionForGvk(group, version, kind string) map[string]interface{} {
	return map[string]interface{}{
		"spec": map[string]interface{}{
			"structureDefinition": map[string]interface{}{
				"rootComponent": map[string]interface{}{
					"kind": map[string]interface{}{
						"group":   group,
						"version": version,
						"kind":    kind,
					},
					"name": "",
				},
			},
			"optimizationInstructions": map[string]interface{}{},
		},
		"metadata": map[string]interface{}{
			"creationTimestamp": nil,
		},
		"status": map[string]interface{}{},
	}
}
