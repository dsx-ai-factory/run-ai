/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"

	rsResources "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/researcher-service/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"

	osv1 "github.com/openshift/api/route/v1"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RouteResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewRouteResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *RouteResource {
	return &RouteResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (r *RouteResource) OperandName() string {
	return r.operandName
}

func (r *RouteResource) Object(ctx context.Context) (client.Object, error) {
	routeName := fmt.Sprintf("runai-%s-route", name)
	namespace := r.runaiConfig.Namespace

	if !r.Enabled() {
		return &osv1.Route{
			ObjectMeta: metav1.ObjectMeta{
				Name:      routeName,
				Namespace: namespace,
			},
		}, nil
	}

	route := &osv1.Route{}
	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Name:      routeName,
		Namespace: namespace,
	}, route)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	route.ObjectMeta.Name = routeName
	route.ObjectMeta.Namespace = namespace

	err = initClusterApiRoute(ctx, r.k8sReader, route, r.runaiConfig)
	if err != nil {
		return nil, err
	}

	return route, nil
}

func (r *RouteResource) Status(ctx context.Context) error {
	if !r.Enabled() {
		return nil
	}

	route, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if route.GetUID() == "" {
		return fmt.Errorf("route for cluster-api [%s] is missing", route.GetName())
	}

	return nil
}

func (r *RouteResource) Enabled() bool {
	return r.runaiConfig.Spec.Internal.Global.ClusterURL != "" &&
		(r.runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1_internal.KubernetesFlavorOpenShift)
}

func initClusterApiRoute(ctx context.Context, k8sReader client.Reader, route *osv1.Route, runaiConfig *runv1.RunaiConfig) error {
	certificate, key, err := getTlsSecretDataIfNeeded(ctx, runaiConfig, k8sReader)
	if err != nil {
		return err
	}

	if route.ObjectMeta.Labels == nil {
		route.ObjectMeta.Labels = map[string]string{}
	}
	route.ObjectMeta.Labels["app"] = name

	domain := utils.ParseClusterDomain(runaiConfig.Spec.Internal.Global.ClusterURL)
	if domain == nil {
		return fmt.Errorf("failed to parse cluster domain '%s'", runaiConfig.Spec.Internal.Global.ClusterURL)
	}

	route.Spec = osv1.RouteSpec{
		Host: domain.Host,
		Path: "/" + clusterApiServicePath,
		TLS: &osv1.TLSConfig{
			InsecureEdgeTerminationPolicy: osv1.InsecureEdgeTerminationPolicyRedirect,
			Termination:                   osv1.TLSTerminationEdge,
			Certificate:                   certificate,
			Key:                           key,
		},
		To: osv1.RouteTargetReference{
			Kind:   "Service",
			Name:   name,
			Weight: ptr.To[int32](100),
		},
		Port:           &osv1.RoutePort{TargetPort: intstr.FromString(servicePortName)},
		WildcardPolicy: osv1.WildcardPolicyNone,
	}

	return nil
}

func getTlsSecretDataIfNeeded(ctx context.Context, runaiConfig *runv1.RunaiConfig, k8sReader client.Reader) (string, string, error) {
	var certificate, key string
	var err error
	if runaiConfig.Spec.ClusterApi.Route.TlsSecret != "" {
		certificate, key, err = rsResources.GetTlsSecretCertAndKey(ctx, k8sReader, runaiConfig.Spec.ClusterApi.Route.TlsSecret, runaiConfig.Namespace)
		if err != nil {
			return "", "", fmt.Errorf("failed to get tls secret for cluster-api route: %w", err)
		}
	}
	return certificate, key, nil
}
