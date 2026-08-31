/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v2alpha1

import (
	"encoding/json"

	jsonq "github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// unmarshal the merged values from the merge stack back to a workload so the builder will
// be able to work with it
func UnmarshalWorkload(values *jsonq.JSONQ, masterValues *jsonq.JSONQ, origTypeMeta *metav1.TypeMeta, origObjectMeta *metav1.ObjectMeta) (*DistributedWorkload, error) {

	jsonDump, err := json.Marshal(values.JumpToTop().Get())
	if err != nil {
		return nil, err
	}

	var masterJsonDump []byte
	if masterValues != nil {
		masterJsonDump, err = json.Marshal(masterValues.JumpToTop().Get())
		if err != nil {
			return nil, err
		}
	}

	var mergedWorkload DistributedWorkload
	err = json.Unmarshal(jsonDump, &mergedWorkload.Spec)
	if err != nil {
		return nil, err
	}

	if masterJsonDump != nil {
		err = json.Unmarshal(masterJsonDump, &mergedWorkload.MasterSpec)
		if err != nil {
			return nil, err
		}
	}

	//
	//   complement the object with some information which is missing from the merge stack
	//
	mergedWorkload.TypeMeta = metav1.TypeMeta{
		Kind:       origTypeMeta.Kind,
		APIVersion: origTypeMeta.APIVersion,
	}
	origObjectMeta.DeepCopyInto(&mergedWorkload.ObjectMeta)

	return &mergedWorkload, nil
}
