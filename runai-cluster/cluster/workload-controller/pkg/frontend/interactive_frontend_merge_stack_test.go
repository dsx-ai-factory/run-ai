/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package frontend

import (
	"context"
	"fmt"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer/factory"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"k8s.io/utils/ptr"

	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
)

func TestSubmitWorkloadResideInProject(t *testing.T) {

	//                              usage              namespace
	testSubmitWorkloadInProject(t, common.UsedForSubmit, "runai-team-a")
	testSubmitWorkloadInProject(t, common.UsedAsTemplate, "runai-team-a")
	testSubmitWorkloadInProject(t, common.UsedForSubmit, constants.NamespaceRunAi)
	testSubmitWorkloadInProject(t, common.UsedAsTemplate, constants.NamespaceRunAi)
}

func testSubmitWorkloadInProject(t *testing.T, usage string, ns string) {

	workloadInRunai := v2alpha1.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mytemp",
			Namespace: ns,
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					Usage: usage,
				},
			},
		},
	}

	fakeClient := NewFakeClient(t, workloadproject.EmulateProjNs("runai-team-a", "team-a"), workloadproject.EmulateRunaiNs(),
		&workloadInRunai)

	desc := fmt.Sprintf("usage=Submit-in-project usage? %v ns/proj? %v", usage, ns)

	_, err := InteractiveCreateMergeStackForFrontend(fakeClient, &workloadInRunai, WebhookOptions{})

	if usage == common.UsedForSubmit && ns == constants.NamespaceRunAi {
		assert.Error(t, err, desc)
		if err != nil {
			assert.Contains(t, err.Error(), "in a project", desc)
		}
	} else {
		assert.NoError(t, err, desc)
	}
}

func TestBaseAllowedOnlyForSubmit(t *testing.T) {

	//
	//                                          usage              base
	testBaseAllowedOnlyForSubmit(t, common.UsedForSubmit, ptr.To("base"))
	testBaseAllowedOnlyForSubmit(t, common.UsedAsTemplate, ptr.To("base"))
	testBaseAllowedOnlyForSubmit(t, common.UsedForSubmit, ptr.To(""))
	testBaseAllowedOnlyForSubmit(t, common.UsedAsTemplate, ptr.To(""))
	testBaseAllowedOnlyForSubmit(t, common.UsedForSubmit, nil)
	testBaseAllowedOnlyForSubmit(t, common.UsedAsTemplate, nil)
}

func testBaseAllowedOnlyForSubmit(t *testing.T, usage string, base *string) {

	baseWorkload := v2alpha1.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "base",
			Namespace: constants.NamespaceRunAi,
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					Usage:        usage,
					BaseWorkload: base,
				},
			},
		},
	}

	workload := v2alpha1.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mytemp",
			Namespace: "runai-team-a",
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					Usage:        usage,
					BaseWorkload: base,
				},
			},
		},
	}

	fackeClient := NewFakeClient(t, workloadproject.EmulateProjNs("runai-team-a", "team-a"), workloadproject.EmulateRunaiNs(),
		&baseWorkload, &workload)

	baseName := ptr.Deref(base, "")

	desc := fmt.Sprintf("base-only-for-submit usage? %v base? '%v'", usage, baseName)

	_, err := InteractiveCreateMergeStackForFrontend(fackeClient, &workload, WebhookOptions{})

	if usage != common.UsedForSubmit && baseName != "" {
		assert.Error(t, err, desc)
		if err != nil {
			assert.NotContains(t, err.Error(), "baseWorkload is only allowed for workload with usage=Submit", desc)
		}
	} else {
		assert.NoError(t, err, desc)
	}
}

