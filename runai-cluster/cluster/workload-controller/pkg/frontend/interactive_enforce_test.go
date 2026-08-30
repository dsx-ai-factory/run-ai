package frontend

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/nodepools"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	"k8s.io/utils/ptr"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type EnforceTestInfo struct {
	policy       *ipv1.InteractivePolicy
	workload     *v2alpha1.InteractiveWorkload
	baseWorkload *v2alpha1.InteractiveWorkload
	err          string // expected error
}

func TestPrimitiveRules(t *testing.T) {

	policy := ipv1.InteractivePolicy{
		ObjectMeta: resource.NewRunaiObject("interactive"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Rules: &flds.StringRules{
							Required: ptr.To(true),
						},
					},
				},

				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("ubuntu"),
						Rules: &flds.StringRules{
							CanEdit: ptr.To(false),
						},
					},
					WorkingDir: &flds.StringField{
						Rules: &flds.StringRules{
							Options: &[]flds.OptionRule{
								{
									Value: ptr.To("a"),
								},
								{
									Value:    ptr.To("b"),
									Disabled: ptr.To(true),
								},
								{
									Value: ptr.To("c"),
								},
							},
						},
					},
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				Gpu: &flds.NumberField{
					Rules: &flds.NumberRules{
						Min:  ptr.To("3.5"),
						Max:  ptr.To("3.8"),
						Step: ptr.To("0.2"),
					},
				},
			},
		},
	}

	cantEditImage := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("cantEditImage"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("changed_image"),
					},
				},
			},
		},
	}

	requiredName := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("requiredName"),
		Spec:       ipv1.InteractiveSpec{},
	}

	optionsAok := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("optionsAok"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
				ContainerFields: field_groups.ContainerFields{
					WorkingDir: &flds.StringField{
						Value: ptr.To("a"),
					},
				},
			},
		},
	}

	optionsBdisabled := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("optionsBdisabled"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
				ContainerFields: field_groups.ContainerFields{
					WorkingDir: &flds.StringField{
						Value: ptr.To("b"),
					},
				},
			},
		},
	}

	optionsXnotOk := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("optionsXnotOk"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
				ContainerFields: field_groups.ContainerFields{
					WorkingDir: &flds.StringField{
						Value: ptr.To("x"),
					},
				},
			},
		},
	}

	gpuTooSmall := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("gpTooSmall"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				Gpu: &flds.NumberField{
					Value: ptr.To("3.3"),
				},
			},
		},
	}

	gpuTooHigh := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("gpuTooHigh"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				Gpu: &flds.NumberField{
					Value: ptr.To("3.9"),
				},
			},
		},
	}

	gpuStepNotValid := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("gpuStepNotValid"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				Gpu: &flds.NumberField{
					Value: ptr.To("3.6"),
				},
			},
		},
	}

	gpuStepValid := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("gpuStepValid"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				Gpu: &flds.NumberField{
					Value: ptr.To("3.7"),
				},
			},
		},
	}

	tests := []EnforceTestInfo{
		{&policy, &requiredName, nil, "missing value for required: name.value"},
		{&policy, &optionsAok, nil, ""},
		{&policy, &cantEditImage, nil, "the administrator prohibited overriding the value of image.value"},
		{&policy, &optionsBdisabled, nil, "value must be one of: a, c"},
		{&policy, &optionsXnotOk, nil, "value must be one of: a, c"},
		{&policy, &gpuTooSmall, nil, "gpu.value: must be no less than 3.5"},
		{&policy, &gpuTooHigh, nil, "gpu.value: must be no greater than 3.8"},
		{&policy, &gpuStepValid, nil, ""},
		{&policy, &gpuStepNotValid, nil, "value must follow 0.2 steps from 3.5 (3.50, 3.70, etc.)"},
	}

	for _, test := range tests {
		interactiveEnforceTest(t, test)
	}
}

