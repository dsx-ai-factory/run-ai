package services

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)

	defer runai_config.FlushAll()

	RunSpecs(t, "Services Suite")
}
