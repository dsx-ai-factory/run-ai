/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ownership

import (
	"encoding/json"

	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type workloadRunner struct {
	object *unstructured.Unstructured
}

func decodeWorkloadRunner(raw []byte) (*workloadRunner, error) {
	object := &unstructured.Unstructured{}
	if err := json.Unmarshal(raw, &object.Object); err != nil {
		return nil, err
	}
	return &workloadRunner{object: object}, nil
}

func (w *workloadRunner) namespace() string {
	return w.object.GetNamespace()
}

func (w *workloadRunner) ownerUID() string {
	return w.object.GetAnnotations()[clusterconstants.WorkloadOwnerUIDAnnotation]
}

func (w *workloadRunner) controlPlaneManaged() bool {
	return w.object.GetLabels()[commonconstants.ControlPlaneManagedLabel] == "true"
}

func (w *workloadRunner) restoreOwner(stored *workloadRunner) {
	annotations := w.object.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[clusterconstants.WorkloadOwnerUIDAnnotation] = stored.ownerUID()
	w.object.SetAnnotations(annotations)
}

func (w *workloadRunner) marshal() ([]byte, error) {
	return json.Marshal(w.object.Object)
}
