package services

import (
	"context"
	"testing"

	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	policyclusters "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/topology"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/clients/clusters"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPyTorchSegmentTopologyEnrichment(t *testing.T) {
	const (
		clusterID    = "segment-cluster"
		projectID    = "segment-project"
		topologyName = "cluster-topology"
		lowestLevel  = "rack"
	)

	newProject := func(nodePools ...string) *orgUnitApi.Project {
		return &orgUnitApi.Project{
			Id:               projectID,
			Name:             projectID,
			ClusterId:        clusterID,
			DefaultNodePools: nodePools,
		}
	}
	setTopologyMock := func(nodePools ...string) {
		t.Helper()
		client := policyclusters.NewSimpleMock().
			WithNodePoolsWithTopology(clusterID, topologyName, nodePools...).
			WithTopology(topologyName, &clusterApi.NetworkTopology{
				Name:   topologyName,
				Levels: []string{"zone", lowestLevel},
			})
		clusters.SetMock(client)
	}
	workerAnnotations := func(t *testing.T, manifest map[string]interface{}) map[string]string {
		t.Helper()
		annotations, found, err := unstructured.NestedStringMap(manifest, pyTorchWorkerTemplateAnnotationsPath...)
		require.NoError(t, err)
		require.True(t, found)
		return annotations
	}
	rootAnnotations := func(t *testing.T, manifest map[string]interface{}) map[string]interface{} {
		t.Helper()
		metadata, ok := manifest["metadata"].(map[string]interface{})
		require.True(t, ok)
		annotations, ok := metadata["annotations"].(map[string]interface{})
		require.True(t, ok)
		return annotations
	}

	t.Run("adds root and Worker defaults to the effective manifest only", func(t *testing.T) {
		setTopologyMock("pool-a")
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})

		effective, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.NoError(t, err)

		root := rootAnnotations(t, effective)
		require.Equal(t, topologyName, root[topology.KAITopologyAnnotationKey])
		require.Equal(t, lowestLevel, root[topology.KAITopologyAnnotationPreferredKey])
		require.Equal(t, lowestLevel, workerAnnotations(t, effective)[topology.KAISegmentTopologyPreferredAnnotationKey])

		require.NotContains(t, workerAnnotations(t, manifest), topology.KAISegmentTopologyPreferredAnnotationKey)
		originalMetadata := manifest["metadata"].(map[string]interface{})
		require.NotContains(t, originalMetadata, "annotations")
	})

	t.Run("uses the Worker manifest node pool instead of the project default", func(t *testing.T) {
		const (
			projectTopology  = "project-topology"
			manifestTopology = "manifest-topology"
			manifestLevel    = "block"
			nodePoolLabelKey = "node-pool.example.com/type"
		)
		setDetailedTopologyMock(
			clusterID,
			[]clusterApi.Nodepool{
				newTopologyNodePool("pool-a", nodePoolLabelKey, "a", projectTopology),
				newTopologyNodePool("pool-b", nodePoolLabelKey, "b", manifestTopology),
			},
			map[string][]string{
				projectTopology:  {"zone", "rack"},
				manifestTopology: {"zone", manifestLevel},
			},
		)
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
			commonconstants.AnnotationNodepools:  "  pool-b  ",
		})
		setPyTorchReplicaAffinity(t, manifest, pyTorchReplicaTypeWorker, requiredNodeAffinity(
			corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{{
				Key:      nodePoolLabelKey,
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"b"},
			}}},
		))

		effective, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.NoError(t, err)

		root := rootAnnotations(t, effective)
		require.Equal(t, manifestTopology, root[topology.KAITopologyAnnotationKey])
		require.Equal(t, manifestLevel, root[topology.KAITopologyAnnotationPreferredKey])
		effectiveWorker := workerAnnotations(t, effective)
		require.Equal(t, manifestLevel, effectiveWorker[topology.KAISegmentTopologyPreferredAnnotationKey])
		require.Equal(t, "pool-b", effectiveWorker[commonconstants.AnnotationNodepools])
		originalWorker := workerAnnotations(t, manifest)
		require.Equal(t, "  pool-b  ", originalWorker[commonconstants.AnnotationNodepools])
		require.NotContains(t, originalWorker, topology.KAISegmentTopologyPreferredAnnotationKey)
	})

	t.Run("accepts an explicit Master pool backed by the Worker topology", func(t *testing.T) {
		const (
			projectTopology  = "project-topology"
			manifestTopology = "manifest-topology"
			nodePoolLabelKey = "node-pool.example.com/type"
		)
		setDetailedTopologyMock(
			clusterID,
			[]clusterApi.Nodepool{
				newTopologyNodePool("pool-a", nodePoolLabelKey, "a", projectTopology),
				newTopologyNodePool("pool-b", nodePoolLabelKey, "b", manifestTopology),
				newTopologyNodePool("pool-c", nodePoolLabelKey, "c", manifestTopology),
			},
			map[string][]string{
				projectTopology:  {"zone", "rack"},
				manifestTopology: {"zone", "block"},
			},
		)
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
		setPyTorchReplicaLabels(t, manifest, pyTorchReplicaTypeWorker, map[string]string{clusterconstants.NodePoolNameLabel: "pool-b"})
		setPyTorchReplicaTemplate(t, manifest, pyTorchReplicaTypeMaster)
		setPyTorchReplicaLabels(t, manifest, pyTorchReplicaTypeMaster, map[string]string{clusterconstants.NodePoolNameLabel: "pool-c"})

		effective, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.NoError(t, err)
		require.Equal(t, manifestTopology, rootAnnotations(t, effective)[topology.KAITopologyAnnotationKey])
	})

	t.Run("rejects a Master fallback backed by a different topology", func(t *testing.T) {
		const nodePoolLabelKey = "node-pool.example.com/type"
		setDetailedTopologyMock(
			clusterID,
			[]clusterApi.Nodepool{
				newTopologyNodePool("pool-a", nodePoolLabelKey, "a", "project-topology"),
				newTopologyNodePool("pool-b", nodePoolLabelKey, "b", "worker-topology"),
			},
			map[string][]string{
				"project-topology": {"zone", "rack"},
				"worker-topology":  {"zone", "block"},
			},
		)
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
		setPyTorchReplicaLabels(t, manifest, pyTorchReplicaTypeWorker, map[string]string{clusterconstants.NodePoolNameLabel: "pool-b"})
		setPyTorchReplicaTemplate(t, manifest, pyTorchReplicaTypeMaster)

		_, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.ErrorContains(t, err, "same topology")
	})

	t.Run("preserves explicit segment placement", func(t *testing.T) {
		setTopologyMock("pool-a")
		for _, test := range []struct {
			name        string
			annotation  string
			value       string
			wantDefault bool
		}{
			{name: "preferred", annotation: topology.KAISegmentTopologyPreferredAnnotationKey, value: "zone"},
			{name: "required", annotation: topology.KAISegmentTopologyRequiredAnnotationKey, value: "zone"},
			{name: "empty preferred is unset", annotation: topology.KAISegmentTopologyPreferredAnnotationKey, value: "", wantDefault: true},
		} {
			t.Run(test.name, func(t *testing.T) {
				manifest := newPyTorchJobManifest(map[string]interface{}{
					topology.KAISegmentSizeAnnotationKey: "2",
					test.annotation:                      test.value,
				})

				effective, err := NewK8sResourceUtil().UpdateResourceMetadata(
					context.Background(), manifest, "workload-id", newProject("pool-a"),
				)
				require.NoError(t, err)
				require.Equal(t, lowestLevel, rootAnnotations(t, effective)[topology.KAITopologyAnnotationPreferredKey])

				worker := workerAnnotations(t, effective)
				if test.wantDefault {
					require.Equal(t, lowestLevel, worker[topology.KAISegmentTopologyPreferredAnnotationKey])
				} else {
					require.Equal(t, test.value, worker[test.annotation])
					if test.annotation == topology.KAISegmentTopologyRequiredAnnotationKey {
						require.NotContains(t, worker, topology.KAISegmentTopologyPreferredAnnotationKey)
					}
				}
			})
		}
	})

	t.Run("adds root and segment placement when only the topology name is set", func(t *testing.T) {
		setTopologyMock("pool-a")
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
		manifest["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
			topology.KAITopologyAnnotationKey: topologyName,
		}

		effective, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.NoError(t, err)

		root := rootAnnotations(t, effective)
		require.Equal(t, topologyName, root[topology.KAITopologyAnnotationKey])
		require.Equal(t, lowestLevel, root[topology.KAITopologyAnnotationPreferredKey])
		require.Equal(t, lowestLevel, workerAnnotations(t, effective)[topology.KAISegmentTopologyPreferredAnnotationKey])
	})

	t.Run("treats empty root placement annotations as unset", func(t *testing.T) {
		setTopologyMock("pool-a")
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
		manifest["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
			topology.KAITopologyAnnotationKey:          topologyName,
			topology.KAITopologyAnnotationPreferredKey: "",
			topology.KAITopologyAnnotationRequiredKey:  "",
		}

		effective, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.NoError(t, err)

		root := rootAnnotations(t, effective)
		require.Equal(t, lowestLevel, root[topology.KAITopologyAnnotationPreferredKey])
		require.Equal(t, lowestLevel, workerAnnotations(t, effective)[topology.KAISegmentTopologyPreferredAnnotationKey])
	})

	t.Run("preserves a complete trusted native manifest when project defaults use another topology", func(t *testing.T) {
		clusters.SetMock(policyclusters.NewSimpleMock().
			WithNodePoolsWithTopology(clusterID, "project-default-topology", "project-default-pool"))
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAITopologyAnnotationKey:                 "workload-topology",
			topology.KAISegmentSizeAnnotationKey:              "2",
			topology.KAISegmentTopologyPreferredAnnotationKey: "workload-rack",
		})
		manifest["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
			topology.KAITopologyAnnotationKey:          "workload-topology",
			topology.KAITopologyAnnotationPreferredKey: "workload-rack",
		}

		effective, err := NewK8sResourceUtil().updateResourceMetadataWithOptions(
			context.Background(),
			manifest,
			"workload-id",
			newProject("project-default-pool"),
			resourceMetadataUpdateOptions{trustedConvertedFromNative: true},
		)
		require.NoError(t, err)

		root := rootAnnotations(t, effective)
		require.Equal(t, "workload-topology", root[topology.KAITopologyAnnotationKey])
		require.Equal(t, "workload-rack", root[topology.KAITopologyAnnotationPreferredKey])
		worker := workerAnnotations(t, effective)
		require.Equal(t, "workload-topology", worker[topology.KAITopologyAnnotationKey])
		require.Equal(t, "workload-rack", worker[topology.KAISegmentTopologyPreferredAnnotationKey])
	})

	t.Run("rejects an incomplete trusted native segmented manifest", func(t *testing.T) {
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})

		_, err := NewK8sResourceUtil().updateResourceMetadataWithOptions(
			context.Background(),
			manifest,
			"workload-id",
			newProject("project-default-pool"),
			resourceMetadataUpdateOptions{trustedConvertedFromNative: true},
		)
		require.ErrorContains(t, err, "converted native segmented PyTorch manifest is missing")
	})

	t.Run("adds a missing topology name without overriding root preferred placement", func(t *testing.T) {
		setTopologyMock("pool-a")
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
		manifest["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
			topology.KAITopologyAnnotationPreferredKey: "zone",
		}

		effective, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.NoError(t, err)

		root := rootAnnotations(t, effective)
		require.Equal(t, topologyName, root[topology.KAITopologyAnnotationKey])
		require.Equal(t, "zone", root[topology.KAITopologyAnnotationPreferredKey])
		require.Equal(t, lowestLevel, workerAnnotations(t, effective)[topology.KAISegmentTopologyPreferredAnnotationKey])
	})

	t.Run("adds a missing topology name without overriding root required placement", func(t *testing.T) {
		setTopologyMock("pool-a")
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
		manifest["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
			topology.KAITopologyAnnotationRequiredKey: "zone",
		}

		effective, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.NoError(t, err)

		root := rootAnnotations(t, effective)
		require.Equal(t, topologyName, root[topology.KAITopologyAnnotationKey])
		require.Equal(t, "zone", root[topology.KAITopologyAnnotationRequiredKey])
		require.NotContains(t, root, topology.KAITopologyAnnotationPreferredKey)
		require.Equal(t, lowestLevel, workerAnnotations(t, effective)[topology.KAISegmentTopologyPreferredAnnotationKey])
	})

	t.Run("does not enrich a lookalike unsupported GVK", func(t *testing.T) {
		setTopologyMock("pool-a")
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
		manifest["kind"] = "TFJob"

		effective, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.NoError(t, err)
		require.NotContains(t, workerAnnotations(t, effective), topology.KAISegmentTopologyPreferredAnnotationKey)
	})

	t.Run("does not enrich an unsupported PyTorchJob API version", func(t *testing.T) {
		setTopologyMock("pool-a")
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
		manifest["apiVersion"] = "kubeflow.org/v2"

		effective, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.NoError(t, err)
		require.NotContains(t, workerAnnotations(t, effective), topology.KAISegmentTopologyPreferredAnnotationKey)
	})

	t.Run("rejects a topology annotation that conflicts with selected node pools", func(t *testing.T) {
		setTopologyMock("pool-a")
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
		manifest["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
			topology.KAITopologyAnnotationKey: "another-topology",
		}

		_, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.ErrorContains(t, err, "does not match the topology")
	})

	t.Run("rejects selected node pools with different topologies", func(t *testing.T) {
		client := policyclusters.NewSimpleMock().
			WithNodePoolsWithTopology(clusterID, "topology-a", "pool-a").
			WithNodePoolsWithTopology(clusterID, "topology-b", "pool-b")
		clusters.SetMock(client)
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})

		_, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a", "pool-b"),
		)
		require.ErrorContains(t, err, "same topology")
	})

	t.Run("validates project node pools even when a public manifest is fully annotated", func(t *testing.T) {
		client := policyclusters.NewSimpleMock().
			WithNodePoolsWithTopology(clusterID, "topology-a", "pool-a").
			WithNodePoolsWithTopology(clusterID, "topology-b", "pool-b")
		clusters.SetMock(client)
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAITopologyAnnotationKey:                 "topology-a",
			topology.KAISegmentSizeAnnotationKey:              "2",
			topology.KAISegmentTopologyPreferredAnnotationKey: "rack",
		})
		manifest["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{
			topology.KAITopologyAnnotationKey:          "topology-a",
			topology.KAITopologyAnnotationPreferredKey: "rack",
		}
		manifest["metadata"].(map[string]interface{})["labels"] = map[string]interface{}{
			ConvertedFromNativeLabel: TrueAsString,
		}

		_, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a", "pool-b"),
		)
		require.ErrorContains(t, err, "same topology")
	})

	t.Run("rejects selected node pools without a network topology", func(t *testing.T) {
		clusters.SetMock(policyclusters.NewSimpleMock().WithNodePools(clusterID, "pool-a"))
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})

		_, err := NewK8sResourceUtil().UpdateResourceMetadata(
			context.Background(), manifest, "workload-id", newProject("pool-a"),
		)
		require.ErrorContains(t, err, "requires the selected node pools to have a network topology")
	})
}

