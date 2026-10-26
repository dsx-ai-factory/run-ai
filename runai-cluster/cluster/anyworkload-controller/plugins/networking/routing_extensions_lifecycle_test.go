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

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	openshiftv1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestRoutingAnnotationsRejectServicesWithoutMutation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runner := newOwnershipRunner()
	service := newOwnershipService(runner, true, map[string]string{"app": "owned"})
	service.Annotations = map[string]string{"traefik.ingress.kubernetes.io/service.nativelb": "true"}
	pod := newOwnershipPod(runner, "owned", service.Spec.Selector, "10.0.0.21")
	slice := newOwnershipEndpointSlice(service, pod, pod.Status.PodIP)
	tests := []struct {
		name      string
		validate  func(client.Reader) error
		wantLists int
	}{
		{
			name: "desired", wantLists: 1,
			validate: func(reader client.Reader) error {
				return validateDesiredServiceOwnership(ctx, reader, runner, service)
			},
		},
		{
			name: "created", wantLists: 2,
			validate: func(reader client.Reader) error {
				return validateCreatedServiceOwnership(ctx, reader, runner, ownershipServiceKey)
			},
		},
		{
			name: "referenced", wantLists: 2,
			validate: func(reader client.Reader) error {
				return validateReferencedServiceOwnership(ctx, reader, runner, ownershipServiceKey)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := newOwnershipFakeClient(t, service, pod, slice)
			before := &corev1.Service{}
			if err := k8sClient.Get(ctx, ownershipServiceKey, before); err != nil {
				t.Fatal(err)
			}
			desiredBefore := service.DeepCopy()
			reader := &countingOwnershipReader{Reader: k8sClient}
			if err := tt.validate(reader); !errors.Is(err, errOwnershipViolation) {
				t.Fatalf("expected annotation ownership violation, got %v", err)
			}
			if reader.listCalls != tt.wantLists {
				t.Fatalf("List calls = %d, want %d; annotation denial must not skip traffic validation", reader.listCalls, tt.wantLists)
			}
			after := &corev1.Service{}
			if err := k8sClient.Get(ctx, ownershipServiceKey, after); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(service, desiredBefore) {
				t.Fatal("validation mutated the desired or live Service")
			}
		})
	}
}

func TestRoutingAnnotationsPreserveErrorPrecedence(t *testing.T) {
	t.Parallel()
	readErr := errors.New("API unavailable")
	tests := []struct {
		name      string
		cause     error
		forbidden bool
	}{
		{
			name: "violation takes precedence over missing Service", forbidden: true,
			cause: apierrors.NewNotFound(schema.GroupResource{Resource: "services"}, "backend"),
		},
		{name: "violation takes precedence over API failure", cause: readErr, forbidden: true},
		{name: "valid annotations preserve API failure", cause: readErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ingress := ingressWithServices("backend")
			if tt.forbidden {
				ingress.Annotations = map[string]string{"nginx.ingress.kubernetes.io/default-backend": "foreign"}
			}
			reader := &countingOwnershipReader{Reader: errorOwnershipReader{err: tt.cause}}
			err := validateExposureServiceOwnership(context.Background(), reader, newOwnershipRunner(), ingress)
			if reader.getCalls != 1 {
				t.Fatalf("Service Get calls = %d, want 1 despite extension validation", reader.getCalls)
			}
			if errors.Is(err, errOwnershipViolation) != tt.forbidden {
				t.Fatalf("error = %v, want violation = %v", err, tt.forbidden)
			}
			if errors.Is(err, tt.cause) == tt.forbidden {
				t.Fatalf("error = %v, cause should be preserved only without a violation", err)
			}
			enforcementErr := ownershipEnforcementError("validate exposure", err)
			if errors.Is(enforcementErr, errOwnershipOperational) == tt.forbidden {
				t.Fatalf("operational classification changed: %v", enforcementErr)
			}
		})
	}
}

