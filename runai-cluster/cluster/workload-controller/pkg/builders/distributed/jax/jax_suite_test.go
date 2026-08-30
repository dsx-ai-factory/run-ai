package jax_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJax(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jax Suite")
}
