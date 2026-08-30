package nim_autoscale_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNimAutoscale(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI NIM Autoscale Tests Suite")
}
