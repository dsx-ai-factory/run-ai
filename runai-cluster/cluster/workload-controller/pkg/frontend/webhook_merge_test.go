package frontend

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/nodepools"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"k8s.io/utils/ptr"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrimitivesInteractive(t *testing.T) {

	fakeClient := NewFakeClient(t)

	//  Following combinations are tested
	//  ---------------------------------
	//                 Temp1     Temp2    Policy3   Expected
	//    Gpu          11                             11
	//    Cpu                    22                   22
	//    CpuLimit                        33          33
	//    Image        T1-12     T2-12                T1-12
	//    Name         T1-13              T3-13       T1-13
	//    Command                T2-23    T3-23       T2-23
	//    WorkingDir   T1-123    T2-123   T3-123      T1-123
	//
	temp1, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
				Name: &flds.StringField{
					Value: ptr.To("T1-13"),
				},
			},
			ContainerFields: field_groups.ContainerFields{
				Image: &flds.StringField{
					Value: ptr.To("T1-12"),
				},
				WorkingDir: &flds.StringField{
					Value: ptr.To("T1-123"),
				},
			},
		},
		ResourceAllocationFields: field_groups.ResourceAllocationFields{
			Gpu: &flds.NumberField{
				Value: ptr.To("11"),
			},
		},
	})

	temp2, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				Image: &flds.StringField{
					Value: ptr.To("T2-12"),
				},
				Command: &flds.StringField{
					Value: ptr.To("T2-23"),
				},
				WorkingDir: &flds.StringField{
					Value: ptr.To("T2-123"),
				},
			},
		},
		ResourceAllocationFields: field_groups.ResourceAllocationFields{
			Cpu: &flds.NumberField{
				Value: ptr.To("22"),
			},
		},
	})

	policy3, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
				Name: &flds.StringField{
					Value: ptr.To("T3-13"),
				},
			},
			ContainerFields: field_groups.ContainerFields{
				Command: &flds.StringField{
					Value: ptr.To("T3-23"),
				},
				Image: &flds.StringField{
					Value: ptr.To("T1-12"),
				},
				WorkingDir: &flds.StringField{
					Value: ptr.To("T3-123"),
				},
			},
		},
		ResourceAllocationFields: field_groups.ResourceAllocationFields{
			CpuLimit: &flds.NumberField{
				Value: ptr.To("33"),
			},
		},
	})

	interactiveRules, err := utils.NewJsonQ(ipv1.GetRunaiRules(context.Background(), nodepools.NewNameLister(fakeClient)))
	assert.NoError(t, err)

	mergeStack := enforcer.NewMergeStack(temp1, "temp1", interactiveRules.JumpToTop(), true)
	mergeStack.AddLevel(temp2, enforcer.LevelPolicy, "temp2")
	mergeStack.AddLevel(policy3, enforcer.LevelPolicy, "interactive")

	err = mergeStack.MergeAndValidate(context.Background(), enforcer.MergeAndValidateOptions{
		DisableEnforce:   true,
		DisableSignining: true,
	})
	require.Nil(t, err)

	actual := mergeStack.MergedValues()

	expected, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			IdentityDeprecatedFields: field_groups.IdentityDeprecatedFields{
				Name: &flds.StringField{
					Value: ptr.To("T1-13"),
				},
			},
			ContainerFields: field_groups.ContainerFields{
				Image: &flds.StringField{
					Value: ptr.To("T1-12"),
				},
				Command: &flds.StringField{
					Value: ptr.To("T2-23"),
				},
				WorkingDir: &flds.StringField{
					Value: ptr.To("T1-123"),
				},
			},
		},
		ResourceAllocationFields: field_groups.ResourceAllocationFields{
			Gpu: &flds.NumberField{
				Value: ptr.To("11"),
			},
			Cpu: &flds.NumberField{
				Value: ptr.To("22"),
			},
			CpuLimit: &flds.NumberField{
				Value: ptr.To("33"),
			},
		},
	})

	compareJsons(t, expected, actual)
}

