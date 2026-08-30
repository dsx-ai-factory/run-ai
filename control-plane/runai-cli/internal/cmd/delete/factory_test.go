package delete

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributed"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
)

var _ = Describe("Factory", func() {
	var (
		factory   *Factory
		instance  *config.Instance
		cmdHelper helper.Cmd
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		factory = NewCommandFactory(instance, cmdHelper, factories.GetCompletionFactoryMock(ctrl))
	})

	Describe("IsSupported", func() {
		DescribeTable("returns correct support status",
			func(framework api.Framework, workloadType api.WorkloadType, expected bool) {
				Expect(factory.IsSupported(framework, workloadType)).To(Equal(expected))
			},
			Entry("MPI framework", api.MPI, api.Distributed, true),
			Entry("XGBoost framework", api.XGBoost, api.Distributed, true),
			Entry("TF framework", api.TF, api.Distributed, true),
			Entry("PyTorch framework", api.PyTorch, api.Distributed, true),
			Entry("JAX framework", api.Jax, api.Distributed, true),
			Entry("Runai framework with Training workload", api.Runai, api.Training, true),
			Entry("Runai framework with Workspace workload", api.Runai, api.Workspace, true),
			Entry("UndefinedFramework with Workspace workload", api.UndefinedFramework, api.Workspace, false),
			Entry("Unsupported framework", api.UndefinedFramework, api.Training, false),
			Entry("Distributed workload with non-distributed framework", api.Runai, api.Distributed, false),
		)
	})

	Describe("IsAliasSupported", func() {
		DescribeTable("should return correct support alias status",
			func(framework api.Framework, workloadType api.WorkloadType, expected bool) {
				Expect(factory.IsAliasSupported(framework, workloadType)).To(Equal(expected))
			},
			Entry("MPI framework", api.MPI, api.Distributed, false),
			Entry("XGBoost framework", api.XGBoost, api.Distributed, false),
			Entry("TF framework", api.TF, api.Distributed, false),
			Entry("PyTorch framework", api.PyTorch, api.Distributed, false),
			Entry("JAX framework", api.Jax, api.Distributed, false),
			Entry("Runai framework with Training workload", api.Runai, api.Training, false),
			Entry("Runai framework with Workspace workload", api.Runai, api.Workspace, false),
		)
	})

	Describe("GetAliasCommand", func() {
		DescribeTable("should return error for all configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				_, err := factory.GetAliasCommand(framework, workloadType)
				Expect(err).To(MatchError(cmd.NewUnsupportedAliasError(framework, workloadType)))
			},
			Entry("MPI framework", api.MPI, api.Distributed),
			Entry("XGBoost framework", api.XGBoost, api.Distributed),
			Entry("TF framework", api.TF, api.Distributed),
			Entry("PyTorch framework", api.PyTorch, api.Distributed),
			Entry("JAX framework", api.Jax, api.Distributed),
			Entry("Runai framework with Training workload", api.Runai, api.Training),
			Entry("Runai framework with Workspace workload", api.Runai, api.Workspace),
		)
	})

	Describe("GetCommand", func() {
		DescribeTable("with supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType) {
				cmd, err := factory.GetCommand(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmd).NotTo(BeNil())
			},
			Entry("returns a command for MPI framework", api.MPI, api.Distributed),
			Entry("returns a command for XGBoost framework", api.XGBoost, api.Distributed),
			Entry("returns a command for TF framework", api.TF, api.Distributed),
			Entry("returns a command for PyTorch framework", api.PyTorch, api.Distributed),
			Entry("returns a command for Runai framework with Training workload", api.Runai, api.Training),
			Entry("returns a command for Runai framework with Workspace workload", api.Runai, api.Workspace),
		)

		Context("with unsupported configurations", func() {
			It("returns an error for unsupported framework and workload combination", func() {
				_, err := factory.GetCommand(api.UndefinedFramework, api.Training)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(api.UndefinedFramework, api.Training)))
			})

			It("returns an error for distributed workload with non-distributed framework", func() {
				_, err := factory.GetCommand(api.Runai, api.Distributed)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(api.Runai, api.Distributed)))
			})
		})
	})

	Describe("getCommandTexts", func() {
		DescribeTable("returns correct texts for supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType, expectWorkloadOnly bool) {
				texts, err := factory.getCommandTexts(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				if expectWorkloadOnly {
					Expect(texts).To(BeAssignableToTypeOf(NewWorkloadOnlyTexts(workloadType)))
				} else {
					Expect(texts).To(BeAssignableToTypeOf(NewFrameworkTexts(workloadType, framework)))
				}
			},
			Entry("returns FrameworkTexts for MPI framework", api.MPI, api.Distributed, false),
			Entry("returns FrameworkTexts for XGBoost framework", api.XGBoost, api.Distributed, false),
			Entry("returns FrameworkTexts for TF framework", api.TF, api.Distributed, false),
			Entry("returns FrameworkTexts for PyTorch framework", api.PyTorch, api.Distributed, false),
			Entry("returns FrameworkTexts for Runai framework with Training workload", api.Runai, api.Training, false),
			Entry("returns WorkloadOnlyTexts for Runai framework with Workspace workload", api.Runai, api.Workspace, true),
		)

		Context("with unsupported configurations", func() {
			It("returns error for unsupported framework and workload combination", func() {
				_, err := factory.getCommandTexts(api.UndefinedFramework, api.Training)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(api.UndefinedFramework, api.Training)))
			})
		})
	})

	Describe("getOperationFunc", func() {
		DescribeTable("returns correct operation function for supported configurations",
			func(framework api.Framework, workloadType api.WorkloadType, expectedType interface{}) {
				fn, err := factory.getOperationFunc(framework, workloadType)
				Expect(err).NotTo(HaveOccurred())
				Expect(fn).NotTo(BeNil())
				Expect(fn).To(BeAssignableToTypeOf(expectedType))
			},
			Entry("returns distributed.Delete for MPI framework", api.MPI, api.Distributed, distributed.Delete),
			Entry("returns distributed.Delete for XGBoost framework", api.XGBoost, api.Distributed, distributed.Delete),
			Entry("returns distributed.Delete for TF framework", api.TF, api.Distributed, distributed.Delete),
			Entry("returns distributed.Delete for PyTorch framework", api.PyTorch, api.Distributed, distributed.Delete),
			Entry("returns training.Delete for Runai framework with Training workload", api.Runai, api.Training, distributed.Delete),
			Entry("returns training.Delete for Runai framework with Workspace workload", api.Runai, api.Workspace, distributed.Delete),
		)

		Context("with unsupported configurations", func() {
			It("returns error for unsupported framework and workload combination", func() {
				_, err := factory.getOperationFunc(api.UndefinedFramework, api.Training)
				Expect(err).To(MatchError(cmd.NewUnsupportedError(api.UndefinedFramework, api.Training)))
			})
		})
	})
})