func TestMapRules(t *testing.T) {

	policyRequiredEnvVars := ipv1.InteractivePolicy{
		ObjectMeta: resource.NewRunaiObject("interactive"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Environment: &flds.DictionaryField{
						Rules: &flds.MapRules{
							Required: ptr.To(true),
						},
					},
				},
			},
		},
	}

	totallyEmpty := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("emptyEnvVars"),
		Spec:       ipv1.InteractiveSpec{},
	}

	emptyEnvVars := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("missingEnvVars"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Environment: &flds.DictionaryField{
						Value: &map[string]flds.DictionaryItemAndRules{},
					},
				},
			},
		},
	}

	policyRequiredVolumes := ipv1.InteractivePolicy{
		ObjectMeta: resource.NewRunaiObject("interactive"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Rules: &flds.MapRules{
							Required: ptr.To(true),
						},
					},
				},
			},
		},
	}

	emptyVolumes := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("emptyVolumes"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Value: &map[string]flds.VolumeItemAndRules{},
					},
				},
			},
		},
	}

	policyCantAddVolumes := ipv1.InteractivePolicy{
		ObjectMeta: resource.NewRunaiObject("interactive"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Rules: &flds.MapRules{
							Required: ptr.To(true),
							CanAdd:   ptr.To(false),
						},
						Value: &map[string]flds.VolumeItemAndRules{
							"adminVolume": {
								Value: &flds.VolumeItemValue{
									TargetPath: ptr.To("/admin/target"),
									SourcePath: ptr.To("/admin/source"),
								},
							},
						},
					},
				},
			},
		},
	}

	addedVolume := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("addedVolume"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Value: &map[string]flds.VolumeItemAndRules{
							"userVolume": {
								Value: &flds.VolumeItemValue{
									TargetPath: ptr.To("/user/target"),
									SourcePath: ptr.To("/user/source"),
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []EnforceTestInfo{
		{&policyRequiredVolumes, &emptyVolumes, nil, "value is not provided required: volumes.value"},
		{&policyRequiredEnvVars, &totallyEmpty, nil, "missing value for required: environment.value"},
		{&policyRequiredEnvVars, &emptyEnvVars, nil, "value is not provided required: environment.value"},
		{&policyRequiredVolumes, &totallyEmpty, nil, "missing value for required: volumes.value"},
		{&policyCantAddVolumes, &addedVolume, nil, "userVolume: the administrator prohibited adding to volumes"},
	}

	for _, test := range tests {
		interactiveEnforceTest(t, test)
	}
}

func TestEditRemoveEnvVars(t *testing.T) {

	policyEnvVars := ipv1.InteractivePolicy{
		ObjectMeta: resource.NewRunaiObject("interactive"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Environment: &flds.DictionaryField{
						Value: &map[string]flds.DictionaryItemAndRules{
							"public": {
								Value: ptr.To("public"),
							},
							"cantRemove": {
								Value: ptr.To("cantRemove"),
								Rules: &flds.MapItemRules{
									CanRemove: ptr.To(false),
								},
							},
							"cantEdit": {
								Value: ptr.To("cantEdit"),
								Rules: &flds.MapItemRules{
									CanEdit: ptr.To(false),
								},
							},
						},
					},
				},
			},
		},
	}

	userRemoveOk := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("userRemovedOk"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
					Environment: &flds.DictionaryField{
						Value: &map[string]flds.DictionaryItemAndRules{
							"public": {
								Deleted: ptr.To(true),
							},
							"cantEdit": {
								Deleted: ptr.To(true),
							},
							"userAdded": {
								Value: ptr.To("userAdded"),
							},
							"userDeleted": {
								Deleted: ptr.To(true),
							},
						},
					},
				},
			},
		},
	}

	userCantRemove := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("userCantRemove"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Environment: &flds.DictionaryField{
						Value: &map[string]flds.DictionaryItemAndRules{
							"cantRemove": {
								Deleted: ptr.To(true),
							},
						},
					},
				},
			},
		},
	}

	userCantEdit := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("userCantEdit"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
					Environment: &flds.DictionaryField{
						Value: &map[string]flds.DictionaryItemAndRules{
							"cantEdit": {
								Value: ptr.To("valueChanged"),
							},
						},
					},
				},
			},
		},
	}

	tests := []EnforceTestInfo{
		{&policyEnvVars, &userCantRemove, nil, "the administrator mandated that environment.cantRemove cannot be removed"},
		{&policyEnvVars, &userCantEdit, nil, "the administrator prohibited overriding the value of environment.cantEdit"},
		{&policyEnvVars, &userRemoveOk, nil, ""},
	}

	for _, test := range tests {
		interactiveEnforceTest(t, test)
	}
}

