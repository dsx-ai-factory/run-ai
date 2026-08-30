package handlers

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	infwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type revisionTestCachedClient struct {
	client.Client
}

func (c *revisionTestCachedClient) Start(context.Context, <-chan struct{}) {}
func (c *revisionTestCachedClient) Shutdown()                              {}
func (c *revisionTestCachedClient) GetInformer(cachedclient.SupportedType) cachedclient.CachedInformer {
	return nil
}
func (c *revisionTestCachedClient) NewRateLimitedController(cachedclient.RateLimiterName, cachedclient.Handler) *cachedclient.RateLimitedController {
	return nil
}

func TestRevisionHandlerResolveWorkloadID(t *testing.T) {
	const namespace = "runai-test"

	tests := []struct {
		name       string
		topOwner   client.Object
		wantID     string
		topOwnerGV schema.GroupVersionKind
	}{
		{
			name: "native InferenceWorkload",
			topOwner: &infwv2.InferenceWorkload{ObjectMeta: metav1.ObjectMeta{
				Name:      "native",
				Namespace: namespace,
				UID:       types.UID("native-uid"),
				Annotations: map[string]string{
					commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: "native-workload-id",
				},
			}},
			wantID: "native-workload-id",
			topOwnerGV: schema.GroupVersionKind{
				Group: "run.ai", Version: "v2alpha1", Kind: "InferenceWorkload",
			},
		},
		{
			name: "V2 WorkloadRunner",
			topOwner: &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
				Name:      "v2-runner",
				Namespace: namespace,
				UID:       types.UID("runner-uid"),
				Annotations: map[string]string{
					commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: "v2-workload-id",
				},
			}},
			wantID: "v2-workload-id",
			topOwnerGV: schema.GroupVersionKind{
				Group:   dromosv1alpha1.GroupVersion.Group,
				Version: dromosv1alpha1.GroupVersion.Version,
				Kind:    dromosv1alpha1.WorkloadRunnerKind,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			if err := knv1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			if err := infwv2.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			if err := dromosv1alpha1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}

			topOwner := tt.topOwner.DeepCopyObject().(client.Object)
			service := &knv1.Service{ObjectMeta: metav1.ObjectMeta{
				Name:      "service",
				Namespace: namespace,
				UID:       types.UID("service-uid"),
				OwnerReferences: []metav1.OwnerReference{ownerReference(
					tt.topOwnerGV, topOwner.GetName(), topOwner.GetUID(),
				)},
			}}
			configuration := &knv1.Configuration{ObjectMeta: metav1.ObjectMeta{
				Name:      "configuration",
				Namespace: namespace,
				UID:       types.UID("configuration-uid"),
				OwnerReferences: []metav1.OwnerReference{ownerReference(
					schema.GroupVersionKind{
						Group:   knv1.SchemeGroupVersion.Group,
						Version: knv1.SchemeGroupVersion.Version,
						Kind:    "Service",
					},
					service.Name,
					service.UID,
				)},
			}}
			revision := &knv1.Revision{ObjectMeta: metav1.ObjectMeta{
				Name:      "revision",
				Namespace: namespace,
				UID:       types.UID("revision-uid"),
				OwnerReferences: []metav1.OwnerReference{ownerReference(
					schema.GroupVersionKind{
						Group:   knv1.SchemeGroupVersion.Group,
						Version: knv1.SchemeGroupVersion.Version,
						Kind:    "Configuration",
					},
					configuration.Name,
					configuration.UID,
				)},
			}}

			k8sClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(topOwner, service, configuration, revision).
				Build()
			handler := &RevisionHandler{CachedClient: &revisionTestCachedClient{Client: k8sClient}}

			got, err := handler.resolveWorkloadID(revision)
			if err != nil {
				t.Fatalf("resolveWorkloadID() error = %v", err)
			}
			if got != tt.wantID {
				t.Fatalf("resolveWorkloadID() = %q, want %q", got, tt.wantID)
			}
		})
	}
}

func TestRevisionHandlerFullSyncRetainsUnmanagedAndTransientFailures(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := knv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	handler := &RevisionHandler{CachedClient: &revisionTestCachedClient{Client: k8sClient}}

	revisions := []knv1.Revision{
		{ObjectMeta: metav1.ObjectMeta{
			UID: types.UID("unmanaged"),
			Labels: map[string]string{
				commonconstants.ControlPlaneWorkloadKindLabelForWorkload: commonconstants.V2InferenceWorkloadKind,
			},
		}},
		{ObjectMeta: metav1.ObjectMeta{
			UID: types.UID("temporarily-unresolved"),
			Labels: map[string]string{
				commonconstants.ControlPlaneWorkloadKindLabelForWorkload: commonconstants.V2InferenceWorkloadKind,
				commonconstants.ControlPlaneWorkloadNameLabelForWorkload: "v2-workload",
			},
			OwnerReferences: []metav1.OwnerReference{ownerReference(
				schema.GroupVersionKind{
					Group:   knv1.SchemeGroupVersion.Group,
					Version: knv1.SchemeGroupVersion.Version,
					Kind:    "Configuration",
				},
				"missing-configuration",
				types.UID("missing-configuration-uid"),
			)},
		}},
	}

	parsed, idsToIgnore := handler.parseRevisionsForSync(revisions, nil)
	if len(parsed) != 0 {
		t.Fatalf("parsed %d revisions, want none", len(parsed))
	}
	wantIDsToIgnore := []string{"unmanaged", "temporarily-unresolved"}
	if len(idsToIgnore) != len(wantIDsToIgnore) {
		t.Fatalf("idsToIgnore = %v, want %v", idsToIgnore, wantIDsToIgnore)
	}
	for i := range wantIDsToIgnore {
		if idsToIgnore[i] != wantIDsToIgnore[i] {
			t.Fatalf("idsToIgnore = %v, want %v", idsToIgnore, wantIDsToIgnore)
		}
	}
}

func TestRevisionHandlerRejectsOwnerlessRevision(t *testing.T) {
	handler := &RevisionHandler{}
	revision := &knv1.Revision{ObjectMeta: metav1.ObjectMeta{
		Name:      "ownerless",
		Namespace: "runai-test",
		UID:       types.UID("ownerless-uid"),
	}}

	if _, err := handler.resolveWorkloadID(revision); err == nil {
		t.Fatal("resolveWorkloadID() expected an error for an ownerless Revision")
	}
}

func TestRevisionHandlerIncrementalIgnoresUnmanagedRevision(t *testing.T) {
	handler := &RevisionHandler{}
	revision := &knv1.Revision{ObjectMeta: metav1.ObjectMeta{
		UID: types.UID("unmanaged"),
		Labels: map[string]string{
			commonconstants.ControlPlaneWorkloadKindLabelForWorkload: commonconstants.V2InferenceWorkloadKind,
		},
	}}

	if err := handler.handleRevisionAddUpdate(revision); err != nil {
		t.Fatalf("handleRevisionAddUpdate() error = %v", err)
	}
}

func ownerReference(gvk schema.GroupVersionKind, name string, uid types.UID) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: gvk.GroupVersion().String(),
		Kind:       gvk.Kind,
		Name:       name,
		UID:        uid,
	}
}
