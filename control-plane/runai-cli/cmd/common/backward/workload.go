package backward

import (
	"context"
	"errors"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	workloadCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	controlApi "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	workloadService "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	instance = config.GetInstance()
)

func HandleUniqWorkloadBackwardCompatiblePreE(ctx context.Context, cmd *cobra.Command, workloadName, workloadType, project, errorCmdMsg string) (*apis.Workload, error) {
	controlUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	token := instance.GetToken()
	clusterID := instance.GetConfigValue(config.ClusterUUID)
	if workloadType != "" {
		parsedWorkloadType, err := api.ParseWorkloadType(workloadType)
		if err != nil {
			return nil, cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToGetWorkloadType, errors.New(consts.FailedToGetWorkloadType))
		}

		return apis.NewWorkload(project, parsedWorkloadType, workloadName, parsedWorkloadType.DefaultFramework()), nil
	}
	reqWorkload := controlApi.Workload{
		Name:      workloadName,
		Project:   project,
		ClusterID: clusterID,
	}
	clusterWorkload, err := workloadService.SearchClusterWorkload(ctx, &reqWorkload, controlUrl, token)
	if err != nil {
		return nil, workloadCommon.DisplayGetWorkloadError(err, cmd, errorCmdMsg)
	}

	return clusterWorkload, nil
}
