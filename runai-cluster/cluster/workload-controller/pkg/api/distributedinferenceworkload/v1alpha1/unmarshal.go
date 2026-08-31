/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	"encoding/json"

	jsonq "github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UnmarshalWorkload unmarshal the merged values from the merge stack back to a workload so the builder will
// be able to work with it
func UnmarshalWorkload(values *jsonq.JSONQ, origObjectMeta *metav1.ObjectMeta) (*DistributedInferenceWorkload, error) {
	jsonDump, err := json.Marshal(values.JumpToTop().Get())
	if err != nil {
		return nil, err
	}

	var mergedWorkload DistributedInferenceWorkload
	err = json.Unmarshal(jsonDump, &mergedWorkload.Spec)
	if err != nil {
		return nil, err
	}

	//
	//   complement the object with some information which is missing from the merge stack
	//
	origObjectMeta.DeepCopyInto(&mergedWorkload.ObjectMeta)

	return &mergedWorkload, nil
}
