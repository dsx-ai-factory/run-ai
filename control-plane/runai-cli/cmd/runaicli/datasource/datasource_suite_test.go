package datasource_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDatasource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Datasource Suite")
}
