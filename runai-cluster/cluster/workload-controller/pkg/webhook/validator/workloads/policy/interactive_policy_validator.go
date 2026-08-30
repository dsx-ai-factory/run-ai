package policy

import (
	"encoding/json"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	interactivepolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type InteractiveAdapter struct {
	policy *interactivepolicy.InteractivePolicy
}

func NewInteractiveAdapter(policy *interactivepolicy.InteractivePolicy) *InteractiveAdapter {
	return &InteractiveAdapter{
		policy,
	}
}

func InitInteractivePolicyFromReq(req admission.Request) (Adapter, error) {
	policy := interactivepolicy.NewEmptyInteractivePolicy()
	err := json.Unmarshal(req.Object.Raw, &policy)
	if err != nil {
		return nil, err
	}
	return NewInteractiveAdapter(policy), nil
}

func (m *InteractiveAdapter) GetMeta() metav1.ObjectMeta {
	return m.policy.ObjectMeta
}

func (m *InteractiveAdapter) NewPolicyList() client.ObjectList {
	return &interactivepolicy.InteractivePolicyList{}
}

func (m *InteractiveAdapter) NewEmptyPolicy() interface{} {
	return interactivepolicy.NewEmptyInteractivePolicy()
}

func (m *InteractiveAdapter) GetSpec() interface{} {
	return m.policy.Spec
}

func (m *InteractiveAdapter) GetContainerFields() *field_groups.ContainerFields {
	return &m.policy.Spec.ContainerFields
}

func (m *InteractiveAdapter) GetAccessControlFields() *field_groups.AccessControlFields {
	return &m.policy.Spec.AccessControlFields
}

func (m *InteractiveAdapter) GetRunaiSchedulerFields() *field_groups.RunaiSchedulerFields {
	return &m.policy.Spec.RunaiSchedulerFields
}