func TestRequestedPyTorchWorkerNodePools(t *testing.T) {
	const nodePoolLabelKey = "node-pool.example.com/type"
	deleting := clusterApi.NODEPOOLPHASE_DELETING
	actualNodePools := []clusterApi.Nodepool{
		{Name: composer.DefaultNodePoolName},
		{Name: "pool-a", LabelKey: nodePoolLabelKey, LabelValue: "a"},
		{Name: "pool-b", LabelKey: nodePoolLabelKey, LabelValue: "b"},
		{Name: "pool-c", LabelKey: "other-node-pool.example.com/type", LabelValue: "c"},
		{Name: "pool-deleting", LabelKey: nodePoolLabelKey, LabelValue: "deleting", Phase: &deleting},
	}
	fallbackNodePools := []string{"pool-a"}
	newManifest := func() map[string]interface{} {
		return newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
	}

	t.Run("annotation wins and normalizes the requested pools", func(t *testing.T) {
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
			commonconstants.AnnotationNodepools:  "  pool-b   pool-a pool-b ",
		})
		setPyTorchReplicaAffinity(t, manifest, pyTorchReplicaTypeWorker, requiredNodeAffinity(
			corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{{
				Key:      nodePoolLabelKey,
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"a", "b"},
			}}},
		))

		requested, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.NoError(t, err)
		require.Equal(t, []string{"pool-b", "pool-a"}, requested)
		annotations, found, err := unstructured.NestedStringMap(manifest, pyTorchWorkerTemplateAnnotationsPath...)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, "pool-b pool-a", annotations[commonconstants.AnnotationNodepools])
	})

	t.Run("rejects an annotation that disagrees with effective pod placement", func(t *testing.T) {
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
			commonconstants.AnnotationNodepools:  "pool-b",
		})

		_, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.ErrorContains(t, err, "conflicts with effective pod placement")
	})

	t.Run("label selects a pool", func(t *testing.T) {
		manifest := newManifest()
		setPyTorchReplicaLabels(t, manifest, pyTorchReplicaTypeWorker, map[string]string{clusterconstants.NodePoolNameLabel: "pool-b"})

		requested, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.NoError(t, err)
		require.Equal(t, []string{"pool-b"}, requested)
	})

	t.Run("label replaces required affinity on the same node label key", func(t *testing.T) {
		manifest := newManifest()
		setPyTorchReplicaLabels(t, manifest, pyTorchReplicaTypeWorker, map[string]string{clusterconstants.NodePoolNameLabel: "pool-b"})
		setPyTorchReplicaAffinity(t, manifest, pyTorchReplicaTypeWorker, requiredNodeAffinity(
			corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{{
				Key:      nodePoolLabelKey,
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"a"},
			}}},
		))

		requested, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.NoError(t, err)
		require.Equal(t, []string{"pool-b"}, requested)
	})

	t.Run("rejects a later different-key node-pool expression retained by label mutation", func(t *testing.T) {
		manifest := newManifest()
		setPyTorchReplicaLabels(t, manifest, pyTorchReplicaTypeWorker, map[string]string{clusterconstants.NodePoolNameLabel: "pool-b"})
		setPyTorchReplicaAffinity(t, manifest, pyTorchReplicaTypeWorker, requiredNodeAffinity(
			corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{
				{Key: nodePoolLabelKey, Operator: corev1.NodeSelectorOpIn, Values: []string{"a"}},
				{Key: "other-node-pool.example.com/type", Operator: corev1.NodeSelectorOpIn, Values: []string{"c"}},
			}},
		))

		_, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.ErrorContains(t, err, "conflicting node-pool label")
	})

	t.Run("resolves and deduplicates required affinity across OR terms", func(t *testing.T) {
		manifest := newManifest()
		setPyTorchReplicaAffinity(t, manifest, pyTorchReplicaTypeWorker, requiredNodeAffinity(
			corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{
				{Key: "unrelated.example.com/key", Operator: corev1.NodeSelectorOpIn, Values: []string{"value"}},
				{Key: nodePoolLabelKey, Operator: corev1.NodeSelectorOpIn, Values: []string{"a", "b"}},
			}},
			corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{{
				Key:      nodePoolLabelKey,
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"b"},
			}}},
		))

		requested, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.NoError(t, err)
		require.Equal(t, []string{"pool-a", "pool-b"}, requested)
	})

	t.Run("uses only the first recognized node-pool expression in a term", func(t *testing.T) {
		manifest := newManifest()
		setPyTorchReplicaAffinity(t, manifest, pyTorchReplicaTypeWorker, requiredNodeAffinity(
			corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{
				{Key: nodePoolLabelKey, Operator: corev1.NodeSelectorOpIn, Values: []string{"a"}},
				{Key: nodePoolLabelKey, Operator: corev1.NodeSelectorOpIn, Values: []string{"b"}},
			}},
		))

		requested, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.NoError(t, err)
		require.Equal(t, []string{"pool-a"}, requested)
	})

	t.Run("recognizes the default pool DoesNotExist expression", func(t *testing.T) {
		manifest := newManifest()
		setPyTorchReplicaAffinity(t, manifest, pyTorchReplicaTypeWorker, requiredNodeAffinity(
			corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{{
				Key:      clusterconstants.NodePoolNameLabel,
				Operator: corev1.NodeSelectorOpDoesNotExist,
			}}},
		))

		requested, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.NoError(t, err)
		require.Equal(t, []string{composer.DefaultNodePoolName}, requested)
	})

	t.Run("ignores deleting pools in affinity", func(t *testing.T) {
		manifest := newManifest()
		setPyTorchReplicaAffinity(t, manifest, pyTorchReplicaTypeWorker, requiredNodeAffinity(
			corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{{
				Key:      nodePoolLabelKey,
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"deleting"},
			}}},
			corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{{
				Key:      nodePoolLabelKey,
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"a"},
			}}},
		))

		requested, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.NoError(t, err)
		require.Equal(t, []string{"pool-a"}, requested)
	})

	t.Run("rejects an explicitly selected deleting pool", func(t *testing.T) {
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
			commonconstants.AnnotationNodepools:  "pool-deleting",
		})

		_, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.ErrorContains(t, err, "deleting node pool")
	})

	t.Run("unrelated affinity and the sentinel label use the fallback", func(t *testing.T) {
		manifest := newManifest()
		setPyTorchReplicaLabels(t, manifest, pyTorchReplicaTypeWorker, map[string]string{clusterconstants.NodePoolNameLabel: unexistingNodePoolLabelValue})
		setPyTorchReplicaAffinity(t, manifest, pyTorchReplicaTypeWorker, requiredNodeAffinity(
			corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{{
				Key:      "unrelated.example.com/key",
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"value"},
			}}},
		))

		requested, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.NoError(t, err)
		require.Equal(t, fallbackNodePools, requested)
	})

	t.Run("rejects malformed affinity", func(t *testing.T) {
		manifest := newManifest()
		require.NoError(t, unstructured.SetNestedField(
			manifest,
			"not-an-object",
			pyTorchReplicaTemplatePath(pyTorchReplicaTypeWorker, "spec", "affinity")...,
		))

		_, err := requestedPyTorchWorkerNodePools(manifest, actualNodePools, fallbackNodePools)
		require.ErrorContains(t, err, "affinity must be an object")
	})
}

