package controller_utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cu "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("RunaiConfig", func() {
	Context("Status", func() {
		It("should add a condition", func() {
			runaiConfig := &runaiv1.RunaiConfig{}
			cu.SetCondition(runaiConfig, runaiv1.ConditionTypeDeployed, metav1.ConditionTrue, "a", "b", runaiConfig.Generation)
			Expect(len(runaiConfig.Status.Conditions)).NotTo(Equal(0))
		})

		It("should update an existing condition", func() {
			runaiConfig := &runaiv1.RunaiConfig{}
			cu.SetCondition(runaiConfig, runaiv1.ConditionTypeDeployed, metav1.ConditionTrue, "a", "b", runaiConfig.Generation)
			Expect(len(runaiConfig.Status.Conditions)).NotTo(Equal(0))

			cu.SetCondition(runaiConfig, runaiv1.ConditionTypeDeployed, metav1.ConditionFalse, "a", "b", runaiConfig.Generation)
			Expect(runaiConfig.Status.Conditions[0].Status).To(Equal(metav1.ConditionFalse))
		})
	})
})
