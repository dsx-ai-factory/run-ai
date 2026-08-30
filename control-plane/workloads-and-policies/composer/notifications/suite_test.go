package notifications

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkloadsComposer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "notifications composer suite")
}
