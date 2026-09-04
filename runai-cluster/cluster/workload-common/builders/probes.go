/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	PORT_MIN = 1
	PORT_MAX = 65535
)

func ValidateProbe(probeSpec *field_types.ProbeMembers, probeType string, allowHttpGetHostField bool) error {
	if probeSpec == nil {
		return nil
	}

	// validate port field
	port := probeSpec.HttpGetPort.Deref()
	if probeSpec.HttpGetPort.Value == nil {
		return fmt.Errorf("%s: httpGetPort is mandatory", probeType)
	} else {
		if port < PORT_MIN || port > PORT_MAX {
			return fmt.Errorf("%s: httpGetPort must be in the range %d to %d", probeType, PORT_MIN, PORT_MAX)
		}
	}

	// validate httpGet fields
	path := probeSpec.HttpGetPath.Deref()
	if path != "" {
		if !strings.HasPrefix(path, "/") {
			return fmt.Errorf("%s: httpGetPath must start with /", probeType)
		}

		if strings.Index(path, " ") != -1 {
			return fmt.Errorf("%s: httpGetPath must not include spaces", probeType)
		}
	}

	host := probeSpec.HttpGetHost.Deref()
	if host != "" {
		if !allowHttpGetHostField {
			return fmt.Errorf("%s: httpGetHost is not allowed for this workload type", probeType)
		}

		if strings.Index(host, " ") != -1 {
			return fmt.Errorf("%s: httpGetHost must not include spaces", probeType)
		}
	}

	scheme := probeSpec.HttpGetScheme.Deref()
	if scheme != "" {
		scheme = strings.ToUpper(scheme)
		if scheme != "HTTP" && scheme != "HTTPS" {
			return fmt.Errorf("%s: httpGetScheme must be either HTTP or HTTPS", probeType)
		}
	}

	// validate optional fields
	if probeSpec.InitialDelaySeconds != nil && probeSpec.InitialDelaySeconds.Value != nil {
		initDelaySeconds := probeSpec.InitialDelaySeconds.Deref()
		if initDelaySeconds < 0 {
			return fmt.Errorf("%s: initialDelaySeconds must be greater than or equal to 0", probeType)
		}
	}

	if probeSpec.TimeoutSeconds != nil && probeSpec.TimeoutSeconds.Value != nil {
		timeoutSeconds := probeSpec.TimeoutSeconds.Deref()
		if timeoutSeconds < 1 {
			return fmt.Errorf("%s: timeoutSeconds must be greater than or equal to 1", probeType)
		}
	}

	if probeSpec.PeriodSeconds != nil && probeSpec.PeriodSeconds.Value != nil {
		periodSeconds := probeSpec.PeriodSeconds.Deref()
		if periodSeconds < 1 {
			return fmt.Errorf("%s: periodSeconds must be greater than or equal to 1", probeType)
		}
	}

	if probeSpec.SuccessThreshold != nil && probeSpec.SuccessThreshold.Value != nil {
		successThreshold := probeSpec.SuccessThreshold.Deref()
		if successThreshold < 1 {
			return fmt.Errorf("%s: successThreshold must be greater than or equal to 1", probeType)
		}
	}

	if probeSpec.FailureThreshold != nil && probeSpec.FailureThreshold.Value != nil {
		failureThreshold := probeSpec.FailureThreshold.Deref()
		if failureThreshold < 0 {
			return fmt.Errorf("%s: failureThreshold must be greater than or equal to 0", probeType)
		}
	}

	return nil
}

func createProbe(probeSpec *field_types.ProbeMembers) *corev1.Probe {
	if probeSpec == nil {
		return nil
	}

	// init httpGet probe handler
	probeHandler := &corev1.ProbeHandler{
		HTTPGet: &corev1.HTTPGetAction{
			Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(probeSpec.HttpGetPort.Deref())},
		},
	}

	// set optional httpGet fields
	if probeSpec.HttpGetPath != nil {
		probeHandler.HTTPGet.Path = probeSpec.HttpGetPath.Deref()
	}

	if probeSpec.HttpGetHost != nil {
		probeHandler.HTTPGet.Host = probeSpec.HttpGetHost.Deref()
	}

	if probeSpec.HttpGetScheme != nil {
		probeHandler.HTTPGet.Scheme = corev1.URIScheme(strings.ToUpper(probeSpec.HttpGetScheme.Deref()))
	}

	// init probe
	probe := &corev1.Probe{
		ProbeHandler: *probeHandler,
	}

	// set optional fields
	if probeSpec.InitialDelaySeconds != nil {
		probe.InitialDelaySeconds = int32(probeSpec.InitialDelaySeconds.Deref())
	}

	if probeSpec.TimeoutSeconds != nil {
		probe.TimeoutSeconds = int32(probeSpec.TimeoutSeconds.Deref())
	}

	if probeSpec.PeriodSeconds != nil {
		probe.PeriodSeconds = int32(probeSpec.PeriodSeconds.Deref())
	}

	if probeSpec.SuccessThreshold != nil {
		probe.SuccessThreshold = int32(probeSpec.SuccessThreshold.Deref())
	}

	if probeSpec.FailureThreshold != nil {
		probe.FailureThreshold = int32(probeSpec.FailureThreshold.Deref())
	}

	return probe
}

func CreateProbes(_ context.Context, containerFields *field_groups.ContainerFields, containerSpec *corev1.Container, allowHttpGetHostField bool) error {
	if containerFields == nil {
		return nil
	}

	if containerFields.ReadinessProbe == nil {
		return nil
	}

	readinessProbeSpec := containerFields.ReadinessProbe.Deref()

	err := ValidateProbe(&readinessProbeSpec, "readinessProbe", allowHttpGetHostField)
	if err != nil {
		return err
	}

	containerSpec.ReadinessProbe = createProbe(&readinessProbeSpec)

	return nil
}

func CreateTcpReadinessProbe(ports []corev1.ContainerPort, container *corev1.Container) error {
	if len(ports) != 1 {
		return fmt.Errorf("tcp readiness probe expects exactly one port")
	}

	container.ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(ports[0].ContainerPort)},
			},
		},
	}

	return nil
}