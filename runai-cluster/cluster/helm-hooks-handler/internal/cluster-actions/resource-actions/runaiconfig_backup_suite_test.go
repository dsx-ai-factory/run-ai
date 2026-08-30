package resource_actions_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRunaiconfigBackup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RunaiconfigBackup Suite")
}
