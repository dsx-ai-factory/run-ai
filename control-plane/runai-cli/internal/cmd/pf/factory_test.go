package pf

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
)

var _ = Describe("Port Forward Factory", func() {
	var (
		factory   *Factory
		instance  *config.Instance
		cmdHelper helper.Cmd
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		instance = config.GetInstance()
		cmdHelper = helper.Cmd{}
		factory = NewCommandFactory(instance, cmdHelper, factories.GetCompletionFactoryMock(ctrl))
	})

	Describe("IsSupported", func() {
		DescribeTable("returns correct support status",
			func(framework api.Framework, workloadType api.WorkloadType, expected bool) {
				Expect(factory.IsSupported(framework, workloadType)).To(Equal(expected))
			},
			Entry("MPI Framework with Distributed", api.MPI, api.Distributed, true),
			Entry("XGBoost Framework with Distributed", api.XGBoost, api.Distributed, true),
			Entry("TF Framework with Distributed", api.TF, api.Distributed, true),
			Entry("PyTorch Framework with Distributed", api.PyTorch, api.Distributed, true),
			Entry("JAX Framework with Distributed", api.Jax, api.Distributed, true),
			Entry("Runai Framework with Training workload", api.Runai, api.Training, true),
			Entry("MPI Framework with Training", api.MPI, api.Training, false),
			Entry("Runai Framework with Workspace workload", api.Runai, api.Workspace, true),
			Entry("Knative Framework with Inference workload", api.Knative, api.Inference, true),
			Entry("Unsupported framework", api.UndefinedFramework, api.Training, false),
		)
	})

	Describe("GetCommand", func() {
		Context("with supported configurations", func() {
			DescribeTable("returns command successfully",
				func(framework api.Framework, workloadType api.WorkloadType) {
					cmd, err := factory.GetCommand(framework, workloadType)
					Expect(err).NotTo(HaveOccurred())
					Expect(cmd).NotTo(BeNil())
				},
				Entry("returns a command for MPI Framework with Distributed", api.MPI, api.Distributed),
				Entry("returns a command for TF Framework with Distributed", api.TF, api.Distributed),
				Entry("returns a command for XGBoost Framework with Distributed", api.XGBoost, api.Distributed),
				Entry("returns a command for PyTorch Framework with Distributed", api.PyTorch, api.Distributed),
				Entry("returns a command for JAX Framework with Distributed", api.Jax, api.Distributed),
				Entry("returns a command for Runai Framework with Training workload", api.Runai, api.Training),
			)
		})

		Context("with unsupported configurations", func() {
			DescribeTable("returns appropriate error",
				func(framework api.Framework, workloadType api.WorkloadType) {
					_, err := factory.GetCommand(framework, workloadType)
					Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
				},
				Entry("MPI with Training", api.MPI, api.Training),
				Entry("Runai with Distributed", api.Runai, api.Distributed),
				Entry("UndefinedFramework", api.UndefinedFramework, api.Training),
			)
		})
	})

	Describe("getCommandTexts", func() {
		DescribeTable("returns correct text type for supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				texts, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				Expect(texts).To(BeAssignableToTypeOf(NewFrameworkTexts(workloadType, framework)))
			},
			Entry("returns FrameworkPFTexts for MPI Framework with Distributed", api.MPI, api.Distributed),
			Entry("returns FrameworkPFTexts for TF Framework with Distributed", api.TF, api.Distributed),
			Entry("returns FrameworkPFTexts for XGBoost Framework with Distributed", api.XGBoost, api.Distributed),
			Entry("returns FrameworkPFTexts for PyTorch Framework with Distributed", api.PyTorch, api.Distributed),
			Entry("returns FrameworkPFTexts for JAX Framework with Distributed", api.Jax, api.Distributed),
			Entry("returns FrameworkPFTexts for Runai Framework with Training", api.Runai, api.Training),
		)

		DescribeTable("returns error for unsupported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(framework, workloadType)))
			},
			Entry("MPI with Training", api.MPI, api.Training),
			Entry("Runai with Distributed", api.Runai, api.Distributed),
			Entry("UndefinedFramework", api.UndefinedFramework, api.Training),
		)
	})

	Describe("IsAliasSupported", func() {
		It("always returns false", func() {
			Expect(factory.IsAliasSupported(api.MPI, api.Distributed)).To(BeFalse())
			Expect(factory.IsAliasSupported(api.Runai, api.Training)).To(BeFalse())
			Expect(factory.IsAliasSupported(api.UndefinedFramework, api.Workspace)).To(BeFalse())
			Expect(factory.IsAliasSupported(api.UndefinedFramework, api.Inference)).To(BeFalse())
		})
	})

	Describe("GetAliasCommand", func() {
		DescribeTable("always returns unsupported error",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.GetAliasCommand(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedAliasError(framework, workloadType)))
			},
			Entry("MPI Framework with Distributed", api.MPI, api.Distributed),
			Entry("Runai Framework with Training", api.Runai, api.Training),
			Entry("UndefinedFramework with Training", api.UndefinedFramework, api.Training),
		)
	})
})