func TestRoutingAnnotationsExposureLifecycle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		exposure func(string) client.Object
	}{
		{name: "Ingress", exposure: func(service string) client.Object { return ingressWithServices(service) }},
		{name: "Route", exposure: func(service string) client.Object { return routeWithServices(service) }},
		{name: "HTTPRoute", exposure: func(service string) client.Object { return httpRouteWithBackend(gatewayBackend(service)) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			runner := newOwnershipRunner()
			service, pod, slice := newNamedOwnershipBackend(runner, "backend", "10.0.1.2")
			desired := tt.exposure(service.Name)
			live := desired.DeepCopyObject().(client.Object)
			live.SetUID("exposure-uid")
			live.SetOwnerReferences([]metav1.OwnerReference{newOwnershipRunnerRef(runner)})
			if route, ok := live.(*openshiftv1.Route); ok {
				route.Spec.Host = "generated.apps.example"
				route.Annotations = map[string]string{"openshift.io/host.generated": "true"}
			}
			k8sClient := enforcementClient(t, service, pod, slice, live)
			plugin := newOwnershipTestPlugin(k8sClient)
			resources := []managedResource{{object: desired, description: tt.name}}

			// Live system metadata is valid; only controller instructions are bounded.
			if err := plugin.validateLiveExternalAccess(ctx, runner, resources); err != nil {
				t.Fatalf("initial live graph was rejected: %v", err)
			}
			if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(live), live); err != nil {
				t.Fatal(err)
			}
			live.SetAnnotations(map[string]string{"custom.example/routing": "unvalidated-destination"})
			if err := k8sClient.Update(ctx, live); err != nil {
				t.Fatal(err)
			}
			if err := plugin.validateLiveExternalAccess(ctx, runner, resources); !errors.Is(err, errOwnershipViolation) {
				t.Fatalf("live annotation mutation was not denied: %v", err)
			}

			desired.SetAnnotations(live.GetAnnotations())
			if err := plugin.reconcileExternalAccess(ctx, runner, runner, resources, k8sClient); !isSettledOwnershipViolation(err) {
				t.Fatalf("unsafe desired exposure was not denied: %v", err)
			}
			assertNotFound(t, k8sClient, desired.GetNamespace(), desired.GetName(), live)

			desired.SetAnnotations(nil)
			if err := plugin.reconcileExternalAccess(ctx, runner, runner, resources, k8sClient); err != nil {
				t.Fatalf("corrected exposure did not recover: %v", err)
			}
			if err := plugin.validateLiveExternalAccess(ctx, runner, resources); err != nil {
				t.Fatalf("recovered live exposure failed validation: %v", err)
			}
		})
	}
}

func TestRoutingAnnotationsReferencedServiceCleanupAndRecovery(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	runner := enforcementRunner()
	config := runner.Spec.Instructions.Networking
	service := config.Services[0].Create.DeepCopy()
	service.Namespace, service.UID = runner.Namespace, "referenced-service-uid"
	service.Annotations = map[string]string{
		"nvidia.com/last-applied-hash":                   "58d4b4b4ff",
		"nvidia.com/parent-spec-hash":                    "5df844fc9c",
		"openshift.io/required-scc":                      "nonroot",
		"traefik.ingress.kubernetes.io/service.nativelb": "true",
	}
	config.Services[0] = dromosv1alpha1.NetworkingServiceItem{
		Ref: &dromosv1alpha1.NetworkingServiceRef{Name: service.Name, Port: 8080},
	}
	exposure := NewPlugin(nil).buildIngress(runner, config.ExternalAccess.Ingress, service.Name, 8080)
	exposure.UID = "exposure-uid"
	exposure.OwnerReferences = []metav1.OwnerReference{newOwnershipRunnerRef(runner)}
	k8sClient := enforcementClient(t, service, enforcementPod(runner, "owned", runner.UID), exposure)
	plugin := newOwnershipTestPlugin(k8sClient)
	for _, forbidden := range []bool{true, true, false} {
		before := &corev1.Service{}
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(service), before); err != nil {
			t.Fatal(err)
		}
		if !forbidden {
			delete(before.Annotations, "traefik.ingress.kubernetes.io/service.nativelb")
			if err := k8sClient.Update(ctx, before); err != nil {
				t.Fatal(err)
			}
		}
		state := reconcile.New()
		if err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state); err != nil {
			t.Fatalf("reconcile with forbidden annotation = %v: %v", forbidden, err)
		}
		after := &corev1.Service{}
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(service), after); err != nil {
			t.Fatalf("referenced Service was removed: %v", err)
		}
		if !reflect.DeepEqual(before, after) {
			t.Fatal("reconciliation mutated the referenced Service, including UID or resource version")
		}
		issues := state.PluginIssues(plugin)
		if forbidden {
			assertNotFound(t, k8sClient, runner.Namespace, exposure.Name, &networkingv1.Ingress{})
			if len(issues) != 1 || !issues[0].IsCritical {
				t.Fatalf("expected one critical networking issue, got %v", issues)
			}
		} else {
			if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(exposure), &networkingv1.Ingress{}); err != nil {
				t.Fatalf("exposure did not recover after annotation correction: %v", err)
			}
			if len(issues) != 0 {
				t.Fatalf("unexpected issues after recovery: %v", issues)
			}
		}
	}
}
