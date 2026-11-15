/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/itchyny/gojq"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	workloadsAPI "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/karta/pkg/jq/execution"
)

type podOwnershipTarget struct{ podInputPath, labelsPath string }

// stampPodOwnershipLabels uses only Karta-declared writable Pod inputs. The runner
// copies the manifest, retaining unknown operator fields and the original input.
func stampPodOwnershipLabels(ctx context.Context, manifest map[string]interface{}, workloadID string,
	workloadType *workloadsAPI.WorkloadTypeConfig, info *ResourceInfo,
) (map[string]interface{}, bool, error) {
	if workloadType == nil || len(workloadType.Kartas) == 0 {
		return manifest, false, nil
	}
	karta, kartaErr := decodeKartaForVersion(workloadType, info.Version)
	if kartaErr != nil {
		if kartaErr.Code == http.StatusNotFound {
			return manifest, false, nil
		}
		return nil, false, fmt.Errorf("ownership resource interface: %s", kartaErr.Message)
	}
	components := append([]kartav1alpha1.ComponentDefinition{karta.Spec.StructureDefinition.RootComponent}, karta.Spec.StructureDefinition.ChildComponents...)
	var targets []podOwnershipTarget
	for _, component := range components {
		// Component Kind describes the generated child; its input paths still address this manifest.
		def := component.SpecDefinition
		if def == nil {
			continue
		}
		switch {
		case def.PodTemplateSpecPath != nil:
			p := ownershipPath(*def.PodTemplateSpecPath)
			targets = append(targets, podOwnershipTarget{p, "(" + p + ").metadata.labels"})
		case def.PodSpecPath != nil && def.MetadataPath != nil:
			targets = append(targets, podOwnershipTarget{ownershipPath(*def.PodSpecPath), "(" + ownershipPath(*def.MetadataPath) + ").labels"})
		case def.FragmentedPodSpecDefinition != nil && def.FragmentedPodSpecDefinition.LabelsPath != nil:
			frag := def.FragmentedPodSpecDefinition
			// Fragmented inputs can use alternate containers or have no container
			// mapping. Their owning raw object determines presence, per label path.
			targets = append(targets, podOwnershipTarget{"", ownershipPath(*frag.LabelsPath)})
		}
	}
	if len(targets) == 0 {
		return manifest, false, nil
	}
	runner := execution.NewDefaultRunner(manifest)
	for _, target := range targets {
		var err error
		var podInputs []interface{}
		if target.podInputPath != "" {
			target.podInputPath, err = nullableOwnershipPath(target.podInputPath)
			if err != nil {
				return nil, false, err
			}
			podInputs, err = runner.Evaluate(ctx, target.podInputPath)
			if err != nil {
				return nil, false, fmt.Errorf("ownership anchor %s: %w", target.podInputPath, err)
			}
		}
		target.labelsPath, err = nullableOwnershipPath(target.labelsPath)
		if err != nil {
			return nil, false, err
		}
		paths, err := runner.Evaluate(ctx, "path("+target.labelsPath+")")
		if err != nil {
			return nil, false, fmt.Errorf("ownership labels %s: %w", target.labelsPath, err)
		}
		if target.podInputPath != "" && len(podInputs) != len(paths) {
			return nil, false, fmt.Errorf("ownership label/Pod input count mismatch for %s", target.labelsPath)
		}
		for i, path := range paths {
			// Optional components may be absent or explicitly null. Do not create them.
			if target.podInputPath != "" && podInputs[i] == nil {
				continue
			}
			if target.podInputPath == "" {
				parts, ok := path.([]interface{})
				if !ok || len(parts) == 0 {
					return nil, false, fmt.Errorf("invalid ownership label path %v", path)
				}
				parent := parts[:len(parts)-1]
				if len(parent) > 0 && parent[len(parent)-1] == "metadata" {
					parent = parent[:len(parent)-1]
				}
				encodedParent, err := json.Marshal(parent)
				if err != nil {
					return nil, false, fmt.Errorf("encode ownership parent: %w", err)
				}
				values, err := runner.Evaluate(ctx, "getpath("+string(encodedParent)+")")
				if err != nil {
					return nil, false, fmt.Errorf("evaluate ownership parent: %w", err)
				}
				if len(values) != 1 {
					return nil, false, fmt.Errorf("ownership parent must select one object")
				}
				if values[0] == nil {
					continue
				}
				if _, ok := values[0].(map[string]interface{}); !ok {
					return nil, false, fmt.Errorf("ownership parent is not an object")
				}
			}
			encoded, err := json.Marshal(path)
			if err != nil {
				return nil, false, err
			}
			if err := writePodOwnership(ctx, runner, "getpath("+string(encoded)+")", workloadID); err != nil {
				return nil, false, err
			}
		}
	}
	if err := writePodOwnership(ctx, runner, ".metadata.labels", workloadID); err != nil {
		return nil, false, err
	}
	object, err := runner.GetObject()
	if err != nil {
		return nil, false, err
	}
	result, ok := object.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Errorf("ownership manifest is not an object")
	}
	if err := validatePodOwnershipSelector(result, info); err != nil {
		return nil, false, err
	}
	return result, true, nil
}

