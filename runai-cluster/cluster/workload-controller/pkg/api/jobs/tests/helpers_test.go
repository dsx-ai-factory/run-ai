package tests

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	distw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	ssv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	tpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	twv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func GetTrainingWorkload(namespace, name, userId string) *twv1.TrainingWorkload {
	return &twv1.TrainingWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: tpv1.TrainingSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("ubuntu"),
					},
				},
				IdentityFields: field_groups.IdentityFields{
					UserId: &flds.StringField{
						Value: ptr.To(userId),
					},
					Usage: common.UsedForSubmit,
				},
			},
		},
	}
}

func GetInteractiveWorkload(namespace, name, userId string) *iwv1.InteractiveWorkload {
	return &iwv1.InteractiveWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: ipv1.InteractiveSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("ubuntu"),
					},
				},
				IdentityFields: field_groups.IdentityFields{
					UserId: &flds.StringField{
						Value: ptr.To(userId),
					},
					Usage: common.UsedForSubmit,
				},
			},
		},
	}
}

func GetInferenceWorkload(namespace, name, userId string) *infv2.InferenceWorkload {
	return &infv2.InferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: ssv1.InferenceSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("ubuntu"),
					},
				},
				IdentityFields: field_groups.IdentityFields{
					UserId: &flds.StringField{
						Value: ptr.To(userId),
					},
					Usage: common.UsedForSubmit,
				},
			},
		},
	}
}

func GetDistributedWorkload(namespace, name, userId string) *distw.DistributedWorkload {
	return &distw.DistributedWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: dp.DistributedSpec{
			WorkloadFields: field_groups.WorkloadFields{
				ContainerFields: field_groups.ContainerFields{
					Image: &flds.StringField{
						Value: ptr.To("ubuntu"),
					},
				},
				IdentityFields: field_groups.IdentityFields{
					UserId: &flds.StringField{
						Value: ptr.To(userId),
					},
					Usage: common.UsedForSubmit,
				},
			},
		},
	}
}

func GetAccessRule(name string) *v1alpha1.AccessRule {
	return &v1alpha1.AccessRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{},
		},
		Spec: v1alpha1.AccessRuleSpec{
			ScopeType:   v1alpha1.ProjectScope,
			SubjectType: v1alpha1.UserSubject,
		},
	}
}
