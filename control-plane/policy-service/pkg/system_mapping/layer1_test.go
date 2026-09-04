/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping_test

// Layer 1 of the system-mapping test plan: structural validation of a system
// workload mapping (spec selectors, field groups, item selectors) against a
// hand-authored "kitchen-sink" reference spec for that GVK.
//
// The walker iterates the merged mapping and, for every field that uses a
// builtin formatter (KeyValue / KeyValueMap), evaluates the spec-selector +
// (optional) item-selector pipeline against the reference spec, then walks the
// declared SpecRef from each emitted root and asserts the resulting value's
// JSON kind matches the field's MappingFieldType.
//
// Two assertions are made per GVK:
//
//   - every field with a builtin formatter must resolve at least once across
//     all (selector × item-selector) combinations (catches typos in SpecRef,
//     dead JQ queries, item selectors pointed at the wrong subtree).
//   - whenever a field does resolve, the value's kind must match its declared
//     MappingFieldType (catches type-system drift between the mapping and the
//     real k8s schema).
//
// Fields whose formatter is a custom system formatter (runai-podSpec etc.)
// are skipped: their SpecRef is a logical handle, not a path, so this layer
// can't validate them. Those formatters are exercised by Layer 3.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/itchyny/gojq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"sigs.k8s.io/yaml"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	sm "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/system_mapping"
)

// systemFieldGroupBuilders is a static name → builder lookup that mirrors the
// switch performed by handleSystemFieldGroups at seed time. Kept here (rather
// than reused) so this test stays decoupled from the seeding code paths and
// from the database — Layer 1 is a pure in-memory check.
var systemFieldGroupBuilders = map[string]sm.SystemFieldGroupBuilder{
	sm.SystemPodSpecFgName:                   sm.BuildSystemPodSpecFg,
	sm.SystemPodSpecContainerFgName:          sm.BuildSystemPodSpecContainerFg,
	sm.SystemPodSpecAndContainerFgName:       sm.BuildSystemPodSpecAndContainerFg,
	sm.SystemJobSpecFgName:                   sm.BuildSystemJobSpecFg,
	sm.SystemCronJobSpecFgName:               sm.BuildSystemCronJobSpecFg,
	sm.SystemSelectorFgName:                  sm.BuildSystemSelectorFg,
	sm.SystemDeploymentSpecFgName:            sm.BuildSystemDeploymentSpecFg,
	sm.SystemReplicaSetSpecFgName:            sm.BuildSystemReplicaSetSpecFg,
	sm.SystemStatefulSetSpecFgName:           sm.BuildSystemStatefulSetSpecFg,
	sm.SystemMetadataFgName:                  sm.BuildSystemMetadataFg,
	sm.SystemNimCacheSpecFgName:              sm.BuildSystemNimCacheSpecFg,
	sm.SystemNimCacheContainerFgName:         sm.BuildSystemNimCacheContainerFg,
	sm.SystemNimServiceSpecFgName:            sm.BuildSystemNimServiceSpecFg,
	sm.SystemNimServiceContainerFgName:       sm.BuildSystemNimServiceContainerFg,
	sm.SystemKubeflowRunPolicyFgName:         sm.BuildSystemKubeflowRunPolicyFg,
	sm.SystemKubeflowReplicaSpecFgName:       sm.BuildSystemKubeflowReplicaSpecFg,
	sm.SystemPyTorchJobSpecFgName:            sm.BuildSystemPyTorchJobSpecFg,
	sm.SystemMPIJobSpecFgName:                sm.BuildSystemMPIJobSpecFg,
	sm.SystemLeaderWorkerSetSpecFgName:       sm.BuildSystemLeaderWorkerSetSpecFg,
	sm.SystemLeaderWorkerTemplateFgName:      sm.BuildSystemLeaderWorkerTemplateFg,
	sm.SystemDynamoGraphDeploymentSpecFgName: sm.BuildSystemDynamoGraphDeploymentSpecFg,
	sm.SystemDynamoServiceSpecFgName:         sm.BuildSystemDynamoServiceSpecFg,
	sm.SystemRayClusterSpecFgName:            sm.BuildSystemRayClusterSpecFg,
	sm.SystemRayWorkerGroupSpecFgName:        sm.BuildSystemRayWorkerGroupSpecFg,
	sm.SystemRayJobSpecFgName:                sm.BuildSystemRayJobSpecFg,
	sm.SystemRayServiceSpecFgName:            sm.BuildSystemRayServiceSpecFg,
	sm.SystemNativeSpecFgName:                sm.BuildSystemNativeSpecFg,
	sm.SystemNativeDistributedSpecFgName:     sm.BuildSystemNativeDistributedSpecFg,
}

