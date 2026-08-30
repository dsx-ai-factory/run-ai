package e2e_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/provisioner-service/e2e/utils"
)

var testEnv *utils.TestEnv

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Provisioner Service E2E Suite")
}

var _ = BeforeSuite(func() {
	var err error
	testEnv, err = utils.SetupTestEnv()
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	if testEnv != nil {
		testEnv.Cleanup()
	}
})
