package orchestrator

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	"github.com/run-ai/runai/runai-installer/installer/internal/install"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq"
)

type FlowBuilderFunc func(deps InstallDeps, runaiData *model.RunAIInstallData, externalEvents chan *flow.FlowStatusReportingMsg) []*FlowGroup

var BuildFlowGroups FlowBuilderFunc = buildFlowGroups

func buildFlowGroups(deps InstallDeps, runaiData *model.RunAIInstallData, externalEvents chan *flow.FlowStatusReportingMsg) []*FlowGroup {
	var flowGroups []*FlowGroup

	if len(runaiData.Prerequisites) > 0 {
		prereqAssets := prereq.GetPrerequisites(deps.K8sClient)
		prereqFlows := prereq.GetFlowsToInstall(deps.K8sClient, deps.HelmInstaller, deps.RunAIClientProvider, runaiData, prereqAssets)
		flowGroups = append(flowGroups, &FlowGroup{
			Flows:   prereqFlows,
			Manager: flow.NewFlowManagerWithExternalEvents(prereqFlows, externalEvents, false),
		})
	} else {
		logging.Infof("No prerequisites configured, skipping prerequisite flows")
	}

	runaiFlows := install.GetFlowsToInstall(deps.K8sClient, deps.HelmInstaller, deps.RunAIClientProvider, runaiData)
	flowGroups = append(flowGroups, &FlowGroup{
		Flows:   runaiFlows,
		Manager: flow.NewFlowManagerWithExternalEvents(runaiFlows, externalEvents, false),
	})

	return flowGroups
}
