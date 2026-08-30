package resources

import (
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	kaischeduler "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/scheduler"
	"k8s.io/utils/ptr"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runai_scheduler "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-scheduler"
)

func convertSchedulerConfig(runaiConfig *runaiv1.RunaiConfig) *kaischeduler.Scheduler {
	schedulerConfig := runaiConfig.Spec.RunAIScheduler
	return &kaischeduler.Scheduler{
		Service: &kaicommon.Service{
			Enabled:   ptr.To(ptr.Deref(schedulerConfig.Enabled, true)),
			Image:     convertImage(schedulerConfig.Image),
			Resources: (*kaicommon.Resources)(schedulerConfig.Resources),
			Affinity:  schedulerConfig.Affinity,
		},
		GOGC:             schedulerConfig.GOGC,
		SchedulerService: convertSchedulerService(schedulerConfig),
		VPA:              convertVPA(schedulerConfig.VPA),
	}
}

func convertSchedulerService(config *runai_scheduler.RunAIScheduler) *kaischeduler.Service {
	if config.Service == nil {
		return nil
	}
	return &kaischeduler.Service{
		Type:       config.Service.Type,
		Port:       config.Service.Port,
		TargetPort: config.Service.TargetPort,
	}
}
