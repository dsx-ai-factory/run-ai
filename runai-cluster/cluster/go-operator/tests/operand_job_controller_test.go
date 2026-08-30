package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	job_controller "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/job-controller"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/job-controller/resources"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Job Controller Operand", func() {
	Context("Resources", func() {
		It("should have resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := job_controller.Operand(hybridClient, runaiConfig)
			Expect(op.Resources()).NotTo(BeEmpty())
		})

		It("should create desired resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := job_controller.Operand(hybridClient, runaiConfig)

			for _, res := range op.Resources() {
				obj, err := res.Object(apiCtx)
				ExpectNoErr(err)
				if res.Enabled() {
					ExpectGetResource(k8s, obj)
				}
			}
		})

		It("should re-create deleted resources", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			op := job_controller.Operand(hybridClient, runaiConfig)

			for _, res := range op.Resources() {
				if res.Enabled() {
					obj, err := res.Object(apiCtx)
					ExpectNoErr(err)
					DeleteAndPollUntilDeleted(k8s, obj)
				}
			}

			EventuallyWithTimeout(func() bool {
				reCreated := true
				for _, res := range op.Resources() {
					if res.Enabled() {
						obj, err := res.Object(apiCtx)
						ExpectNoErr(err)
						if !ResourceExists(k8s, obj) {
							reCreated = false
						}
					}
				}
				return reCreated
			}).Should(BeTrue())
		})

		Context("Deployment", func() {
			It("should reconcile a modified deployment", func() {
				hybridClient := kc.New(k8s, k8s, nil)
				depResource := resources.NewDeploymentResource("job-controller", hybridClient, runaiConfig)
				deployment, err := depResource.Object(apiCtx)
				ExpectNoErr(err)

				changedScheduler := "run.ai/test"

				deployment.(*appsv1.Deployment).Spec.
					Template.Spec.SchedulerName = changedScheduler

				EventuallyUpdateResource(k8s, deployment)

				EventuallyWithTimeout(func() bool {
					ExpectGetResource(k8s, deployment)
					return deployment.(*appsv1.Deployment).Spec.
						Template.Spec.SchedulerName == changedScheduler
				}).Should(BeFalse())
			})
		})

		Context("Service", func() {
			It("should reconcile a modified service", func() {
				hybridClient := kc.New(k8s, k8s, nil)
				svcResource := resources.NewServiceResource("job-controller", hybridClient, runaiConfig)
				svc, err := svcResource.Object(apiCtx)
				ExpectNoErr(err)

				svc.(*v1.Service).Spec.Selector = map[string]string{}

				EventuallyUpdateResource(k8s, svc)

				EventuallyWithTimeout(func() bool {
					ExpectGetResource(k8s, svc)
					return len(svc.(*v1.Service).Spec.Selector) != 0
				}).Should(BeTrue())
			})
		})
	})
})
