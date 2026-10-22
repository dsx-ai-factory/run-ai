/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package render turns one InferenceEndpoint into the Kubernetes objects that
// make up its routing stack: an InferencePool, the endpoint-picker (EPP) stack
// and its ingress NetworkPolicy, an optional InferenceModelRewrite, optional
// InferenceObjectives, and an HTTPRoute.
//
// The objects are produced from YAML templates that were derived directly from
// the hand-written Track A lab manifests (test/testdata/golden/03, 04, 05, 06, 07). Keeping
// them as templates rather than typed builders is a deliberate v0 choice: it
// lets the tests assert that the controller emits exactly what the lab applied
// by hand.
package render

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
	// EndpointLabel is set on every generated object and names the owning
	// InferenceEndpoint. It lets humans and tooling find a stack by endpoint name.
	EndpointLabel = "inference.platform.run.ai/endpoint"

	// ConfigHashAnnotation is set on the EPP Deployment's pod template and holds
	// a hash of the EPP ConfigMap content. The EPP reads its configuration once
	// at startup, so a changed hash is what makes Kubernetes roll the pod.
	ConfigHashAnnotation = "inference.platform.run.ai/config-hash"
)

//go:embed templates/*.yaml
var templateFS embed.FS

var templates = template.Must(
	template.New("").Option("missingkey=error").Funcs(template.FuncMap{
		// JSON strings are valid YAML scalars and cannot introduce another document.
		"toJSON": func(value string) (string, error) {
			encoded, err := json.Marshal(value)
			return string(encoded), err
		},
		"dnsLabel": DNSLabel,
	}).ParseFS(templateFS, "templates/*.yaml"),
)

// templateOrder is the apply order: the pool first so the EPP can find it, the
// EPP's NetworkPolicy right after the EPP it selects, the rewrite and objectives
// after the pool they reference, the route last so the Gateway never points at
// a pool that does not exist yet.
var templateOrder = []string{"pool.yaml", "epp.yaml", "networkpolicy.yaml", "rewrite.yaml", "objective.yaml", "route.yaml"}

// maxDNSLabel is the RFC 1123 label length limit Kubernetes applies to names.
const maxDNSLabel = 63

var notDNSLabelChars = regexp.MustCompile(`[^a-z0-9]+`)

// DNSLabel turns a free-form model name into a DNS-1123 label so it can be
// part of an object name: lower case, non-alphanumerics collapsed to one dash,
// no leading or trailing dash, at most 63 characters. Distinct names can
// collide ("Llama3" and "llama3"); validate rejects such input.
func DNSLabel(value string) string {
	label := notDNSLabelChars.ReplaceAllString(strings.ToLower(value), "-")
	label = strings.Trim(label, "-")
	if len(label) > maxDNSLabel {
		label = strings.TrimRight(label[:maxDNSLabel], "-")
	}
	return label
}

// ObjectiveName is the name of the InferenceObjective rendered for one model of
// the given endpoint. The controller uses it to find lanes that left the spec.
func ObjectiveName(endpoint, model string) string {
	return endpoint + "-objective-" + DNSLabel(model)
}

// Scorer is one GIE scoring plugin and its weight in the default profile.
type Scorer struct {
	// Plugin is the GIE plugin type, for example "queue-scorer".
	Plugin string
	// Weight is the relative weight, 0 to 100.
	Weight int32
}

// DefaultScorers are the Track A weights: queue 30, KV cache 20, prefix cache 10.
func DefaultScorers() []Scorer {
	return []Scorer{
		{Plugin: "queue-scorer", Weight: 30},
		{Plugin: "kv-cache-utilization-scorer", Weight: 20},
		{Plugin: "prefix-cache-scorer", Weight: 10},
	}
}

// RewriteTarget is one destination of a rewrite rule.
type RewriteTarget struct {
	Model  string
	Weight int32
}

// Rewrite maps one requested model name to weighted targets. All rewrites of
// an endpoint become rules of a single InferenceModelRewrite.
type Rewrite struct {
	Model   string
	Targets []RewriteTarget
}

// Objective is one priority lane. Each objective becomes its own
// InferenceObjective, because the EPP selects a lane by object name.
type Objective struct {
	Model    string
	Priority int32
}

