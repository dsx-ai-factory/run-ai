package reconcilers

import (
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type IntraEventSender struct {
	projectEvents chan event.GenericEvent
}

// Sends events for each project in the list to the ProjectReconciler via the dedicated channel
func (eventSender IntraEventSender) triggerEventsForAllProjects(
	projectsList kaiv1alpha1.ProjectList,
) (ctrl.Result, error) {
	for _, project := range projectsList.Items {
		go func(event event.GenericEvent) {
			eventSender.projectEvents <- event
		}(event.GenericEvent{
			Object: project.DeepCopy(),
		})
	}
	return ctrl.Result{}, nil
}
