package services_crud

import (
	"context"
	"strings"
	"testing"

	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	clientsclusterservice "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/topology"
	serviceclusters "github.com/run-ai/runai/control-plane/workloads/pkg/clients/clusterservice"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type testAnnotater struct {
	annotations []workloadsApi.Annotation
}

func (a *testAnnotater) GetAnnotations() []workloadsApi.Annotation {
	return a.annotations
}

func (a *testAnnotater) SetAnnotations(annotations []workloadsApi.Annotation) {
	a.annotations = annotations
}

func TestSetActiveAnnotationReactivatesLastMatchingAnnotation(t *testing.T) {
	active := workloadsApi.NewAnnotation()
	active.SetName("example.com/placement")
	active.SetValue("old-active-value")

	excluded := workloadsApi.NewAnnotation()
	excluded.SetName("example.com/placement")
	excluded.SetValue("excluded-policy-value")
	excluded.SetExclude(true)

	annotations := &testAnnotater{annotations: []workloadsApi.Annotation{*active, *excluded}}
	setActiveAnnotation(annotations, "example.com/placement", "system-value")

	value, found := topology.ActiveAnnotationValue(annotations.GetAnnotations(), "example.com/placement")
	if !found {
		t.Fatal("expected the system annotation to be active")
	}
	if value != "system-value" {
		t.Fatalf("expected last matching annotation value %q, got %q", "system-value", value)
	}
	if annotations.annotations[1].GetExclude() {
		t.Fatal("expected the last matching annotation to be reactivated")
	}
	if annotations.annotations[0].GetValue() != "old-active-value" {
		t.Fatal("expected earlier duplicate annotation to remain unchanged")
	}
}

func TestUpdateWorkloadTopologyAnnotationsForFixedSizeSegmentedPyTorch(t *testing.T) {
	actualNodePools := setupSegmentTopologyMock(t)

	t.Run("adds workload and segment preferences from the selected node-pool topology", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAITopologyAnnotationKey, "have-topology")
		assertActiveAnnotation(t, spec, topology.KAISegmentSizeAnnotationKey, "2")
		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")
		assertActiveAnnotation(t, spec, topology.KAITopologyAnnotationPreferredKey, "level2")
	})

	t.Run("preserves explicit segment preferred placement", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(testWorkloadAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "level1"))

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level1")
		assertActiveAnnotation(t, spec, topology.KAITopologyAnnotationPreferredKey, "level2")
	})

	t.Run("preserves explicit segment required placement without adding preferred", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(testWorkloadAnnotation(topology.KAISegmentTopologyRequiredAnnotationKey, "level1"))

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyRequiredAnnotationKey, "level1")
		assertNoActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey)
		assertActiveAnnotation(t, spec, topology.KAITopologyAnnotationPreferredKey, "level2")
	})

	t.Run("preserves explicit workload preferred placement", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(testWorkloadAnnotation(topology.KAITopologyAnnotationPreferredKey, "level1"))

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAITopologyAnnotationPreferredKey, "level1")
		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")
	})

	t.Run("preserves explicit workload required placement without adding preferred", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(testWorkloadAnnotation(topology.KAITopologyAnnotationRequiredKey, "level1"))

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAITopologyAnnotationRequiredKey, "level1")
		assertNoActiveAnnotation(t, spec, topology.KAITopologyAnnotationPreferredKey)
		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")
	})

	t.Run("reactivates automatic preference when a policy preference is excluded", func(t *testing.T) {
		excluded := testWorkloadAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "level1")
		excluded.SetExclude(true)
		spec := newSegmentedPyTorchSpec(excluded)

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")
	})

	t.Run("replaces an empty explicit segment preference with the automatic preference", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(testWorkloadAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, ""))

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")
	})

	t.Run("treats an empty explicit segment requirement as absent", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(testWorkloadAnnotation(topology.KAISegmentTopologyRequiredAnnotationKey, ""))

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")
	})
}

