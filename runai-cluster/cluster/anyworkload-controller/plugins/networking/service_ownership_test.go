/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const ownershipNamespace = "project"

var ownershipServiceKey = types.NamespacedName{Namespace: ownershipNamespace, Name: "service"}

func TestValidateServiceOwnership(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		created       bool
		objects       func(*dromosv1alpha1.WorkloadRunner) []client.Object
		key           types.NamespacedName
		wantNotFound  bool
		wantViolation bool
	}{
		{
			name:         "missing Service is retryable",
			created:      true,
			key:          types.NamespacedName{Namespace: ownershipNamespace, Name: "missing"},
			wantNotFound: true,
		},
		{
			name:    "created Service with owned Pod endpoint",
			created: true,
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, true, map[string]string{"app": "owned"})
				pod := newOwnershipPod(runner, "owned", map[string]string{"app": "owned"}, "10.0.0.1")
				return []client.Object{service, pod, newOwnershipEndpointSlice(service, pod, "10.0.0.1")}
			},
		},
		{
			name: "referenced ownerless Service with owned targets",
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, false, map[string]string{"app": "owned"})
				pod := newOwnershipPod(runner, "owned", map[string]string{"app": "owned"}, "10.0.0.2")
				return []client.Object{service, pod, newOwnershipEndpointSlice(service, pod, "10.0.0.2")}
			},
		},
		{
			name:    "created ownerless Service",
			created: true,
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, false, map[string]string{"app": "owned"})
				pod := newOwnershipPod(runner, "owned", map[string]string{"app": "owned"}, "10.0.0.3")
				return []client.Object{service, pod, newOwnershipEndpointSlice(service, pod, "10.0.0.3")}
			},
			wantViolation: true,
		},
		{
			name: "foreign selected Pod",
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, false, map[string]string{"app": "foreign"})
				pod := newOwnershipPod(runner, "foreign", map[string]string{"app": "foreign"}, "10.0.0.4")
				pod.OwnerReferences[0].UID = "foreign-runner-uid"
				return []client.Object{service, pod, newOwnershipEndpointSlice(service, pod, "10.0.0.4")}
			},
			wantViolation: true,
		},
		{
			name:    "mixed selected Pod ownership",
			created: true,
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, true, map[string]string{"app": "mixed"})
				owned := newOwnershipPod(runner, "owned", map[string]string{"app": "mixed"}, "10.0.0.5")
				foreign := newOwnershipPod(runner, "foreign", map[string]string{"app": "mixed"}, "10.0.0.6")
				foreign.OwnerReferences[0].UID = "foreign-runner-uid"
				return []client.Object{service, owned, foreign, newOwnershipEndpointSlice(service, owned, "10.0.0.5")}
			},
			wantViolation: true,
		},
		{
			name:    "all selector terms use Kubernetes AND semantics",
			created: true,
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				selector := map[string]string{"app": "shared", "identity": "runner"}
				service := newOwnershipService(runner, true, selector)
				owned := newOwnershipPod(runner, "owned", selector, "10.0.0.7")
				foreign := newOwnershipPod(
					runner, "foreign", map[string]string{"app": "shared", "identity": "foreign"}, "10.0.0.8",
				)
				foreign.OwnerReferences[0].UID = "foreign-runner-uid"
				return []client.Object{service, owned, foreign, newOwnershipEndpointSlice(service, owned, "10.0.0.7")}
			},
		},
		{
			name:    "selected Pods without endpoints are safe",
			created: true,
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				return []client.Object{
					newOwnershipService(runner, true, map[string]string{"app": "pending"}),
					newOwnershipPod(runner, "pending", map[string]string{"app": "pending"}, ""),
				}
			},
		},
		{
			name: "Service without destinations is safe",
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				return []client.Object{newOwnershipService(runner, false, map[string]string{"app": "missing"})}
			},
		},
		{
			name: "ExternalName Service",
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, false, map[string]string{"app": "owned"})
				service.Spec.Type = corev1.ServiceTypeExternalName
				service.Spec.ExternalName = "example.test"
				return []client.Object{service}
			},
			wantViolation: true,
		},
		{
			name: "selectorless Service",
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				return []client.Object{newOwnershipService(runner, false, nil)}
			},
			wantViolation: true,
		},
		{
			name:    "EndpointSlice with stale Service UID",
			created: true,
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, true, map[string]string{"app": "owned"})
				pod := newOwnershipPod(runner, "owned", map[string]string{"app": "owned"}, "10.0.0.9")
				slice := newOwnershipEndpointSlice(service, pod, "10.0.0.9")
				slice.OwnerReferences[0].UID = "stale-service-uid"
				return []client.Object{service, pod, slice}
			},
			wantViolation: true,
		},
		{
			name: "EndpointSlice bypasses selector to a foreign Pod",
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, false, map[string]string{"app": "owned"})
				foreign := newOwnershipPod(runner, "foreign", map[string]string{"app": "foreign"}, "10.0.0.14")
				foreign.OwnerReferences[0].UID = "foreign-runner-uid"
				return []client.Object{service, foreign, newOwnershipEndpointSlice(service, foreign, "10.0.0.14")}
			},
			wantViolation: true,
		},
		{
			name:    "IP-only EndpointSlice target",
			created: true,
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, true, map[string]string{"app": "owned"})
				pod := newOwnershipPod(runner, "owned", map[string]string{"app": "owned"}, "10.0.0.10")
				slice := newOwnershipEndpointSlice(service, pod, "10.0.0.10")
				slice.Endpoints[0].TargetRef = nil
				return []client.Object{service, pod, slice}
			},
			wantViolation: true,
		},
		{
			name:    "stale endpoint Pod UID",
			created: true,
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, true, map[string]string{"app": "owned"})
				pod := newOwnershipPod(runner, "owned", map[string]string{"app": "owned"}, "10.0.0.11")
				slice := newOwnershipEndpointSlice(service, pod, "10.0.0.11")
				slice.Endpoints[0].TargetRef.UID = "stale-pod-uid"
				return []client.Object{service, pod, slice}
			},
			wantViolation: true,
		},
		{
			name: "missing endpoint Pod",
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, false, map[string]string{"app": "none"})
				missing := newOwnershipPod(runner, "missing", map[string]string{"app": "other"}, "10.0.0.12")
				return []client.Object{service, newOwnershipEndpointSlice(service, missing, "10.0.0.12")}
			},
			wantViolation: true,
		},
		{
			name:    "endpoint address outside Pod IPs",
			created: true,
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, true, map[string]string{"app": "owned"})
				pod := newOwnershipPod(runner, "owned", map[string]string{"app": "owned"}, "10.0.0.13")
				return []client.Object{service, pod, newOwnershipEndpointSlice(service, pod, "10.0.0.99")}
			},
			wantViolation: true,
		},
		{
			name:    "unsupported EndpointSlice address type",
			created: true,
			objects: func(runner *dromosv1alpha1.WorkloadRunner) []client.Object {
				service := newOwnershipService(runner, true, map[string]string{"app": "owned"})
				pod := newOwnershipPod(runner, "owned", map[string]string{"app": "owned"}, "10.0.0.15")
				slice := newOwnershipEndpointSlice(service, pod, "10.0.0.15")
				slice.AddressType = discoveryv1.AddressType("FQDN")
				return []client.Object{service, pod, slice}
			},
			wantViolation: true,
		},
		{
			name:          "cross-namespace Service key",
			created:       true,
			key:           types.NamespacedName{Namespace: "foreign", Name: "service"},
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			runner := newOwnershipRunner()
			var objects []client.Object
			if tt.objects != nil {
				objects = tt.objects(runner)
			}
			key := tt.key
			if key == (types.NamespacedName{}) {
				key = ownershipServiceKey
			}
			reader := newOwnershipFakeClient(t, objects...)

			var err error
			if tt.created {
				err = validateCreatedServiceOwnership(context.Background(), reader, runner, key)
			} else {
				err = validateReferencedServiceOwnership(context.Background(), reader, runner, key)
			}
			assertOwnershipResult(t, err, tt.wantNotFound, tt.wantViolation)
		})
	}
}

