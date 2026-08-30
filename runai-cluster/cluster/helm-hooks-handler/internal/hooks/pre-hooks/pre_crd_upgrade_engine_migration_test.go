package pre_hooks

import (
	"context"
	"errors"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const queueCRDName = "queues.scheduling.run.ai"

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := apiextensionsv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add apiextensions scheme: %v", err)
	}
	return scheme
}

func queueCRD(storedVersion string) *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: queueCRDName},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{Name: storedVersion, Storage: true},
			},
		},
	}
}

func engineShard(name string) *unstructured.Unstructured {
	shard := &unstructured.Unstructured{}
	shard.SetAPIVersion("engine.run.ai/v1alpha1")
	shard.SetKind("SchedulingShard")
	shard.SetName(name)
	return shard
}

func engineDynamicClient(shardNames ...string) *dynamicfake.FakeDynamicClient {
	objects := make([]runtime.Object, 0, len(shardNames))
	for _, name := range shardNames {
		objects = append(objects, engineShard(name))
	}
	return dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{legacyEngineSchedulingShardGVR: "SchedulingShardList"},
		objects...,
	)
}

func engineDynamicClientWithListError(err error) dynamic.Interface {
	dynamicClient := engineDynamicClient()
	dynamicClient.PrependReactor("list", legacyEngineSchedulingShardGVR.Resource,
		func(k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, err
		})
	return dynamicClient
}

func TestShouldTearDownEngineSchedulers(t *testing.T) {
	testCases := []struct {
		name                 string
		queueVersion         string
		engineShards         []string
		engineListError      error
		engineToKaiMigration bool
		expected             bool
	}{
		{
			name:         "queue v1 stored -> fires regardless of migration flag",
			queueVersion: "v1",
			expected:     true,
		},
		{
			name:                 "queue v2 stored, migration flag off -> no-op",
			queueVersion:         "v2",
			engineShards:         []string{"np-1"},
			engineToKaiMigration: false,
			expected:             false,
		},
		{
			name:                 "queue v2 stored, migration on, engine shards exist -> fires",
			queueVersion:         "v2",
			engineShards:         []string{"np-1"},
			engineToKaiMigration: true,
			expected:             true,
		},
		{
			name:                 "queue v2 stored, migration on, no engine shards -> no-op",
			queueVersion:         "v2",
			engineToKaiMigration: true,
			expected:             false,
		},
		{
			name: "no queue CRD, migration on, engine CRD absent -> no-op",
			engineListError: apierrors.NewNotFound(
				schema.GroupResource{Group: legacyEngineSchedulingShardGVR.Group, Resource: legacyEngineSchedulingShardGVR.Resource},
				"",
			),
			engineToKaiMigration: true,
			expected:             false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clientBuilder := fake.NewClientBuilder().WithScheme(testScheme(t))
			if tc.queueVersion != "" {
				clientBuilder = clientBuilder.WithObjects(queueCRD(tc.queueVersion))
			}
			k8sClient := clientBuilder.Build()
			dynamicClient := dynamic.Interface(engineDynamicClient(tc.engineShards...))
			if tc.engineListError != nil {
				dynamicClient = engineDynamicClientWithListError(tc.engineListError)
			}

			hook := &PreCrdUpgradeHook{
				params: &hooks.ClusterHookParams{EngineToKAIMigrationEnabled: tc.engineToKaiMigration},
			}

			got, err := hook.shouldTearDownEngineSchedulers(context.Background(), k8sClient, dynamicClient)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Errorf("shouldTearDownEngineSchedulers = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestEngineSchedulingShardsExist(t *testing.T) {
	t.Run("returns true when shards exist", func(t *testing.T) {
		hook := &PreCrdUpgradeHook{params: &hooks.ClusterHookParams{}}
		got, err := hook.engineSchedulingShardsExist(context.Background(), engineDynamicClient("np-1"))
		if err != nil || !got {
			t.Fatalf("got (%v, %v), want (true, nil)", got, err)
		}
	})

	t.Run("returns false when no shards exist", func(t *testing.T) {
		hook := &PreCrdUpgradeHook{params: &hooks.ClusterHookParams{}}
		got, err := hook.engineSchedulingShardsExist(context.Background(), engineDynamicClient())
		if err != nil || got {
			t.Fatalf("got (%v, %v), want (false, nil)", got, err)
		}
	})

	t.Run("returns false (no error) when engine CRD is absent", func(t *testing.T) {
		hook := &PreCrdUpgradeHook{params: &hooks.ClusterHookParams{}}
		dynamicClient := engineDynamicClientWithListError(apierrors.NewNotFound(
			schema.GroupResource{Group: legacyEngineSchedulingShardGVR.Group, Resource: legacyEngineSchedulingShardGVR.Resource},
			"",
		))
		got, err := hook.engineSchedulingShardsExist(context.Background(), dynamicClient)
		if err != nil || got {
			t.Fatalf("got (%v, %v), want (false, nil)", got, err)
		}
	})
}

func TestDeleteSchedulingShards(t *testing.T) {
	t.Run("deletes all shards", func(t *testing.T) {
		dynamicClient := engineDynamicClient("np-1", "np-2")
		hook := &PreCrdUpgradeHook{params: &hooks.ClusterHookParams{}}
		err := hook.deleteSchedulingShards(context.Background(), &clients.Clients{
			RemoteClusterDynamicClient: dynamicClient,
		})
		if err != nil {
			t.Fatalf("unexpected error deleting scheduling shards: %v", err)
		}

		remaining, err := dynamicClient.Resource(legacyEngineSchedulingShardGVR).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			t.Fatalf("unexpected error listing scheduling shards: %v", err)
		}
		if len(remaining.Items) != 0 {
			t.Fatalf("got %d remaining scheduling shards, want 0", len(remaining.Items))
		}
	})

	t.Run("does nothing when engine CRD is absent", func(t *testing.T) {
		dynamicClient := engineDynamicClientWithListError(apierrors.NewNotFound(
			schema.GroupResource{Group: legacyEngineSchedulingShardGVR.Group, Resource: legacyEngineSchedulingShardGVR.Resource},
			"",
		))
		hook := &PreCrdUpgradeHook{params: &hooks.ClusterHookParams{}}
		err := hook.deleteSchedulingShards(context.Background(), &clients.Clients{
			RemoteClusterDynamicClient: dynamicClient,
		})
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}
	})

	t.Run("returns delete errors", func(t *testing.T) {
		dynamicClient := engineDynamicClient("np-1")
		expectedErr := errors.New("delete failed")
		dynamicClient.PrependReactor("delete", legacyEngineSchedulingShardGVR.Resource,
			func(k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, expectedErr
			})
		hook := &PreCrdUpgradeHook{params: &hooks.ClusterHookParams{}}
		err := hook.deleteSchedulingShards(context.Background(), &clients.Clients{
			RemoteClusterDynamicClient: dynamicClient,
		})
		if !errors.Is(err, expectedErr) {
			t.Fatalf("got error %v, want wrapped error %v", err, expectedErr)
		}
	})
}
