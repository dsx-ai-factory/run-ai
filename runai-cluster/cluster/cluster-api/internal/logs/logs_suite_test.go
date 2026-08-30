package logs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLogs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logs Suite")
}
