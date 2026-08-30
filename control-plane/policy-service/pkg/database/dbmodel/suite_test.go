package dbmodel

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDbModel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DbModel Suite")
}
