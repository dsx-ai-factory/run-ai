/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	mnnvltypes "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	"os"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	dinfp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	jobstesting "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/jobs/testing"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

const (
	projectName   = "test-project"
	namespaceName = "runai-test-project"
)

var _ = Describe("DistributedInferenceBuilder GetPreview", func() {
	var (
		ctx       context.Context
		k8sClient client.Client
		builder   *DistributedInferenceBuilder
	)

	BeforeEach(func() {
		ctx = workloadproject.SetProjectInContext(context.Background(), projectName)
		err := os.Setenv(workloadconfig.EnvVarClusterDomain, "test.com")
		Expect(err).ToNot(HaveOccurred())
		scheme := runtime.NewScheme()
		_ = dinfw.AddToScheme(scheme)
		_ = corev1.AddToScheme(scheme)
		_ = networkingv1.AddToScheme(scheme)
		_ = lwsv1.AddToScheme(scheme)
		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		builder = NewDistributedInferenceBuilder(k8sClient, false, false)

		utils.ActiveUuidGenerator = &jobstesting.MockUuidGenerator{}
	})

	Context("when comparing GetPreview and Build outputs", func() {
		It("should return same resources for basic workload", func() {
			workload := &dinfw.DistributedInferenceWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workload",
					Namespace: namespaceName,
				},
				LeaderSpec: dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField("test-image:latest"),
					},
				},
			}

			previewObjects, err := builder.GetPreview(ctx, workload)
			Expect(err).NotTo(HaveOccurred())

			buildObjects, err := builder.Build(ctx, workload, workload)
			Expect(err).NotTo(HaveOccurred())

			compareResourceStructures(previewObjects, buildObjects)
		})

		It("should return same resources for basic workload with uid", func() {
			workload := &dinfw.DistributedInferenceWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workload",
					Namespace: namespaceName,
					UID:       "12345-67890",
				},
				LeaderSpec: dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField("test-image:latest"),
					},
				},
			}

			previewObjects, err := builder.GetPreview(ctx, workload)
			Expect(err).NotTo(HaveOccurred())

			buildObjects, err := builder.Build(ctx, workload, workload)
			Expect(err).NotTo(HaveOccurred())

			compareResourceStructures(previewObjects, buildObjects)

			// ensure LWS name uses the workload UID
			for _, obj := range buildObjects {
				if lws, ok := obj.(*lwsv1.LeaderWorkerSet); ok {
					Expect(lws.Name).To(Equal(GetUniqueLwsName(workload)))
					// verify that the LWS name is longer than just the workload name (which is the case with no UID)
					Expect(len(lws.Name)).To(BeNumerically(">", len(workload.Name)))
				}
			}
		})

		It("should return same resources for workload with serving ports", func() {
			workload := &dinfw.DistributedInferenceWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workload-with-ports",
					Namespace: namespaceName,
				},
				LeaderSpec: dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField("test-image:latest"),
					},
				},
				Spec: dinfp.DistributedInferenceSpec{
					DistributedInferenceFields: field_groups.DistributedInferenceFields{
						ServingPorts: &field_types.ServingPortsField{
							Value: &map[string]field_types.ServingPortItemAndRules{
								"serving-port": {
									Value: &field_types.ServingPortItemValue{
										Port:              ptr.To(8080),
										ExposeExternally:  ptr.To(true),
										AuthorizationType: ptr.To("public"),
									},
								},
							},
						},
					},
				},
			}

			previewObjects, err := builder.GetPreview(ctx, workload)
			Expect(err).NotTo(HaveOccurred())

			buildObjects, err := builder.Build(ctx, workload, workload)
			Expect(err).NotTo(HaveOccurred())

			compareResourceStructures(previewObjects, buildObjects)
		})

		It("should return same resources for workload with PVCs", func() {
			workload := &dinfw.DistributedInferenceWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workload-with-pvcs",
					Namespace: namespaceName,
				},
				LeaderSpec: dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField("test-image:latest"),
					},
					StorageFields: field_groups.StorageFields{
						Pvcs: &field_types.PvcsField{
							Value: &map[string]field_types.PvcsItemAndRules{
								"leader-pvc-1": {
									Value: &field_types.PvcsItemValue{
										Size: ptr.To("10Gi"),
									},
								},
								"leader-pvc-2": {
									Value: &field_types.PvcsItemValue{
										Size: ptr.To("20Gi"),
									},
								},
							},
						},
					},
				},
			}

			previewObjects, err := builder.GetPreview(ctx, workload)
			Expect(err).NotTo(HaveOccurred())

			buildObjects, err := builder.Build(ctx, workload, workload)
			Expect(err).NotTo(HaveOccurred())

			compareResourceStructures(previewObjects, buildObjects)
		})

		It("should return same resources for workload with MNNVL", func() {
			workload := &dinfw.DistributedInferenceWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workload-with-mnnvl",
					Namespace: namespaceName,
					Labels: map[string]string{
						mnnvltypes.FeatureLabelKey: mnnvltypes.Required,
					},
				},
				LeaderSpec: dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField("test-image:latest"),
					},
					ResourceAllocationFields: field_groups.ResourceAllocationFields{
						Gpu: field_types.NewOptionalNumberField(ptr.To(1.0)),
					},
				},
			}

			previewObjects, err := builder.GetPreview(ctx, workload)
			Expect(err).NotTo(HaveOccurred())

			buildObjects, err := builder.Build(ctx, workload, workload)
			Expect(err).NotTo(HaveOccurred())

			compareResourceStructures(previewObjects, buildObjects)
		})

		It("should return same resources for workload with worker spec", func() {
			workload := &dinfw.DistributedInferenceWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-workload-with-workers",
					Namespace: namespaceName,
				},
				LeaderSpec: dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField("test-image:latest"),
					},
					StorageFields: field_groups.StorageFields{
						Pvcs: &field_types.PvcsField{
							Value: &map[string]field_types.PvcsItemAndRules{
								"leader-pvc": {
									Value: &field_types.PvcsItemValue{
										Size: ptr.To("10Gi"),
									},
								},
							},
						},
					},
				},
				WorkerSpec: &dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField("worker-image:latest"),
					},
					StorageFields: field_groups.StorageFields{
						Pvcs: &field_types.PvcsField{
							Value: &map[string]field_types.PvcsItemAndRules{
								"worker-pvc-1": {
									Value: &field_types.PvcsItemValue{
										Size: ptr.To("15Gi"),
									},
								},
								"worker-pvc-2": {
									Value: &field_types.PvcsItemValue{
										Size: ptr.To("25Gi"),
									},
								},
							},
						},
					},
				},
			}

			previewObjects, err := builder.GetPreview(ctx, workload)
			Expect(err).NotTo(HaveOccurred())

			buildObjects, err := builder.Build(ctx, workload, workload)
			Expect(err).NotTo(HaveOccurred())

			compareResourceStructures(previewObjects, buildObjects)
		})

		It("should return same resources for complex workload with all resource types", func() {
			workload := &dinfw.DistributedInferenceWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-complex-workload",
					Namespace: namespaceName,
				},
				LeaderSpec: dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField("leader-image:latest"),
					},
					ResourceAllocationFields: field_groups.ResourceAllocationFields{
						Gpu: field_types.NewOptionalNumberField(ptr.To(1.0)),
					},
					StorageFields: field_groups.StorageFields{
						Pvcs: &field_types.PvcsField{
							Value: &map[string]field_types.PvcsItemAndRules{
								"leader-data-pvc": {
									Value: &field_types.PvcsItemValue{
										Size: ptr.To("100Gi"),
									},
								},
								"leader-logs-pvc": {
									Value: &field_types.PvcsItemValue{
										Size: ptr.To("50Gi"),
									},
								},
							},
						},
					},
				},
				WorkerSpec: &dinfp.DistributedInferenceEntitySpec{
					ContainerFields: field_groups.ContainerFields{
						Image: field_types.NewStringField("worker-image:latest"),
					},
					ResourceAllocationFields: field_groups.ResourceAllocationFields{
						Gpu: field_types.NewOptionalNumberField(ptr.To(2.0)),
					},
					StorageFields: field_groups.StorageFields{
						Pvcs: &field_types.PvcsField{
							Value: &map[string]field_types.PvcsItemAndRules{
								"worker-data-pvc": {
									Value: &field_types.PvcsItemValue{
										Size: ptr.To("200Gi"),
									},
								},
							},
						},
					},
				},
				Spec: dinfp.DistributedInferenceSpec{
					DistributedInferenceFields: field_groups.DistributedInferenceFields{
						ServingPorts: &field_types.ServingPortsField{
							Value: &map[string]field_types.ServingPortItemAndRules{
								"serving-port": {
									Value: &field_types.ServingPortItemValue{
										Port:              ptr.To(9090),
										ExposeExternally:  ptr.To(true),
										AuthorizationType: ptr.To("authorizedUsers"),
										AuthorizedUsers:   ptr.To("user1,user2"),
										ExposedUrl:        ptr.To("my-custom-url"),
									},
								},
							},
						},
					},
				},
			}

			previewObjects, err := builder.GetPreview(ctx, workload)
			Expect(err).NotTo(HaveOccurred())

			buildObjects, err := builder.Build(ctx, workload, workload)
			Expect(err).NotTo(HaveOccurred())

			compareResourceStructures(previewObjects, buildObjects)
		})
	})
})

