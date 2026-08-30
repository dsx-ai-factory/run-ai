package handlers_test

// calling the import here to ensure that the helper init takes place before anything else
// see inside the helper package for more info
import (
	"testing"

	_ "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/tests/helper"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Suite")
}
