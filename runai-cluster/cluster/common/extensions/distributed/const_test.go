package distributed_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/extensions"
	"github.com/run-ai/runai/runai-cluster/cluster/common/extensions/distributed"
)

var _ = Describe("Extensions", func() {
	var ext extensions.Group
	BeforeEach(func() {
		ext = extensions.Group{
			distributed.Jax:        extensions.Extension{},
			distributed.Pytorch:    extensions.Extension{},
			distributed.Tensorflow: extensions.Extension{},
			distributed.MPI:        extensions.Extension{},
			distributed.XGBoost:    extensions.Extension{},
		}
	})
	It("should equal to distributed extensions", func() {
		Expect(distributed.SupportedExtensions).Should(HaveLen(5))
		Expect(ext.Equals(distributed.SupportedExtensions)).To(BeTrue())
	})
})
