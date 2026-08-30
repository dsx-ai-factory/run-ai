package common

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
)

func TestRunAIPodPhaseAndMessage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RunAIPodPhaseAndMessage Test Suite")
}

func generatePod(phase v1.PodPhase, reasonStr string, conditions *[]v1.PodCondition) *v1.Pod {
	pod := v1.Pod{
		Spec: v1.PodSpec{},
		Status: v1.PodStatus{
			Phase:   phase,
			Message: "",
			Reason:  reasonStr,
		},
	}

	if conditions != nil {
		for _, condition := range *conditions {
			pod.Status.Conditions = append(pod.Status.Conditions, condition)
		}
	}

	return &pod
}
