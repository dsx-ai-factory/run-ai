/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestServiceAnnotationsUseRunnerSubmissionProfile(t *testing.T) {
	t.Parallel()
	for _, profile := range []string{generatedNativeProfile, generatedNIMProfile, "direct-v2", "", "unknown"} {
		for _, stage := range []string{"desired", "created", "referenced"} {
			t.Run(profile+"/"+stage, func(t *testing.T) {
				ctx := context.Background()
				runner := newOwnershipRunner()
				runner.Annotations = map[string]string{clusterconstants.NetworkingSubmissionProfileAnnotation: profile}
				service := newOwnershipService(runner, stage != "referenced", map[string]string{"app": "owned"})
				service.Annotations = map[string]string{
					"custom.example/routing":                               "unvalidated-destination",
					clusterconstants.NetworkingSubmissionProfileAnnotation: "generated-nim",
				}
				pod := newOwnershipPod(runner, "owned", service.Spec.Selector, "10.0.0.21")
				k8sClient := newOwnershipFakeClient(t, service, pod, newOwnershipEndpointSlice(service, pod, pod.Status.PodIP))
				before := &corev1.Service{}
				if err := k8sClient.Get(ctx, ownershipServiceKey, before); err != nil {
					t.Fatal(err)
				}
				var err error
				switch stage {
				case "desired":
					err = validateDesiredServiceOwnership(ctx, k8sClient, runner, service)
				case "created":
					err = validateCreatedServiceOwnership(ctx, k8sClient, runner, ownershipServiceKey)
				case "referenced":
					err = validateReferencedServiceOwnership(ctx, k8sClient, runner, ownershipServiceKey)
				}
				wantViolation := profile != generatedNativeProfile && profile != generatedNIMProfile
				assertOwnershipResult(t, err, false, wantViolation)
				after := &corev1.Service{}
				if err := k8sClient.Get(ctx, ownershipServiceKey, after); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(before, after) {
					t.Fatal("submission profile changed read-only Service validation")
				}
			})
		}
	}
}

func TestExposureAnnotationsUseRunnerSubmissionProfile(t *testing.T) {
	t.Parallel()
	for _, profile := range []string{generatedNativeProfile, generatedNIMProfile, "direct-v2", "", "unknown"} {
		for _, tt := range []struct {
			name      string
			construct func(string) client.Object
		}{
			{"Ingress", func(name string) client.Object { return ingressWithServices(name) }},
			{"Route", func(name string) client.Object { return routeWithServices(name) }},
			{"HTTPRoute", func(name string) client.Object { return httpRouteWithServices(name, "", "") }},
		} {
			t.Run(profile+"/"+tt.name, func(t *testing.T) {
				runner := newOwnershipRunner()
				runner.Annotations = map[string]string{clusterconstants.NetworkingSubmissionProfileAnnotation: profile}
				service, pod, slice := newNamedOwnershipBackend(runner, "backend", "10.0.0.21")
				exposure := tt.construct(service.Name)
				exposure.SetAnnotations(map[string]string{
					"custom.example/routing":                               "unvalidated-destination",
					clusterconstants.NetworkingSubmissionProfileAnnotation: "generated-native",
				})
				err := validateExposureServiceOwnership(context.Background(), newOwnershipFakeClient(t, service, pod, slice), runner, exposure)
				wantViolation := profile != generatedNativeProfile && profile != generatedNIMProfile
				assertOwnershipResult(t, err, false, wantViolation)
			})
		}
	}
}

func TestNativeSubmissionProfilesStillValidatePodOwnership(t *testing.T) {
	t.Parallel()
	for _, profile := range []string{generatedNativeProfile, generatedNIMProfile} {
		for _, destination := range []string{"selector", "EndpointSlice"} {
			t.Run(profile+"/"+destination, func(t *testing.T) {
				runner := newOwnershipRunner()
				runner.Annotations = map[string]string{clusterconstants.NetworkingSubmissionProfileAnnotation: profile}
				service, owned, slice := newNamedOwnershipBackend(runner, "backend", "10.0.0.21")
				service.Annotations = map[string]string{"custom.example/routing": "unvalidated-destination"}
				foreign := newOwnershipPod(runner, "foreign", map[string]string{"app": "foreign"}, "10.0.0.22")
				foreign.OwnerReferences[0].UID = "foreign-runner-uid"
				if destination == "selector" {
					foreign.Labels = service.Spec.Selector
				} else {
					slice = newOwnershipEndpointSlice(service, foreign, foreign.Status.PodIP)
				}
				reader := newOwnershipFakeClient(t, service, owned, foreign, slice)
				for name, err := range map[string]error{
					"referenced Service": validateReferencedServiceOwnership(context.Background(), reader, runner, client.ObjectKeyFromObject(service)),
					"Ingress": validateExposureServiceOwnership(context.Background(), reader, runner,
						ingressWithServices(service.Name)),
					"Route": validateExposureServiceOwnership(context.Background(), reader, runner,
						routeWithServices(service.Name)),
					"HTTPRoute": validateExposureServiceOwnership(context.Background(), reader, runner,
						httpRouteWithServices(service.Name, "", "")),
				} {
					t.Run(name, func(t *testing.T) {
						if !errors.Is(err, errOwnershipViolation) || !strings.Contains(err.Error(), foreign.Name) {
							t.Fatalf("expected foreign Pod ownership violation, got %v", err)
						}
					})
				}
			})
		}
	}
}

func TestNativeSubmissionProfilesStillValidateRoutingBindingsAndFilters(t *testing.T) {
	t.Parallel()
	ingress := ingressWithServices("backend")
	ingress.Spec.IngressClassName = ptr.To("foreign")
	route := httpRouteWithServices("backend", "", "")
	route.Spec.ParentRefs = []gatewayv1.ParentReference{{Name: "foreign"}}
	constraints := routingConstraints{ingressClass: "nginx", gatewayName: "approved", gatewayNamespace: ownershipNamespace}
	for _, profile := range []string{generatedNativeProfile, generatedNIMProfile} {
		for _, tt := range []struct {
			name     string
			exposure client.Object
		}{
			{"foreign Ingress class", ingress},
			{"foreign Gateway parent", route},
			{"unsupported HTTPRoute filter", httpRouteWithFilter("rule", gatewayv1.HTTPRouteFilter{Type: "Unknown"})},
		} {
			t.Run(profile+"/"+tt.name, func(t *testing.T) {
				runner := newOwnershipRunner()
				runner.Annotations = map[string]string{clusterconstants.NetworkingSubmissionProfileAnnotation: profile}
				if err := validateExposureExtensions(runner, tt.exposure, constraints); !errors.Is(err, errOwnershipViolation) {
					t.Fatalf("native annotation exemption bypassed routing validation: %v", err)
				}
			})
		}
	}
}
