package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mps_server "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/mps-server"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/mps-server/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
)

var _ = Describe("RunAI MPS Server Operand", func() {
	BeforeEach(func() {
		runaiConfig.Spec.MPSServer.Enabled = true
		EventuallyUpdateResource(kc.New(k8s, k8s, nil), runaiConfig)
		PollForTrueCondition(k8s, runaiv1.ConditionTypeDeployed, runaiConfig)
	})

	Context("Resources", func() {
		It("should have resources", func() {
			resources, err := mps_server.Operand().ResourcesForRunaiConfig(apiCtx,
				kc.New(k8s, k8s, nil), runaiConfig)
			ExpectNoErr(err)
			Expect(resources).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			resources, err := mps_server.Operand().ResourcesForRunaiConfig(apiCtx,
				kc.New(k8s, k8s, nil), runaiConfig)
			ExpectNoErr(err)

			for _, resource := range resources {
				ExpectGetResource(k8s, resource.Object)
			}
		})

		It("should re-create deleted resources", func() {
			resources, err := mps_server.Operand().ResourcesForRunaiConfig(apiCtx,
				kc.New(k8s, k8s, nil), runaiConfig)
			ExpectNoErr(err)

			for _, resource := range resources {
				DeleteAndPollUntilDeleted(k8s, resource.Object)
			}

			EventuallyWithTimeout(func() bool {
				reCreated := true
				for _, resource := range resources {
					if !ResourceExists(k8s, resource.Object) {
						reCreated = false
					}
				}

				return reCreated
			}).Should(BeTrue())
		})

		Context("DaemonSet", func() {
			It("should reconcile a modified daemonset", func() {
				daemonSet, err := resources.DaemonSetForRunaiConfig(apiCtx,
					kc.New(k8s, k8s, nil), runaiConfig)
				ExpectNoErr(err)

				changedScheduler := "run.ai/test"

				daemonSet.(*appsv1.DaemonSet).Spec.
					Template.Spec.SchedulerName = changedScheduler

				EventuallyUpdateResource(k8s, daemonSet)

				EventuallyWithTimeout(func() bool {
					ExpectGetResource(k8s, daemonSet)
					return daemonSet.(*appsv1.DaemonSet).Spec.
						Template.Spec.SchedulerName == changedScheduler
				}).Should(BeFalse())
			})
		})
	})
})
