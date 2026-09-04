/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ids

const (
	PrerequisiteIdentifierIngress          string = "Ingress NGINX"
	PrerequisiteIdentifierKnative          string = "Knative Serving"
	PrerequisiteIdentifierPrometheus       string = "Prometheus"
	PrerequisiteIdentifierTrainingOperator string = "Kubeflow Training Operator"
	PrerequisiteIdentifierNetworkOperator  string = "NVIDIA Network Operator"
	PrerequisiteIdentifierLWS              string = "LWS - Distributed Inference"
	PrerequisiteIdentifierGpuOperator      string = "NVIDIA Gpu Operator"
)