func TestEditRemoveVolumes(t *testing.T) {

	policyVolumes := ipv1.InteractivePolicy{
		ObjectMeta: resource.NewRunaiObject("interactive"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Value: &map[string]flds.VolumeItemAndRules{
							"public": {
								Value: &flds.VolumeItemValue{
									SourcePath: ptr.To("/public/source"),
									TargetPath: ptr.To("/public/target"),
								},
							},
							"cantRemove": {
								Value: &flds.VolumeItemValue{
									SourcePath: ptr.To("/cant/remove/source"),
									TargetPath: ptr.To("/cant/remove/target"),
								},
								Rules: &flds.MapItemRules{
									CanRemove: ptr.To(false),
								},
							},
							"cantEdit": {
								Value: &flds.VolumeItemValue{
									SourcePath: ptr.To("/cant/edit/source"),
									TargetPath: ptr.To("/cant/edit/target"),
								},
								Rules: &flds.MapItemRules{
									CanEdit: ptr.To(false),
								},
							},
						},
					},
				},
			},
		},
	}

	userRemoveOk := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("userRemovedOk"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Value: &map[string]flds.VolumeItemAndRules{
							"public": {
								Deleted: ptr.To(true),
							},
							"cantEdit": {
								Deleted: ptr.To(true),
							},
							"userDeleted": {
								Deleted: ptr.To(true),
							},
							"userAdded": {
								Value: &flds.VolumeItemValue{
									SourcePath: ptr.To("/user/source"),
									TargetPath: ptr.To("/user/target"),
								},
							},
						},
					},
				},
			},
		},
	}

	userCantRemove := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("userCantRemove"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Value: &map[string]flds.VolumeItemAndRules{
							"cantRemove": {
								Deleted: ptr.To(true),
							},
						},
					},
				},
			},
		},
	}

	userCantEdit := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("userCantEdit"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Value: &map[string]flds.VolumeItemAndRules{
							"cantEdit": {
								Value: &flds.VolumeItemValue{
									TargetPath: ptr.To("/changed/value"),
								},
							},
						},
					},
				},
			},
		},
	}

	userCanEdit := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("userCanEdit"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Value: &map[string]flds.VolumeItemAndRules{
							"public": {
								Value: &flds.VolumeItemValue{
									SourcePath: ptr.To("/changed/value"),
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []EnforceTestInfo{
		{&policyVolumes, &userRemoveOk, nil, ""},
		{&policyVolumes, &userCantRemove, nil, "the administrator mandated that volumes.cantRemove cannot be removed"},
		{&policyVolumes, &userCanEdit, nil, ""},
		{&policyVolumes, &userCantEdit, nil, "the administrator prohibited overriding the value of volumes.cantEdit"},
	}

	for _, test := range tests {
		interactiveEnforceTest(t, test)
	}
}

// in this special case we test policy + two templates, one in which a value has been
// deleted and the other in which it was added. The enforcement code needs to figure
// out if it has been deleted first (=in base workload) then added (=in our workload)
// or is it the other way around
func TestDeletedPlusAdded(t *testing.T) {

	policyVolumes := ipv1.InteractivePolicy{
		ObjectMeta: resource.NewRunaiObject("interactive"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Value: &map[string]flds.VolumeItemAndRules{
							"myvolume": {
								Value: &flds.VolumeItemValue{
									SourcePath: ptr.To("/cant/remove/source"),
									TargetPath: ptr.To("/cant/remove/target"),
								},
								Rules: &flds.MapItemRules{
									CanRemove: ptr.To(false),
								},
							},
						},
					},
				},
			},
		},
	}

	volumeDeleted := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("volumeDeleted"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Value: &map[string]flds.VolumeItemAndRules{
							"myvolume": {
								Deleted: ptr.To(true),
							},
						},
					},
				},
			},
		},
	}

	volumeAdded := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("volumeAdded"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Value: &map[string]flds.VolumeItemAndRules{
							"myvolume": {
								Value: &flds.VolumeItemValue{
									SourcePath: ptr.To("/added"),
									TargetPath: ptr.To("/added"),
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []EnforceTestInfo{
		{&policyVolumes, &volumeAdded, &volumeDeleted, ""},
		{&policyVolumes, &volumeDeleted, &volumeAdded, "the administrator mandated that volumes.myvolume cannot be removed"},
	}

	for _, test := range tests {
		interactiveEnforceTest(t, test)
	}
}

func TestItemRules(t *testing.T) {

	policyPvcs := ipv1.InteractivePolicy{
		ObjectMeta: resource.NewRunaiObject("interactive"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Pvcs: &flds.PvcsField{
						ItemRules: &flds.PvcsItemRules{
							Size: &flds.StringRules{
								Options: &[]flds.OptionRule{
									{Name: ptr.To("big"), Value: ptr.To("10G")},
									{Name: ptr.To("small"), Value: ptr.To("500M")},
								},
							},
						},
						Value: &map[string]flds.PvcsItemAndRules{
							"default": {
								Value: &flds.PvcsItemValue{
									Size: ptr.To("500M"),
								},
							},
						},
					},
				},
			},
		},
	}

	validPvcSize := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("validPvcSize"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Pvcs: &flds.PvcsField{
						Value: &map[string]flds.PvcsItemAndRules{
							"pvc1": {
								Value: &flds.PvcsItemValue{
									Size: ptr.To("10G"),
								},
							},
							"pvc2": {
								Value: &flds.PvcsItemValue{
									Size: ptr.To("500M"),
								},
							},
						},
					},
				},
			},
		},
	}

	wrongPvcSize := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("wrongPvcSize"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Pvcs: &flds.PvcsField{
						Value: &map[string]flds.PvcsItemAndRules{
							"pvc3": {
								Value: &flds.PvcsItemValue{
									Size: ptr.To("200M"),
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []EnforceTestInfo{
		{&policyPvcs, &validPvcSize, nil, ""},
		{&policyPvcs, &wrongPvcSize, nil, "value must be one of: 10G, 500M"},
	}

	for _, test := range tests {
		interactiveEnforceTest(t, test)
	}
}

func interactiveEnforceTest(t *testing.T, info EnforceTestInfo) {

	fakeClient := NewFakeClient(t, info.policy)

	//
	//    load the workload and policy to json
	//
	workloadQ, err := utils.NewJsonQ(info.workload)
	require.Nil(t, err, "workload-to-json")
	policyQ, err := utils.NewJsonQ(info.policy)
	require.Nil(t, err, "policy-to-json")

	//
	//    create a merge stach containing the policy and workload + optional base workload
	//
	interactiveRules, err := utils.NewJsonQ(ipv1.GetRunaiRules(context.Background(), nodepools.NewNameLister(fakeClient)))
	assert.NoError(t, err)

	mergeStack := enforcer.NewMergeStack(workloadQ.JumpTo(constants.Spec), info.workload.Name, interactiveRules.JumpToTop(), true)
	if info.baseWorkload != nil {
		baseTemplateQ, err := utils.NewJsonQ(info.baseWorkload)
		require.Nil(t, err, "base-workload-to-json")
		mergeStack.AddLevel(baseTemplateQ.JumpTo(constants.Spec), enforcer.LevelWorkload, info.baseWorkload.Name)
	}
	mergeStack.AddLevel(policyQ.JumpTo(constants.Spec), enforcer.LevelPolicy, info.policy.Name)

	//
	//    process the stack and check the returned error against our expectation
	//
	actualErr := mergeStack.MergeAndValidate(context.Background(), enforcer.MergeAndValidateOptions{})

	if info.err == "" {
		assert.NoError(t, actualErr, info.workload.Name)
	} else {
		msg := fmt.Sprintf("%v", actualErr)
		assert.Contains(t, msg, info.err, info.workload.Name)
	}
}
