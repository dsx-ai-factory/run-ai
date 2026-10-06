/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

// calling the import here to ensure that the helper init takes place before anything else
// see inside the helper package for more info
import (
	_ "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/tests/helper"

	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("rctBuilder", func() {
	var builder *rctBuilder

	BeforeEach(func() {
		builder = &rctBuilder{}
	})

	// buildMeta returns a project-scoped RCT asset meta populated with the
	// fields Build / applyRunaiK8sObjectMetadata look at. Pulled into a
	// helper so each It stays focused on the behavior under test rather
	// than wire-shape plumbing.
	buildMeta := func(name, id string) assetsApi.AssetMetaForSync {
		projectID := int32(42)
		return assetsApi.AssetMetaForSync{
			Scope:     assetsApi.SCOPE_PROJECT,
			Name:      name,
			Id:        id,
			Kind:      assetsApi.ASSETKIND_RESOURCE_CLAIM_TEMPLATE,
			ProjectId: *assetsApi.NewNullableInt32(&projectID),
		}
	}

	// buildSpec wraps a raw ResourceClaimSpec payload (as the control
	// plane sends it) into the openapi K8sObjectSpec envelope that the
	// agent decodes back into a typed RCT.
	buildSpec := func(objectSpec map[string]interface{}) *assetsApi.K8sObjectSpec {
		spec := assetsApi.NewK8sObjectSpec(objectSpec)
		return spec
	}

	Describe("Build", func() {
		It("populates name, namespace, Runai metadata, and spec", func() {
			update := K8sObjectFromBackend{
				Meta:        buildMeta("my-rct", "rct-id-1"),
				UpdateCount: 7,
				Spec: buildSpec(map[string]interface{}{
					"devices": map[string]interface{}{
						"requests": []interface{}{
							map[string]interface{}{
								"name": "gpu",
								"exactly": map[string]interface{}{
									"deviceClassName": "nvidia.com/gpu",
									"count":           2,
									"allocationMode":  "ExactCount",
								},
							},
						},
					},
				}),
			}

			obj, err := builder.Build(update, "runai-teama")
			Expect(err).NotTo(HaveOccurred())

			rct, ok := obj.(*resourcev1.ResourceClaimTemplate)
			Expect(ok).To(BeTrue())
			Expect(rct.Name).To(Equal("my-rct"))
			Expect(rct.Namespace).To(Equal("runai-teama"))

			// Runai-stamped labels/annotations are present on the outer RCT metadata.
			Expect(rct.Labels).To(HaveKeyWithValue(constants.RunaiAssetLabel, resourceClaimTemplateKindLabel))
			Expect(rct.Labels).To(HaveKeyWithValue(assetIdLabel, "rct-id-1"))
			Expect(rct.Labels).To(HaveKeyWithValue(projectScopeLabel, "42"))
			Expect(rct.Annotations).To(HaveKeyWithValue(assetNameAnnotation, "my-rct"))
			Expect(rct.Annotations).To(HaveKeyWithValue(UpdateCountAnnotation, "7"))

			// Inner spec was decoded through the generic JSON path.
			Expect(rct.Spec.Spec.Devices.Requests).To(HaveLen(1))
			Expect(rct.Spec.Spec.Devices.Requests[0].Name).To(Equal("gpu"))
			Expect(rct.Spec.Spec.Devices.Requests[0].Exactly).NotTo(BeNil())
			Expect(rct.Spec.Spec.Devices.Requests[0].Exactly.DeviceClassName).To(Equal("nvidia.com/gpu"))
			Expect(rct.Spec.Spec.Devices.Requests[0].Exactly.Count).To(Equal(int64(2)))
		})

		It("returns an error when the control-plane spec is missing", func() {
			update := K8sObjectFromBackend{
				Meta:        buildMeta("broken", "rct-id-broken"),
				UpdateCount: 1,
				Spec:        nil,
			}

			_, err := builder.Build(update, "runai-teama")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing spec"))
		})

		It("copies the inner ObjectMeta onto the spawned-ResourceClaim template", func() {
			innerMeta := assetsApi.NewK8sObjectMeta()
			innerMeta.Labels = map[string]string{"pod-selector": "gpu-worker"}
			innerMeta.Annotations = map[string]string{"doc.run.ai/notes": "shared pool"}

			spec := buildSpec(map[string]interface{}{})
			spec.SetObjectMeta(*innerMeta)

			update := K8sObjectFromBackend{
				Meta:        buildMeta("with-meta", "rct-id-2"),
				UpdateCount: 1,
				Spec:        spec,
			}

			obj, err := builder.Build(update, "runai-teama")
			Expect(err).NotTo(HaveOccurred())
			rct := obj.(*resourcev1.ResourceClaimTemplate)

			// Inner template ObjectMeta — this is what each spawned
			// ResourceClaim inherits, not the outer RCT itself.
			Expect(rct.Spec.Labels).To(HaveKeyWithValue("pod-selector", "gpu-worker"))
			Expect(rct.Spec.Annotations).To(HaveKeyWithValue("doc.run.ai/notes", "shared pool"))

			// Outer RCT labels do NOT get the inner user-provided
			// labels — the two surfaces are disjoint by design.
			Expect(rct.Labels).NotTo(HaveKey("pod-selector"))
		})
	})

	Describe("CopySpec", func() {
		It("deep-copies the whole template spec from src onto dst", func() {
			src := &resourcev1.ResourceClaimTemplate{
				Spec: resourcev1.ResourceClaimTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"pod-selector": "gpu-worker"}},
					Spec: resourcev1.ResourceClaimSpec{
						Devices: resourcev1.DeviceClaim{
							Requests: []resourcev1.DeviceRequest{{
								Name: "gpu",
								Exactly: &resourcev1.ExactDeviceRequest{
									DeviceClassName: "nvidia.com/gpu",
									Count:           3,
								},
							}},
						},
					},
				},
			}
			dst := &resourcev1.ResourceClaimTemplate{}

			builder.CopySpec(dst, src)

			Expect(dst.Spec.Labels).To(HaveKeyWithValue("pod-selector", "gpu-worker"))
			Expect(dst.Spec.Spec.Devices.Requests).To(HaveLen(1))
			Expect(dst.Spec.Spec.Devices.Requests[0].Exactly.Count).To(Equal(int64(3)))

			// DeepCopy — mutating src.Spec after must not affect dst.
			src.Spec.Spec.Devices.Requests[0].Exactly.Count = 99
			Expect(dst.Spec.Spec.Devices.Requests[0].Exactly.Count).To(Equal(int64(3)))
		})

		It("is a no-op when either argument is not an RCT", func() {
			notAnRct := &corev1.ConfigMap{}
			rct := &resourcev1.ResourceClaimTemplate{
				Spec: resourcev1.ResourceClaimTemplateSpec{
					Spec: resourcev1.ResourceClaimSpec{
						Devices: resourcev1.DeviceClaim{
							Requests: []resourcev1.DeviceRequest{{Name: "gpu"}},
						},
					},
				},
			}

			// Both directions tolerate wrong types without panicking.
			builder.CopySpec(notAnRct, rct)
			builder.CopySpec(rct, notAnRct)
		})
	})

	Describe("ShouldReCreate", func() {
		// baseRct is the typed RCT the control plane payload decodes
		// into when the user sends a minimal "1 GPU" request without
		// filling in server-defaulted fields. The apiserver fills in
		// AllocationMode=ExactCount, Count=1, Toleration.Operator=Equal
		// during storage; the one we build client-side leaves them zero.
		baseRct := func() *resourcev1.ResourceClaimTemplate {
			return &resourcev1.ResourceClaimTemplate{
				Spec: resourcev1.ResourceClaimTemplateSpec{
					Spec: resourcev1.ResourceClaimSpec{
						Devices: resourcev1.DeviceClaim{
							Requests: []resourcev1.DeviceRequest{{
								Name: "gpu",
								Exactly: &resourcev1.ExactDeviceRequest{
									DeviceClassName: "nvidia.com/gpu",
									Tolerations: []resourcev1.DeviceToleration{{
										Key:   "nvidia.com/gpu",
										Value: "present",
									}},
								},
							}},
						},
					},
				},
			}
		}

		It("returns false when specs are byte-equal", func() {
			expected := baseRct()
			existing := baseRct()
			Expect(builder.ShouldReCreate(expected, existing)).To(BeFalse())
		})

		It("returns false when the only diff is server-applied defaults (AllocationMode, Count, Toleration.Operator)", func() {
			// expected is what the control plane sends: zero values on
			// AllocationMode, Count, and Toleration.Operator.
			expected := baseRct()

			// existing is what the apiserver stored and the agent reads
			// back: the same fields are populated with their defaults.
			existing := baseRct()
			existing.Spec.Spec.Devices.Requests[0].Exactly.AllocationMode = resourcev1.DeviceAllocationModeExactCount
			existing.Spec.Spec.Devices.Requests[0].Exactly.Count = 1
			existing.Spec.Spec.Devices.Requests[0].Exactly.Tolerations[0].Operator = resourcev1.DeviceTolerationOpEqual

			// Without the defaulter pass this would fire false positives
			// on every tick and churn the RCT in a delete+recreate loop.
			Expect(builder.ShouldReCreate(expected, existing)).To(BeFalse())
		})

		It("returns true on a real spec diff (Count bumped from 1 to 4)", func() {
			expected := baseRct()
			expected.Spec.Spec.Devices.Requests[0].Exactly.Count = 4

			existing := baseRct()
			existing.Spec.Spec.Devices.Requests[0].Exactly.AllocationMode = resourcev1.DeviceAllocationModeExactCount
			existing.Spec.Spec.Devices.Requests[0].Exactly.Count = 1
			existing.Spec.Spec.Devices.Requests[0].Exactly.Tolerations[0].Operator = resourcev1.DeviceTolerationOpEqual

			Expect(builder.ShouldReCreate(expected, existing)).To(BeTrue())
		})

		It("returns true on a real spec diff in a nested field (DeviceClassName)", func() {
			expected := baseRct()
			expected.Spec.Spec.Devices.Requests[0].Exactly.DeviceClassName = "amd.com/gpu"

			existing := baseRct()

			Expect(builder.ShouldReCreate(expected, existing)).To(BeTrue())
		})

		It("does not mutate the inputs (defaulter runs on DeepCopies)", func() {
			expected := baseRct()
			existing := baseRct()

			_ = builder.ShouldReCreate(expected, existing)

			// Both callers' originals must still have the zero values
			// they came in with; otherwise a retry-on-conflict loop
			// would compare already-defaulted objects on the next pass.
			Expect(expected.Spec.Spec.Devices.Requests[0].Exactly.AllocationMode).To(BeEquivalentTo(""))
			Expect(expected.Spec.Spec.Devices.Requests[0].Exactly.Count).To(Equal(int64(0)))
			Expect(expected.Spec.Spec.Devices.Requests[0].Exactly.Tolerations[0].Operator).To(BeEquivalentTo(""))
			Expect(existing.Spec.Spec.Devices.Requests[0].Exactly.AllocationMode).To(BeEquivalentTo(""))
			Expect(existing.Spec.Spec.Devices.Requests[0].Exactly.Count).To(Equal(int64(0)))
			Expect(existing.Spec.Spec.Devices.Requests[0].Exactly.Tolerations[0].Operator).To(BeEquivalentTo(""))
		})

		It("returns false when either argument is not an RCT", func() {
			notAnRct := client.Object(&corev1.ConfigMap{})
			rct := client.Object(baseRct())
			Expect(builder.ShouldReCreate(notAnRct, rct)).To(BeFalse())
			Expect(builder.ShouldReCreate(rct, notAnRct)).To(BeFalse())
		})
	})
})