func TestPyTorchSegmentTopologyClusterCapability(t *testing.T) {
	const clusterID = "segment-capability-cluster"
	manifest := newPyTorchJobManifest(map[string]interface{}{
		topology.KAISegmentSizeAnnotationKey: "2",
	})
	impl := &WorkloadsV2Impl{k8sUtil: NewK8sResourceUtil()}

	setClusterVersion := func(version string) {
		t.Helper()
		cluster := clusterApi.DisplayedCluster{Uuid: clusterID}
		cluster.Version.Set(&version)
		checkerClusterClient.SetMock(
			checkerClusterClient.NewSimpleMockClient().WithCluster(clusterID, cluster),
		)
		checker.InitClusterCapabilitiesChecker("", &checker.ClustersCacheConfig{
			ExpirationSec: 60,
			CleanupSec:    60,
		}, "")
	}

	t.Run("rejects clusters older than 2.27", func(t *testing.T) {
		setClusterVersion("2.26.9")

		submissionErr := impl.validateSegmentTopologyClusterSupport(context.Background(), clusterID, manifest)
		require.NotNil(t, submissionErr)
		require.EqualValues(t, 400, submissionErr.Code)
		require.Contains(t, submissionErr.Message, "requires cluster version 2.27.0 or later")
		require.Len(t, submissionErr.ComplianceIssues, 1)
		require.Equal(t, ErrorFieldCluster, submissionErr.ComplianceIssues[0].Field)
	})

	t.Run("accepts cluster 2.27", func(t *testing.T) {
		setClusterVersion("2.27.0")
		require.Nil(t, impl.validateSegmentTopologyClusterSupport(context.Background(), clusterID, manifest))
	})

	t.Run("does not gate a PyTorchJob without segment intent", func(t *testing.T) {
		setClusterVersion("2.26.9")
		unsegmented := newPyTorchJobManifest(map[string]interface{}{})
		require.Nil(t, impl.validateSegmentTopologyClusterSupport(context.Background(), clusterID, unsegmented))
	})

	t.Run("reports a capability lookup failure as a cluster validation error", func(t *testing.T) {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient())
		checker.InitClusterCapabilitiesChecker("", &checker.ClustersCacheConfig{
			ExpirationSec: 60,
			CleanupSec:    60,
		}, "")

		submissionErr := impl.validateSegmentTopologyClusterSupport(context.Background(), clusterID, manifest)
		require.NotNil(t, submissionErr)
		require.EqualValues(t, 400, submissionErr.Code)
		require.Contains(t, submissionErr.Message, "failed to verify cluster support")
	})
}