func ownershipPath(path string) string {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, ".") || strings.HasPrefix(path, "(") {
		return path
	}
	return "." + path
}

func writePodOwnership(ctx context.Context, runner execution.Runner, path, workloadID string) error {
	values, err := runner.Evaluate(ctx, path)
	if err != nil {
		return err
	}
	if len(values) != 1 {
		return fmt.Errorf("ownership label path %s must select one object", path)
	}
	labels, ok := values[0].(map[string]interface{})
	if values[0] == nil {
		labels = map[string]interface{}{}
	} else if !ok {
		return fmt.Errorf("ownership labels at %s are not an object", path)
	}
	//nolint:staticcheck // Keep the deprecated label for grouping consumers that still select top-owner-uid.
	labels[commonconstants.TopOwnerWorkloadIdLabel], labels[commonconstants.TopOwnerUidLabel] = workloadID, workloadID
	// A copied Kubernetes UID cannot identify the resource that is about to be
	// created. Leave this key absent so EWI can compute its actual owner UID.
	delete(labels, commonconstants.ClusterTopOwnerUidLabel)
	return runner.Assign(ctx, path, labels)
}

// Skip absent optional arrays without suppressing invalid path/type
// errors. Parse paths so quoted keys containing "[]" are never rewritten.
func nullableOwnershipPath(path string) (string, error) {
	query, err := gojq.Parse(path)
	if err != nil {
		return "", err
	}
	var visit func(*gojq.Query)
	visit = func(q *gojq.Query) {
		if q == nil {
			return
		}
		visit(q.Left)
		visit(q.Right)
		term := q.Term
		if term == nil {
			return
		}
		visit(term.Query)
		if term.Func != nil {
			for _, arg := range term.Func.Args {
				visit(arg)
			}
		}
		if term.Index != nil {
			visit(term.Index.Start)
			visit(term.Index.End)
		}
		for _, suffix := range term.SuffixList {
			if suffix.Index != nil {
				visit(suffix.Index.Start)
				visit(suffix.Index.End)
			}
		}
		for i := 0; i < len(term.SuffixList); i++ {
			if !term.SuffixList[i].Iter {
				continue
			}
			prefix := *term
			prefix.SuffixList = term.SuffixList[:i]
			// prefix.String is generated from a valid AST, so this wrapper parses.
			wrapped, _ := gojq.Parse("(" + prefix.String() + " | select(. != null))")
			wrapped.Term.SuffixList = append(wrapped.Term.SuffixList, term.SuffixList[i:]...)
			*term = *wrapped.Term
			i = 0
		}
	}
	visit(query)
	return query.String(), nil
}

// Kubernetes requires these controllers' selectors to match their Pod templates.
// Reject reserved-label conflicts before persisting an unusable workload. Do not
// reinterpret arbitrary custom resources' selector fields as Pod selectors.
func validatePodOwnershipSelector(manifest map[string]interface{}, info *ResourceInfo) error {
	switch info.Group + "/" + info.Kind {
	case "apps/Deployment", "apps/StatefulSet", "apps/DaemonSet", "apps/ReplicaSet", "batch/Job":
	default:
		return nil
	}
	raw, found, err := unstructured.NestedMap(manifest, "spec", "selector")
	if err != nil {
		return fmt.Errorf("ownership selector: %w", err)
	}
	if !found || raw == nil {
		return nil
	}
	var selector metav1.LabelSelector
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw, &selector); err != nil {
		return fmt.Errorf("ownership selector: %w", err)
	}
	compiled, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		return fmt.Errorf("ownership selector: %w", err)
	}
	requirements, _ := compiled.Requirements()
	usesOwnership := false
	for _, requirement := range requirements {
		switch requirement.Key() {
		//nolint:staticcheck // Validate selectors for both ownership labels while legacy grouping consumers remain.
		case commonconstants.TopOwnerWorkloadIdLabel, commonconstants.TopOwnerUidLabel, commonconstants.ClusterTopOwnerUidLabel:
			usesOwnership = true
		}
	}
	if !usesOwnership {
		return nil
	}
	templateLabels, _, err := unstructured.NestedStringMap(manifest, "spec", "template", "metadata", "labels")
	if err != nil {
		return fmt.Errorf("ownership template labels: %w", err)
	}
	if !compiled.Matches(labels.Set(templateLabels)) {
		return fmt.Errorf("spec.selector conflicts with canonical pod ownership labels; use application labels instead of conflicting run.ai ownership labels in the selector")
	}
	return nil
}
