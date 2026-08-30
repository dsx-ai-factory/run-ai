package resources

import (
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	kaiqueuecontroller "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/queue_controller"
	"k8s.io/utils/ptr"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	queue_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/queue-controller"
)

func convertQueueControllerConfig(runaiConfig *runaiv1.RunaiConfig) *kaiqueuecontroller.QueueController {
	config := runaiConfig.Spec.QueueController
	return &kaiqueuecontroller.QueueController{
		Service: &kaicommon.Service{
			Enabled:   config.Enabled,
			Image:     convertImage(config.Image),
			Resources: (*kaicommon.Resources)(config.Resources),
			Affinity:  config.Affinity,
		},
		ControllerService: convertQueueControllerService(config),
		Webhooks: &kaiqueuecontroller.QueueControllerWebhooks{
			WebhookConfigurationNamePrefix: ptr.To("runai-queue-validation-"),
		},
		VPA:                            convertVPA(config.VPA),
		MetricsNamespace:               ptr.To("runai"),
		QueueLabelToMetricLabel:        ptr.To("project=project,runai/department-name=department,runai/node-pool=nodepool"),
		QueueLabelToDefaultMetricValue: ptr.To("runai/node-pool=default"),
	}
}

func convertQueueControllerService(config *queue_controller.QueueController) *kaiqueuecontroller.Service {
	if config.Service == nil {
		return nil
	}
	serviceConfig := &kaiqueuecontroller.Service{}

	if config.Service.Metrics != nil {
		serviceConfig.Metrics = &kaiqueuecontroller.PortMapping{
			Port:       config.Service.Metrics.Port,
			TargetPort: config.Service.Metrics.TargetPort,
			Name:       config.Service.Metrics.Name,
		}
	}
	if config.Service.Webhook != nil {
		serviceConfig.Webhook = &kaiqueuecontroller.PortMapping{
			Port:       config.Service.Webhook.Port,
			TargetPort: config.Service.Webhook.TargetPort,
		}
	}
	return serviceConfig
}
