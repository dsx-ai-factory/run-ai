package resources

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	k8sClient        client.Client
	testEnv          *envtest.Environment
	apiAccessContext context.Context
	suite            = "Project Controller"
)

func TestResources(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, suite)
}
