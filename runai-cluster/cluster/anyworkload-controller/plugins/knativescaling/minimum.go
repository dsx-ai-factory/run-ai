/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package knativescaling

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/version"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	servingAPI = "serving.knative.dev/v1"
	minScale   = "autoscaling.knative.dev/min-scale"
)

var (
	minimumPath           = []string{"spec", "template", "metadata", "annotations", minScale}
	mutableMinimumVersion = version.MustParseSemantic("1.21.0")
)

// Prepare applies a minimum-only decrease to the serving Revision. The caller
// still applies the Service. Its template keeps the original minimum until a
// change to the workload actually requires a rollout. Run:ai's workload spec
// continues to hold the requested minimum across retries and future rollouts.
func Prepare(ctx context.Context, c client.Client, desired *unstructured.Unstructured, preview func(context.Context, *unstructured.Unstructured) (*unstructured.Unstructured, error)) error {
	if desired.GetAPIVersion() != servingAPI || desired.GetKind() != "Service" {
		return nil
	}
	requested, exists, err := unstructured.NestedString(desired.Object, minimumPath...)
	if err != nil || !exists {
		return err
	}
	service := object(servingAPI, "Service")
	if err := c.Get(ctx, client.ObjectKeyFromObject(desired), service); err != nil {
		return client.IgnoreNotFound(err)
	}
	owner := metav1.GetControllerOf(desired)
	if owner == nil || owner.UID == "" || !reflect.DeepEqual(owner, metav1.GetControllerOf(service)) || service.GetDeletionTimestamp() != nil {
		return nil
	}
	baseline, exists, err := unstructured.NestedString(service.Object, minimumPath...)
	if err != nil || !exists {
		return err
	}
	if !canLowerFrom(baseline, requested) {
		return nil
	}
	name, ready, err := servingRevision(service)
	if err != nil || name == "" {
		return err
	}
	revision := object(servingAPI, "Revision")
	if err := c.Get(ctx, client.ObjectKey{Namespace: service.GetNamespace(), Name: name}, revision); err != nil {
		// Status can still name a deleted Revision. A normal Service apply
		// must remain available to recover the workload.
		return client.IgnoreNotFound(err)
	}
	annotations := revision.GetAnnotations()
	if class := annotations["autoscaling.knative.dev/class"]; class != "" && class != "kpa.autoscaling.knative.dev" {
		return nil
	}
	effective := annotations[minScale]
	if (requested == baseline && effective == requested) || (!ready && effective == baseline) {
		return nil
	}

	enabled, err := supportsMutableMinimum(ctx, c)
	if err != nil || !enabled {
		return err
	}

	candidate := desired.DeepCopy()
	candidate.SetResourceVersion(service.GetResourceVersion())
	result, err := preview(ctx, candidate.DeepCopy())
	if err != nil {
		return fmt.Errorf("preview Knative Service update: %w", err)
	}
	matches, err := onlyMinimumChanged(service, result, baseline, requested)
	if err != nil || !matches {
		return err
	}
	configuration := object(servingAPI, "Configuration")
	if err := c.Get(ctx, client.ObjectKeyFromObject(service), configuration); err != nil {
		return fmt.Errorf("get serving Configuration: %w", err)
	}
	if configuration.GetDeletionTimestamp() != nil || revision.GetDeletionTimestamp() != nil ||
		!metav1.IsControlledBy(configuration, service) || !metav1.IsControlledBy(revision, configuration) {
		return fmt.Errorf("serving Revision %s no longer belongs to Service %s", name, service.GetName())
	}
	if ready && effective != requested {
		before := revision.DeepCopy()
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations[minScale] = requested
		revision.SetAnnotations(annotations)
		if err := c.Patch(ctx, revision, client.MergeFromWithOptions(before, client.MergeFromWithOptimisticLock{})); err != nil {
			return fmt.Errorf("update minimum on serving Revision: %w", err)
		}
	}
	// Pin the subsequent Service apply to the snapshot used to select the
	// serving Revision, including when retrying an already-applied decrease.
	if err := unstructured.SetNestedField(candidate.Object, baseline, minimumPath...); err != nil {
		return err
	}
	*desired = *candidate
	return nil
}

func canLowerFrom(baseline, requested string) bool {
	start, startErr := strconv.ParseInt(baseline, 10, 32)
	end, endErr := strconv.ParseInt(requested, 10, 32)
	return startErr == nil && endErr == nil && start > 0 && end >= 0 && end <= start
}

// The preview includes admission defaults. Excluding just the requested
// minimum from its spec comparison preserves normal image and traffic rollouts.
func onlyMinimumChanged(service, result *unstructured.Unstructured, baseline, requested string) (bool, error) {
	applied, _, err := unstructured.NestedString(result.Object, minimumPath...)
	if err != nil || applied != requested || result.GetUID() != service.GetUID() {
		return false, err
	}
	if err := unstructured.SetNestedField(result.Object, baseline, minimumPath...); err != nil {
		return false, err
	}
	return reflect.DeepEqual(result.Object["spec"], service.Object["spec"]), nil
}

func object(apiVersion, kind string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(apiVersion)
	obj.SetKind(kind)
	return obj
}

func supportsMutableMinimum(ctx context.Context, c client.Reader) (bool, error) {
	// From Knative 1.21, the Revision controller propagates autoscaling
	// annotations to the existing PodAutoscaler and omits them from Pods.
	crd := object("apiextensions.k8s.io/v1", "CustomResourceDefinition")
	if err := c.Get(ctx, client.ObjectKey{Name: "revisions.serving.knative.dev"}, crd); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("get Knative Revision CRD version: %w", err)
	}
	v, err := version.ParseSemantic(crd.GetLabels()["app.kubernetes.io/version"])
	return err == nil && v.AtLeast(mutableMinimumVersion), nil
}

func servingRevision(service *unstructured.Unstructured) (string, bool, error) {
	type trafficTarget struct {
		RevisionName string `json:"revisionName"`
		Percent      int64  `json:"percent"`
	}
	var status struct {
		ObservedGeneration        int64              `json:"observedGeneration"`
		LatestReadyRevisionName   string             `json:"latestReadyRevisionName"`
		LatestCreatedRevisionName string             `json:"latestCreatedRevisionName"`
		Conditions                []metav1.Condition `json:"conditions"`
		Traffic                   []trafficTarget    `json:"traffic"`
	}
	values, found, err := unstructured.NestedMap(service.Object, "status")
	if err != nil || !found {
		return "", false, err
	}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(values, &status); err != nil {
		return "", false, fmt.Errorf("read Knative Service status: %w", err)
	}
	// A different latest-created Revision or explicitly split traffic belongs to
	// a rollout. Leave that rollout to Knative. During a transient loss of
	// readiness, the last stable Revision can still identify a prior override.
	name := status.LatestReadyRevisionName
	if name == "" || name != status.LatestCreatedRevisionName || len(status.Traffic) > 1 {
		return "", false, nil
	}
	if len(status.Traffic) == 0 || status.ObservedGeneration != service.GetGeneration() {
		return name, false, nil
	}
	if target := status.Traffic[0]; target.RevisionName != name || target.Percent != 100 {
		return "", false, nil
	}
	for _, condition := range status.Conditions {
		if condition.Type == "Ready" && condition.Status == metav1.ConditionTrue {
			return name, true, nil
		}
	}
	return name, false, nil
}
