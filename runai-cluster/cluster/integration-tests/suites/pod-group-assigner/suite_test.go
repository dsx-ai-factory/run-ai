package pod_group_assigner

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPodGroupAssigner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pod Group Assigner Integration Tests Suite")
}