func TestSegmentTopologyAutomationScope(t *testing.T) {
	actualNodePools := setupSegmentTopologyMock(t)

	t.Run("enriches segment intent for the validated native path", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{
			validated: true,
			method:    SubmissionMethodNative,
		})
		if err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")
	})

	t.Run("normalizes typed-only segment intent for the validated native path", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.Annotations = nil
		spec.SetSegmentSize(2)

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{
			validated: true,
			method:    SubmissionMethodNative,
		})
		if err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
		assertActiveAnnotation(t, spec, topology.KAISegmentSizeAnnotationKey, "2")
		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")
	})

	t.Run("accepts elastic PyTorch", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.SetMinReplicas(2)
		spec.SetMaxReplicas(4)

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAISegmentSizeAnnotationKey, "2")
		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")
	})

	t.Run("treats an empty segment size as malformed explicit intent", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.Annotations[0].SetValue("")

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "must be a positive integer")
	})

	t.Run("rejects non-PyTorch frameworks", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.SetDistributedFramework(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "supported only for PyTorch")
	})
}

func TestSegmentTopologyValidation(t *testing.T) {
	actualNodePools := setupSegmentTopologyMock(t)

	t.Run("canonicalizes an annotation segment size", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.Annotations[0].SetValue("02")

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAISegmentSizeAnnotationKey, "2")
		if !spec.HasSegmentSize() || spec.GetSegmentSize() != 2 {
			t.Fatalf("expected annotation segment size to be reflected in spec.segmentSize")
		}
	})

	t.Run("normalizes a typed segment size to the KAI annotation", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.Annotations = nil
		spec.SetSegmentSize(2)

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAISegmentSizeAnnotationKey, "2")
		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")
	})

	t.Run("accepts matching typed and annotation segment sizes", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.SetSegmentSize(2)

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
	})

	t.Run("rejects conflicting typed and annotation segment sizes", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.SetSegmentSize(3)

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "must match")
	})

	t.Run("accepts a segment size larger than the worker count as one partial segment", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.SetNumWorkers(3)
		spec.Annotations[0].SetValue("4")

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
	})

	t.Run("accepts a segment size larger than the default worker count", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.UnsetNumWorkers()

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
	})

	t.Run("accepts 10000 segments", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.SetNumWorkers(10_000)
		spec.Annotations[0].SetValue("1")

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
	})

	t.Run("rejects 10001 segments", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.SetNumWorkers(10_001)
		spec.Annotations[0].SetValue("1")

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "exceeding the maximum of 10000")
	})

	t.Run("accepts exactly the maximum number of segments without overflowing", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.SetNumWorkers(2_147_483_647)
		spec.Annotations[0].SetValue("214749")

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
	})

	t.Run("rejects more than the maximum number of segments without overflowing", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.SetNumWorkers(2_147_483_647)
		spec.Annotations[0].SetValue("214748")

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "exceeding the maximum of 10000")
	})

	t.Run("rejects a non-positive segment size", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.Annotations[0].SetValue("0")

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "must be a positive integer")
	})

	t.Run("rejects placement without segment size", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.Annotations = []workloadsApi.Annotation{
			testWorkloadAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "level2"),
		}

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "must be provided when segment placement is configured")
	})

	t.Run("rejects an unknown preferred level", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(testWorkloadAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "unknown"))

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "is not a level")
	})

	t.Run("rejects a preferred level broader than the required level", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(
			testWorkloadAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "level1"),
			testWorkloadAnnotation(topology.KAISegmentTopologyRequiredAnnotationKey, "level2"),
		)

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "must be the same as or lower")
	})

	t.Run("accepts a preferred level narrower than the required level", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(
			testWorkloadAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "level2"),
			testWorkloadAnnotation(topology.KAISegmentTopologyRequiredAnnotationKey, "level1"),
		)

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
	})

	t.Run("rejects a topology annotation conflicting with the selected node pools", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(testWorkloadAnnotation(topology.KAITopologyAnnotationKey, "another-topology"))

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "does not match the topology")
	})

	t.Run("treats an empty topology annotation as absent", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(testWorkloadAnnotation(topology.KAITopologyAnnotationKey, ""))

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}

		assertActiveAnnotation(t, spec, topology.KAITopologyAnnotationKey, "have-topology")
	})

	t.Run("rejects node pools without a topology", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.NodePools = []string{"without-topology"}
		pools := []clusterApi.Nodepool{{Name: "without-topology"}}

		err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), pools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "requires the selected node pools to have a network topology")
	})
}

