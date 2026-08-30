package portforward

import (
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/capability"

	argshandler "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/args"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/cluster"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/portforward/forward"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/portforward/options"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/resolve"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	ColonStr                    = ":"
	SamePortForLocalRemote      = 0
	DifferentPortForLocalRemote = 1

	ParseBase = 10
	BitSize   = 32
)

var (
	instance  = config.GetInstance()
	cmdHelper helper.Cmd
)

func FlagOutputType(cmd *cobra.Command, options *options.Options, framework api.Framework) {
	cmd.Flags().StringVarP(&options.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = cmd.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	podDescription := getPodFlagDescription(framework)
	cmd.Flags().StringVar(&options.Pod, consts.Pod, "", podDescription)

	cmd.Flags().StringVar(&options.Address, consts.AttributeNameAddress, "localhost", consts.AttributeUsageAddress)
	cmd.Flags().DurationVar(&options.PodRunningTimeout, consts.AttributeNamePodRunningTimeout, 0, consts.AttributeUsagePodRunningTimeout)
	cmd.Flags().StringArrayVar(&options.Ports, consts.AttributeNamePorts, []string{}, consts.AttributeNamePorts)
	cmd.MarkFlagsOneRequired(consts.AttributeNamePorts)
}

func getPodFlagDescription(framework api.Framework) string {
	switch framework {
	case api.MPI, api.XGBoost, api.TF, api.PyTorch:
		return consts.PodIdAttributePortForwardMaster
	case api.Runai, api.Knative, api.Jax, api.LWS:
		return consts.PodIdAttributePortForwardRandom
	default:
		return consts.PodIdAttributePortForward
	}
}

func Pre(cmd *cobra.Command, args []string, options *options.Options) error {
	token := instance.GetToken()
	txn := instance.GetTxn()
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	controlUrl := instance.GetConfigValue(config.ControlPlaneUrl)

	options.WorkloadName = argshandler.GetNameFromArgs(&args)
	if options.Project == "" {
		options.Project = instance.GetConfigValue(config.ClusterProjectName)
	}

	capabilityChecker := capability.NewChecker(controlUrl, token, clusterId)
	isSupported, isSupportedErr := capabilityChecker.IsSupported(txn, options.WorkloadType, capability.PortForward)
	if isSupportedErr != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedPfCommandError, isSupportedErr)
	}
	if !isSupported {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedPfCommandError, fmt.Errorf(capability.ClusterVersionError, capability.MinimalClusterVersionForInfwClusterApiActions))
	}

	cmdHelper.AddValidation(validators.NewProjectValidator(options.Project))
	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}
	return nil
}

func Run(cmd *cobra.Command, options *options.Options) error {
	var (
		local, remote int32
		err           error
	)
	txn := instance.GetTxn()
	token := instance.GetToken()

	metadata, err := cluster.GetStatus()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedPfCommandError, err)
	}
	clusterURL, err := cluster.GetClusterApiUrl()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedPfCommandError, err)
	}
	instance.GetLogger().Debug("cluster metadata", zap.Any("metadata", metadata))

	options.WorkloadType, options.WorkloadFramework, options.WorkloadName, err = resolve.ClusterWorkload(
		instance, txn, options.WorkloadType, options.WorkloadFramework, options.Project, options.WorkloadName)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedPfCommandError, err)
	}

	workload := apis.NewWorkloadForClusterAPI(options.Project, options.WorkloadType, options.WorkloadName, options.WorkloadFramework)

	forwardList := forward.NewForwardList(txn.GetCtx(), workload, options, token, clusterURL)
	for _, p := range options.Ports {
		local, remote, err = parseForwardingPorts(p)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedPfCommandError, err)
		}
		fp := forward.NewForward(token, options, local, remote, workload, clusterURL.String(), metadata)
		forwardList.Add(fp)
	}

	err = cluster.WaitForWorkloadInClusterRunning(txn.GetCtx(), workload, clusterURL.String(), token, options.PodRunningTimeout)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedPfCommandError, err)
	}

	cmd.Printf(consts.StartingPortForward, options.Ports)
	err = forwardList.BindLocalPort()
	defer forwardList.Close()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedPfCommandError, err)
	}
	err = forwardList.Run()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedPfCommandError, err)
	}
	cmd.Printf(consts.ClosingPortForward, options.Ports)
	return nil
}

func parseForwardingPorts(port string) (int32, int32, error) {
	var (
		local, remote int64
		err           error
	)
	cc := strings.Count(port, ColonStr)

	switch cc {
	case SamePortForLocalRemote:
		local, err = strconv.ParseInt(port, ParseBase, BitSize)
		if err != nil {
			return 0, 0, err
		}
		remote = local
	case DifferentPortForLocalRemote:
		portSplit := strings.Split(port, ColonStr)
		local, err = strconv.ParseInt(portSplit[0], ParseBase, BitSize)
		if err != nil {
			return 0, 0, err
		}
		remote, err = strconv.ParseInt(portSplit[1], ParseBase, BitSize)
		if err != nil {
			return 0, 0, err
		}
	default:
		return 0, 0, fmt.Errorf("there can be only one colon")
	}

	return int32(local), int32(remote), nil
}
