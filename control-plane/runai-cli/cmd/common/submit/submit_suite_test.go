package submit

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSubmit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Submit Suite")
}
