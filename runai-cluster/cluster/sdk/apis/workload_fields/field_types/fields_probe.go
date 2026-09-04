/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_types

// +k8s:deepcopy-gen=true
type ProbeField struct {
	Values *ProbeMembers `json:"values,omitempty"`
}

// +k8s:deepcopy-gen=true
type ProbeMembers struct {
	// Number of seconds after the container has started before the probe is initiated.
	InitialDelaySeconds *IntField `json:"initialDelaySeconds,omitempty"`

	// Number of seconds after which the probe times out.
	// Defaults to 1 second. Minimum value is 1.
	TimeoutSeconds *IntField `json:"timeoutSeconds,omitempty"`

	// How often (in seconds) to perform the probe.
	// Default to 10 seconds. Minimum value is 1.
	PeriodSeconds *IntField `json:"periodSeconds,omitempty"`

	// Minimum consecutive successes for the probe to be considered successful after having failed.
	// Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
	SuccessThreshold *IntField `json:"successThreshold,omitempty"`

	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	// Defaults to 3. Minimum value is 1.
	FailureThreshold *IntField `json:"failureThreshold,omitempty"`

	// Path to access on the HTTP server.
	HttpGetPath *StringField `json:"httpGetPath,omitempty"`

	// Number of the port to access on the container.
	// Must be in the range 1 to 65535.
	HttpGetPort IntField `json:"httpGetPort"`

	// Host name to connect to, defaults to the pod IP.
	HttpGetHost *StringField `json:"httpGetHost,omitempty"`

	// Scheme to use for connecting to the host: HTTP or HTTPS.
	// Defaults to HTTP.
	HttpGetScheme *StringField `json:"httpGetScheme,omitempty"`
}

func (self *ProbeField) Deref() ProbeMembers {
	if self != nil && self.Values != nil {
		return *self.Values
	} else {
		return ProbeMembers{}
	}
}