var _ = Describe("system mapping layer 1: structural validation against reference spec", func() {
	Context("walker self-tests", func() {
		It("catches a typo in SpecRef as NeverResolved", func() {
			const fgName = "selfTestFG"
			builders := map[string]sm.SystemFieldGroupBuilder{
				fgName: func() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
					return map[string]api.FieldGroupField{
							"goodField": {Formatter: dbmodel.BuiltinFormatterKeyValue, SpecRef: "replicas", Type: api.MAPPINGFIELDTYPE_INTEGER},
							"badField":  {Formatter: dbmodel.BuiltinFormatterKeyValue, SpecRef: "replicaz", Type: api.MAPPINGFIELDTYPE_INTEGER},
						},
						map[string]api.ItemSelector{}
				},
			}
			specSelectors := map[string]api.SpecSelector{
				"selfTest": {Query: ".spec", FieldGroups: []string{fgName}},
			}
			refSpec := map[string]any{
				"spec": map[string]any{"replicas": float64(3)},
			}

			report, err := runLayer1(specSelectors, builders, refSpec)
			Expect(err).To(BeNil())
			Expect(report.NeverResolved).To(HaveLen(1))
			Expect(report.NeverResolved[0].fieldGroup).To(Equal(fgName))
			Expect(report.NeverResolved[0].field).To(Equal("badField"))
			Expect(report.TypeMismatch).To(BeEmpty())
		})

		It("catches a type mismatch as TypeMismatch", func() {
			const fgName = "selfTestFG"
			builders := map[string]sm.SystemFieldGroupBuilder{
				fgName: func() (map[string]api.FieldGroupField, map[string]api.ItemSelector) {
					return map[string]api.FieldGroupField{
							"shouldBeInt": {Formatter: dbmodel.BuiltinFormatterKeyValue, SpecRef: "name", Type: api.MAPPINGFIELDTYPE_INTEGER},
						},
						map[string]api.ItemSelector{}
				},
			}
			specSelectors := map[string]api.SpecSelector{
				"selfTest": {Query: ".spec", FieldGroups: []string{fgName}},
			}
			refSpec := map[string]any{
				"spec": map[string]any{"name": "actually-a-string"},
			}

			report, err := runLayer1(specSelectors, builders, refSpec)
			Expect(err).To(BeNil())
			Expect(report.TypeMismatch).To(HaveLen(1))
			Expect(report.TypeMismatch[0].field).To(Equal("shouldBeInt"))
			Expect(report.NeverResolved).To(BeEmpty())
		})
	})

	Context("apps/Deployment/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemAppsDeploymentV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/apps_Deployment_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("batch/Job/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemBatchJobV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/batch_Job_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("apps/ReplicaSet/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemAppsReplicaSetV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/apps_ReplicaSet_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("apps/StatefulSet/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemAppsStatefulSetV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/apps_StatefulSet_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("batch/CronJob/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemBatchCronJobV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/batch_CronJob_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("/Pod/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemCorePodV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/core_Pod_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("apps.nvidia.com/NIMCache/v1alpha1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemAppsNvidiaComNimCacheV1alpha1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/apps_nvidia_com_NIMCache_v1alpha1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("apps.nvidia.com/NIMService/v1alpha1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemAppsNvidiaComNimServiceV1alpha1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/apps_nvidia_com_NIMService_v1alpha1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("kubeflow.org/PyTorchJob/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemKubeflowOrgPyTorchJobV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/kubeflow_org_PyTorchJob_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("kubeflow.org/MPIJob/v2beta1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemKubeflowOrgMPIJobV2beta1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/kubeflow_org_MPIJob_v2beta1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("kubeflow.org/TFJob/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemKubeflowOrgTFJobV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/kubeflow_org_TFJob_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("kubeflow.org/XGBoostJob/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemKubeflowOrgXGBoostJobV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/kubeflow_org_XGBoostJob_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("kubeflow.org/JAXJob/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemKubeflowOrgJAXJobV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/kubeflow_org_JAXJob_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("leaderworkerset.x-k8s.io/LeaderWorkerSet/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemLeaderWorkerSetXK8sIoV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/leaderworkerset_x_k8s_io_LeaderWorkerSet_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("nvidia.com/DynamoGraphDeployment/v1alpha1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemNvidiaComDynamoGraphDeploymentV1alpha1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/nvidia_com_DynamoGraphDeployment_v1alpha1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("ray.io/RayCluster/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemRayIoRayClusterV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/ray_io_RayCluster_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("ray.io/RayJob/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemRayIoRayJobV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/ray_io_RayJob_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("ray.io/RayService/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemRayIoRayServiceV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/ray_io_RayService_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("serving.knative.dev/Service/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemServingKnativeDevServiceV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/serving_knative_dev_Service_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("run.ai/Workspace/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemRunAiWorkspaceV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/run_ai_Workspace_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})

	Context("run.ai/Distributed/v1", func() {
		It("resolves and is type-correct against the reference spec", func() {
			_, specSelectors := sm.BuildSystemRunAiDistributedV1Mapping()
			refSpec := loadReferenceSpecGinkgo("testing/inputs/run_ai_Distributed_v1/reference.yaml")

			report, err := runLayer1(specSelectors, systemFieldGroupBuilders, refSpec)
			Expect(err).To(BeNil())
			Expect(formatReport(report)).To(BeEmpty())
		})
	})
})

// fieldKey identifies a single field-group field across the whole mapping.
// Used to track resolution status: a field is "resolved" if at least one
// (specSelector × itemSelector) combination produced a value at its SpecRef.
type fieldKey struct {
	fieldGroup string
	field      string
}

// typeMismatch records a path that resolved to a value whose JSON kind does
// not match the field's declared MappingFieldType.
type typeMismatch struct {
	specSelector string
	itemSelector string
	fieldGroup   string
	field        string
	specRef      string
	wantType     api.MappingFieldType
	got          any
}

// layer1Report aggregates a single GVK walk's results.
type layer1Report struct {
	NeverResolved []fieldKey
	TypeMismatch  []typeMismatch
}

// runLayer1 walks the (specSelectors, fieldGroups) mapping over refSpec and
// reports unresolved fields and type mismatches.
//
// Builders are passed in so each test can choose between the production set
// (`systemFieldGroupBuilders`) and a synthetic registry for self-tests; this
// keeps the walker pure and avoids mutating shared state between tests.
func runLayer1(
	specSelectors map[string]api.SpecSelector,
	builders map[string]sm.SystemFieldGroupBuilder,
	refSpec map[string]any,
) (*layer1Report, error) {
	report := &layer1Report{}

	tracked := map[fieldKey]bool{}
	resolved := map[fieldKey]bool{}

	specSelectorNames := sm.SortedKeys(specSelectors)

	for _, selName := range specSelectorNames {
		sel := specSelectors[selName]

		selOutputs, err := evalJQ(sel.Query, refSpec, sampleArgs(sel.ArgsInfo, sel.ArgsEnum))
		if err != nil {
			return nil, fmt.Errorf("spec selector %q: %w", selName, err)
		}
		if len(selOutputs) == 0 {
			return nil, fmt.Errorf("spec selector %q: produced no outputs from reference spec", selName)
		}

		for _, fgName := range sel.FieldGroups {
			builder, ok := builders[fgName]
			if !ok {
				return nil, fmt.Errorf("spec selector %q: field group %q has no registered builder", selName, fgName)
			}
			fields, itemSelectors := builder()

			for _, fName := range sm.SortedKeys(fields) {
				f := fields[fName]
				if !isBuiltinFormatter(f.Formatter) {
					// Custom formatter: SpecRef is a logical handle, not a
					// path. Layer 1 can't validate it; Layer 3 (set-everything
					// round-trip) will.
					continue
				}

				key := fieldKey{fieldGroup: fgName, field: fName}
				tracked[key] = true

				// First mode: resolve directly from each spec-selector output
				// (no item selector). This is how a rule that references the
				// field without an itemSelector evaluates.
				for _, root := range selOutputs {
					value, ok := walkSpecRef(root, f.SpecRef)
					if !ok {
						continue
					}
					resolved[key] = true
					if !valueMatchesType(value, f.Type) {
						report.TypeMismatch = append(report.TypeMismatch, typeMismatch{
							specSelector: selName,
							fieldGroup:   fgName,
							field:        fName,
							specRef:      f.SpecRef,
							wantType:     f.Type,
							got:          value,
						})
					}
				}

				// Second mode: for each compatible item selector, descend into
				// the items it produces and re-evaluate the SpecRef from each
				// item root. This is how a rule with itemSelector evaluates.
				for _, isName := range f.ItemSelectors {
					isq, found := itemSelectors[isName]
					if !found {
						return nil, fmt.Errorf("field group %q field %q references missing item selector %q", fgName, fName, isName)
					}

					isArgs := sampleArgs(isq.ArgsInfo, isq.ArgsEnum)
					for _, root := range selOutputs {
						itemOutputs, err := evalJQ(isq.Query, root, isArgs)
						if err != nil {
							return nil, fmt.Errorf("item selector %q (under %q): %w", isName, selName, err)
						}
						for _, item := range itemOutputs {
							value, ok := walkSpecRef(item, f.SpecRef)
							if !ok {
								continue
							}
							resolved[key] = true
							if !valueMatchesType(value, f.Type) {
								report.TypeMismatch = append(report.TypeMismatch, typeMismatch{
									specSelector: selName,
									itemSelector: isName,
									fieldGroup:   fgName,
									field:        fName,
									specRef:      f.SpecRef,
									wantType:     f.Type,
									got:          value,
								})
							}
						}
					}
				}
			}
		}
	}

	for k := range tracked {
		if !resolved[k] {
			report.NeverResolved = append(report.NeverResolved, k)
		}
	}
	sort.Slice(report.NeverResolved, func(i, j int) bool {
		if report.NeverResolved[i].fieldGroup != report.NeverResolved[j].fieldGroup {
			return report.NeverResolved[i].fieldGroup < report.NeverResolved[j].fieldGroup
		}
		return report.NeverResolved[i].field < report.NeverResolved[j].field
	})
	sort.Slice(report.TypeMismatch, func(i, j int) bool {
		a, b := report.TypeMismatch[i], report.TypeMismatch[j]
		if a.fieldGroup != b.fieldGroup {
			return a.fieldGroup < b.fieldGroup
		}
		if a.field != b.field {
			return a.field < b.field
		}
		if a.specSelector != b.specSelector {
			return a.specSelector < b.specSelector
		}
		return a.itemSelector < b.itemSelector
	})

	return report, nil
}

// evalJQ parses, compiles and runs a JQ query over input. args populates the
// $ARGS.named.* namespace; passing an empty map is fine even when the query
// does not reference any args.
func evalJQ(query string, input any, args map[string]any) ([]any, error) {
	parsed, err := gojq.Parse(strings.TrimSpace(query))
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	code, err := gojq.Compile(parsed, gojq.WithVariables([]string{"$ARGS"}))
	if err != nil {
		return nil, fmt.Errorf("compile: %w", err)
	}

	iter := code.Run(input, map[string]any{"named": args})
	var out []any
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if e, isErr := v.(error); isErr {
			return nil, fmt.Errorf("run: %w", e)
		}
		out = append(out, v)
	}
	return out, nil
}

// walkSpecRef descends the dotted spec-ref path from root. Returns (value, true)
// when every segment was found; (nil, false) the moment a segment is missing.
// A trimmed-empty path means "use root as-is" (currently unused but harmless).
func walkSpecRef(root any, specRef string) (any, bool) {
	if strings.TrimSpace(specRef) == "" {
		return root, true
	}
	cur := root
	for _, seg := range strings.Split(specRef, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, exists := m[seg]
		if !exists {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

// valueMatchesType is the per-MappingFieldType kind check. nil never matches:
// a leaf value of nil is treated the same as "field absent".
func valueMatchesType(v any, t api.MappingFieldType) bool {
	if v == nil {
		return false
	}
	switch t {
	case api.MAPPINGFIELDTYPE_STRING, api.MAPPINGFIELDTYPE_IPADDRESS, api.MAPPINGFIELDTYPE_QUANTITY:
		_, ok := v.(string)
		return ok
	case api.MAPPINGFIELDTYPE_BOOLEAN:
		_, ok := v.(bool)
		return ok
	case api.MAPPINGFIELDTYPE_INTEGER:
		return isJSONInteger(v)
	case api.MAPPINGFIELDTYPE_NUMBER:
		return isJSONNumber(v)
	case api.MAPPINGFIELDTYPE_STRINGS:
		return isHomogeneousArray(v, func(x any) bool { _, ok := x.(string); return ok })
	case api.MAPPINGFIELDTYPE_INTEGERS:
		return isHomogeneousArray(v, isJSONInteger)
	case api.MAPPINGFIELDTYPE_NUMBERS:
		return isHomogeneousArray(v, isJSONNumber)
	case api.MAPPINGFIELDTYPE_NAMES_ARRAY:
		// Array of objects with a non-empty .name string (e.g. imagePullSecrets,
		// schedulingGates). An empty array satisfies the type.
		return isHomogeneousArray(v, func(x any) bool {
			m, ok := x.(map[string]any)
			if !ok {
				return false
			}
			n, ok := m["name"].(string)
			return ok && n != ""
		})
	case api.MAPPINGFIELDTYPE_OBJECT:
		// "Object" is intentionally permissive — it covers nested maps and
		// (ad-hoc) array-shaped sub-objects like configMap.items.
		switch v.(type) {
		case map[string]any, []any:
			return true
		}
	}
	return false
}

// isJSONInteger accepts ints in their native form plus float64 with no
// fractional part (sigs.k8s.io/yaml routes through JSON, where every numeric
// scalar arrives as float64).
func isJSONInteger(v any) bool {
	switch n := v.(type) {
	case int, int32, int64:
		return true
	case float64:
		return n == float64(int64(n))
	}
	return false
}

func isJSONNumber(v any) bool {
	switch v.(type) {
	case int, int32, int64, float32, float64:
		return true
	}
	return false
}

func isHomogeneousArray(v any, elemOK func(any) bool) bool {
	s, ok := v.([]any)
	if !ok {
		return false
	}
	for _, x := range s {
		if !elemOK(x) {
			return false
		}
	}
	return true
}

func isBuiltinFormatter(name string) bool {
	return name == dbmodel.BuiltinFormatterKeyValue || name == dbmodel.BuiltinFormatterKeyValueMap
}

// sampleArgs returns a deterministic value for every declared arg. Strings
// are picked from the first enum option when one exists, and from a fixed
// "test" sentinel otherwise. Whenever the reference spec needs to satisfy a
// parameterised selector (e.g. `appContainerByName` matching a container
// named "test"), it must use the same sentinel.
//
// Returning a non-nil empty map is intentional: every spec selector and item
// selector is compiled with `WithVariables(["$ARGS"])` so we always pass an
// args object on Run, even when the query references nothing.
func sampleArgs(info map[string]api.MappingFieldType, enum map[string][]string) map[string]any {
	args := map[string]any{}
	for name, t := range info {
		args[name] = sampleValue(t, enum[name])
	}
	return args
}

func sampleValue(t api.MappingFieldType, enum []string) any {
	if len(enum) > 0 {
		return enum[0]
	}
	switch t {
	case api.MAPPINGFIELDTYPE_STRING, api.MAPPINGFIELDTYPE_IPADDRESS, api.MAPPINGFIELDTYPE_QUANTITY:
		return "test"
	case api.MAPPINGFIELDTYPE_BOOLEAN:
		return true
	case api.MAPPINGFIELDTYPE_INTEGER:
		return 1
	case api.MAPPINGFIELDTYPE_NUMBER:
		return 1.0
	case api.MAPPINGFIELDTYPE_STRINGS, api.MAPPINGFIELDTYPE_NAMES_ARRAY:
		return []any{"test"}
	case api.MAPPINGFIELDTYPE_INTEGERS:
		return []any{1}
	case api.MAPPINGFIELDTYPE_NUMBERS:
		return []any{1.0}
	}
	return nil
}

func loadReferenceSpecGinkgo(path string) map[string]any {
	abs, err := filepath.Abs(path)
	Expect(err).To(BeNil(), "resolve %s", path)

	bytes, err := os.ReadFile(abs)
	Expect(err).To(BeNil(), "read %s", abs)

	var out map[string]any
	Expect(yaml.Unmarshal(bytes, &out)).To(Succeed(), "unmarshal %s", abs)
	Expect(out).ToNot(BeEmpty(), "%s is empty", abs)
	return out
}

// formatReport renders a layer1Report into a single human-readable block.
// Returns the empty string when the report is clean, so callers can drive a
// pass/fail decision off `len(formatReport(r)) > 0`.
func formatReport(r *layer1Report) string {
	if len(r.NeverResolved) == 0 && len(r.TypeMismatch) == 0 {
		return ""
	}
	var b strings.Builder
	if len(r.NeverResolved) > 0 {
		b.WriteString("fields with builtin formatter that never resolved against the reference spec\n")
		b.WriteString("(typo in SpecRef? wrong spec selector binding? reference spec missing the field?)\n")
		for _, k := range r.NeverResolved {
			fmt.Fprintf(&b, "  - %s.%s\n", k.fieldGroup, k.field)
		}
	}
	if len(r.TypeMismatch) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("fields whose value at SpecRef does not match the declared MappingFieldType:\n")
		for _, m := range r.TypeMismatch {
			fmt.Fprintf(&b, "  - %s.%s [selector=%s itemSelector=%s] specRef=%s want=%s got=%T(%v)\n",
				m.fieldGroup, m.field, m.specSelector, m.itemSelector, m.specRef, m.wantType, m.got, m.got)
		}
	}
	return b.String()
}
