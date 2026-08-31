/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"reflect"

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"golang.org/x/exp/maps"

	prom_common "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/common/prometheus"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const prometheusServicePath = "prometheus"

type IngressResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewIngressResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *IngressResource {
	return &IngressResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (r *IngressResource) OperandName() string {
	return r.operandName
}

func (r *IngressResource) Object(ctx context.Context) (client.Object, error) {
	var (
		ingressName = "nmc-mgmt-prometheus"
		global      = r.runaiConfig.Spec.Global
		nmcConfig   = r.runaiConfig.Spec.Internal.Global.NMC
	)

	ingress := &v1.Ingress{}
	// Get the existing ingress if it exists to consume any cluster-set values
	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Name:      ingressName,
		Namespace: r.runaiConfig.Namespace,
	}, ingress)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	ingress.ObjectMeta.Name = ingressName
	ingress.ObjectMeta.Namespace = r.runaiConfig.Namespace

	ingress.ObjectMeta.Annotations = make(map[string]string)
	if global.Ingress != nil {
		maps.Copy(ingress.ObjectMeta.Annotations, global.Ingress.Annotations)
	}
	if nmcConfig.MgmtPrometheusIngress != nil {
		maps.Copy(ingress.ObjectMeta.Annotations, nmcConfig.MgmtPrometheusIngress.Annotations)
	}
	var ingressClassName string
	if nmcConfig.MgmtPrometheusIngress != nil && nmcConfig.MgmtPrometheusIngress.Class != "" {
		ingressClassName = nmcConfig.MgmtPrometheusIngress.Class
	} else if global.Ingress != nil && global.Ingress.Class != "" {
		ingressClassName = global.Ingress.Class
	} else {
		ingressClassName = "nginx"
	}
	ingress.ObjectMeta.Annotations["kubernetes.io/ingress.class"] = ingressClassName
	ingress.Spec.IngressClassName = &ingressClassName

	// Rules
	ingress.Spec.Rules = []v1.IngressRule{
		{
			Host: prom_common.MgmtPrometheusDomain,
			IngressRuleValue: v1.IngressRuleValue{
				HTTP: &v1.HTTPIngressRuleValue{
					Paths: []v1.HTTPIngressPath{
						{
							Path: "/",
							Backend: v1.IngressBackend{
								Service: &v1.IngressServiceBackend{
									Name: prom_common.MgmtPrometheusServiceName,
									Port: v1.ServiceBackendPort{Number: 9090},
								},
							},
							PathType: ptr.To(v1.PathTypePrefix),
						},
					},
				},
			},
		},
	}

	// TLS
	ingress.Spec.TLS = []v1.IngressTLS{
		{
			Hosts:      []string{prom_common.MgmtPrometheusDomain},
			SecretName: tlsSecretName,
		},
	}

	return ingress, nil
}

func (r *IngressResource) Status(ctx context.Context) error {
	res, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if res == nil {
		return nil
	}

	if res.GetUID() == "" {
		return fmt.Errorf("ingress [%s] is missing", res.GetName())
	}

	ingress := res.(*v1.Ingress)

	if reflect.DeepEqual(ingress.Status, v1.IngressStatus{}) {
		return fmt.Errorf("no ingress controller has reported status for [%s]", ingress.GetName())
	}

	return nil
}

func (r *IngressResource) Enabled() bool {
	global := r.runaiConfig.Spec.Global
	return global.Ingress != nil &&
		global.Ingress.Enabled != nil &&
		*global.Ingress.Enabled
}