// compareResourceStructures compares the structure of two resource lists
// checking for same number, types, and metadata
func compareResourceStructures(previewObjects, buildObjects []client.Object) {
	// Check same number of objects
	Expect(len(previewObjects)).To(Equal(len(buildObjects)),
		"Preview and Build should return same number of resources")

	// Group objects by type for comparison
	previewByType := groupObjectsByType(previewObjects)
	buildByType := groupObjectsByType(buildObjects)

	// Check same resource types
	Expect(len(previewByType)).To(Equal(len(buildByType)),
		"Preview and Build should return same resource types")

	for resourceType := range previewByType {
		Expect(buildByType).To(HaveKey(resourceType),
			"Build should contain same resource type as Preview: %s", resourceType)

		previewResources := previewByType[resourceType]
		buildResources := buildByType[resourceType]

		// Check same number of resources per type
		Expect(len(previewResources)).To(Equal(len(buildResources)),
			"Preview and Build should return same number of %s resources", resourceType)

		// Sort both slices by name for consistent comparison
		sortObjectsByName(previewResources)
		sortObjectsByName(buildResources)

		// Compare metadata for each resource
		for i := range previewResources {
			compareObjectMetadata(previewResources[i], buildResources[i], resourceType)
		}
	}
}

// groupObjectsByType groups objects by their Kind using reflection. not relying on the object's TypeMeta
// as it is not necessarily set in all cases
func groupObjectsByType(objects []client.Object) map[string][]client.Object {
	grouped := make(map[string][]client.Object)
	for _, obj := range objects {
		objType := reflect.TypeOf(obj)
		// Handle pointer types by getting the element type
		if objType.Kind() == reflect.Ptr {
			objType = objType.Elem()
		}
		kind := objType.Name()
		grouped[kind] = append(grouped[kind], obj)
	}
	return grouped
}

// sortObjectsByName sorts objects by their name for consistent comparison
func sortObjectsByName(objects []client.Object) {
	for i := 0; i < len(objects)-1; i++ {
		for j := i + 1; j < len(objects); j++ {
			if objects[i].GetName() > objects[j].GetName() {
				objects[i], objects[j] = objects[j], objects[i]
			}
		}
	}
}

// compareObjectMetadata compares the metadata of two objects
func compareObjectMetadata(previewObj, buildObj client.Object, resourceType string) {
	// Compare basic metadata
	Expect(previewObj.GetName()).To(Equal(buildObj.GetName()),
		"Preview and Build %s should have same name", resourceType)
	Expect(previewObj.GetNamespace()).To(Equal(buildObj.GetNamespace()),
		"Preview and Build %s should have same namespace", resourceType)
}
