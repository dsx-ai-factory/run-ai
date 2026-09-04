/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/shared-objects-controller/pkg/controller/utils"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ServiceMonitorAPIVersion = "monitoring.coreos.com/v1"
	ServiceMonitorKind       = "ServiceMonitor"

	KnativeServingServiceMonitorName = "runai-knative-serving"
	NIMServiceMonitorInferenceName   = "runai-nim-inference"
	NIMServiceMonitor                = "runai-nim"
	VLLMServiceMonitorName           = "runai-vllm"

	PortHTTPIstio      = "http-istio"
	PortServingPort    = "serving-port"
	PortHTTPUsermetric = "http-usermetric"
	PortMetrics        = "metrics"

	PathV1Metrics = "/v1/metrics"
	PathMetrics   = "/metrics"

	LabelNIMWorkload       = "run.ai/nim-workload"
	LabelWorkloadKind      = "workloadKind"
	ValueInferenceWorkload = "InferenceWorkload"

	LabelKnativeServiceType = "networking.internal.knative.dev/serviceType"
	ValuePrivate            = "Private"

	appLabel = "app"

	serviceMonitorHandlerFailed = "ServiceMonitorHandlerFailed"
)

type MetricRelabelConfig struct {
	Regex  string
	Action string
}

type ServiceMonitorEndpoint struct {
	Port              string
	Path              string
	MetricRelabelings []MetricRelabelConfig
}

type LabelSelectorRequirement struct {
	Key      string
	Operator string
	Values   []string
}

type ServiceMonitorConfig struct {
	Name                string
	App                 string
	Endpoints           []ServiceMonitorEndpoint
	LabelSelector       map[string]string
	SelectorExpressions []LabelSelectorRequirement
}

var nimMetricsEndpoints = []ServiceMonitorEndpoint{
	{Port: PortMetrics, Path: PathV1Metrics},
	{Port: PortMetrics, Path: PathMetrics},
	{Port: PortServingPort, Path: PathV1Metrics},
	{Port: PortServingPort, Path: PathMetrics},
}

var (
	knativeServingServiceMonitorConfig = ServiceMonitorConfig{
		Name: KnativeServingServiceMonitorName,
		Endpoints: []ServiceMonitorEndpoint{
			{Port: PortHTTPUsermetric},
		},
		LabelSelector: map[string]string{
			LabelKnativeServiceType: ValuePrivate,
		},
	}

	nimServiceMonitorConfigInference = ServiceMonitorConfig{
		Name:      NIMServiceMonitorInferenceName,
		Endpoints: nimMetricsEndpoints,
		LabelSelector: map[string]string{
			LabelWorkloadKind:                        ValueInferenceWorkload,
			LabelNIMWorkload:                         constants.BoolTrueAsString,
			constants.LabelNIMWorkloadMetricsService: constants.BoolTrueAsString,
		},
	}

	nimServiceMonitorConfig = ServiceMonitorConfig{
		Name:      NIMServiceMonitor,
		Endpoints: nimMetricsEndpoints,
		LabelSelector: map[string]string{
			LabelNIMWorkload: "true",
		},
		SelectorExpressions: []LabelSelectorRequirement{
			{
				Key:      LabelWorkloadKind,
				Operator: "NotIn",
				Values:   []string{ValueInferenceWorkload},
			},
		},
	}

	vllmServiceMonitorConfig = ServiceMonitorConfig{
		Name: VLLMServiceMonitorName,
		Endpoints: []ServiceMonitorEndpoint{
			{Port: PortMetrics, Path: PathMetrics},
		},
		LabelSelector: map[string]string{
			LabelWorkloadKind:                         ValueInferenceWorkload,
			constants.LabelVLLMWorkload:               constants.BoolTrueAsString,
			constants.LabelVLLMWorkloadMetricsService: constants.BoolTrueAsString,
		},
	}
)

type ServiceMonitorReconciler struct {
	client.Client
}

func NewServiceMonitorReconciler(c client.Client) *ServiceMonitorReconciler {
	return &ServiceMonitorReconciler{Client: c}
}

