/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	podutils "github.com/run-ai/runai/runai-cluster/cluster/workload-common/pod-utils"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	"strings"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"

	"github.com/go-logr/logr"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

const (
	TypeInferenceWorkload            workloadCommon.WorkloadType = "Inference"
	TypeDistributedInferenceWorkload workloadCommon.WorkloadType = "DistributedInference"

	LabelInferenceWorkload            string = "run.ai/inferenceworkload"
	LabelDistributedInferenceWorkload string = "run.ai/distributedinferenceworkload"
	NimImagePrefix                           = "nvcr.io/nim/"
)

func InferencePodsFilter() func(ctx context.Context, obj client.Object) []reconcile.Request {
	// returns a function that filters pods and returns the inference workload that the pod belongs to
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		// filter pods that are not part of inference workloads
		name := obj.GetLabels()[LabelInferenceWorkload]
		if name != "" {
			return []reconcile.Request{
				{NamespacedName: types.NamespacedName{
					Name:      name,
					Namespace: obj.GetNamespace(),
				}},
			}
		}
		return []reconcile.Request{}
	}
}

func DistributedInferencePodsFilter() func(ctx context.Context, obj client.Object) []reconcile.Request {
	return func(ctx context.Context, obj client.Object) []reconcile.Request {
		// Check if workloadKind label exists and has the expected value
		workloadKind := obj.GetLabels()[constants.LabelWorkloadKind]
		if workloadKind != constants.KindDistributedInferenceWorkload {
			return []reconcile.Request{}
		}

		// Check if workloadName label exists
		workloadName := obj.GetLabels()[constants.LabelWorkloadName]
		if workloadName == "" {
			return []reconcile.Request{}
		}

		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{
				Name:      workloadName,
				Namespace: obj.GetNamespace(),
			}},
		}
	}
}

func InferencePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			_, ok := e.ObjectOld.(*corev1.Pod)
			if ok {
				return podutils.PodShouldUpdate(e)
			}
			return e.ObjectOld.GetResourceVersion() != e.ObjectNew.GetResourceVersion()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}

// LwsStatusChangesOnlyPredicate returns a predicate that only triggers on status changes of LeaderWorkerSet resources.
// this is required, since without it we get an infinite reconcile loop, as when upserting a LWS, the LWS controller
// makes some changes in the spec (changing some nil entries to their defaults), which triggers the reconcile
func LwsStatusChangesOnlyPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld == nil || e.ObjectNew == nil {
				return false
			}

			oldLWS, ok := e.ObjectOld.(*lwsv1.LeaderWorkerSet)
			if !ok {
				return false
			}
			newLWS, ok := e.ObjectNew.(*lwsv1.LeaderWorkerSet)
			if !ok {
				return false
			}

			// only trigger if status has changed
			return !reflect.DeepEqual(oldLWS.Status, newLWS.Status)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			// always handle create events
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// always handle delete events
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			// handle generic events
			return true
		},
	}
}

// IsNimImage checks if the given image is a NIM image based on the nvcr.io/nim/<framework>:<tag> pattern
func IsNimImage(image string) bool {
	if image == "" {
		return false
	}
	return strings.HasPrefix(image, NimImagePrefix)
}

// AddNimWorkloadLabel adds the NIM workload label if the image is a NIM image
func AddNimWorkloadLabel(labels *flds.DictionaryField, image string) {
	if labels == nil {
		return
	}
	if !IsNimImage(image) {
		return
	}
	labels.Put(constants.LabelNIMWorkload, "true")
}

func PatchWorkloadLabelsAndAnnotations(ctx context.Context, k8sClient client.Client, workload client.Object, log logr.Logger) error {
	log.V(4).Info(fmt.Sprintf("workload %s labels or annotations have changed, patching the workload", workload.GetName()))

	type patchMetadata struct {
		Metadata struct {
			Labels      map[string]string `json:"labels"`
			Annotations map[string]string `json:"annotations"`
		} `json:"metadata"`
	}
	patch := patchMetadata{}
	patch.Metadata.Labels = workload.GetLabels()
	patch.Metadata.Annotations = workload.GetAnnotations()

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshal patch for workload %s: %w", workload.GetName(), err)
	}

	err = k8sClient.Patch(ctx, workload, client.RawPatch(types.MergePatchType, patchBytes))
	if err != nil {
		if apierrors.IsNotFound(err) { // workload not found, skip patching
			log.V(4).Info(fmt.Sprintf("workload %s not found, skipping patching", workload.GetName()))
			return nil
		}
		return fmt.Errorf("failed to patch workload %s: %w", workload.GetName(), err)
	}
	return nil
}
