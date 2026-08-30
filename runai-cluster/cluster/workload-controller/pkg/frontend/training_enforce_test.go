package frontend

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/nodepools"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	"k8s.io/utils/ptr"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TrainingEnforceTestInfo struct {
	policy       *tpv1.TrainingPolicy
	workload     *v2alpha1.TrainingWorkload
	baseWorkload *v2alpha1.TrainingWorkload
	err          string // expected error
}

func TestPrimitiveRulesTraining(t *testing.T) {

	policy := tpv1.TrainingPolicy{
		ObjectMeta: resource.NewRunaiObject("training"),
		Spec: tpv1.TrainingSpec{
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
				Memory: &flds.QuantityField{
					Rules: &flds.QuantityRules{
						Min: ptr.To("10M"),
						Max: ptr.To("1G"),
					},
				},
			},
		},
	}

	cantEditImage := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("cantEditImage"),
		Spec: tpv1.TrainingSpec{
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

	cantEditImageSameValue := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("cantEditImage"),
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("ubuntu"),
					},
				},
			},
		},
	}

	requiredName := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("requiredName"),
		Spec:       tpv1.TrainingSpec{},
	}

	optionsAok := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("optionsAok"),
		Spec: tpv1.TrainingSpec{
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

	optionsBdisabled := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("optionsBdisabled"),
		Spec: tpv1.TrainingSpec{
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

	optionsXnotOk := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("optionsXnotOk"),
		Spec: tpv1.TrainingSpec{
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

	gpuTooSmall := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("gpTooSmall"),
		Spec: tpv1.TrainingSpec{
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

	gpuTooHigh := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("gpuTooHigh"),
		Spec: tpv1.TrainingSpec{
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

	gpuStepNotValid := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("gpuStepNotValid"),
		Spec: tpv1.TrainingSpec{
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

	gpuStepValid := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("gpuStepValid"),
		Spec: tpv1.TrainingSpec{
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

	memoryTooSmall := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("memoryTooSmall"),
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				Memory: &flds.QuantityField{
					Value: ptr.To("9M"),
				},
			},
		},
	}

	memoryTooHighMb := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("memoryToo/High"),
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				Memory: &flds.QuantityField{
					Value: ptr.To("1025M"),
				},
			},
		},
	}

	memoryTooHighGb := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("memoryToo/High"),
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
					Name: &flds.StringField{
						Value: ptr.To("name"),
					},
				},
			},
			ResourceAllocationFields: field_groups.ResourceAllocationFields{
				Memory: &flds.QuantityField{
					Value: ptr.To("2G"),
				},
			},
		},
	}

	tests := []TrainingEnforceTestInfo{
		{&policy, &requiredName, nil, "missing value for required: name.value"},
		{&policy, &optionsAok, nil, ""},
		{&policy, &cantEditImage, nil, "the administrator prohibited overriding the value of image.value"},
		{&policy, &cantEditImageSameValue, nil, ""},
		{&policy, &optionsBdisabled, nil, "value must be one of: a, c"},
		{&policy, &optionsXnotOk, nil, "value must be one of: a, c"},
		{&policy, &gpuTooSmall, nil, "gpu.value: must be no less than 3.5"},
		{&policy, &gpuTooHigh, nil, "gpu.value: must be no greater than 3.8"},
		{&policy, &gpuStepValid, nil, ""},
		{&policy, &gpuStepNotValid, nil, "value must follow 0.2 steps from 3.5 (3.50, 3.70, etc.)"},
		{&policy, &memoryTooSmall, nil, "memory.value: must be no less than 10M"},
		{&policy, &memoryTooHighMb, nil, "memory.value: must be no greater than 1G"},
		{&policy, &memoryTooHighGb, nil, "memory.value: must be no greater than 1G"},
	}

	for _, test := range tests {
		enforceTest(t, test)
	}
}