func TestPyTorchSegmentTopologyValidation(t *testing.T) {
	k8sUtil := NewK8sResourceUtil()

	t.Run("rejects invalid segment sizes", func(t *testing.T) {
		tests := []struct {
			name        string
			segmentSize string
		}{
			{name: "empty", segmentSize: ""},
			{name: "non-numeric", segmentSize: "abc"},
			{name: "zero", segmentSize: "0"},
			{name: "negative", segmentSize: "-5"},
			{name: "int32 overflow", segmentSize: "2147483648"},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				manifest := newPyTorchJobManifest(map[string]interface{}{
					topology.KAISegmentSizeAnnotationKey: test.segmentSize,
				})

				target, err := k8sUtil.pyTorchSegmentAnnotationTarget(manifest)
				require.Nil(t, target)
				require.ErrorContains(t, err, "must be a positive integer")
			})
		}
	})

	t.Run("does not treat empty placement without segment size as intent", func(t *testing.T) {
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentTopologyPreferredAnnotationKey: "",
		})

		target, err := k8sUtil.pyTorchSegmentAnnotationTarget(manifest)
		require.NoError(t, err)
		require.Nil(t, target)
	})

	t.Run("rejects segment placement without segment size", func(t *testing.T) {
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentTopologyPreferredAnnotationKey: "rack",
		})

		target, err := k8sUtil.pyTorchSegmentAnnotationTarget(manifest)
		require.Nil(t, target)
		require.ErrorContains(t, err, "must be provided when segment placement is configured")
	})

	t.Run("accepts elastic PyTorch", func(t *testing.T) {
		tests := []struct {
			name          string
			elasticPolicy map[string]interface{}
		}{
			{name: "empty policy", elasticPolicy: map[string]interface{}{}},
			{name: "replica bounds", elasticPolicy: map[string]interface{}{
				"minReplicas": float64(2),
				"maxReplicas": float64(4),
			}},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				manifest := newPyTorchJobManifest(map[string]interface{}{
					topology.KAISegmentSizeAnnotationKey: "2",
				})
				require.NoError(t, unstructured.SetNestedMap(
					manifest,
					test.elasticPolicy,
					"spec",
					"elasticPolicy",
				))

				target, err := k8sUtil.pyTorchSegmentAnnotationTarget(manifest)
				require.NoError(t, err)
				require.NotNil(t, target)
			})
		}
	})

	t.Run("accepts the maximum segment count", func(t *testing.T) {
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "1",
		})
		require.NoError(t, unstructured.SetNestedField(
			manifest,
			float64(10_000),
			pyTorchReplicaSpecPath(pyTorchReplicaTypeWorker, "replicas")...,
		))

		target, err := k8sUtil.pyTorchSegmentAnnotationTarget(manifest)
		require.NoError(t, err)
		require.NotNil(t, target)
	})

	t.Run("rejects more than the maximum segment count", func(t *testing.T) {
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "1",
		})
		require.NoError(t, unstructured.SetNestedField(
			manifest,
			float64(10_001),
			pyTorchReplicaSpecPath(pyTorchReplicaTypeWorker, "replicas")...,
		))

		target, err := k8sUtil.pyTorchSegmentAnnotationTarget(manifest)
		require.Nil(t, target)
		require.ErrorContains(t, err, "exceeding the maximum of 10000")
	})

	t.Run("defaults missing Worker replicas to one", func(t *testing.T) {
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
		unstructured.RemoveNestedField(
			manifest,
			pyTorchReplicaSpecPath(pyTorchReplicaTypeWorker, "replicas")...,
		)

		target, err := k8sUtil.pyTorchSegmentAnnotationTarget(manifest)
		require.NoError(t, err)
		require.NotNil(t, target)
	})

	t.Run("rejects a non-integral Worker replica count", func(t *testing.T) {
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "2",
		})
		require.NoError(t, unstructured.SetNestedField(
			manifest,
			float64(4.5),
			pyTorchReplicaSpecPath(pyTorchReplicaTypeWorker, "replicas")...,
		))

		target, err := k8sUtil.pyTorchSegmentAnnotationTarget(manifest)
		require.Nil(t, target)
		require.ErrorContains(t, err, "replicas must be a non-negative integer")
	})

	t.Run("maps invalid segment input to a manifest submission error", func(t *testing.T) {
		manifest := newPyTorchJobManifest(map[string]interface{}{
			topology.KAISegmentSizeAnnotationKey: "abc",
		})
		impl := &WorkloadsV2Impl{k8sUtil: k8sUtil}

		submissionErr := impl.validateSegmentTopologyClusterSupport(
			context.Background(),
			"unused-cluster",
			manifest,
		)

		require.NotNil(t, submissionErr)
		require.EqualValues(t, 400, submissionErr.Code)
		require.Contains(t, submissionErr.Message, "must be a positive integer")
		require.Len(t, submissionErr.ComplianceIssues, 1)
		require.Equal(t, ErrorFieldManifest, submissionErr.ComplianceIssues[0].Field)
	})
}