func TestSegmentedNativeResponseSanitization(t *testing.T) {
	actualNodePools := setupSegmentTopologyMock(t)

	t.Run("typed request hides derived annotations from response and snapshot", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.Annotations = []workloadsApi.Annotation{
			testWorkloadAnnotation(topology.KAITopologyAnnotationPreferredKey, "level1"),
			testWorkloadAnnotation("example.com/unrelated", "kept"),
		}
		spec.SetSegmentSize(2)
		state := captureSegmentResponseSanitizationState(spec, nil)

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
		assertActiveAnnotation(t, spec, topology.KAITopologyAnnotationKey, "have-topology")
		assertActiveAnnotation(t, spec, topology.KAISegmentSizeAnnotationKey, "2")
		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")

		sanitizeSegmentedNativeResponse(spec, nil, state)

		if !spec.HasSegmentSize() || spec.GetSegmentSize() != 2 {
			t.Fatalf("expected canonical spec.segmentSize to remain 2")
		}
		assertNoActiveAnnotation(t, spec, topology.KAITopologyAnnotationKey)
		assertNoActiveAnnotation(t, spec, topology.KAISegmentSizeAnnotationKey)
		assertNoActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey)
		assertActiveAnnotation(t, spec, topology.KAITopologyAnnotationPreferredKey, "level1")
		assertActiveAnnotation(t, spec, "example.com/unrelated", "kept")

		params := &createWorkloadParams{Response: workloadsApi.SupersetWorkload{Spec: spec}}
		snapshot := NewWorkloadV2SnapshotFromWorkloadParams(params, nil, "runai-test")
		if !snapshot.Spec.HasSegmentSize() || snapshot.Spec.GetSegmentSize() != 2 {
			t.Fatalf("expected snapshot to retain canonical spec.segmentSize")
		}
		assertNoActiveAnnotation(t, snapshot.Spec, topology.KAISegmentSizeAnnotationKey)
		assertNoActiveAnnotation(t, snapshot.Spec, topology.KAITopologyAnnotationKey)
	})

	t.Run("annotation-only request is returned as the typed field", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		master := &workloadsApi.MasterSpec{
			Annotations: []workloadsApi.Annotation{
				testWorkloadAnnotation(topology.KAISegmentSizeAnnotationKey, "2"),
				testWorkloadAnnotation("example.com/master", "kept"),
			},
		}
		state := captureSegmentResponseSanitizationState(spec, master)

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
		sanitizeSegmentedNativeResponse(spec, master, state)

		if !spec.HasSegmentSize() || spec.GetSegmentSize() != 2 {
			t.Fatalf("expected annotation-only size to be normalized to spec.segmentSize")
		}
		assertNoActiveAnnotation(t, spec, topology.KAISegmentSizeAnnotationKey)
		assertNoActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey)
		assertNoActiveAnnotation(t, master, topology.KAISegmentSizeAnnotationKey)
		assertActiveAnnotation(t, master, "example.com/master", "kept")
	})

	t.Run("explicit segment and workload placement is preserved", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec(
			testWorkloadAnnotation(topology.KAITopologyAnnotationKey, "have-topology"),
			testWorkloadAnnotation(topology.KAITopologyAnnotationRequiredKey, "level1"),
			testWorkloadAnnotation(topology.KAISegmentTopologyPreferredAnnotationKey, "level2"),
			testWorkloadAnnotation(topology.KAISegmentTopologyRequiredAnnotationKey, "level1"),
		)
		state := captureSegmentResponseSanitizationState(spec, nil)

		if err := updateWorkloadTopologyAnnotationsForRouting(context.Background(), actualNodePools, spec, getSpecAnnotations(spec), segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
		sanitizeSegmentedNativeResponse(spec, nil, state)

		assertNoActiveAnnotation(t, spec, topology.KAISegmentSizeAnnotationKey)
		assertActiveAnnotation(t, spec, topology.KAITopologyAnnotationKey, "have-topology")
		assertActiveAnnotation(t, spec, topology.KAITopologyAnnotationRequiredKey, "level1")
		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyPreferredAnnotationKey, "level2")
		assertActiveAnnotation(t, spec, topology.KAISegmentTopologyRequiredAnnotationKey, "level1")
	})
}

