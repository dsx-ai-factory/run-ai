package frontend

import (
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	"k8s.io/utils/ptr"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	"gotest.tools/v3/assert"
)

type TestInfo struct {
	tag string                        // identifies the test in the log
	obj *v2alpha1.InteractiveWorkload // an object to run on
	err string                        // expected error
}

//
//    forbidden elements in workload
//

func TestForbiddenElements(t *testing.T) {

	forbiddenJupyterRule := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("forbiddenRuleJupyter"),
		Spec: ipv1.InteractiveSpec{
			InteractiveFields: field_groups.InteractiveFields{
				IsJupyter: &flds.BoolField{
					Rules: &flds.BoolRules{},
				},
			},
		},
	}

	forbiddenTensorboardRule := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("forbiddenRuleTensorboard"),
		Spec: ipv1.InteractiveSpec{
			InteractiveFields: field_groups.InteractiveFields{
				IsTensorboard: &flds.BoolField{
					Rules: &flds.BoolRules{},
				},
			},
		},
	}

	forbiddenItemRules := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("forbiddenItemRules"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					Usage: common.UsedAsTemplate,
				},
				StorageFields: field_groups.StorageFields{
					Pvcs: &flds.PvcsField{
						ItemRules: &flds.PvcsItemRules{},
					},
				},
			},
		},
	}

	forbiddenKeyRules := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("forbiddenKeyRules"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					Usage: common.UsedAsTemplate,
				},
				ContainerFields: field_groups.ContainerFields{
					Environment: &flds.DictionaryField{
						KeyRules: &flds.StringRules{},
					},
				},
			},
		},
	}

	valueAndDeletedNotOk := v2alpha1.InteractiveWorkload{
		ObjectMeta: resource.NewRunaiObject("valueAndDeletedNotOk"),
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				IdentityFields: field_groups.IdentityFields{
					Usage: common.UsedAsTemplate,
				},
				StorageFields: field_groups.StorageFields{
					Pvcs: &flds.PvcsField{
						Value: &map[string]flds.PvcsItemAndRules{
							"notOkRule": {
								Value:   &flds.PvcsItemValue{},
								Deleted: ptr.To(true),
							},
						},
					},
				},
			},
		},
	}

	tests := []TestInfo{
		{"", &valueAndDeletedNotOk, "pvcs.notOkRule: 'value' and 'deleted' are mutual exclusive"},
		{"", &forbiddenJupyterRule, "jupyter: Element 'rules' is forbidden"},
		{"", &forbiddenTensorboardRule, "tensorboard: Element 'rules' is forbidden"},
		{"", &forbiddenItemRules, "pvcs: Element 'itemRules' is forbidden"},
		{"", &forbiddenKeyRules, "environment: Element 'keyRules' is forbidden"},
	}

	for _, test := range tests {

		fakeClient := NewFakeClient(t, test.obj, workloadproject.EmulateRunaiNs())
		_, err := InteractiveCreateMergeStackForFrontend(fakeClient, test.obj, WebhookOptions{})
		assertError(t, test, err)
	}
}

func assertError(t *testing.T, testInfo TestInfo, err error) {
	tag := utils.DotUs(testInfo.obj.Name, testInfo.tag)
	if testInfo.err == "" {
		assert.NilError(t, err, tag)
	} else {
		assert.ErrorContains(t, err, testInfo.err, tag)
	}
}