func (r *ServiceMonitorReconciler) Reconcile(ctx context.Context, project *kaiorg.Project) error {
	namespace := project.Status.Namespace
	if namespace == "" {
		log.Ctx(ctx).Info().Msgf("Skipping ServiceMonitor reconciliation for project %s: namespace not yet set", project.Name)
		return nil
	}

	var err error
	if err = utils.ValidateNamespaceExists(ctx, r.Client, namespace, project.Name); err != nil {
		return err
	}

	for _, cfg := range []ServiceMonitorConfig{
		knativeServingServiceMonitorConfig,
		nimServiceMonitorConfigInference,
		nimServiceMonitorConfig,
		vllmServiceMonitorConfig,
	} {
		err := r.createServiceMonitorIfNecessary(ctx, project, namespace, cfg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ServiceMonitorReconciler) createServiceMonitorIfNecessary(
	ctx context.Context, project *kaiorg.Project, namespace string, cfg ServiceMonitorConfig,
) error {
	sm := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": ServiceMonitorAPIVersion,
			"kind":       ServiceMonitorKind,
		},
	}

	err := r.Get(ctx, client.ObjectKey{Name: cfg.Name, Namespace: namespace}, sm)
	if err == nil {
		desired := r.buildServiceMonitor(namespace, project.Name, project.UID, cfg)
		if r.needsUpdate(sm, desired) {
			log.Ctx(ctx).Info().Msgf("ServiceMonitor spec differs from desired, updating: %s/%s", namespace, cfg.Name)
			sm.Object["spec"] = desired.Object["spec"]
			if updateErr := r.Update(ctx, sm); updateErr != nil {
				log.Ctx(ctx).Error().Err(updateErr).Msgf("Failed to update ServiceMonitor: %s/%s", namespace, cfg.Name)
				return updateErr
			}
		}
		return nil
	}

	if !errors.IsNotFound(err) {
		log.Ctx(ctx).Error().Err(err).Msgf("Error getting ServiceMonitor: %s/%s (project %s)", namespace, cfg.Name, project.Name)
		return err
	}

	return r.createServiceMonitor(ctx, namespace, project, cfg)
}

func (r *ServiceMonitorReconciler) createServiceMonitor(
	ctx context.Context, namespace string, project *kaiorg.Project, cfg ServiceMonitorConfig,
) error {
	sm := r.buildServiceMonitor(namespace, project.Name, project.UID, cfg)
	log.Ctx(ctx).Info().Msgf("Creating ServiceMonitor: %s/%s (project %s)", namespace, cfg.Name, project.Name)
	err := r.Create(ctx, sm)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Ctx(ctx).Error().Err(err).Msgf("Error creating ServiceMonitor: %s/%s (project %s)", namespace, cfg.Name, project.Name)
		return err
	}
	return nil
}

func (r *ServiceMonitorReconciler) buildServiceMonitor(namespace, projectName string, projectUID types.UID, cfg ServiceMonitorConfig) *unstructured.Unstructured {
	endpoints := make([]interface{}, len(cfg.Endpoints))
	for i, ep := range cfg.Endpoints {
		epMap := map[string]interface{}{
			"port": ep.Port,
		}
		if ep.Path != "" {
			epMap["path"] = ep.Path
		}
		if len(ep.MetricRelabelings) > 0 {
			relabelings := make([]interface{}, len(ep.MetricRelabelings))
			for j, r := range ep.MetricRelabelings {
				rm := make(map[string]interface{})
				if r.Regex != "" {
					rm["regex"] = r.Regex
				}
				if r.Action != "" {
					rm["action"] = r.Action
				}
				relabelings[j] = rm
			}
			epMap["metricRelabelings"] = relabelings
		}
		endpoints[i] = epMap
	}

	matchLabels := make(map[string]interface{})
	for k, v := range cfg.LabelSelector {
		matchLabels[k] = v
	}

	selector := map[string]interface{}{
		"matchLabels": matchLabels,
	}
	if len(cfg.SelectorExpressions) > 0 {
		exprs := make([]interface{}, len(cfg.SelectorExpressions))
		for i, expr := range cfg.SelectorExpressions {
			em := map[string]interface{}{
				"key":      expr.Key,
				"operator": expr.Operator,
			}
			if len(expr.Values) > 0 {
				em["values"] = expr.Values
			}
			exprs[i] = em
		}
		selector["matchExpressions"] = exprs
	}

	app := cfg.App
	if app == "" {
		app = cfg.Name
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": ServiceMonitorAPIVersion,
			"kind":       ServiceMonitorKind,
			"metadata": map[string]interface{}{
				"name":      cfg.Name,
				"namespace": namespace,
				"labels": map[string]interface{}{
					appLabel: app,
				},
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": kaiorg.GroupVersion.Identifier(),
						"kind":       "Project",
						"name":       projectName,
						"uid":        string(projectUID),
					},
				},
			},
			"spec": map[string]interface{}{
				"endpoints": endpoints,
				"selector":  selector,
			},
		},
	}
}

func (r *ServiceMonitorReconciler) needsUpdate(existing, desired *unstructured.Unstructured) bool {
	existingSpec, existingSpecOk := existing.Object["spec"].(map[string]interface{})
	desiredSpec, desiredSpecOk := desired.Object["spec"].(map[string]interface{})
	if !existingSpecOk || !desiredSpecOk {
		return true
	}
	return !equality.Semantic.DeepEqual(existingSpec["endpoints"], desiredSpec["endpoints"]) ||
		!equality.Semantic.DeepEqual(existingSpec["selector"], desiredSpec["selector"])
}