func TestValidateDesiredServiceOwnership(t *testing.T) {
	t.Parallel()

	runner := newOwnershipRunner()
	desired := newOwnershipService(runner, false, map[string]string{"app": "owned"})
	desired.UID = ""
	ownedPod := newOwnershipPod(runner, "owned", desired.Spec.Selector, "10.0.0.20")
	foreignPod := ownedPod.DeepCopy()
	foreignPod.Name = "foreign"
	foreignPod.UID = "foreign-pod-uid"
	foreignPod.OwnerReferences[0].UID = "foreign-runner-uid"

	tests := []struct {
		name          string
		objects       []client.Object
		wantViolation bool
	}{
		{name: "owned selected Pod", objects: []client.Object{ownedPod}},
		{name: "no selected Pods is safe"},
		{name: "foreign selected Pod", objects: []client.Object{foreignPod}, wantViolation: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateDesiredServiceOwnership(
				context.Background(), newOwnershipFakeClient(t, tt.objects...), runner, desired.DeepCopy(),
			)
			assertOwnershipResult(t, err, false, tt.wantViolation)
		})
	}
}

func TestValidateReferencedServiceOwnershipDoesNotMutateIt(t *testing.T) {
	t.Parallel()

	runner := newOwnershipRunner()
	service := newOwnershipService(runner, false, map[string]string{
		"app":                     "existing",
		"custom.example/selector": "preserved",
	})
	pod := newOwnershipPod(runner, "owned", service.Spec.Selector, "10.0.0.21")
	slice := newOwnershipEndpointSlice(service, pod, "10.0.0.21")
	want := service.DeepCopy()
	reader := newOwnershipFakeClient(t, service, pod, slice)

	if err := validateReferencedServiceOwnership(
		context.Background(), reader, runner, ownershipServiceKey,
	); err != nil {
		t.Fatalf("validateReferencedServiceOwnership() error = %v", err)
	}

	got := &corev1.Service{}
	if err := reader.Get(context.Background(), ownershipServiceKey, got); err != nil {
		t.Fatalf("get Service after validation: %v", err)
	}
	if !reflect.DeepEqual(got.Spec, want.Spec) || !reflect.DeepEqual(got.Labels, want.Labels) {
		t.Fatalf("referenced Service was mutated: got %#v, want %#v", got, want)
	}
}

