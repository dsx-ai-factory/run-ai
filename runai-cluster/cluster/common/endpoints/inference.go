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

const InferenceEndpointAnnotation = "run.ai/inference-endpoint"

const InferenceEndpointPluginName = "inference-endpoint"
const InferenceEndpointReadyCondition = "InferenceEndpointReady"

var InferenceEndpointGVK = schema.GroupVersionKind{Group: "inference.platform.run.ai", Version: "v1alpha1", Kind: "InferenceEndpoint"}

// InferenceEndpointConfig carries resolved routing intent. The spec remains JSON
// so consumers share the installed CRD contract without importing a controller
// executable module or maintaining a second copy of its nested API types.
// Identity, ownership and status cannot be supplied through this envelope.
type InferenceEndpointConfig struct {
	SchemaVersion string          `json:"schemaVersion"`
	Spec          json.RawMessage `json:"spec"`
}

func GetInferenceEndpoint(obj metav1.Object) (*InferenceEndpointConfig, error) {
	raw, found := obj.GetAnnotations()[InferenceEndpointAnnotation]
	if !found {
		return nil, nil
	}
	var cfg *InferenceEndpointConfig
	dec := json.NewDecoder(bytes.NewBufferString(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode %s: %w", InferenceEndpointAnnotation, err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("%s must contain one JSON object", InferenceEndpointAnnotation)
	}
	if _, err := inferenceObject(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func SetInferenceEndpoint(obj metav1.Object, cfg *InferenceEndpointConfig) error {
	if cfg == nil {
		return fmt.Errorf("inference endpoint configuration is required")
	}
	copy := *cfg
	if copy.SchemaVersion == "" {
		copy.SchemaVersion = SchemaVersionV1
	}
	if _, err := inferenceObject(&copy); err != nil {
		return err
	}
	raw, err := json.Marshal(copy)
	if err != nil {
		return fmt.Errorf("marshal inference endpoint: %w", err)
	}
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[InferenceEndpointAnnotation] = string(raw)
	obj.SetAnnotations(annotations)
	return nil
}

// InferenceEndpointName binds identity to the immutable runner UID, keeping all
// generated child names within their DNS-label limit even for long workload names.
func InferenceEndpointName(runner metav1.Object) (string, error) {
	name := "inference-" + string(runner.GetUID())
	if runner.GetUID() == "" || len(name) > 53 || len(validation.IsDNS1123Label(name)) > 0 || runner.GetNamespace() == "" {
		return "", fmt.Errorf("inference endpoint requires a saved namespaced runner with a valid UID")
	}
	return name, nil
}

func BuildInferenceEndpoint(runner metav1.Object, cfg *InferenceEndpointConfig) (*unstructured.Unstructured, error) {
	name, err := InferenceEndpointName(runner)
	if err != nil {
		return nil, err
	}
	obj, err := inferenceObject(cfg)
	if err != nil {
		return nil, err
	}
	obj.SetName(name)
	obj.SetNamespace(runner.GetNamespace())
	return obj, nil
}

func inferenceObject(cfg *InferenceEndpointConfig) (*unstructured.Unstructured, error) {
	if cfg == nil || cfg.SchemaVersion != SchemaVersionV1 {
		return nil, fmt.Errorf("inference endpoint schemaVersion must be %q", SchemaVersionV1)
	}
	raw, err := json.Marshal(struct {
		APIVersion string          `json:"apiVersion"`
		Kind       string          `json:"kind"`
		Spec       json.RawMessage `json:"spec"`
	}{InferenceEndpointGVK.GroupVersion().String(), InferenceEndpointGVK.Kind, cfg.Spec})
	if err != nil {
		return nil, fmt.Errorf("encode inference endpoint spec: %w", err)
	}
	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(raw); err != nil {
		return nil, fmt.Errorf("decode inference endpoint spec: %w", err)
	}
	if err := rejectInferenceNulls(obj.Object["spec"], "spec"); err != nil {
		return nil, err
	}
	// Validate the envelope here; Kubernetes CRD admission validates all nested
	// values (including profiles, weights, hostnames and selectors) on apply.
	for _, section := range []string{"workload", "routing", "epp"} {
		value, found, err := unstructured.NestedMap(obj.Object, "spec", section)
		if err != nil || !found || len(value) == 0 {
			return nil, fmt.Errorf("inference endpoint spec.%s must be a non-empty object", section)
		}
	}
	return obj, nil
}

// No fields in the current CRD are nullable. Reject explicit nulls before the
// API server prunes them, so desired and persisted specs can converge.
func rejectInferenceNulls(value any, path string) error {
	switch v := value.(type) {
	case nil:
		return fmt.Errorf("inference endpoint %s must omit unset fields instead of using null", path)
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