func TestLoadPreviousJob(t *testing.T) {

	workingDirFromBase := "working-dir-from-base"
	baseName := "base"
	basedOnSpecName := "based-on-spec"

	//
	//   setup base template
	//
	baseTemplate := v2alpha1.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      baseName,
			Namespace: "runai-team-a",
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name-from-base"),
					},
				},
				ContainerFields: field_groups.ContainerFields{
					WorkingDir: flds.NewStringField(workingDirFromBase),
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				Gpu: flds.NewOptionalNumberField(ptr.To[float64](4.0)),
			},
		},
	}

	//
	//    setup workload
	//
	basedOnSpec := v2alpha1.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      basedOnSpecName,
			Namespace: "runai-team-a",
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					BaseWorkload: ptr.To(baseName),
				},
			},
		},
	}

	fakeClient := NewFakeClient(t,
		&baseTemplate, &basedOnSpec)

	/////////////////////////////////////////////////////////////////////////////////////////
	//    Part 1 --> no value set in status so take from the base spec itself
	/////////////////////////////////////////////////////////////////////////////////////////
	interactiveMergeCreator := factory.NewInteractiveWorkloadMergeCreator(fakeClient, false)
	mergeStackMap, err := interactiveMergeCreator.CreateMergeStack(context.TODO(), &basedOnSpec)
	assert.NoError(t, err, "part1 -> build the stack")
	mergeStack1 := mergeStackMap[api.SpecSubResource]
	err = mergeStack1.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{
		DisableEnforce: true,
	})
	assert.NoError(t, err, "part1 -> merge")

	//
	//   the expected results are:
	//   name - not given (we do not take name from base)
	//   gpu - taken from the workload (2.0)
	//   workingDir - taken from the base
	//
	merged1, err := jobs.UnmarshalWorkload(mergeStack1.MergedValues(), &basedOnSpec.TypeMeta, &basedOnSpec.ObjectMeta)
	assert.NoError(t, err, "part1 - unmarshal")

	assert.Nil(t, err, merged1.Name, "part1 - name - nil")
	assert.Equal(t, "4", ptr.Deref(merged1.Spec.Gpu.Value, ""), "part1 - gpu - from base")
	assert.Equal(t, workingDirFromBase, merged1.Spec.WorkingDir.Deref(), "part1 - workingDir - from workload")
}

func TestLoadPolicy(t *testing.T) {

	completionFromSett := 3
	completionFromBase := 5
	completionFromTemp := 7
	maxCompletionValue := 6

	policy := ipv1.InteractivePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "policy",
			Namespace: constants.NamespaceRunAi,
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					RunAsUid: &flds.IntField{
						Value: ptr.To(completionFromSett),
						Rules: &flds.IntRules{
							Max: ptr.To(maxCompletionValue),
						},
					},
					RunAsUser: flds.NewOptionalBoolField(ptr.To(true)),
				},
			},
		},
	}

	anotherPolicy := tpv1.TrainingPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ignored-policy",
			Namespace: constants.NamespaceRunAi,
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					RunAsUid: &flds.IntField{
						Value: ptr.To(completionFromSett),
						Rules: &flds.IntRules{
							Max: ptr.To(1),
						},
					},
					RunAsUser: flds.NewOptionalBoolField(ptr.To(false)),
				},
			},
		},
	}

	base := v2alpha1.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "base",
			Namespace: "runai-team-a",
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					RunAsUid: flds.NewOptionalIntField(ptr.To(completionFromBase)),
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				GpuMemory: flds.NewQuantityField("123M"),
			},
		},
	}

	temp := v2alpha1.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "temp",
			Namespace: "runai-team-a",
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					Usage:        common.UsedForSubmit,
					BaseWorkload: ptr.To("base"),
				},
				ContainerFields: field_groups.ContainerFields{
					Image:    flds.NewStringField("ubuntu"),
					RunAsUid: flds.NewOptionalIntField(ptr.To(completionFromTemp)),
				},
			},
		},
	}

	fakeClient := NewFakeClient(t, &policy, &anotherPolicy, &base, &temp,
		workloadproject.EmulateProjNs("runai-team-a", "team-a"))

	interactiveMergeStack := factory.NewInteractiveWorkloadMergeCreator(fakeClient, false)
	mergeStackMap, err := interactiveMergeStack.CreateMergeStack(context.TODO(), &temp)
	assert.NoError(t, err, "policy -> build stack")
	mergeStack := mergeStackMap[api.SpecSubResource]
	err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{})
	assert.Error(t, err, "policy -> merge")
	if err != nil {
		assert.Contains(t, err.Error(), "runAsUid.value: must be no greater than 6")
	}

	mergedTemp, err := jobs.UnmarshalWorkload(mergeStack.MergedValues(), &temp.TypeMeta, &temp.ObjectMeta)
	assert.NoError(t, err)

	assert.Equal(t, true, mergedTemp.Spec.RunAsUser.Deref(), "runAsUser -> true from policy")
	assert.Equal(t, "123M", mergedTemp.Spec.GpuMemory.Deref(), "gpu memory -> 123M from base")

	//
	//   try the webhook on this workload - should give the same error
	//
	_, err = InteractiveCreateMergeStackForFrontend(fakeClient, &temp, WebhookOptions{})
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "runAsUid.value: must be no greater than 6")
	}
}
