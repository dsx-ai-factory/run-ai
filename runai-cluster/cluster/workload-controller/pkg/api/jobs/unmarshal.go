/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

import (
	"encoding/json"

	jsonq "github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UnmarshalWorkload unmarshal the merged values from the merge stack back to a workload so the builder will
// be able to work with it
func UnmarshalWorkload(values *jsonq.JSONQ, origTypeMeta *metav1.TypeMeta, origObjectMeta *metav1.ObjectMeta) (*JobWorkload, error) {
	jsonDump, err := json.Marshal(values.JumpToTop().Get())
	if err != nil {
		return nil, err
	}

	var mergedWorkload JobWorkload
	err = json.Unmarshal(jsonDump, &mergedWorkload.Spec)
	if err != nil {
		return nil, err
	}

	//
	//   complement the object with information which is missing from the merge stack
	//
	mergedWorkload.TypeMeta = metav1.TypeMeta{
		Kind:       origTypeMeta.Kind,
		APIVersion: origTypeMeta.APIVersion,
	}
	origObjectMeta.DeepCopyInto(&mergedWorkload.ObjectMeta)
	return &mergedWorkload, nil
}
