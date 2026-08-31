/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package endpoints

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// StatusSchemaVersionV1 is the current networking status details schema version.
	StatusSchemaVersionV1 = "v1"
)

// Phase mirrors the plugin phase values at networking-resource granularity.
type Phase string

const (
	PhaseDisabled     Phase = "Disabled"
	PhaseReady        Phase = "Ready"
	PhaseCreating     Phase = "Creating"
	PhasePending      Phase = "Pending"
	PhaseInitializing Phase = "Initializing"
	PhaseFailed       Phase = "Failed"
	PhaseUnknown      Phase = "Unknown"
)

// ExternalAccessKind identifies the external access resource kind that produced status.
type ExternalAccessKind string

const (
	ExternalAccessKindIngress   ExternalAccessKind = "Ingress"
	ExternalAccessKindRoute     ExternalAccessKind = "Route"
	ExternalAccessKindHTTPRoute ExternalAccessKind = "HTTPRoute"
)

// NetworkingStatus is the networking plugin's structured status details payload.
type NetworkingStatus struct {
	SchemaVersion      string           `json:"schemaVersion"`
	ObservedGeneration int64            `json:"observedGeneration,omitempty"`
	Endpoints          []EndpointStatus `json:"endpoints,omitempty"`
}

// EndpointStatus is the status of one configured networking Endpoint.
type EndpointStatus struct {
	Phase          Phase                 `json:"phase,omitempty"`
	Services       []ServiceStatus       `json:"services,omitempty"`
	ExternalAccess *ExternalAccessStatus `json:"externalAccess,omitempty"`
}

// ServiceStatus is the status of one service exposed by an Endpoint.
type ServiceStatus struct {
	// Name is the main correlation key back to CP networking[].services[].spec.name.
	Name string `json:"name"`

	Phase   Phase  `json:"phase,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`

	Type         string              `json:"type,omitempty"`
	ClusterIP    string              `json:"clusterIP,omitempty"`
	Ports        []ServicePortStatus `json:"ports,omitempty"`
	LoadBalancer *LoadBalancerStatus `json:"loadBalancer,omitempty"`

	// AuthProxyReady is nil when auth is not configured for this service.
	AuthProxyReady *bool `json:"authProxyReady,omitempty"`

	// URLs are attached to the service so CP can correlate URL to service/tool.
	URLs []ServiceURLStatus `json:"urls,omitempty"`
}

// ServicePortStatus represents a port exposed by a service.
type ServicePortStatus struct {
	Name       string             `json:"name,omitempty"`
	Port       int32              `json:"port,omitempty"`
	TargetPort intstr.IntOrString `json:"targetPort,omitempty"`
	Protocol   string             `json:"protocol,omitempty"`
}

// ServiceURLStatus is a URL reachable through a service, with external resource provenance.
type ServiceURLStatus struct {
	URL        string             `json:"url"`
	SourceKind ExternalAccessKind `json:"sourceKind,omitempty"`
	SourceName string             `json:"sourceName,omitempty"`
}

// LoadBalancerStatus represents the status of a load-balancer.
type LoadBalancerStatus struct {
	Ingress []LoadBalancerIngress `json:"ingress,omitempty"`
}

// LoadBalancerIngress represents a load-balancer ingress point.
type LoadBalancerIngress struct {
	IP       string `json:"ip,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

// ExternalAccessStatus is the status of external access resources for an Endpoint.
type ExternalAccessStatus struct {
	Phase   Phase  `json:"phase,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`

	Resources []ExternalAccessResourceStatus `json:"resources,omitempty"`
}

// ExternalAccessResourceStatus is the status of one external access resource.
type ExternalAccessResourceStatus struct {
	Kind ExternalAccessKind `json:"kind,omitempty"`
	Name string             `json:"name,omitempty"`

	Phase   Phase  `json:"phase,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`

	// Accepted is mainly used for HTTPRoute gateway acceptance.
	Accepted *bool `json:"accepted,omitempty"`

	// Hosts is diagnostic/provenance data. CP should primarily use ServiceStatus.URLs.
	Hosts []string `json:"hosts,omitempty"`
}

// NewNetworkingStatus creates a NetworkingStatus with the current schema version.
func NewNetworkingStatus(observedGeneration int64) *NetworkingStatus {
	return &NetworkingStatus{
		SchemaVersion:      StatusSchemaVersionV1,
		ObservedGeneration: observedGeneration,
	}
}

// NetworkingStatusToRawExtension serializes status into a RawExtension suitable for PluginStatus.AdditionalInfo.
func NetworkingStatusToRawExtension(status *NetworkingStatus) (*runtime.RawExtension, error) {
	if status == nil {
		return nil, nil
	}
	status.SchemaVersion = StatusSchemaVersionV1
	raw, err := json.Marshal(status)
	if err != nil {
		return nil, fmt.Errorf("marshal networking status: %w", err)
	}
	return &runtime.RawExtension{Raw: raw}, nil
}

// NetworkingStatusFromRawExtension parses PluginStatus.AdditionalInfo as networking status.
func NetworkingStatusFromRawExtension(additionalInfo *runtime.RawExtension) (*NetworkingStatus, error) {
	if additionalInfo == nil || len(additionalInfo.Raw) == 0 {
		return nil, nil
	}
	var status NetworkingStatus
	if err := json.Unmarshal(additionalInfo.Raw, &status); err != nil {
		return nil, fmt.Errorf("unmarshal networking status: %w", err)
	}
	if status.SchemaVersion != StatusSchemaVersionV1 {
		return nil, fmt.Errorf("%w %q (expected %q)",
			ErrUnsupportedSchemaVersion, status.SchemaVersion, StatusSchemaVersionV1)
	}
	return &status, nil
}
