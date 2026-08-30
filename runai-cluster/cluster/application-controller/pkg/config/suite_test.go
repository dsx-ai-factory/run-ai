package config

import (
	ginkgoTesting "testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}