func TestSegmentTopologyMasterAndWorkloadScope(t *testing.T) {
	actualNodePools := setupSegmentTopologyMock(t)

	t.Run("rejects segment annotations on a separately configured master", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.Annotations = nil
		master := &workloadsApi.MasterSpec{
			Annotations: []workloadsApi.Annotation{
				testWorkloadAnnotation(topology.KAISegmentSizeAnnotationKey, "2"),
			},
		}

		err := handleNodePoolsAndTopology(context.Background(), Distributed, actualNodePools, spec, master, false, segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "not on a separate master spec")
	})

	t.Run("rejects segment annotations on a non-PyTorch distributed workload", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.SetDistributedFramework(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)

		err := handleNodePoolsAndTopology(context.Background(), Distributed, actualNodePools, spec, nil, false, segmentedDistributedRoutingDecision{validated: true})
		assertTopologyErrorContains(t, err, "supported only for PyTorch distributed workloads")
	})

	t.Run("detects segment intent configured only on a non-PyTorch master", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.Annotations = nil
		spec.SetDistributedFramework(workloadsApi.DISTRIBUTEDFRAMEWORK_MPI)
		master := &workloadsApi.MasterSpec{
			Annotations: []workloadsApi.Annotation{
				testWorkloadAnnotation(topology.KAISegmentSizeAnnotationKey, "2"),
			},
		}

		err := handleNodePoolsAndTopology(context.Background(), Distributed, actualNodePools, spec, master, false, segmentedDistributedRoutingDecision{})
		assertTopologyErrorContains(t, err, "supported only for PyTorch distributed workloads")
	})

	t.Run("allows copied worker annotations when master is same as worker", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		master := &workloadsApi.MasterSpec{
			Annotations: []workloadsApi.Annotation{
				testWorkloadAnnotation(topology.KAISegmentSizeAnnotationKey, "2"),
			},
		}

		if err := handleNodePoolsAndTopology(context.Background(), Distributed, actualNodePools, spec, master, true, segmentedDistributedRoutingDecision{validated: true}); err != nil {
			t.Fatalf("unexpected topology enrichment error: %v", err)
		}
	})

	t.Run("rejects segment annotations on unsupported native workload types", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()

		err := handleNodePoolsAndTopology(context.Background(), DistributedInference, actualNodePools, spec, nil, false, segmentedDistributedRoutingDecision{})
		assertTopologyErrorContains(t, err, "supported only for PyTorch distributed workloads")
	})

	t.Run("rejects typed segment size on unsupported native workload types", func(t *testing.T) {
		spec := newSegmentedPyTorchSpec()
		spec.Annotations = nil
		spec.SetSegmentSize(2)

		err := handleNodePoolsAndTopology(context.Background(), Training, actualNodePools, spec, nil, false, segmentedDistributedRoutingDecision{})
		assertTopologyErrorContains(t, err, "supported only for PyTorch distributed workloads")
	})

	t.Run("rejects segment annotations nested on a distributed inference leader", func(t *testing.T) {
		spec := &workloadsApi.SupersetSpec{}
		spec.SetLeader(workloadsApi.DistributedInferenceLeaderWorkerSpec{
			Annotations: []workloadsApi.Annotation{
				testWorkloadAnnotation(topology.KAISegmentSizeAnnotationKey, "2"),
			},
		})

		err := handleNodePoolsAndTopology(context.Background(), DistributedInference, actualNodePools, spec, nil, false, segmentedDistributedRoutingDecision{})
		assertTopologyErrorContains(t, err, "supported only for PyTorch distributed workloads")
	})

	t.Run("rejects segment annotations nested on a distributed inference worker", func(t *testing.T) {
		spec := &workloadsApi.SupersetSpec{}
		spec.SetWorker(workloadsApi.DistributedInferenceLeaderWorkerSpec{
			Annotations: []workloadsApi.Annotation{
				testWorkloadAnnotation(topology.KAISegmentSizeAnnotationKey, "2"),
			},
		})

		err := handleNodePoolsAndTopology(context.Background(), DistributedInference, actualNodePools, spec, nil, false, segmentedDistributedRoutingDecision{})
		assertTopologyErrorContains(t, err, "supported only for PyTorch distributed workloads")
	})
}

