package helm_deployer

import (
	ginkgoTesting "testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHelmDeployer(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helm Deployer Suite")
}
