/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_workloads

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	openapi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func int32ToKnativeContainerPorts(src *int64, target any, _ any) *workloadsApi.Error {
	if src == nil || *src <= 0 {
		return nil
	}

	container, ok := target.(*corev1.Container)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("int32ToKnativeContainerPorts: target type %T is not *corev1.Container", target))
	}

	container.Ports = []corev1.ContainerPort{{ContainerPort: int32(*src)}}
	return nil
}

func initTimeoutToProgressDeadlineAnnotation(src *int64, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	podMeta, ok := target.(*metav1.ObjectMeta)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("initTimeoutToProgressDeadlineAnnotation: target type %T is not *metav1.ObjectMeta", target))
	}

	if podMeta.Annotations == nil {
		podMeta.Annotations = map[string]string{}
	}
	podMeta.Annotations[annotationProgressDeadline] = strconv.FormatInt(*src, 10) + "s"
	return nil
}

func buildMetricsEndpoint(engine inferenceEngine, workloadName string, servingPort *workloadsApi.ServingPort) *openapi.NetworkingEndpoint {
	if engine == inferenceEngineUnknown {
		return nil
	}

	port := int32(8000)
	if servingPort != nil {
		if configuredPort := servingPort.Container.Get(); configuredPort != nil && *configuredPort > 0 {
			port = *configuredPort
		}
	}

	servicePort := openapi.NewServicePort(port)
	servicePort.SetName("metrics")
	servicePort.SetTargetPort(port)
	servicePort.SetProtocol(openapi.SERVICEPROTOCOL_TCP)

	serviceLabels := map[string]string{
		constants.LabelWorkloadMetricsService: "true",
		engine.identityLabel():                "true",
		engine.metricsLabel():                 "true",
		constants.LabelWorkloadKind:           constants.KindInferenceWorkload,
	}
	serviceSpec := openapi.NewServiceSpec(
		workloadName+"-metrics",
		map[string]string{constants.LabelWorkloadName: workloadName},
		[]openapi.ServicePort{*servicePort},
	)
	serviceSpec.SetType(openapi.SERVICETYPE_CLUSTER_IP)
	serviceSpec.SetLabels(serviceLabels)

	service := openapi.NewNetworkingService()
	service.SetSpec(*serviceSpec)

	endpoint := openapi.NewNetworkingEndpoint()
	endpoint.Services = []openapi.NetworkingService{*service}
	return endpoint
}

func appendMetricsEndpoint(cfg *openapi.WorkloadV2Configuration, endpoint *openapi.NetworkingEndpoint) {
	if endpoint != nil {
		cfg.Networking = append(cfg.Networking, *endpoint)
	}
}

func autoScalingToKnative(src *workloadsApi.AutoScaling, target any, _ any) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	svc, ok := target.(*knservingv1.Service)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("autoScalingToKnative: target type %T is not *knservingv1.Service", target))
	}

	podMeta := &svc.Spec.Template.ObjectMeta
	if podMeta.Annotations == nil {
		podMeta.Annotations = map[string]string{}
	}
	ann := podMeta.Annotations
	ann[annotationAutoscalingClass] = autoscalingClassKPA

	if v := src.MinReplicas.Get(); v != nil {
		ann[annotationAutoscalingMin] = strconv.FormatInt(int64(*v), 10)
	}
	if v := src.MaxReplicas.Get(); v != nil {
		ann[annotationAutoscalingMax] = strconv.FormatInt(int64(*v), 10)
	}
	if v := src.Metric.Get(); v != nil && *v != "" {
		metric := *v
		switch metric {
		case composer.AutoScalingMetric_THROUGHPUT:
			metric = autoscalingMetricRPS
		case composer.AutoScalingMetric_LATENCY:
			metric = autoscalingMetricLatency
		}
		ann[annotationAutoscalingMetric] = metric

		if metric != autoscalingMetricRPS && metric != composer.AutoScalingMetric_CONCURRENCY {
			ann[annotationAutoscalingClass] = autoscalingClassHPA
		}
	}
	if v := src.MetricThreshold.Get(); v != nil {
		ann[annotationAutoscalingTarget] = strconv.FormatInt(int64(*v), 10)
	}
	if v := src.ScaleToZeroRetentionSeconds.Get(); v != nil {
		ann[annotationScaleToZeroRetention] = strconv.FormatInt(int64(*v), 10) + "s"
	}

	return nil
}
