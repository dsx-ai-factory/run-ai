package project_controller

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProjectController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Project Controller Integration Tests Suite")
}