func TestEnvironmentMergeInteractive(t *testing.T) {

	fakeClient := NewFakeClient(t)

	//  Following combinations are tested
	//  ---------------------------------
	//                 Temp1     Temp2    Policy3   Expected
	//    env1         T1-1                           T1-1
	//    env2                   T2-2                 T2-2
	//    env3                            T3-3        T3-3
	//    env12        T1-12     T2-12                T1-12
	//    env13        T1-13              T3-13       T1-13
	//    env23                  T2-23    T3-23       T2-23
	//    env123       T1-123    T2-123   T3-123      T1-123
	//
	temp1, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				Environment: &flds.DictionaryField{
					Value: &map[string]flds.DictionaryItemAndRules{
						"env1": {
							Value: ptr.To("T1-1"),
						},
						"env12": {
							Value: ptr.To("T1-12"),
						},
						"env13": {
							Value: ptr.To("T1-13"),
						},
						"env123": {
							Value: ptr.To("T1-123"),
						},
					},
				},
			},
		},
	})

	temp2, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				Environment: &flds.DictionaryField{
					Value: &map[string]flds.DictionaryItemAndRules{
						"env2": {
							Value: ptr.To("T2-2"),
						},
						"env12": {
							Value: ptr.To("T2-12"),
						},
						"env23": {
							Value: ptr.To("T2-23"),
						},
						"env123": {
							Value: ptr.To("T2-123"),
						},
					},
				},
			},
		},
	})

	policy3, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				Environment: &flds.DictionaryField{
					Value: &map[string]flds.DictionaryItemAndRules{
						"env3": {
							Value: ptr.To("T3-3"),
						},
						"env13": {
							Value: ptr.To("T3-13"),
						},
						"env23": {
							Value: ptr.To("T3-23"),
						},
						"env123": {
							Value: ptr.To("T3-123"),
						},
					},
				},
			},
		},
	})

	interactiveRules, err := utils.NewJsonQ(ipv1.GetRunaiRules(context.Background(), nodepools.NewNameLister(fakeClient)))
	assert.NoError(t, err)

	mergeStack := enforcer.NewMergeStack(temp1, "temp1", interactiveRules.JumpToTop(), true)
	mergeStack.AddLevel(temp2, enforcer.LevelPolicy, "temp2")
	mergeStack.AddLevel(policy3, enforcer.LevelPolicy, "interactive")

	err = mergeStack.MergeAndValidate(context.Background(), enforcer.MergeAndValidateOptions{
		DisableEnforce:   true,
		DisableSignining: true,
	})
	require.Nil(t, err)

	actual := mergeStack.MergedValues()

	expected, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				Environment: &flds.DictionaryField{
					Value: &map[string]flds.DictionaryItemAndRules{
						"env1": {
							Value: ptr.To("T1-1"),
						},
						"env2": {
							Value: ptr.To("T2-2"),
						},
						"env3": {
							Value: ptr.To("T3-3"),
						},
						"env12": {
							Value: ptr.To("T1-12"),
						},
						"env13": {
							Value: ptr.To("T1-13"),
						},
						"env23": {
							Value: ptr.To("T2-23"),
						},
						"env123": {
							Value: ptr.To("T1-123"),
						},
					},
				},
			},
		},
	})

	compareJsons(t, expected, actual)
}

func TestVolumesMergeInteractive(t *testing.T) {

	fakeClient := NewFakeClient(t)

	//  Following combinations are tested
	//  ---------------------------------
	//                 Temp1     Temp2    Policy3   Expected
	//    vol1         T1-1                           T1-1
	//    vol2                   T2-2                 T2-2
	//    vol3                            T3-3        T3-3
	//    vol12        T1-12     T2-12                T1-12
	//    vol13        T1-13              T3-13       T1-13
	//    vol23                  T2-23    T3-23       T2-23
	//    vol123       T1-123    T2-123   T3-123      T1-123
	//
	temp1, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			StorageFields: field_groups.StorageFields{
				Volumes: &flds.VolumesField{
					Value: &map[string]flds.VolumeItemAndRules{
						"vol1": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T1-1"),
							},
						},
						"vol12": {
							Value: &flds.VolumeItemValue{
								TargetPath: ptr.To("T1-12"),
							},
						},
						"vol13": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T1-13"),
							},
						},
						"vol123": {
							Value: &flds.VolumeItemValue{
								TargetPath: ptr.To("T1-123"),
							},
						},
					},
				},
			},
		},
	})

	temp2, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			StorageFields: field_groups.StorageFields{
				Volumes: &flds.VolumesField{
					Value: &map[string]flds.VolumeItemAndRules{
						"vol2": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T2-2"),
							},
						},
						"vol12": {
							Value: &flds.VolumeItemValue{
								TargetPath: ptr.To("T2-12"),
							},
						},
						"vol23": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T2-23"),
							},
						},
						"vol123": {
							Value: &flds.VolumeItemValue{
								TargetPath: ptr.To("T2-123"),
							},
						},
					},
				},
			},
		},
	})

	policy3, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			StorageFields: field_groups.StorageFields{
				Volumes: &flds.VolumesField{
					Value: &map[string]flds.VolumeItemAndRules{
						"vol3": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T3-3"),
							},
						},
						"vol13": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T3-13"),
							},
						},
						"vol23": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T3-23"),
							},
						},
						"vol123": {
							Value: &flds.VolumeItemValue{
								TargetPath: ptr.To("T3-123"),
							},
						},
					},
				},
			},
		},
	})

	interactiveRules, err := utils.NewJsonQ(ipv1.GetRunaiRules(context.Background(), nodepools.NewNameLister(fakeClient)))
	assert.NoError(t, err)

	mergeStack := enforcer.NewMergeStack(temp1, "temp1", interactiveRules.JumpToTop(), true)
	mergeStack.AddLevel(temp2, enforcer.LevelPolicy, "temp2")
	mergeStack.AddLevel(policy3, enforcer.LevelPolicy, "interactive")

	err = mergeStack.MergeAndValidate(context.Background(), enforcer.MergeAndValidateOptions{
		DisableEnforce:   true,
		DisableSignining: true,
	})
	require.Nil(t, err)

	actual := mergeStack.MergedValues()

	expected, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			StorageFields: field_groups.StorageFields{
				Volumes: &flds.VolumesField{
					Value: &map[string]flds.VolumeItemAndRules{
						"vol1": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T1-1"),
							},
						},
						"vol2": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T2-2"),
							},
						},
						"vol3": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T3-3"),
							},
						},
						"vol12": {
							Value: &flds.VolumeItemValue{
								TargetPath: ptr.To("T1-12"),
							},
						},
						"vol13": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T1-13"),
							},
						},
						"vol23": {
							Value: &flds.VolumeItemValue{
								SourcePath: ptr.To("T2-23"),
							},
						},
						"vol123": {
							Value: &flds.VolumeItemValue{
								TargetPath: ptr.To("T1-123"),
							},
						},
					},
				},
			},
		},
	})

	compareJsons(t, expected, actual)
}