func TestMapRulesTraining(t *testing.T) {

	policyRequiredEnvVars := tpv1.TrainingPolicy{
		ObjectMeta: resource.NewRunaiObject("training"),
		Spec: tpv1.TrainingSpec{
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

	totallyEmpty := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("emptyEnvVars"),
		Spec:       tpv1.TrainingSpec{},
	}

	emptyEnvVars := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("missingEnvVars"),
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Environment: &flds.DictionaryField{
						Value: &map[string]flds.DictionaryItemAndRules{},
					},
				},
			},
		},
	}

	policyRequiredVolumes := tpv1.TrainingPolicy{
		ObjectMeta: resource.NewRunaiObject("training"),
		Spec: tpv1.TrainingSpec{
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

	emptyVolumes := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("emptyVolumes"),
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				StorageFields: field_groups.StorageFields{
					Volumes: &flds.VolumesField{
						Value: &map[string]flds.VolumeItemAndRules{},
					},
				},
			},
		},
	}

	policyCantAddVolumes := tpv1.TrainingPolicy{
		ObjectMeta: resource.NewRunaiObject("training"),
		Spec: tpv1.TrainingSpec{
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

	addedVolume := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("addedVolume"),
		Spec: tpv1.TrainingSpec{
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

	tests := []TrainingEnforceTestInfo{
		{&policyRequiredVolumes, &emptyVolumes, nil, "value is not provided required: volumes.value"},
		{&policyRequiredEnvVars, &totallyEmpty, nil, "missing value for required: environment.value"},
		{&policyRequiredEnvVars, &emptyEnvVars, nil, "value is not provided required: environment.value"},
		{&policyRequiredVolumes, &totallyEmpty, nil, "missing value for required: volumes.value"},
		{&policyCantAddVolumes, &addedVolume, nil, "userVolume: the administrator prohibited adding to volumes"},
	}

	for _, test := range tests {
		enforceTest(t, test)
	}
}

func TestEditRemoveEnvVarsTraining(t *testing.T) {

	policyEnvVars := tpv1.TrainingPolicy{
		ObjectMeta: resource.NewRunaiObject("training"),
		Spec: tpv1.TrainingSpec{
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

	userRemoveOk := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("userRemovedOk"),
		Spec: tpv1.TrainingSpec{
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

	userCantRemove := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("userCantRemove"),
		Spec: tpv1.TrainingSpec{
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

	userCantEdit := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("userCantEdit"),
		Spec: tpv1.TrainingSpec{
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

	tests := []TrainingEnforceTestInfo{
		{&policyEnvVars, &userCantRemove, nil, "the administrator mandated that environment.cantRemove cannot be removed"},
		{&policyEnvVars, &userCantEdit, nil, "the administrator prohibited overriding the value of environment.cantEdit"},
		{&policyEnvVars, &userRemoveOk, nil, ""},
	}

	for _, test := range tests {
		enforceTest(t, test)
	}
}

func TestEditRemoveVolumesTraining(t *testing.T) {

	policyVolumes := tpv1.TrainingPolicy{
		ObjectMeta: resource.NewRunaiObject("training"),
		Spec: tpv1.TrainingSpec{
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

	userRemoveOk := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("userRemovedOk"),
		Spec: tpv1.TrainingSpec{
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

	userCantRemove := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("userCantRemove"),
		Spec: tpv1.TrainingSpec{
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

	userCantEdit := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("userCantEdit"),
		Spec: tpv1.TrainingSpec{
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

	userCanEdit := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("userCanEdit"),
		Spec: tpv1.TrainingSpec{
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

	tests := []TrainingEnforceTestInfo{
		{&policyVolumes, &userRemoveOk, nil, ""},
		{&policyVolumes, &userCantRemove, nil, "the administrator mandated that volumes.cantRemove cannot be removed"},
		{&policyVolumes, &userCanEdit, nil, ""},
		{&policyVolumes, &userCantEdit, nil, "the administrator prohibited overriding the value of volumes.cantEdit"},
	}

	for _, test := range tests {
		enforceTest(t, test)
	}
}

// in this special case we test policy + two templates, one in which a value has been
// deleted and the other in which it was added. The enforcement code needs to figure
// out if it has been deleted first (=in base workload) then added (=in our workload)
// or is it the other way around
func TestDeletedPlusAddedTraining(t *testing.T) {

	policyVolumes := tpv1.TrainingPolicy{
		ObjectMeta: resource.NewRunaiObject("training"),
		Spec: tpv1.TrainingSpec{
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

	volumeDeleted := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("volumeDeleted"),
		Spec: tpv1.TrainingSpec{
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

	volumeAdded := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("volumeAdded"),
		Spec: tpv1.TrainingSpec{
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

	tests := []TrainingEnforceTestInfo{
		{&policyVolumes, &volumeAdded, &volumeDeleted, ""},
		{&policyVolumes, &volumeDeleted, &volumeAdded, "the administrator mandated that volumes.myvolume cannot be removed"},
	}

	for _, test := range tests {
		enforceTest(t, test)
	}
}

func TestItemRulesTraining(t *testing.T) {

	policyPvcs := tpv1.TrainingPolicy{
		ObjectMeta: resource.NewRunaiObject("training"),
		Spec: tpv1.TrainingSpec{
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

	validPvcSize := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("validPvcSize"),
		Spec: tpv1.TrainingSpec{
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

	wrongPvcSize := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("wrongPvcSize"),
		Spec: tpv1.TrainingSpec{
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

	tests := []TrainingEnforceTestInfo{
		{&policyPvcs, &validPvcSize, nil, ""},
		{&policyPvcs, &wrongPvcSize, nil, "value must be one of: 10G, 500M"},
	}

	for _, test := range tests {
		enforceTest(t, test)
	}
}

func TestItemRulesCantEditTraining(t *testing.T) {

	policyPvcs := tpv1.TrainingPolicy{
		ObjectMeta: resource.NewRunaiObject("training"),
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				StorageFields: field_groups.StorageFields{
					Pvcs: &flds.PvcsField{
						ItemRules: &flds.PvcsItemRules{
							ExistingPvc: &flds.BoolRules{
								CanEdit: ptr.To(false),
							},
						},
						Value: &map[string]flds.PvcsItemAndRules{
							"DEFAULTS": {
								Value: &flds.PvcsItemValue{
									ExistingPvc: ptr.To(true),
								},
							},
						},
					},
				},
			},
		},
	}

	validPvc := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("validPvc"),
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Pvcs: &flds.PvcsField{
						Value: &map[string]flds.PvcsItemAndRules{
							"pvc1": {
								Value: &flds.PvcsItemValue{
									ClaimName: ptr.To("existing"),
								},
							},
							"pvc2": {
								Value: &flds.PvcsItemValue{
									ExistingPvc: ptr.To(true),
									ClaimName:   ptr.To("existing"),
								},
							},
						},
					},
				},
			},
		},
	}

	invalidPvc := v2alpha1.TrainingWorkload{
		ObjectMeta: resource.NewRunaiObject("validPvc"),
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{Value: ptr.To("ubuntu")},
				},
				StorageFields: field_groups.StorageFields{
					Pvcs: &flds.PvcsField{
						Value: &map[string]flds.PvcsItemAndRules{
							"pvc1": {
								Value: &flds.PvcsItemValue{
									ExistingPvc:  ptr.To(false),
									StorageClass: ptr.To("default"),
									Size:         ptr.To("10G"),
									Path:         ptr.To("/data"),
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []TrainingEnforceTestInfo{
		{&policyPvcs, &validPvc, nil, ""},
		{&policyPvcs, &invalidPvc, nil, "the administrator prohibited overriding the value of pvcs.pvc1.existingPvc"},
	}

	for _, test := range tests {
		enforceTest(t, test)
	}
}

func enforceTest(t *testing.T, info TrainingEnforceTestInfo) {

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
	trainingRules, err := utils.NewJsonQ(tpv1.GetRunaiRules(context.Background(), nodepools.NewNameLister(fakeClient)))
	assert.NoError(t, err)

	mergeStack := enforcer.NewMergeStack(workloadQ.JumpTo(constants.Spec), info.workload.Name, trainingRules.JumpToTop(), true)
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