func newPyTorchJobManifest(workerAnnotations map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "kubeflow.org/v1",
		"kind":       "PyTorchJob",
		"metadata": map[string]interface{}{
			"name": "pytorch-job",
		},
		"spec": map[string]interface{}{
			"pytorchReplicaSpecs": map[string]interface{}{
				"Worker": map[string]interface{}{
					"replicas": float64(4),
					"template": map[string]interface{}{
						"metadata": map[string]interface{}{
							"annotations": workerAnnotations,
						},
						"spec": map[string]interface{}{},
					},
				},
			},
		},
	}
}

type detailedNodePoolClusterClient struct {
	*policyclusters.SimpleMock
	nodePools []clusterApi.Nodepool
}

func (c *detailedNodePoolClusterClient) GetNodePoolsInternalByCluster(
	_ context.Context,
	_ string,
) (*clusterApi.GetNodepoolsInternal200Response, error) {
	return &clusterApi.GetNodepoolsInternal200Response{Nodepools: c.nodePools}, nil
}

func setDetailedTopologyMock(
	clusterID string,
	nodePools []clusterApi.Nodepool,
	topologies map[string][]string,
) {
	client := policyclusters.NewSimpleMock()
	for name, levels := range topologies {
		client.WithTopology(name, &clusterApi.NetworkTopology{Name: name, Levels: levels})
	}
	clusters.SetMock(&detailedNodePoolClusterClient{
		SimpleMock: client,
		nodePools:  nodePools,
	})
}

