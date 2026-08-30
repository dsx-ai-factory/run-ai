package controller

import (
	ginkgoTesting "testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCrud(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller tests")
}
