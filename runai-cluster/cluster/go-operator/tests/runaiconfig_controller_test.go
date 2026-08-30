package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ccontroller "github.com/run-ai/runai/runai-cluster/cluster/go-operator/controllers/run"
	controller "github.com/run-ai/runai/runai-cluster/cluster/go-operator/controllers/run"
	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/run-ai/runai/runai-cluster/cluster/go-operator/tests/utils"
)

var _ = Describe("RunaiConfig Controller", func() {
	Context("Resources", func() {
		It("should have resources to control", func() {
			hybridClient := kc.New(k8s, k8s, nil)
			managedOperands := ccontroller.NewManagedOperands(apiCtx, hybridClient, hybridClient, runaiConfig)
			resources, err := controller.ResourcesForOperands(apiCtx,
				managedOperands.GetRemoteClusterEnabled())
			ExpectNoErr(err)
			Expect(resources.CreateOrUpdate).NotTo(BeEmpty())
		})
	})
})
