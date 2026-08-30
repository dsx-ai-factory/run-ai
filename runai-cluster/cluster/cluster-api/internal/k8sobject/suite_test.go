package k8sobject

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestK8sObject(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8sObject Suite")
}