func TestValidateServiceOwnershipPreservesReadFailure(t *testing.T) {
	t.Parallel()

	readErr := errors.New("API unavailable")
	err := validateCreatedServiceOwnership(
		context.Background(), errorOwnershipReader{err: readErr}, newOwnershipRunner(), ownershipServiceKey,
	)
	if !errors.Is(err, readErr) {
		t.Fatalf("validation error = %v, want wrapped %v", err, readErr)
	}
	if errors.Is(err, errOwnershipViolation) {
		t.Fatalf("read failure must not be an ownership violation: %v", err)
	}
}

func TestValidateServiceOwnershipChecksEndpointsAfterSelectorReadFailure(t *testing.T) {
	t.Parallel()

	runner := newOwnershipRunner()
	service := newOwnershipService(runner, false, map[string]string{"app": "owned"})
	foreign := newOwnershipPod(runner, "foreign", map[string]string{"app": "foreign"}, "10.0.0.31")
	foreign.OwnerReferences[0].UID = "foreign-runner-uid"
	slice := newOwnershipEndpointSlice(service, foreign, "10.0.0.31")
	readErr := errors.New("Pod list temporarily unavailable")
	reader := selectiveOwnershipReader{
		Reader:     newOwnershipFakeClient(t, service, foreign, slice),
		podListErr: readErr,
	}

	err := validateReferencedServiceOwnership(context.Background(), reader, runner, ownershipServiceKey)
	if !errors.Is(err, errOwnershipViolation) {
		t.Fatalf("validation error = %v, want endpoint ownership violation", err)
	}
	if errors.Is(err, readErr) {
		t.Fatalf("lower-priority selector read failure leaked into denial result: %v", err)
	}
}

