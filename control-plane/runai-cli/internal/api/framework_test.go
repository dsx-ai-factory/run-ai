package api_test

import (
	"testing"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

func TestFramework(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Framework Suite")
}

var _ = Describe("Framework", func() {
	DescribeTable("String method",
		func(frame api.Framework, expected string) {
			Expect(frame.String()).To(Equal(expected))
		},
		Entry("Runai", api.Runai, api.RunaiFrameworkName),
		Entry("Knative", api.Knative, api.KnativeFrameworkName),
		Entry("MPI", api.MPI, api.MPIFrameworkName),
		Entry("PyTorch", api.PyTorch, api.PyTorchFrameworkName),
		Entry("JAX", api.Jax, api.JaxFrameworkName),
		Entry("TF", api.TF, api.TensorFlowFrameworkName),
		Entry("XGBoost", api.XGBoost, api.XGBoostFrameworkName),
		Entry("AnyDistributed", api.AnyDistributed, api.AnyDistributedFrameworkName),
		Entry("NimService", api.NimService, api.NimServiceFrameworkName),
		Entry("UndefinedFramework", api.UndefinedFramework, api.UndefinedFrameworkName),
		Entry("OutOfRange", api.Framework(100), api.UndefinedFrameworkName),
	)

	DescribeTable("ClusterAPIIdentifiers",
		func(f api.Framework, userName string, wantName string, wantFramework api.Framework) {
			gotName, gotFramework := f.ClusterAPIIdentifiers(userName)
			Expect(gotName).To(Equal(wantName))
			Expect(gotFramework).To(Equal(wantFramework))
		},
		Entry("NimService → prefixed name and NimService", api.NimService, "my-svc", "ew-nimservice-my-svc", api.NimService),
		Entry("Runai unchanged", api.Runai, "my-workspace", "my-workspace", api.Runai),
		Entry("Knative unchanged", api.Knative, "my-inf", "my-inf", api.Knative),
	)

	DescribeTable("CLIType method",
		func(frame api.Framework, expected string) {
			Expect(frame.CLIType()).To(Equal(expected))
		},
		Entry("Runai", api.Runai, "standard"),
		Entry("Knative", api.Knative, "knative"),
		Entry("MPI", api.MPI, "mpi"),
		Entry("PyTorch", api.PyTorch, "pytorch"),
		Entry("JAX", api.Jax, "jax"),
		Entry("TF", api.TF, "tf"),
		Entry("XGBoost", api.XGBoost, "xgboost"),
		Entry("AnyDistributed", api.AnyDistributed, "any"),
		Entry("NimService", api.NimService, "nim"),
		Entry("UndefinedFramework", api.UndefinedFramework, "undefined"),
	)

	DescribeTable("JobKind method",
		func(frame api.Framework, expected string) {
			Expect(frame.JobKind()).To(Equal(expected))
		},
		Entry("Runai", api.Runai, clusterconstants.KindRunaiJob),
		Entry("MPI", api.MPI, clusterconstants.KindMpiJob),
		Entry("PyTorch", api.PyTorch, clusterconstants.KindPyTorchJob),
		Entry("JAX", api.Jax, clusterconstants.KindJAXJob),
		Entry("TF", api.TF, clusterconstants.KindTFJob),
		Entry("XGBoost", api.XGBoost, clusterconstants.KindXGBoostJob),
		Entry("NimService", api.NimService, clusterconstants.KindNIMService),
		Entry("AnyDistributed", api.AnyDistributed, api.UndefinedFrameworkName),
		Entry("UndefinedFramework", api.UndefinedFramework, api.UndefinedFrameworkName),
	)
})
