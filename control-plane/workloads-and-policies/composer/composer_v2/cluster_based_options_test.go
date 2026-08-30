package composer_v2

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
)

var _ = Describe("cluster-based composer options", func() {
	DescribeTable("gates ServiceSpec labels by cluster version",
		func(version string, expected bool) {
			capabilities, err := checker.NewClusterCapabilitiesCheckerForVersion(version)
			Expect(err).NotTo(HaveOccurred())
			options := composer.FieldsComposerOptions{}

			Expect(SetClusterDependentOptions(context.Background(), &options, capabilities)).To(Succeed())
			Expect(options.WorkloadV2ServiceSpecLabelsSupported).To(Equal(expected))
		},
		Entry("unsupported before 3.0", "2.99.99", false),
		Entry("supported at 3.0", "3.0.0", true),
	)
})
