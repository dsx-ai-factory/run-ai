package main

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	diwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/controller/utils"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestAvailableLegacyReconcilers(t *testing.T) {
	training := twv2a1.GroupVersion.WithKind("TrainingWorkload")
	interactive := iwv2a1.GroupVersion.WithKind("InteractiveWorkload")
	distributed := diwv2.GroupVersion.WithKind("DistributedWorkload")
	inference := infwv2a1.GroupVersion.WithKind("InferenceWorkload")
	distributedInference := dinfw.GroupVersion.WithKind("DistributedInferenceWorkload")
	allGVKs := []schema.GroupVersionKind{training, interactive, distributed, inference, distributedInference}

	candidates := []legacyReconcilerRegistration{
		{gvk: training, reconciler: testReconciler{name: "training"}},
		{gvk: interactive, reconciler: testReconciler{name: "interactive"}},
		{gvk: distributed, reconciler: testReconciler{name: "distributed"}},
		{gvk: inference, reconciler: testReconciler{name: "inference"}},
		{gvk: distributedInference, reconciler: testReconciler{name: "distributed-inference"}},
	}

	nonNoMatchErr := errors.New("discovery unavailable")
	tests := []struct {
		name       string
		errors     map[schema.GroupVersionKind]error
		wantNames  []string
		wantErr    error
		wantMapped []schema.GroupVersionKind
	}{
		{
			name:       "returns every reconciler when exact mappings exist",
			wantNames:  []string{"training", "interactive", "distributed", "inference", "distributed-inference"},
			wantMapped: allGVKs,
		},
		{
			name: "filters reconcilers without matching CRDs",
			errors: map[schema.GroupVersionKind]error{
				interactive: &meta.NoKindMatchError{GroupKind: interactive.GroupKind(), SearchedVersions: []string{interactive.Version}},
				inference:   &meta.NoKindMatchError{GroupKind: inference.GroupKind(), SearchedVersions: []string{inference.Version}},
			},
			wantNames:  []string{"training", "distributed", "distributed-inference"},
			wantMapped: allGVKs,
		},
		{
			name: "returns no reconcilers when every legacy CRD is absent",
			errors: map[schema.GroupVersionKind]error{
				training:             &meta.NoKindMatchError{GroupKind: training.GroupKind(), SearchedVersions: []string{training.Version}},
				interactive:          &meta.NoKindMatchError{GroupKind: interactive.GroupKind(), SearchedVersions: []string{interactive.Version}},
				distributed:          &meta.NoKindMatchError{GroupKind: distributed.GroupKind(), SearchedVersions: []string{distributed.Version}},
				inference:            &meta.NoKindMatchError{GroupKind: inference.GroupKind(), SearchedVersions: []string{inference.Version}},
				distributedInference: &meta.NoKindMatchError{GroupKind: distributedInference.GroupKind(), SearchedVersions: []string{distributedInference.Version}},
			},
			wantNames:  []string{},
			wantMapped: allGVKs,
		},
		{
			name: "returns non NoMatch REST mapping errors",
			errors: map[schema.GroupVersionKind]error{
				distributed: nonNoMatchErr,
			},
			wantErr:    nonNoMatchErr,
			wantMapped: allGVKs[:3],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &testRESTMappingResolver{errors: tt.errors}

			got, err := availableLegacyReconcilers(resolver, candidates)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("availableLegacyReconcilers() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !reflect.DeepEqual(reconcilerNames(got), tt.wantNames) {
				t.Errorf("availableLegacyReconcilers() = %v, want %v", reconcilerNames(got), tt.wantNames)
			}
			if !reflect.DeepEqual(resolver.mappedGVKs, tt.wantMapped) {
				t.Errorf("RESTMapping() GVKs = %v, want %v", resolver.mappedGVKs, tt.wantMapped)
			}
		})
	}
}

func TestAvailableLegacyReconcilersLogsAbsentGVK(t *testing.T) {
	absentGVK := iwv2a1.GroupVersion.WithKind("InteractiveWorkload")
	resolver := &testRESTMappingResolver{errors: map[schema.GroupVersionKind]error{
		absentGVK: &meta.NoKindMatchError{GroupKind: absentGVK.GroupKind(), SearchedVersions: []string{absentGVK.Version}},
	}}
	var output bytes.Buffer
	previousLogger := log.Logger
	log.Logger = zerolog.New(&output)
	t.Cleanup(func() { log.Logger = previousLogger })

	_, err := availableLegacyReconcilers(resolver, []legacyReconcilerRegistration{{
		gvk:        absentGVK,
		reconciler: testReconciler{name: "interactive"},
	}})

	if err != nil {
		t.Fatalf("availableLegacyReconcilers() error = %v, want nil", err)
	}
	if !strings.Contains(output.String(), absentGVK.String()) {
		t.Fatalf("availableLegacyReconcilers() log = %q, want absent GVK %q", output.String(), absentGVK.String())
	}
}

type testRESTMappingResolver struct {
	errors     map[schema.GroupVersionKind]error
	mappedGVKs []schema.GroupVersionKind
}

func (r *testRESTMappingResolver) RESTMapping(groupKind schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	gvk := schema.GroupVersionKind{Group: groupKind.Group, Kind: groupKind.Kind}
	if len(versions) > 0 {
		gvk.Version = versions[0]
	}
	r.mappedGVKs = append(r.mappedGVKs, gvk)
	return &meta.RESTMapping{}, r.errors[gvk]
}

type testReconciler struct {
	name string
}

func (testReconciler) SetupWithManager(ctrl.Manager) error {
	return nil
}

func (testReconciler) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func reconcilerNames(reconcilers []utils.Reconciler) []string {
	names := make([]string, 0, len(reconcilers))
	for _, reconciler := range reconcilers {
		names = append(names, reconciler.(testReconciler).name)
	}
	return names
}