func TestEmptyValuesInteractive(t *testing.T) {

	fakeClient := NewFakeClient(t)

	//
	//   we expect all the values to be taken from the workload even though they "look" empty
	//   this is especially important for complex items in which workload can set only part of
	//   the fields (only SourceTarget, for example) but still all the other members will by set to
	//   default values such as "" or zeros. There is no "half full" item. item is either fully
	//   filled, with all members have values, or not filled at all.
	//
	temp, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				Image: &flds.StringField{
					Value: ptr.To(""),
				},
				RunAsUid: &flds.IntField{
					Value: ptr.To(0),
				},
				RunAsUser: &flds.BoolField{
					Value: ptr.To(false),
				},
			},
		},
		ResourceAllocationFields: field_groups.ResourceAllocationFields{
			Gpu: &flds.NumberField{
				Value: ptr.To("0"),
			},
		},
	})

	policy, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				Image: &flds.StringField{
					Value: ptr.To("aaabbb"),
				},
				RunAsUid: &flds.IntField{
					Value: ptr.To(123),
				},
				RunAsUser: &flds.BoolField{
					Value: ptr.To(true),
				},
			},
		},
		ResourceAllocationFields: field_groups.ResourceAllocationFields{
			Gpu: &flds.NumberField{
				Value: ptr.To("123"),
			},
		},
	})

	interactiveRules, err := utils.NewJsonQ(ipv1.GetRunaiRules(context.Background(), nodepools.NewNameLister(fakeClient)))
	assert.NoError(t, err)

	mergeStack := enforcer.NewMergeStack(temp, "temp1", interactiveRules.JumpToTop(), true)
	mergeStack.AddLevel(policy, enforcer.LevelPolicy, "interactive")

	err = mergeStack.MergeAndValidate(context.Background(), enforcer.MergeAndValidateOptions{
		DisableEnforce:   true,
		DisableSignining: true,
	})
	require.Nil(t, err)

	actual := mergeStack.MergedValues()

	compareJsons(t, temp, actual)
}

