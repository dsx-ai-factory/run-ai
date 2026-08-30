package integration_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestComposerIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "composer_v2 integration tests suite")
}
