package describe_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEnvironmentDescribe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Environment Describe Suite")
}