func TestValidateObjectOwnerClassifiesForbiddenWalkAndPreservesCause(t *testing.T) {
	t.Parallel()

	runner := newOwnershipRunner()
	pod := newOwnershipPod(runner, "pod", nil, "")
	pod.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: "batch/v1",
		Kind:       "Job",
		Name:       "job",
		UID:        "job-uid",
		Controller: ptr.To(true),
	}}
	forbidden := apierrors.NewForbidden(
		schema.GroupResource{Group: "batch", Resource: "jobs"}, "job", errors.New("denied"),
	)

	err := validateObjectOwner(context.Background(), errorOwnershipReader{err: forbidden}, runner, pod, "Pod")
	if !errors.Is(err, errOwnershipViolation) {
		t.Fatalf("validation error = %v, want ownership violation", err)
	}
	if !apierrors.IsForbidden(err) {
		t.Fatalf("validation error = %v, want preserved Forbidden cause", err)
	}
}

func TestValidateObjectOwnerClassifiesDeletedIntermediateOwnerAndPreservesCause(t *testing.T) {
	t.Parallel()

	runner := newOwnershipRunner()
	pod := newOwnershipPod(runner, "pod", nil, "")
	pod.OwnerReferences = []metav1.OwnerReference{ownerReferenceFor(
		"batch/v1", "Job", "deleted-job", "job-uid", true,
	)}
	notFound := apierrors.NewNotFound(
		schema.GroupResource{Group: "batch", Resource: "jobs"}, "deleted-job",
	)

	err := validateObjectOwner(context.Background(), errorOwnershipReader{err: notFound}, runner, pod, "Pod")
	if !errors.Is(err, errOwnershipViolation) {
		t.Fatalf("validation error = %v, want ownership violation", err)
	}
	if !apierrors.IsNotFound(err) {
		t.Fatalf("validation error = %v, want preserved NotFound cause", err)
	}
}

func TestOwnershipValidationAppliesDeadline(t *testing.T) {
	t.Parallel()

	readErr := errors.New("stop")
	reader := &deadlineOwnershipReader{err: readErr}
	err := validateCreatedServiceOwnership(
		context.Background(), reader, newOwnershipRunner(), ownershipServiceKey,
	)
	if !errors.Is(err, readErr) {
		t.Fatalf("validation error = %v, want wrapped %v", err, readErr)
	}
	if !reader.hasDeadline {
		t.Fatal("validation reader context has no deadline")
	}
	remaining := reader.deadline.Sub(reader.observedAt)
	if remaining <= 0 || remaining > ownershipValidationTimeout {
		t.Fatalf("validation deadline remaining = %v, want within (0, %v]", remaining, ownershipValidationTimeout)
	}
}

