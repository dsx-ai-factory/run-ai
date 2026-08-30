package describe_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDatasourceDescribe(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Datasource Describe Suite")
}
