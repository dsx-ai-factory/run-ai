/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package issue

import (
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	openshiftV1 "github.com/openshift/api/route/v1"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"strings"
)

type Group string

const (
	Internal Group = "Internal"
	Network  Group = "Network"
	Compute  Group = "Compute"
	Storage  Group = "Storage"
)

func (g Group) String() string {
	return string(g)
}

type Issue struct {
	// name resource name
	name string
	// resource type v1.GroupVersionKind
	resource schema.GroupVersionKind
	// the error that caused the issue
	error
	// isCritical whether the issue has critical implications on the workload's lifecycle
	isCritical bool
	// description of the error (can be empty)
	description string
	// group the resource group it relevant to (Network/Storage/Compute/Internal)
	group Group
}

func New(object client.Object, err error, isCritical bool, desc string, rg ResourceGrouper) *Issue {
	if rg == nil {
		rg = &DefaultResourceGrouper{}
	}

	gvk, objErr := rg.GetObjectResource(object)
	if err != nil {
		log.Debug("failed to generate kind for object", zap.Error(objErr), zap.String("name", object.GetName()))
	}

	group := rg.GetIssueGroupByGvk(gvk)

	return &Issue{
		name:        object.GetName(),
		resource:    schema.GroupVersionKind(gvk),
		isCritical:  isCritical,
		description: desc,
		error:       err,
		group:       group,
	}
}

func (i *Issue) GetError() error {
	return i.error
}

func (i *Issue) GetGroup() Group {
	return i.group
}

func (i *Issue) Name() string {
	return i.name
}

func (i *Issue) Kind() string {
	return i.resource.Kind
}

func (i *Issue) IsCritical() bool {
	return i.isCritical
}

func (i *Issue) Message() string {
	var parts []string
	if i.description != "" {
		parts = append(parts, i.description)
	}
	if i.error != nil {
		parts = append(parts, i.error.Error())
	}
	return message(i.name, i.Kind(), strings.Join(parts, ": "))
}

func message(name, kind, msg string) string {
	return fmt.Sprintf("[%s:%s] %s", kind, name, msg)
}

// ResourceGrouper is an interface for determining the issue group for a resource.
// it is required in order not to include here workload-type specific logic, especially for the Compute group.
type ResourceGrouper interface {
	GetObjectResource(resource client.Object) (metav1.GroupVersionKind, error)
	GetIssueGroupByGvk(gvk metav1.GroupVersionKind) Group
	GetIssueGroupByResource(resource client.Object) Group
}

type DefaultResourceGrouper struct{}

func (rg *DefaultResourceGrouper) GetObjectResource(resource client.Object) (metav1.GroupVersionKind, error) {
	if resource == nil {
		return metav1.GroupVersionKind{}, fmt.Errorf("object is nil %v", resource)
	}

	var objectResource metav1.GroupVersionKind

	switch resource.(type) {
	case *networkingv1.Ingress:
		objectResource.Kind = constants.KindIngress
	case *v1.Service:
		objectResource.Kind = constants.KindService
	case *openshiftV1.Route:
		objectResource.Kind = constants.KindRoute
	case *v1.PersistentVolumeClaim:
		objectResource.Kind = constants.KindPvc
	case *nvapi.ComputeDomain:
		objectResource.Kind = constants.KindComputeDomain
	case *v1.Secret:
		objectResource.Kind = constants.KindSecret
	case *gatewayv1.HTTPRoute:
		objectResource.Kind = constants.KindHTTPRoute
	default:
		return metav1.GroupVersionKind{Kind: resource.GetObjectKind().GroupVersionKind().Kind}, fmt.Errorf("unknown resource type %v", resource)
	}
	return objectResource, nil
}

func (rg *DefaultResourceGrouper) GetIssueGroupByGvk(gvk metav1.GroupVersionKind) Group {
	switch gvk.Kind {
	case constants.KindService, constants.KindRoute, constants.KindIngress, constants.KindHTTPRoute:
		return Network
	case constants.KindPvc, constants.KindSecret:
		return Storage
	case constants.KindComputeDomain:
		return Compute
	default:
		return Internal
	}
}

func (rg *DefaultResourceGrouper) GetIssueGroupByResource(resource client.Object) Group {
	gvk, err := rg.GetObjectResource(resource)
	if err != nil {
		return ""
	}

	return rg.GetIssueGroupByGvk(gvk)
}