func TestValidateCreatedServiceOwnershipWithProductionOwnerShape(t *testing.T) {
	t.Parallel()

	runner := newOwnershipRunner()
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       ownershipNamespace,
			Name:            "primary",
			UID:             "deployment-uid",
			OwnerReferences: []metav1.OwnerReference{newOwnershipRunnerRef(runner)},
		},
	}
	replicaSet := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "ReplicaSet"},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ownershipNamespace,
			Name:      "primary-rs",
			UID:       "replica-set-uid",
			OwnerReferences: []metav1.OwnerReference{ownerReferenceFor(
				appsv1.SchemeGroupVersion.String(), "Deployment", deployment.Name, deployment.UID, true,
			)},
		},
	}
	selector := map[string]string{"app": "primary"}
	service := newOwnershipService(runner, false, selector)
	service.OwnerReferences = []metav1.OwnerReference{ownerReferenceFor(
		appsv1.SchemeGroupVersion.String(), "Deployment", deployment.Name, deployment.UID, false,
	)}
	pod := newOwnershipPod(runner, "primary", selector, "10.0.0.32")
	pod.TypeMeta = metav1.TypeMeta{}
	pod.OwnerReferences = []metav1.OwnerReference{ownerReferenceFor(
		appsv1.SchemeGroupVersion.String(), "ReplicaSet", replicaSet.Name, replicaSet.UID, true,
	)}
	pod2 := newOwnershipPod(runner, "secondary", selector, "10.0.0.33")
	pod2.TypeMeta = metav1.TypeMeta{}
	pod2.OwnerReferences = []metav1.OwnerReference{ownerReferenceFor(
		appsv1.SchemeGroupVersion.String(), "ReplicaSet", replicaSet.Name, replicaSet.UID, true,
	)}
	slice := newOwnershipEndpointSlice(service, pod, "10.0.0.32")
	slice.Endpoints = append(
		slice.Endpoints,
		newOwnershipEndpointSlice(service, pod2, "10.0.0.33").Endpoints[0],
	)
	reader := &countingOwnershipReader{Reader: newOwnershipFakeClient(
		t, service, pod, pod2, slice, deployment, replicaSet,
	)}

	err := validateCreatedServiceOwnership(
		context.Background(),
		reader,
		runner,
		ownershipServiceKey,
	)
	if err != nil {
		t.Fatalf("validateCreatedServiceOwnership() error = %v", err)
	}
	// One Service read plus one live read for each shared intermediate owner.
	// EndpointSlice targets reuse the Pods already verified by the selector.
	if reader.getCalls != 3 {
		t.Fatalf("live Get calls = %d, want 3 unique-object reads", reader.getCalls)
	}
	if reader.listCalls != 2 {
		t.Fatalf("live List calls = %d, want one Pod and one EndpointSlice list", reader.listCalls)
	}
	if reader.podGetCalls != 0 {
		t.Fatalf("live Pod Get calls = %d, want selector-verified Pods to be reused", reader.podGetCalls)
	}
	// Before selector-Pod reuse and intermediate-owner memoization, this graph
	// required 12 Gets and 2 Lists. It now requires 3 Gets and the same 2 Lists.
}

func TestEndpointPodCacheMissPerformsLivePodRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		endpointPod   func(*dromosv1alpha1.WorkloadRunner) *corev1.Pod
		endpointUID   types.UID
		wantViolation bool
	}{
		{
			name: "stale endpoint UID",
			endpointPod: func(runner *dromosv1alpha1.WorkloadRunner) *corev1.Pod {
				return newOwnershipPod(runner, "selected", map[string]string{"app": "selected"}, "10.0.0.34")
			},
			endpointUID:   "stale-pod-uid",
			wantViolation: true,
		},
		{
			name: "owned endpoint Pod outside selector result",
			endpointPod: func(runner *dromosv1alpha1.WorkloadRunner) *corev1.Pod {
				return newOwnershipPod(runner, "endpoint-only", map[string]string{"app": "endpoint-only"}, "10.0.0.35")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := newOwnershipRunner()
			service := newOwnershipService(runner, false, map[string]string{"app": "selected"})
			selected := newOwnershipPod(runner, "selected", service.Spec.Selector, "10.0.0.34")
			endpointPod := tt.endpointPod(runner)
			slice := newOwnershipEndpointSlice(service, endpointPod, endpointPod.Status.PodIP)
			if tt.endpointUID != "" {
				slice.Endpoints[0].TargetRef.UID = tt.endpointUID
			}

			objects := []client.Object{service, selected, slice}
			if endpointPod.Name != selected.Name {
				objects = append(objects, endpointPod)
			}
			reader := &countingOwnershipReader{Reader: newOwnershipFakeClient(t, objects...)}

			err := validateReferencedServiceOwnership(
				context.Background(), reader, runner, ownershipServiceKey,
			)
			if tt.wantViolation {
				if !errors.Is(err, errOwnershipViolation) {
					t.Fatalf("validation error = %v, want ownership violation", err)
				}
			} else if err != nil {
				t.Fatalf("validateReferencedServiceOwnership() error = %v", err)
			}
			if reader.podGetCalls != 1 {
				t.Fatalf("live Pod Get calls = %d, want 1 cache-miss read", reader.podGetCalls)
			}
			if reader.getCalls != 2 {
				t.Fatalf("live Get calls = %d, want one Service and one Pod read", reader.getCalls)
			}
			if reader.listCalls != 2 {
				t.Fatalf("live List calls = %d, want one Pod and one EndpointSlice list", reader.listCalls)
			}
		})
	}
}

