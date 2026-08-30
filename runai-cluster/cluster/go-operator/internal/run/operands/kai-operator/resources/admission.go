package resources

import (
	kaiadmission "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/admission"
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	"k8s.io/utils/ptr"
)

func convertAdmissionConfig() *kaiadmission.Admission {
	return &kaiadmission.Admission{
		Service:                            &kaicommon.Service{},
		GPUSharing:                         ptr.To(true),
		QueueLabelSelector:                 ptr.To(true),
		ValidatingWebhookConfigurationName: ptr.To("validating-engine-admission"),
		MutatingWebhookConfigurationName:   ptr.To("mutating-engine-admission"),
	}
}