// make sure the workload cleaner cleans the merged workload from all the un-necessary
// stuff like rules, deleted items, etc
func TestCleanWorkloadInteractive(t *testing.T) {

	fakeClient := NewFakeClient(t)

	policy, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				RunAsUid: &flds.IntField{
					Value: ptr.To(5),
					Rules: &flds.IntRules{
						Min: ptr.To(0),
					},
				},
				Environment: &flds.DictionaryField{
					ItemRules: &flds.StringRules{
						Required: ptr.To(true),
					},
					Value: &map[string]flds.DictionaryItemAndRules{
						"HOME": {
							Value: ptr.To("myhome"),
						},
						"USER": {
							Value: ptr.To("myuser"),
						},
						"ORG": {
							Value: ptr.To("myorg"),
						},
					},
				},
			},
		},
	})

	base, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			ContainerFields: field_groups.ContainerFields{
				Environment: &flds.DictionaryField{
					Value: &map[string]flds.DictionaryItemAndRules{
						"ORG": {
							Deleted: ptr.To(true),
						},
					},
				},
			},
		},
	})

	temp, _ := utils.NewJsonQ(ipv1.InteractiveSpec{
		WorkloadFields: field_groups.WorkloadFields{
			IdentityFields: field_groups.IdentityFields{
				BaseWorkload: ptr.To("base"),
			},
			ContainerFields: field_groups.ContainerFields{
				RunAsUid: &flds.IntField{
					Value: ptr.To(0),
				},
				Environment: &flds.DictionaryField{
					Value: &map[string]flds.DictionaryItemAndRules{
						"HOME": {
							Deleted: ptr.To(true),
						},
						"ORG": {
							Value: ptr.To("neworg"),
						},
					},
				},
			},
		},
	})

	//
	//   build the merge stack and merge it. the merged workload becomes dirty as part of merging
	//
	interactiveRules, err := utils.NewJsonQ(ipv1.GetRunaiRules(context.Background(), nodepools.NewNameLister(fakeClient)))
	assert.NoError(t, err)

	mergeStack := enforcer.NewMergeStack(temp, "temp1", interactiveRules.JumpToTop(), true)
	mergeStack.AddLevel(base, enforcer.LevelWorkload, "interactive")
	mergeStack.AddLevel(policy, enforcer.LevelPolicy, "interactive")

	err = mergeStack.MergeAndValidate(context.TODO(), enforcer.MergeAndValidateOptions{
		DisableEnforce: true,
	})
	assert.NoError(t, err)

	//
	//    before cleanup we expect to find rules, value-from tag, etc. make sure they are there
	//
	dirtyTemp := mergeStack.MergedValues()

	environmentNode := "environment"
	environmentHomeNode := "environment.items.HOME"
	environmentUserNode := "environment.items.USER"
	environmentOrgNode := "environment.items.ORG"

	dirtyItemRules := dirtyTemp.JumpTo(environmentNode).From(common.ItemRules).Get()
	assert.NotNil(t, dirtyItemRules, "dirty %s %s", environmentNode, common.ItemRules)

	dirtyHomeValueFromLevel := dirtyTemp.JumpTo(environmentHomeNode).From(common.ValueFromLevel).Get()
	assert.NotNil(t, dirtyHomeValueFromLevel, "dirty %s %s", environmentHomeNode, common.ValueFromLevel)

	dirtyHomeDeletedLevel := dirtyTemp.JumpTo(environmentHomeNode).From(common.DeletedLevel).Get()
	assert.NotNil(t, dirtyHomeDeletedLevel, "dirty %s %s", environmentHomeNode, common.DeletedLevel)

	dirtyUserValueFromLevel := dirtyTemp.JumpTo(environmentUserNode).From(common.ValueFromLevel).Get()
	assert.NotNil(t, dirtyUserValueFromLevel, "dirty %s %s", environmentHomeNode, common.ValueFromLevel)

	//
	//   note that org has both deletion flag and value
	//
	dirtyOrgValueFromLevel := dirtyTemp.JumpTo(environmentOrgNode).From(common.Value).Get()
	assert.NotNil(t, dirtyOrgValueFromLevel, "dirty %s %s", environmentOrgNode, common.Value)

	dirtyOrgDeletedFromLevel := dirtyTemp.JumpTo(environmentOrgNode).From(common.Deleted).Get()
	assert.NotNil(t, dirtyOrgDeletedFromLevel, "dirty %s %s", environmentOrgNode, common.Deleted)

	//
	//   now clean everything
	//
	cleanedTemp := mergeStack.CleanTheWorkload(mergeStack.MergedValuesLevel(), enforcer.CleanOptions{})

	//
	//   and validate that everything has been deleted
	//
	assert.NotNil(t, cleanedTemp)

	cleanedItemRules := cleanedTemp.JumpTo(environmentNode).From(common.ItemRules).Get()
	assert.Nil(t, cleanedItemRules, "clean %s %s", environmentNode, common.ItemRules)

	cleanedHomeValue := cleanedTemp.JumpTo(environmentNode).From("HOME").Get()
	assert.Nil(t, cleanedHomeValue, "clean %s %s", environmentHomeNode, "value totally deleted")

	cleanedUserValueFromLevel := cleanedTemp.JumpTo(environmentUserNode).From(common.ValueFromLevel).Get()
	assert.Nil(t, cleanedUserValueFromLevel, "cleaned %s %s", environmentUserNode, common.ValueFromLevel)

	//
	//   since org has been deleted in base but restored in workload - value should remain
	//
	cleanedOrgValueFromLevel := cleanedTemp.JumpTo(environmentOrgNode).From(common.Value).Get()
	assert.NotNil(t, cleanedOrgValueFromLevel, "cleaned %s %s - value remain", environmentOrgNode, common.Value)

	cleanedOrgDeletedFromLevel := cleanedTemp.JumpTo(environmentOrgNode).From(common.DeletedLevel).Get()
	assert.Nil(t, cleanedOrgDeletedFromLevel, "cleaned %s %s - deletion gone", environmentOrgNode, common.DeletedLevel)
}
