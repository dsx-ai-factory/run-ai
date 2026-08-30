package resources

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/kai/schedulingshardargs"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runai_scheduler "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-scheduler"
)

var _ = Describe("buildShardArgs", func() {
	Context("when both inputs are nil", func() {
		It("returns an empty map", func() {
			Expect(buildShardArgs(nil, nil)).To(BeEmpty())
		})
	})

	Context("node affinity", func() {
		It("translates RestrictScheduling into restrict-node-scheduling", func() {
			args := buildShardArgs(nil, &global.NodeAffinity{RestrictScheduling: true})
			Expect(args).To(HaveKeyWithValue(schedulingshardargs.RestrictNodeScheduling, "true"))
		})

		It("records restrict-node-scheduling=false explicitly", func() {
			args := buildShardArgs(nil, &global.NodeAffinity{RestrictScheduling: false})
			Expect(args).To(HaveKeyWithValue(schedulingshardargs.RestrictNodeScheduling, "false"))
		})

		It("does not translate NFDGPULabel (it has no scheduler arg)", func() {
			args := buildShardArgs(nil, &global.NodeAffinity{NFDGPULabel: "nvidia.com/gpu.present"})
			Expect(args).To(HaveLen(1))
			Expect(args).To(HaveKey(schedulingshardargs.RestrictNodeScheduling))
		})
	})

	Context("RunAIScheduler top-level fields", func() {
		It("translates the SchedulingShard-facing booleans", func() {
			config := &runai_scheduler.RunAIScheduler{
				DetailedFitErrors:     ptr.To(true),
				AdvancedCSIScheduling: ptr.To(false),
				Profiling:             &runai_scheduler.Profiling{Enabled: true},
			}
			args := buildShardArgs(config, nil)
			Expect(args).To(HaveKeyWithValue(schedulingshardargs.DetailedFitErrors, "true"))
			Expect(args).To(HaveKeyWithValue(schedulingshardargs.ScheduleCSIStorage, "false"))
			Expect(args).To(HaveKeyWithValue(schedulingshardargs.EnableProfiler, "true"))
		})

		It("omits unset fields", func() {
			Expect(buildShardArgs(&runai_scheduler.RunAIScheduler{}, nil)).To(BeEmpty())
		})

		It("does not translate scheduler-deployment fields (resources, vpa, ...)", func() {
			config := &runai_scheduler.RunAIScheduler{
				Enabled:              ptr.To(true),
				GOGC:                 ptr.To(200),
				InferencePreemptible: ptr.To(true),
			}
			Expect(buildShardArgs(config, nil)).To(BeEmpty())
		})
	})

	Context("RunAIScheduler.Args", func() {
		It("translates every arg into its KAI key", func() {
			config := &runai_scheduler.RunAIScheduler{
				Args: &runai_scheduler.Args{
					Verbosity:                        ptr.To(6),
					DefaultSchedulerName:             ptr.To("kai-scheduler"),
					DefaultSchedulerPeriod:           ptr.To("1s"),
					MaxNumberConsolidationPreemptees: ptr.To(4),
					ClientQPS:                        ptr.To(50),
					ClientBurst:                      ptr.To(100),
					UseSchedulingSignatures:          ptr.To(true),
					FullHierarchyFairness:            ptr.To(false),
					AllowConsolidatingReclaim:        ptr.To(true),
					DefaultStalenessGracePeriod:      ptr.To("60s"),
					NumOfStatusRecordingWorkers:      ptr.To(5),
				},
			}

			Expect(buildShardArgs(config, nil)).To(Equal(map[string]string{
				schedulingshardargs.Verbosity:                   "6",
				schedulingshardargs.SchedulerName:               "kai-scheduler",
				schedulingshardargs.SchedulePeriod:              "1s",
				schedulingshardargs.MaxConsolidationPreemptees:  "4",
				schedulingshardargs.QPS:                         "50",
				schedulingshardargs.Burst:                       "100",
				schedulingshardargs.UseSchedulingSignatures:     "true",
				schedulingshardargs.FullHierarchyFairness:       "false",
				schedulingshardargs.AllowConsolidatingReclaim:   "true",
				schedulingshardargs.DefaultStalenessGracePeriod: "60s",
				schedulingshardargs.NumOfStatusRecordingWorkers: "5",
			}))
		})

		It("omits args that are unset", func() {
			config := &runai_scheduler.RunAIScheduler{Args: &runai_scheduler.Args{Verbosity: ptr.To(6)}}
			args := buildShardArgs(config, nil)
			Expect(args).To(HaveLen(1))
			Expect(args).To(HaveKeyWithValue(schedulingshardargs.Verbosity, "6"))
		})
	})

	Context("full conversion", func() {
		It("combines the node-affinity policy with the scheduler config", func() {
			config := &runai_scheduler.RunAIScheduler{
				DetailedFitErrors: ptr.To(true),
				Args:              &runai_scheduler.Args{Verbosity: ptr.To(6)},
			}

			Expect(buildShardArgs(config, &global.NodeAffinity{RestrictScheduling: true})).To(Equal(map[string]string{
				schedulingshardargs.RestrictNodeScheduling: "true",
				schedulingshardargs.DetailedFitErrors:      "true",
				schedulingshardargs.Verbosity:              "6",
			}))
		})
	})
})
