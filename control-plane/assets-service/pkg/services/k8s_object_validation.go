/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package services (file k8s_object_validation.go): Shared validation for the
// k8s-object family of asset kinds (resource claim template today; device
// classes, GRES, etc. tomorrow). All k8s-object assets carry the same
// K8sObjectMeta shape and store a free-form ObjectSpec that must conform to a
// static K8s OpenAPI schema plus a typed struct for strict-unknown-field
// checks. Kind-specific service handlers call these helpers; they own no
// per-kind state.

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	jsschema "github.com/santhosh-tekuri/jsonschema/v6"
	"k8s.io/apimachinery/pkg/util/validation"
)

// JSON paths echoed into validation error messages. Kept as constants so
// tests can reference the same strings without drift, and future k8s-object
// asset kinds can point to a single canonical spelling.
const (
	PathObjectMetaLabels      = "spec.objectMeta.labels"
	PathObjectMetaAnnotations = "spec.objectMeta.annotations"
)

// Kubernetes label/annotation key + label value regexes. The K8s
// apimachinery package exposes IsDNS1123Label / IsDNS1123Subdomain that
// enforce the same rules but produce a slice-of-strings error shape that's
// awkward to plumb through *api.Error, so we own the regex here and reuse
// the package's length constants (validation.DNS1123LabelMaxLength = 63,
// validation.DNS1123SubdomainMaxLength = 253).
//
//	Qualified-name key: [prefix/]name where
//	  prefix (optional) is a DNS subdomain (≤DNS1123SubdomainMaxLength chars)
//	  name   (required) is [A-Za-z0-9] with `.` `_` `-` inside (≤DNS1123LabelMaxLength chars)
//	Label value: empty OR same character class as name (≤DNS1123LabelMaxLength chars)
//
// See https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set
var (
	dnsLabelSubsetRegex = regexp.MustCompile(`^[A-Za-z0-9]([-A-Za-z0-9_.]*[A-Za-z0-9])?$`)
	dnsSubdomainRegex   = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
)

// validateK8sObjectMeta enforces Kubernetes-compatible key syntax on labels
// and annotations, and value syntax on labels. The K8sObjectMeta payload is
// sync'd to the cluster as-is, so rejecting invalid entries at asset-create /
// asset-update time surfaces the error to the user before the sync agent
// tries to POST the malformed resource to the K8s API server. OpenAPI 3.0
// can't express key patterns on map types (that's a 3.1 propertyNames
// feature), so this is where key validation lives.
func validateK8sObjectMeta(meta *api.K8sObjectMeta) *api.Error {
	if meta == nil {
		return nil
	}
	if meta.Labels != nil {
		for key, value := range *meta.Labels {
			if err := validateQualifiedName(key, PathObjectMetaLabels); err != nil {
				return err
			}
			if err := validateLabelValue(key, value); err != nil {
				return err
			}
		}
	}
	if meta.Annotations != nil {
		// Annotation values have no syntactic constraint (arbitrary strings
		// up to the maxLength enforced by the generated model constraints);
		// only the key is validated here.
		for key := range *meta.Annotations {
			if err := validateQualifiedName(key, PathObjectMetaAnnotations); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateQualifiedName(key, path string) *api.Error {
	if len(key) == 0 {
		return api.NewValidationError(fmt.Sprintf("%s: key must not be empty", path))
	}
	var prefix, name string
	if p, n, found := strings.Cut(key, "/"); found {
		prefix, name = p, n
	} else {
		name = key
	}
	if len(name) == 0 || len(name) > validation.DNS1123LabelMaxLength || !dnsLabelSubsetRegex.MatchString(name) {
		return api.NewValidationError(fmt.Sprintf(
			"%s: key %q name component is not a valid Kubernetes qualified name (must be [A-Za-z0-9] with `.`, `_`, `-` allowed inside, ≤%d chars)",
			path, key, validation.DNS1123LabelMaxLength))
	}
	if prefix != "" {
		if len(prefix) > validation.DNS1123SubdomainMaxLength || !dnsSubdomainRegex.MatchString(prefix) {
			return api.NewValidationError(fmt.Sprintf(
				"%s: key %q prefix component is not a valid DNS subdomain (≤%d chars, lowercase alphanumeric + `-`, dot-separated)",
				path, key, validation.DNS1123SubdomainMaxLength))
		}
	}
	return nil
}

func validateLabelValue(key, value string) *api.Error {
	if value == "" {
		return nil // empty label value is valid per K8s spec
	}
	if len(value) > validation.DNS1123LabelMaxLength || !dnsLabelSubsetRegex.MatchString(value) {
		return api.NewValidationError(fmt.Sprintf(
			"%s[%q]: value %q is not a valid Kubernetes label value (must be [A-Za-z0-9] with `.`, `_`, `-` allowed inside, ≤%d chars)",
			PathObjectMetaLabels, key, value, validation.DNS1123LabelMaxLength))
	}
	return nil
}

// validateK8sObjectSpec checks that the free-form objectSpec conforms to a
// compiled Kubernetes OpenAPI schema (layer 1) and can decode into the typed
// K8s struct T with strict-unknown-field checking (layer 2). Callers specialise
// the type parameter to their kind (e.g. resourcev1.ResourceClaimSpec for RCT).
//
// Two-layer validation:
//  1. OpenAPI JSON Schema (santhosh-tekuri) — types, enums, required fields,
//     and constraints declared in the schema (min/max, length, ...).
//  2. Typed struct decode with json.DisallowUnknownFields — rejects
//     misspelled fields. Necessary because K8s publishes permissive schemas:
//     without `additionalProperties: false` (which they never set), JSON
//     Schema allows any extra key, so e.g. `{"devics": {}}` would pass layer 1.
//
// Does NOT catch cluster-side concerns (referenced-resource existence, CEL
// selector evaluation) — those are checked at apply time by the K8s API server
// and surfaced back through the asset's cluster-status.
//
// pathPrefix is the JSON path echoed into error messages ("spec.objectSpec"
// for every k8s-object asset today). kindLabel identifies the target K8s
// schema for the layer-1 error message ("resource.k8s.io/v1 ResourceClaimSpec").
func validateK8sObjectSpec[T any](
	objectSpec map[string]any,
	schema *jsschema.Schema,
	pathPrefix string,
	kindLabel string,
) *api.Error {
	if len(objectSpec) == 0 {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("%s is required", pathPrefix), nil)
	}
	if schema == nil {
		return api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("%s: schema-based validator is not initialized; check server logs", pathPrefix), nil)
	}

	// Layer 1: JSON Schema.
	if err := schema.Validate(objectSpec); err != nil {
		return api.NewError(http.StatusBadRequest,
			fmt.Sprintf("%s does not conform to %s: %v", pathPrefix, kindLabel, err), nil)
	}

	// Layer 2: typed struct decode with strict unknown-field checking.
	raw, err := json.Marshal(objectSpec)
	if err != nil {
		return api.NewError(http.StatusBadRequest,
			fmt.Sprintf("%s: marshal failed: %v", pathPrefix, err), nil)
	}

	var typed T
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&typed); err != nil {
		return api.NewError(http.StatusBadRequest,
			fmt.Sprintf("%s: %v", pathPrefix, err), nil)
	}

	return nil
}
