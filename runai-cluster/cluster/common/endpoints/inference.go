/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
)

const InferenceEndpointReadyCondition = "InferenceEndpointReady"

var InferenceEndpointGVK = schema.GroupVersionKind{Group: "inference.platform.run.ai", Version: "v1alpha1", Kind: "InferenceEndpoint"}

// InferenceConfig is the externalAccess.inference block. It carries only the
// InferenceEndpoint fields that the networking plugin cannot derive from the
// endpoint's Service and HTTPRoute. The values stay JSON so consumers share the
// installed CRD contract without importing the controller's API module.
type InferenceConfig struct {
	Epp        json.RawMessage `json:"epp"`
	Objectives json.RawMessage `json:"objectives,omitempty"`
	Rewrites   json.RawMessage `json:"rewrites,omitempty"`
}

// UnmarshalJSON rejects unknown keys so a field the plugin does not apply is
// never silently ignored.
func (c *InferenceConfig) UnmarshalJSON(data []byte) error {
	type plain InferenceConfig
	var decoded plain
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&decoded); err != nil {
		return fmt.Errorf("decode externalAccess.inference: %w", err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("externalAccess.inference must contain one JSON object")
	}
	*c = InferenceConfig(decoded)
	return nil
}

// Validate checks the shape of the inference block. Kubernetes CRD admission
// validates nested values (profiles, weights, priorities) when the endpoint is applied.
func (c *InferenceConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("externalAccess.inference is required")
	}
	_, err := c.specFields()
	return err
}

func (c *InferenceConfig) specFields() (map[string]any, error) {
	fields := map[string]any{}
	for _, field := range []struct {
		name     string
		raw      json.RawMessage
		required bool
		array    bool
	}{
		{name: "epp", raw: c.Epp, required: true},
		{name: "objectives", raw: c.Objectives, array: true},
		{name: "rewrites", raw: c.Rewrites, array: true},
	} {
		if len(field.raw) == 0 {
			if field.required {
				return nil, fmt.Errorf("externalAccess.inference.%s is required", field.name)
			}
			continue
		}
		path := "externalAccess.inference." + field.name
		var value any
		if err := json.Unmarshal(field.raw, &value); err != nil {
			return nil, fmt.Errorf("decode %s: %w", path, err)
		}
		if !nonEmptyJSON(value, field.array) {
			return nil, fmt.Errorf("%s must be a non-empty %s", path, map[bool]string{true: "array", false: "object"}[field.array])
		}
		if err := rejectInferenceNulls(value, path); err != nil {
			return nil, err
		}
		fields[field.name] = value
	}
	return fields, nil
}

func nonEmptyJSON(value any, array bool) bool {
	switch v := value.(type) {
	case map[string]any:
		return !array && len(v) > 0
	case []any:
		return array && len(v) > 0
	default:
		return false
	}
}

// InferenceEndpointTarget is the part of the InferenceEndpoint spec that the
// networking plugin derives from the endpoint's Service and HTTPRoute.
type InferenceEndpointTarget struct {
	Selector           map[string]string
	Port               int32
	GatewayName        string
	GatewayNamespace   string
	Hostnames          []string
	DefaultBackendName string
	DefaultBackendPort int32
}

// InferenceEndpointName binds identity to the immutable runner UID, keeping all
// generated child names within their DNS-label limit even for long workload names.
// A runner therefore owns at most one InferenceEndpoint.
func InferenceEndpointName(runner metav1.Object) (string, error) {
	name := "inference-" + string(runner.GetUID())
	if runner.GetUID() == "" || len(name) > 53 || len(validation.IsDNS1123Label(name)) > 0 || runner.GetNamespace() == "" {
		return "", fmt.Errorf("inference endpoint requires a saved namespaced runner with a valid UID")
	}
	return name, nil
}

// BuildInferenceEndpoint returns the desired InferenceEndpoint for runner. The
// caller sets the controller reference.
func BuildInferenceEndpoint(runner metav1.Object, cfg *InferenceConfig, target InferenceEndpointTarget) (*unstructured.Unstructured, error) {
	name, err := InferenceEndpointName(runner)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, fmt.Errorf("externalAccess.inference is required")
	}
	fields, err := cfg.specFields()
	if err != nil {
		return nil, err
	}
	if len(target.Selector) == 0 || target.Port < 1 || target.Port > 65535 || target.GatewayName == "" ||
		len(target.Hostnames) == 0 || target.DefaultBackendName == "" || target.DefaultBackendPort < 1 || target.DefaultBackendPort > 65535 {
		return nil, fmt.Errorf("inference endpoint requires a workload selector, port, Gateway, hostname and default backend")
	}
	matchLabels := make(map[string]any, len(target.Selector))
	for key, value := range target.Selector {
		matchLabels[key] = value
	}
	hostnames := make([]any, 0, len(target.Hostnames))
	for _, hostname := range target.Hostnames {
		hostnames = append(hostnames, hostname)
	}
	gatewayRef := map[string]any{"name": target.GatewayName}
	if target.GatewayNamespace != "" {
		gatewayRef["namespace"] = target.GatewayNamespace
	}
	spec := map[string]any{
		"workload": map[string]any{
			"selector": map[string]any{"matchLabels": matchLabels},
			"port":     int64(target.Port),
		},
		"routing": map[string]any{
			"gatewayRef": gatewayRef,
			"hostnames":  hostnames,
			"defaultBackendRef": map[string]any{
				"name": target.DefaultBackendName,
				"port": int64(target.DefaultBackendPort),
			},
		},
	}
	for key, value := range fields {
		spec[key] = value
	}
	// Round-trip through JSON so numbers match what the API server returns (int64),
	// which keeps the desired and live specs comparable.
	raw, err := json.Marshal(map[string]any{
		"apiVersion": InferenceEndpointGVK.GroupVersion().String(),
		"kind":       InferenceEndpointGVK.Kind,
		"metadata":   map[string]any{"name": name, "namespace": runner.GetNamespace()},
		"spec":       spec,
	})
	if err != nil {
		return nil, fmt.Errorf("encode inference endpoint: %w", err)
	}
	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(raw); err != nil {
		return nil, fmt.Errorf("decode inference endpoint: %w", err)
	}
	return obj, nil
}

// InferenceEndpointIndex returns the index of the first endpoint that carries
// externalAccess.inference, or -1 when none does.
func (c *Config) InferenceEndpointIndex() int {
	if c == nil {
		return -1
	}
	for i := range c.Endpoints {
		if ea := c.Endpoints[i].ExternalAccess; ea != nil && ea.Inference != nil {
			return i
		}
	}
	return -1
}

// HasInference reports whether obj's run.ai/endpoints annotation requests
// inference routing. A malformed annotation reports false.
func HasInference(obj metav1.Object) bool {
	cfg, err := Get(obj)
	return err == nil && cfg.InferenceEndpointIndex() >= 0
}

// No fields in the current CRD are nullable. Reject explicit nulls before the
// API server prunes them, so desired and persisted specs can converge.
func rejectInferenceNulls(value any, path string) error {
	switch v := value.(type) {
	case nil:
		return fmt.Errorf("%s must omit unset fields instead of using null", path)
	case map[string]any:
		for key, child := range v {
			if err := rejectInferenceNulls(child, path+"."+key); err != nil {
				return err
			}
		}
	case []any:
		for i, child := range v {
			if err := rejectInferenceNulls(child, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}
