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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const clusterApiHTTPRouteName = "cluster-api-httproute"

type HTTPRouteResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewHTTPRouteResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *HTTPRouteResource {
	return &HTTPRouteResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (h *HTTPRouteResource) OperandName() string {
	return h.operandName
}

func (h *HTTPRouteResource) Enabled() bool {
	global := h.runaiConfig.Spec.Global
	if global == nil || global.GatewayAPI == nil {
		return false
	}
	gw := global.GatewayAPI
	return gw.Enabled != nil &&
		*gw.Enabled &&
		gw.Name != "" &&
		h.runaiConfig.Spec.Internal.Global.ClusterURL != ""
}

func (h *HTTPRouteResource) Object(ctx context.Context) (client.Object, error) {
	if !h.Enabled() {
		return &gwv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterApiHTTPRouteName,
				Namespace: h.runaiConfig.Namespace,
			},
		}, nil
	}

	route := &gwv1.HTTPRoute{}
	err := h.k8sReader.Get(ctx, types.NamespacedName{
		Name:      clusterApiHTTPRouteName,
		Namespace: h.runaiConfig.Namespace,
	}, route)
	if err != nil && !errors.IsNotFound(err) && !meta.IsNoMatchError(err) {
		return nil, err
	}

	if err := initClusterApiHTTPRoute(h.runaiConfig, route); err != nil {
		return nil, err
	}
	return route, nil
}

func (h *HTTPRouteResource) Status(ctx context.Context) error {
	if !h.Enabled() {
		return nil
	}

	res, err := h.Object(ctx)
	if err != nil {
		return err
	}

	if res.GetUID() == "" {
		return fmt.Errorf("httproute [%s] is missing", res.GetName())
	}

	route := res.(*gwv1.HTTPRoute)
	accepted, msg := utils.HTTPRouteAccepted(route.Status)
	if !accepted {
		return fmt.Errorf("httproute [%s] not accepted: %s", route.GetName(), msg)
	}

	return nil
}

func initClusterApiHTTPRoute(runaiConfig *runv1.RunaiConfig, route *gwv1.HTTPRoute) error {
	gw := runaiConfig.Spec.Global.GatewayAPI
	clusterURL := runaiConfig.Spec.Internal.Global.ClusterURL

	hostname := utils.HTTPRouteHostname(clusterURL)
	if hostname == "" {
		return fmt.Errorf("failed to derive hostname from cluster URL '%s'", clusterURL)
	}

	route.Name = clusterApiHTTPRouteName
	route.Namespace = runaiConfig.Namespace

	if route.Labels == nil {
		route.Labels = map[string]string{}
	}
	route.Labels["app"] = name
	pathPrefix := gwv1.PathMatchPathPrefix
	clusterApiPath := utils.GetIngressPath(clusterURL, clusterApiServicePath)
	port := gwv1.PortNumber(serviceInPort)

	route.Spec = gwv1.HTTPRouteSpec{
		CommonRouteSpec: gwv1.CommonRouteSpec{
			ParentRefs: []gwv1.ParentReference{utils.BuildParentRef(gw, runaiConfig.Spec.ClusterApi.GatewayAPI)},
		},
		Hostnames: []gwv1.Hostname{hostname},
		Rules: []gwv1.HTTPRouteRule{{
			Matches: []gwv1.HTTPRouteMatch{
				{Path: &gwv1.HTTPPathMatch{Type: &pathPrefix, Value: &clusterApiPath}},
			},
			Filters: []gwv1.HTTPRouteFilter{utils.SecurityHeaderFilter()},
			BackendRefs: []gwv1.HTTPBackendRef{{
				BackendRef: gwv1.BackendRef{
					BackendObjectReference: gwv1.BackendObjectReference{
						Name: gwv1.ObjectName(name),
						Port: &port,
					},
				},
			}},
		}},
	}

	return nil
}
