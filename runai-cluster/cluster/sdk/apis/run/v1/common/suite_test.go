package common

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	RegistryTestValue = "registry_env"
	VersionTestValue  = "version_env"
)

func TestAPIs(t *testing.T) {
	t.Setenv(DefaultRegistryEnvVarName, RegistryTestValue)
	t.Setenv(DefaultVersionEnvVarName, VersionTestValue)
	RegisterFailHandler(Fail)
	RunSpecs(t, "run.ai/v1 Common Types")
}