// Input is everything the templates need. It is a flat projection of the
// InferenceEndpoint spec plus the object's name and namespace.
type Input struct {
	// Name of the InferenceEndpoint. Child objects are named <Name>-pool,
	// <Name>-epp, <Name>-rewrite, <Name>-objective-<model> and <Name>-route.
	Name string
	// Namespace of the InferenceEndpoint and of every child object.
	Namespace string
	// Selector is the pod label selector that becomes the InferencePool selector.
	Selector map[string]string
	// Port is the container port the pods serve on.
	Port int32
	// GatewayName is the admin-owned Gateway the HTTPRoute attaches to.
	GatewayName string
	// GatewayNamespace defaults to Namespace when empty.
	GatewayNamespace string
	// MetricsNamespace is the namespace Prometheus scrapes from; the EPP
	// NetworkPolicy admits metrics traffic only from there. The controller
	// supplies the default, so the renderer requires it.
	MetricsNamespace string
	// Hostnames the HTTPRoute matches.
	Hostnames []string
	// EppImage is the endpoint-picker container image. The controller resolves
	// epp.profile to its pinned image before rendering.
	EppImage string
	// Scorers and their weights. Empty means DefaultScorers.
	Scorers []Scorer
	// RuntimeProfiles selects the llm-d v0.10 configuration contract.
	RuntimeProfiles bool
	// Profiles are loaded at startup; ActiveProfile only controls the route.
	Profiles       []SchedulingProfile
	DefaultProfile string
	ActiveProfile  string
	// ProfilePlugins is the sorted union of scoring plugins used by Profiles.
	// It is computed during validation; callers need not populate it.
	ProfilePlugins []string
	// Rewrites are model aliases and canary splits. Empty means no
	// InferenceModelRewrite is rendered.
	Rewrites []Rewrite
	// Objectives are priority lanes. Empty means no InferenceObjective is rendered.
	Objectives []Objective
}

func (in *Input) validate() error {
	var errs []error
	if in.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if in.Namespace == "" {
		errs = append(errs, errors.New("namespace is required"))
	}
	if len(in.Selector) == 0 {
		errs = append(errs, errors.New("workload.selector needs at least one label"))
	}
	if in.Port < 1 || in.Port > 65535 {
		errs = append(errs, fmt.Errorf("workload.port %d is outside 1-65535", in.Port))
	}
	if in.GatewayName == "" {
		errs = append(errs, errors.New("routing.gatewayRef.name is required"))
	}
	if len(in.Hostnames) == 0 {
		errs = append(errs, errors.New("routing.hostnames needs at least one entry"))
	}
	if in.EppImage == "" {
		errs = append(errs, errors.New("epp.image is required"))
	}
	// A namespace is a DNS label, so a valid value is unchanged by DNSLabel.
	if in.MetricsNamespace == "" || DNSLabel(in.MetricsNamespace) != in.MetricsNamespace {
		errs = append(errs, fmt.Errorf("metrics namespace %q is not a DNS label", in.MetricsNamespace))
	}
	for i, s := range in.Scorers {
		if s.Plugin == "" {
			errs = append(errs, fmt.Errorf("epp.scheduling.scorers[%d] has no plugin", i))
		}
		if s.Weight < 0 || s.Weight > 100 {
			errs = append(errs, fmt.Errorf("epp.scheduling.scorers[%d] weight %d is outside 0-100", i, s.Weight))
		}
	}
	for i, rw := range in.Rewrites {
		if rw.Model == "" {
			errs = append(errs, fmt.Errorf("rewrites[%d] has no model", i))
		}
		if len(rw.Targets) == 0 {
			errs = append(errs, fmt.Errorf("rewrites[%d] has no targets", i))
		}
		for j, t := range rw.Targets {
			if t.Model == "" || t.Weight < 1 {
				errs = append(errs, fmt.Errorf("rewrites[%d].targets[%d] needs a model and a weight >= 1", i, j))
			}
		}
	}
	seenModels := map[string]int{}
	seenNames := map[string]int{}
	for i, o := range in.Objectives {
		if o.Model == "" {
			errs = append(errs, fmt.Errorf("objectives[%d] has no model", i))
			continue
		}
		if o.Priority < 0 {
			errs = append(errs, fmt.Errorf("objectives[%d] priority %d is below 0", i, o.Priority))
		}
		if j, dup := seenModels[o.Model]; dup {
			errs = append(errs, fmt.Errorf("objectives[%d] repeats the model of objectives[%d]", i, j))
			continue
		}
		seenModels[o.Model] = i
		name := ObjectiveName(in.Name, o.Model)
		if DNSLabel(o.Model) == "" {
			errs = append(errs, fmt.Errorf("objectives[%d] model %q has no DNS label characters", i, o.Model))
		} else if j, dup := seenNames[name]; dup {
			errs = append(errs, fmt.Errorf("objectives[%d] and objectives[%d] both map to InferenceObjective %s", i, j, name))
		}
		seenNames[name] = i
	}
	if in.GatewayNamespace == "" {
		in.GatewayNamespace = in.Namespace
	}
	if err := in.prepareProfiles(); err != nil {
		errs = append(errs, err)
	}
	if !in.RuntimeProfiles && len(in.Scorers) == 0 {
		in.Scorers = DefaultScorers()
	}
	return errors.Join(errs...)
}