func setupSegmentTopologyMock(t *testing.T) []clusterApi.Nodepool {
	t.Helper()
	const clusterID = "segment-topology-test-cluster"
	mockClient := clientsclusterservice.NewSimpleMock().
		WithNodePoolsWithTopology(clusterID, "have-topology", "with-topology").
		WithTopology("have-topology", &clusterApi.NetworkTopology{
			Name:   "have-topology",
			Levels: []string{"level1", "level2"},
		})
	serviceclusters.SetMock(mockClient)

	response, err := mockClient.GetNodePoolsInternalByCluster(context.Background(), clusterID)
	if err != nil {
		t.Fatalf("failed to build node-pool fixture: %v", err)
	}
	return response.Nodepools
}

func newSegmentedPyTorchSpec(additionalAnnotations ...workloadsApi.Annotation) *workloadsApi.SupersetSpec {
	framework := workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH
	numWorkers := int32(4)
	annotations := []workloadsApi.Annotation{testWorkloadAnnotation(topology.KAISegmentSizeAnnotationKey, "2")}
	annotations = append(annotations, additionalAnnotations...)
	return &workloadsApi.SupersetSpec{
		DistributedFramework: *workloadsApi.NewNullableDistributedFramework(&framework),
		NumWorkers:           *workloadsApi.NewNullableInt32(&numWorkers),
		NodePools:            []string{"with-topology"},
		Annotations:          annotations,
	}
}

func assertTopologyErrorContains(t *testing.T, err *workloadsApi.Error, expected string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected topology error containing %q", expected)
	}
	if !strings.Contains(err.Message, expected) {
		t.Fatalf("expected topology error containing %q, got %q", expected, err.Message)
	}
}

func testWorkloadAnnotation(name, value string) workloadsApi.Annotation {
	annotation := workloadsApi.NewAnnotation()
	annotation.SetName(name)
	annotation.SetValue(value)
	return *annotation
}

func assertActiveAnnotation(t *testing.T, annotations Annotater, name, expectedValue string) {
	t.Helper()
	value, found := topology.ActiveAnnotationValue(annotations.GetAnnotations(), name)
	if !found {
		t.Fatalf("expected active annotation %q", name)
	}
	if value != expectedValue {
		t.Fatalf("expected annotation %q value %q, got %q", name, expectedValue, value)
	}
}

func assertNoActiveAnnotation(t *testing.T, annotations Annotater, name string) {
	t.Helper()
	if value, found := topology.ActiveAnnotationValue(annotations.GetAnnotations(), name); found {
		t.Fatalf("expected annotation %q to be absent, got %q", name, value)
	}
}