func newTopologyNodePool(name string, labelKey string, labelValue string, topologyName string) clusterApi.Nodepool {
	nodePool := clusterApi.Nodepool{
		Name:       name,
		LabelKey:   labelKey,
		LabelValue: labelValue,
	}
	nodePool.SetNetworkTopologyName(topologyName)
	return nodePool
}

func setPyTorchReplicaTemplate(t *testing.T, manifest map[string]interface{}, replicaType string) {
	t.Helper()
	require.NoError(t, unstructured.SetNestedMap(
		manifest,
		map[string]interface{}{
			"metadata": map[string]interface{}{},
			"spec":     map[string]interface{}{},
		},
		pyTorchReplicaTemplatePath(replicaType)...,
	))
}

func setPyTorchReplicaLabels(
	t *testing.T,
	manifest map[string]interface{},
	replicaType string,
	labels map[string]string,
) {
	t.Helper()
	require.NoError(t, unstructured.SetNestedStringMap(
		manifest,
		labels,
		pyTorchReplicaTemplatePath(replicaType, "metadata", "labels")...,
	))
}

func setPyTorchReplicaAffinity(
	t *testing.T,
	manifest map[string]interface{},
	replicaType string,
	affinity *corev1.Affinity,
) {
	t.Helper()
	affinityMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(affinity)
	require.NoError(t, err)
	require.NoError(t, unstructured.SetNestedMap(
		manifest,
		affinityMap,
		pyTorchReplicaTemplatePath(replicaType, "spec", "affinity")...,
	))
}

func requiredNodeAffinity(terms ...corev1.NodeSelectorTerm) *corev1.Affinity {
	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: terms,
			},
		},
	}
}
