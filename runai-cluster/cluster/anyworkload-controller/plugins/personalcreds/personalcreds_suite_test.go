package personalcreds

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPersonalcreds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Personalcreds Plugin Suite")
}