func TestPrioritizeOwnershipErrors(t *testing.T) {
	t.Parallel()

	transient := errors.New("API unavailable")
	missing := apierrors.NewNotFound(schema.GroupResource{Resource: "services"}, "missing")
	violation := ownershipViolationf("foreign destination")

	err := prioritizeOwnershipErrors(missing, transient, violation)
	if !errors.Is(err, errOwnershipViolation) {
		t.Fatalf("prioritized error = %v, want ownership violation", err)
	}
	if errors.Is(err, transient) || apierrors.IsNotFound(err) {
		t.Fatalf("prioritized error retains lower-priority categories: %v", err)
	}

	err = prioritizeOwnershipErrors(missing, transient)
	if !errors.Is(err, transient) || apierrors.IsNotFound(err) {
		t.Fatalf("prioritized error = %v, want transient only", err)
	}

	if err = prioritizeOwnershipErrors(missing); !apierrors.IsNotFound(err) {
		t.Fatalf("prioritized error = %v, want NotFound", err)
	}

	if err = prioritizeOwnershipErrors(); err != nil {
		t.Fatalf("prioritized empty error set = %v, want nil", err)
	}
}

func assertOwnershipResult(t *testing.T, err error, wantNotFound, wantViolation bool) {
	t.Helper()
	switch {
	case wantNotFound && !apierrors.IsNotFound(err):
		t.Fatalf("validation error = %v, want Kubernetes NotFound", err)
	case wantViolation && !errors.Is(err, errOwnershipViolation):
		t.Fatalf("validation error = %v, want ownership violation", err)
	case !wantNotFound && !wantViolation && err != nil:
		t.Fatalf("validation error = %v, want nil", err)
	}
}

func newOwnershipRunner() *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		TypeMeta: metav1.TypeMeta{
			APIVersion: dromosv1alpha1.GroupVersion.String(),
			Kind:       dromosv1alpha1.WorkloadRunnerKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ownershipNamespace,
			Name:      "runner",
			UID:       "runner-uid",
		},
	}
}

func newOwnershipService(
	runner *dromosv1alpha1.WorkloadRunner,
	owned bool,
	selector map[string]string,
) *corev1.Service {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ownershipNamespace,
			Name:      ownershipServiceKey.Name,
			UID:       "service-uid",
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports:    []corev1.ServicePort{{Name: "http", Port: 8080}},
		},
	}
	if owned {
		service.OwnerReferences = []metav1.OwnerReference{newOwnershipRunnerRef(runner)}
	}
	return service
}

func newOwnershipPod(
	runner *dromosv1alpha1.WorkloadRunner,
	name string,
	labels map[string]string,
	ip string,
) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Pod"},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       ownershipNamespace,
			Name:            name,
			UID:             types.UID(name + "-uid"),
			Labels:          labels,
			OwnerReferences: []metav1.OwnerReference{newOwnershipRunnerRef(runner)},
		},
		Status: corev1.PodStatus{PodIP: ip, PodIPs: []corev1.PodIP{{IP: ip}}},
	}
}

