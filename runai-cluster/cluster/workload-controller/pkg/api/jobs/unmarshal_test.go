/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"k8s.io/utils/ptr"

	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {

	origWorkload := JobWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name: "myname",
		},
		Spec: JobSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Environment: flds.NewDictionaryField(map[string]string{
						"HOME": "/home",
					}),
				},
				NetworkingFields: field_groups.NetworkingFields{
					Ports: flds.NewPortsField(map[string]flds.PortItemAndRules{
						"myport": {
							Value: &flds.PortItemValue{
								ContainerPort: ptr.To(1000),
							},
						},
					}),
				},
			},
		},
	}

	origWorkloadQ, err := utils.NewJsonQ(origWorkload.Spec)
	assert.NoError(t, err, "origWorkload -> json")

	mergeStack := enforcer.NewMergeStack(origWorkloadQ, "workload-name", nil, false)

	marshalled, err := UnmarshalWorkload(mergeStack.MergedValues(), &origWorkload.TypeMeta, &origWorkload.ObjectMeta)
	assert.NoError(t, err)
	assert.Equal(t, &origWorkload, marshalled, "unmarshal workload")

}