// Render returns the child objects for the given input in apply order. Every
// object carries EndpointLabel=<Name>, and the EPP Deployment's pod template
// carries ConfigHashAnnotation so that a ConfigMap change rolls the EPP. Owner
// references are the caller's job, because they need the parent's UID.
func Render(in Input) ([]*unstructured.Unstructured, error) {
	if err := in.validate(); err != nil {
		return nil, fmt.Errorf("invalid InferenceEndpoint %s/%s: %w", in.Namespace, in.Name, err)
	}

	var out []*unstructured.Unstructured
	for _, name := range templateOrder {
		var buf bytes.Buffer
		if err := templates.ExecuteTemplate(&buf, name, in); err != nil {
			return nil, fmt.Errorf("render %s: %w", name, err)
		}
		objs, err := decodeAll(buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", name, err)
		}
		for _, obj := range objs {
			labels := obj.GetLabels()
			if labels == nil {
				labels = map[string]string{}
			}
			labels[EndpointLabel] = in.Name
			obj.SetLabels(labels)
		}
		out = append(out, objs...)
	}

	if err := stampConfigHash(out); err != nil {
		return nil, err
	}
	return out, nil
}

// stampConfigHash hashes the EPP ConfigMap data and writes it onto the EPP
// Deployment's pod template as an annotation.
func stampConfigHash(objs []*unstructured.Unstructured) error {
	var configMap, deployment *unstructured.Unstructured
	for _, obj := range objs {
		switch obj.GetKind() {
		case "ConfigMap":
			configMap = obj
		case "Deployment":
			deployment = obj
		}
	}
	if configMap == nil || deployment == nil {
		return errors.New("rendered stack has no EPP ConfigMap or Deployment")
	}

	data, _, _ := unstructured.NestedStringMap(configMap.Object, "data")
	canonical, err := json.Marshal(data) // map keys are sorted by encoding/json
	if err != nil {
		return fmt.Errorf("hash EPP config: %w", err)
	}
	sum := sha256.Sum256(canonical)

	annotations, _, _ := unstructured.NestedStringMap(deployment.Object, "spec", "template", "metadata", "annotations")
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[ConfigHashAnnotation] = hex.EncodeToString(sum[:])
	return unstructured.SetNestedStringMap(deployment.Object, annotations, "spec", "template", "metadata", "annotations")
}

// decodeAll splits a multi-document YAML stream into unstructured objects,
// skipping empty documents.
func decodeAll(data []byte) ([]*unstructured.Unstructured, error) {
	dec := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 4096)
	var out []*unstructured.Unstructured
	for {
		obj := &unstructured.Unstructured{}
		if err := dec.Decode(obj); err != nil {
			if errors.Is(err, io.EOF) {
				return out, nil
			}
			return nil, err
		}
		if obj.Object == nil || obj.GetKind() == "" {
			continue
		}
		out = append(out, obj)
	}
}
