/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	osv1 "github.com/openshift/api/route/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CliRouteResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewCliRouteResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *CliRouteResource {
	return &CliRouteResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (r *CliRouteResource) OperandName() string {
	return r.operandName
}

func (r *CliRouteResource) Object(ctx context.Context) (client.Object, error) {
	routeName := "researcher-cli-route"
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

	err = initCLIRoute(ctx, r.k8sReader, route, r.runaiConfig)
	if err != nil {
		return nil, err
	}

	return route, nil
}

func (r *CliRouteResource) Status(ctx context.Context) error {
	if !r.Enabled() {
		return nil
	}

	route, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if route.GetUID() == "" {
		return fmt.Errorf("route for researcher cli is [%s] is missing", route.GetName())
	}

	return nil
}

func (r *CliRouteResource) Enabled() bool {
	return componentEnabled
}

type ResearcherRouteResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewResearcherRouteResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *ResearcherRouteResource {
	return &ResearcherRouteResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (r *ResearcherRouteResource) OperandName() string {
	return r.operandName
}

func (r *ResearcherRouteResource) Object(ctx context.Context) (client.Object, error) {
	routeName := "runai-researcher-route"
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

	err = initResearcherRoute(ctx, r.k8sReader, route, r.runaiConfig)
	if err != nil {
		return nil, err
	}

	return route, nil
}

func (r *ResearcherRouteResource) Status(ctx context.Context) error {
	if !r.Enabled() {
		return nil
	}

	route, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if route.GetUID() == "" {
		return fmt.Errorf("route for researcher [%s] is missing", route.GetName())
	}

	return nil
}

func (r *ResearcherRouteResource) Enabled() bool {
	return componentEnabled
}

func initCLIRoute(ctx context.Context, k8sReader client.Reader, route *osv1.Route, runaiConfig *runv1.RunaiConfig) error {
	certificate, key, err := getTlsSecretDataIfNeeded(ctx, runaiConfig, k8sReader)
	if err != nil {
		return err
	}

	if route.ObjectMeta.Labels == nil {
		route.ObjectMeta.Labels = map[string]string{}
	}
	route.ObjectMeta.Labels["app"] = DeploymentName

	domain := utils.ParseClusterDomain(runaiConfig.Spec.Internal.Global.ClusterURL)
	if domain == nil {
		return fmt.Errorf("failed to parse cluster domain '%s'", runaiConfig.Spec.Internal.Global.ClusterURL)
	}

	route.Spec = osv1.RouteSpec{
		Host: domain.Host,
		Path: "/cli",
		TLS: &osv1.TLSConfig{
			InsecureEdgeTerminationPolicy: osv1.InsecureEdgeTerminationPolicyRedirect,
			Termination:                   osv1.TLSTerminationEdge,
			Certificate:                   certificate,
			Key:                           key,
		},
		To: osv1.RouteTargetReference{
			Kind:   "Service",
			Name:   DeploymentName,
			Weight: ptr.To[int32](100),
		},
		WildcardPolicy: osv1.WildcardPolicyNone,
	}

	return nil
}

func initResearcherRoute(ctx context.Context, k8sReader client.Reader, route *osv1.Route, runaiConfig *runv1.RunaiConfig) error {
	certificate, key, err := getTlsSecretDataIfNeeded(ctx, runaiConfig, k8sReader)
	if err != nil {
		return err
	}

	if route.ObjectMeta.Labels == nil {
		route.ObjectMeta.Labels = map[string]string{}
	}
	route.ObjectMeta.Labels["app"] = DeploymentName

	domain := utils.ParseClusterDomain(runaiConfig.Spec.Internal.Global.ClusterURL)
	if domain == nil {
		return fmt.Errorf("failed to parse cluster domain '%s'", runaiConfig.Spec.Internal.Global.ClusterURL)
	}

	route.Spec = osv1.RouteSpec{
		Host: domain.Host,
		Path: "/researcher",
		TLS: &osv1.TLSConfig{
			InsecureEdgeTerminationPolicy: osv1.InsecureEdgeTerminationPolicyRedirect,
			Termination:                   osv1.TLSTerminationEdge,
			Certificate:                   certificate,
			Key:                           key,
		},
		To: osv1.RouteTargetReference{
			Kind:   "Service",
			Name:   DeploymentName,
			Weight: ptr.To[int32](100),
		},
		WildcardPolicy: osv1.WildcardPolicyNone,
	}

	return nil
}

func getTlsSecretDataIfNeeded(ctx context.Context, runaiConfig *runv1.RunaiConfig, k8sReader client.Reader) (string, string, error) {
	var certificate, key string
	var err error
	if runaiConfig.Spec.ResearcherService.Route.TlsSecret != "" {
		certificate, key, err = GetTlsSecretCertAndKey(ctx, k8sReader, runaiConfig.Spec.ResearcherService.Route.TlsSecret, runaiConfig.Namespace)
		if err != nil {
			return "", "", fmt.Errorf("failed to get tls secret for researcher route: %w", err)
		}
	}
	return certificate, key, nil
}
