package datavolume_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDatavolume(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Datavolume Suite")
}
