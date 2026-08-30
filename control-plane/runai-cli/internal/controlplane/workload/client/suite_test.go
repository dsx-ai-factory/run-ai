package client

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClients(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Describe Clients Suite")
}
