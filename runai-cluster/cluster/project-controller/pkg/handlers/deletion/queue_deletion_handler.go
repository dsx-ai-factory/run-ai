package deletion

import (
	multierror "github.com/hashicorp/go-multierror"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/handlers"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type QueueDeletionHandler struct {
	CommonResourceDeletionHandler
}

func NewQueueDeletionHandler(client client.Client) QueueDeletionHandler {
	return QueueDeletionHandler{
		CommonResourceDeletionHandler: CommonResourceDeletionHandler{
			WithLoggerAndCli: common.WithLoggerAndCli{
				Client: client,
				Log:    ctrl.Log.WithName("deletion_handlers").WithName(common.LogQueueTag),
			},
		},
	}
}

func (handler QueueDeletionHandler) OnDelete(project *kaiv1alpha1.Project) ([]kaiv1alpha1.ProjectCondition, error) {
	err := handler.onDeleteInner(project)

	return []kaiv1alpha1.ProjectCondition{{
		Type:    kaiv1alpha1.QueuesReady,
		Status:  handlers.GetStatusFromError(err),
		Reason:  getDeletionReasonFromError(err, QueuesDeletionHandlerFailed),
		Message: handlers.GetMessageFromError(err),
	}}, err
}

func (handler QueueDeletionHandler) onDeleteInner(project *kaiv1alpha1.Project) error {
	queues, err := handler.GetQueuesForProjectByLabel(project.Name)
	if err != nil {
		handler.Log.Error(err, "Failed to list queues for project", common.LogProjectTag, project.Name)

		return err
	}

	for _, queue := range queues.Items {
		existingQueue := queue.DeepCopy()
		innerErr := handler.DeleteExistingResourceIfNeeded(
			existingQueue, queue.Name,
			common.LogQueueTag, project.Name,
			&client.DeleteOptions{PropagationPolicy: &common.OrphanDeletePolicyRef})

		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}
	}

	return err
}