func newOwnershipEndpointSlice(
	service *corev1.Service,
	pod *corev1.Pod,
	address string,
) *discoveryv1.EndpointSlice {
	return &discoveryv1.EndpointSlice{
		TypeMeta: metav1.TypeMeta{APIVersion: discoveryv1.SchemeGroupVersion.String(), Kind: "EndpointSlice"},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ownershipNamespace,
			Name:      service.Name + "-slice",
			Labels:    map[string]string{discoveryv1.LabelServiceName: service.Name},
			OwnerReferences: []metav1.OwnerReference{ownerReferenceFor(
				"v1", "Service", service.Name, service.UID, true,
			)},
		},
		AddressType: discoveryv1.AddressTypeIPv4,
		Endpoints: []discoveryv1.Endpoint{{
			Addresses: []string{address},
			TargetRef: &corev1.ObjectReference{
				APIVersion: "v1",
				Kind:       "Pod",
				Namespace:  ownershipNamespace,
				Name:       pod.Name,
				UID:        pod.UID,
			},
		}},
	}
}

func newOwnershipRunnerRef(runner *dromosv1alpha1.WorkloadRunner) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: dromosv1alpha1.GroupVersion.String(),
		Kind:       dromosv1alpha1.WorkloadRunnerKind,
		Name:       runner.Name,
		UID:        runner.UID,
	}
}

func ownerReferenceFor(
	apiVersion, kind, name string,
	uid types.UID,
	controller bool,
) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		UID:        uid,
		Controller: ptr.To(controller),
	}
}

func newOwnershipFakeClient(t *testing.T, objects ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(discoveryv1.AddToScheme(scheme))
	utilruntime.Must(dromosv1alpha1.AddToScheme(scheme))
	utilruntime.Must(appsv1.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

type selectiveOwnershipReader struct {
	client.Reader
	podListErr error
}

type countingOwnershipReader struct {
	client.Reader
	getCalls    int
	listCalls   int
	podGetCalls int
}

func (r *countingOwnershipReader) Get(
	ctx context.Context,
	key client.ObjectKey,
	object client.Object,
	opts ...client.GetOption,
) error {
	r.getCalls++
	if _, ok := object.(*corev1.Pod); ok {
		r.podGetCalls++
	}
	return r.Reader.Get(ctx, key, object, opts...)
}

func (r *countingOwnershipReader) List(
	ctx context.Context,
	list client.ObjectList,
	opts ...client.ListOption,
) error {
	r.listCalls++
	return r.Reader.List(ctx, list, opts...)
}

func (r selectiveOwnershipReader) List(
	ctx context.Context,
	list client.ObjectList,
	opts ...client.ListOption,
) error {
	if _, ok := list.(*corev1.PodList); ok && r.podListErr != nil {
		return r.podListErr
	}
	return r.Reader.List(ctx, list, opts...)
}

type deadlineOwnershipReader struct {
	err         error
	deadline    time.Time
	observedAt  time.Time
	hasDeadline bool
}

func (r *deadlineOwnershipReader) Get(
	ctx context.Context,
	_ client.ObjectKey,
	_ client.Object,
	_ ...client.GetOption,
) error {
	r.observedAt = time.Now()
	r.deadline, r.hasDeadline = ctx.Deadline()
	return r.err
}

func (r *deadlineOwnershipReader) List(
	context.Context,
	client.ObjectList,
	...client.ListOption,
) error {
	return r.err
}

type errorOwnershipReader struct {
	err error
}

func (r errorOwnershipReader) Get(
	context.Context,
	client.ObjectKey,
	client.Object,
	...client.GetOption,
) error {
	return r.err
}

func (r errorOwnershipReader) List(
	context.Context,
	client.ObjectList,
	...client.ListOption,
) error {
	return r.err
}
