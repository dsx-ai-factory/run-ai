package runai_rules

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

var ctx = context.Background()

var runAsUserNil = field_groups.ContainerFields{
	RunAsUser: &field_types.BoolField{
		Value: nil,
	},
}

var runAsUserFalse = field_groups.ContainerFields{
	RunAsUser: field_types.NewOptionalBoolField(ptr.To(false)),
}

var runAsUserTrue = field_groups.ContainerFields{
	RunAsUser: field_types.NewOptionalBoolField(ptr.To(true)),
}

var privilegeEscallationNil = field_groups.AccessControlFields{
	AllowPrivilegeEscalation: &field_types.BoolField{
		Value: nil,
	},
}

var privilegeEscallationFalse = field_groups.AccessControlFields{
	AllowPrivilegeEscalation: field_types.NewOptionalBoolField(ptr.To(false)),
}

var privilegeEscallationTrue = field_groups.AccessControlFields{
	AllowPrivilegeEscalation: field_types.NewOptionalBoolField(ptr.To(true)),
}

func TestRulesPatchingNoRunaiRules(t *testing.T) {
	var resultContainer field_groups.ContainerFields
	var resultAcl field_groups.AccessControlFields
	PatchPolicyWithRunaiDefaults(ctx, &runAsUserNil, &resultContainer, &privilegeEscallationNil, &resultAcl, nil)
	assert.Nil(t, resultContainer.RunAsUser.Value, "runAsUser: no runai rules -> nil")
	assert.Nil(t, resultAcl.AllowPrivilegeEscalation.Value, "privilegeEscalation: no runai rules -> nil")
}

func TestRulesPatchingFalseFalse(t *testing.T) {
	var resultContainer field_groups.ContainerFields
	var resultAcl = field_groups.AccessControlFields{
		AllowPrivilegeEscalation: &field_types.BoolField{
			Value: nil,
		},
	}
	PatchPolicyWithRunaiDefaults(ctx, &runAsUserFalse, &resultContainer, &privilegeEscallationFalse, &resultAcl, nil)
	assert.Equal(t, false, ptr.Deref(resultContainer.RunAsUser.Value, false), "runAsUser: false -> false")
	assert.Equal(t, false, ptr.Deref(resultAcl.AllowPrivilegeEscalation.Value, false), "privilegeEscalation: false -> false")
}

func TestRulesPatchingTrueTrue(t *testing.T) {
	var resultContainer field_groups.ContainerFields
	var resultAcl = field_groups.AccessControlFields{
		AllowPrivilegeEscalation: &field_types.BoolField{
			Value: nil,
		},
	}
	PatchPolicyWithRunaiDefaults(ctx, &runAsUserTrue, &resultContainer, &privilegeEscallationTrue, &resultAcl, nil)
	assert.Equal(t, true, ptr.Deref(resultContainer.RunAsUser.Value, false), "runAsUser: true -> true")
	assert.Equal(t, true, ptr.Deref(resultAcl.AllowPrivilegeEscalation.Value, false), "privilegeEscalation: true -> true")
}

func TestRulesPatchingTrueFalse(t *testing.T) {
	var resultContainer = field_groups.ContainerFields{
		RunAsUser: &field_types.BoolField{
			Value: nil,
		},
	}
	var resultAcl field_groups.AccessControlFields
	PatchPolicyWithRunaiDefaults(ctx, &runAsUserTrue, &resultContainer, &privilegeEscallationFalse, &resultAcl, nil)
	assert.Equal(t, true, ptr.Deref(resultContainer.RunAsUser.Value, false), "runAsUser: true -> true")
	assert.Equal(t, false, ptr.Deref(resultAcl.AllowPrivilegeEscalation.Value, false), "privilegeEscalation: false -> false")
}

func TestRulesPatchingFalseTrue(t *testing.T) {
	var resultContainer = field_groups.ContainerFields{
		RunAsUser: &field_types.BoolField{
			Value: nil,
		},
	}
	var resultAcl field_groups.AccessControlFields
	PatchPolicyWithRunaiDefaults(ctx, &runAsUserFalse, &resultContainer, &privilegeEscallationTrue, &resultAcl, nil)
	assert.Equal(t, false, ptr.Deref(resultContainer.RunAsUser.Value, false), "runAsUser: false -> false")
	assert.Equal(t, true, ptr.Deref(resultAcl.AllowPrivilegeEscalation.Value, false), "privilegeEscalation: true -> true")
}
