package application_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestApplication(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Application Tests Suite")
}
