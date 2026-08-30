package policy

import (
	"encoding/json"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	trainingpolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type TrainingAdapter struct {
	policy *trainingpolicy.TrainingPolicy
}

func NewTrainingAdapter(policy *trainingpolicy.TrainingPolicy) *TrainingAdapter {
	return &TrainingAdapter{
		policy,
	}
}

func InitTrainingPolicyFromReq(req admission.Request) (Adapter, error) {
	policy := trainingpolicy.NewEmptyTrainingPolicy()
	err := json.Unmarshal(req.Object.Raw, &policy)
	if err != nil {
		return nil, err
	}
	return NewTrainingAdapter(policy), nil
}

func (m *TrainingAdapter) GetMeta() metav1.ObjectMeta {
	return m.policy.ObjectMeta
}

func (m *TrainingAdapter) NewPolicyList() client.ObjectList {
	return &trainingpolicy.TrainingPolicyList{}
}

func (m *TrainingAdapter) NewEmptyPolicy() interface{} {
	return trainingpolicy.NewEmptyTrainingPolicy()
}

func (m *TrainingAdapter) GetSpec() interface{} {
	return m.policy.Spec
}

func (m *TrainingAdapter) GetContainerFields() *field_groups.ContainerFields {
	return &m.policy.Spec.ContainerFields
}

func (m *TrainingAdapter) GetAccessControlFields() *field_groups.AccessControlFields {
	return &m.policy.Spec.AccessControlFields
}

func (m *TrainingAdapter) GetRunaiSchedulerFields() *field_groups.RunaiSchedulerFields {
	return &m.policy.Spec.RunaiSchedulerFields
}

func (m *TrainingAdapter) Kind() client.Object {
	return &trainingpolicy.TrainingPolicy{}
}
