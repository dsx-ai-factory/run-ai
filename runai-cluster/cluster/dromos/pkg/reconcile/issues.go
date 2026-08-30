package reconcile

import (
	"context"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func SetControllerIssues(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, state *State) {
	logger := log.FromContext(ctx)
	if runner == nil || state == nil {
		logger.V(2).Info("Runner or state is nil")
		return
	}

	if runner.Status.Plugins == nil {
		runner.Status.Plugins = make(map[string]*dromosv1alpha1.PluginStatus)
	}

	// Clear all plugin issues lists - we don't want to keep the old ones.
	for pluginName, pluginStatus := range runner.Status.Plugins {
		pluginStatus.Issues = nil
		logger.V(3).Info("Cleared issues for plugin", "plugin", pluginName)
	}

	for pluginName, issues := range state.PluginsMap() {
		if len(issues) == 0 {
			continue
		}

		// Convert internal issues to dromosv1alpha1.Issue
		v1alpha1Issues := make([]dromosv1alpha1.Issue, 0, len(issues))
		for _, issue := range issues {
			v1alpha1Issues = append(v1alpha1Issues, dromosv1alpha1.Issue{
				Kind:        issue.GVK.Kind,
				Name:        issue.Name,
				Description: issue.Error(),
				Group:       issue.Group.String(),
				IsCritical:  issue.IsCritical,
			})
		}

		_, exists := runner.Status.Plugins[pluginName]
		if !exists {
			runner.Status.Plugins[pluginName] = &dromosv1alpha1.PluginStatus{
				Phase: dromosv1alpha1.PluginPhaseUnknown,
			}
			logger.V(3).Info("Created new plugin with issues", "plugin", pluginName, "issueCount", len(v1alpha1Issues))
		}

		runner.Status.Plugins[pluginName].Issues = v1alpha1Issues
		logger.V(3).Info("Updated issues for existing plugin", "plugin", pluginName, "issueCount", len(v1alpha1Issues))
	}

	// Set controller issues to the runner status.
	controllerIssues := state.Controller()
	if len(controllerIssues) > 0 {
		v1alpha1Issues := make([]dromosv1alpha1.Issue, 0, len(controllerIssues))
		for _, issue := range controllerIssues {
			v1alpha1Issues = append(v1alpha1Issues, dromosv1alpha1.Issue{
				Kind:        issue.GVK.Kind,
				Name:        issue.Name,
				Description: issue.Error(),
				Group:       issue.Group.String(),
				IsCritical:  issue.IsCritical,
			})
		}
		runner.Status.Issues = v1alpha1Issues
		logger.V(3).Info("Set controller issues", "issueCount", len(v1alpha1Issues))
	} else {
		runner.Status.Issues = nil
		logger.V(3).Info("Cleared controller issues")
	}
}
