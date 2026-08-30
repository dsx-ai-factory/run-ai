package runaijob

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	featuregate "github.com/run-ai/runai/runai-cluster/cluster/workload-common/feature-gate"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/security"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	commonutils "github.com/run-ai/runai/runai-cluster/common/utils"
	"k8s.io/utils/ptr"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	jupyterImage              = "jupyter/scipy-notebook"
	tensorboardImage          = "tensorflow/tensorflow:latest"
	jupyterCommand            = "start-notebook.sh"
	tensorboardCommand        = "tensorboard"
	jupyterBaseUrl            = "--NotebookApp.base_url=%s" // base url is concatenation of project + job name
	annotationClusterId       = "clusterId"
	envVarWorkloadUrlForWandb = "WANDB_NOTES"
)

// PodTemplateSpecBuilder - as deployments and runaijobs share the same pod template spec, it is isolated here to a sub-builder
// once deployment is replaced by the new inference the pod template spec builder can be removed,
// and become integral part of runaijob builder, as sohuld be. so this is actually temporary
type PodTemplateSpecBuilder struct {
	//
	//   the inputs which produce the pod template spec
	//
	WorkloadKind     string
	MergedSpec       *job.JobSpec
	ObjectType       *v1.TypeMeta   // to distinguish runaijob from deployments
	ObjectNaming     *v1.ObjectMeta // job/deployment naming and annotations
	IsJupyter        bool           // always false on deployment, true only in interactive jupyter runaijobs
	IsTensorboard    bool
	ExitingPodSpec   *corev1.PodSpec
	ResourceBaseName string
	PodLabels        map[string]string
	PodAnnotations   map[string]string

	K8sClient client.Client

	// always false on deployment, true only in interactive tensorboard runaijobs

	//
	//   the pointer to the pod template spec at the runaijob or the deployment
	//
	Pts *corev1.PodTemplateSpec
}

func (psb *PodTemplateSpecBuilder) BuildPodTemplateSpec(ctx context.Context, k8sClient client.Client) error {
	if psb.ExitingPodSpec != nil {
		psb.Pts.Spec = *psb.ExitingPodSpec
		return nil
	}

	if psb.MergedSpec.TerminationGracePeriodSeconds != nil {
		psb.Pts.Spec.TerminationGracePeriodSeconds = ptr.To(int64(psb.MergedSpec.TerminationGracePeriodSeconds.Deref()))
	}

	if err := psb.BuildPodSpecInitContainers(ctx); err != nil {
		return err
	}

	err := psb.BuildPodSpecContainers(ctx, k8sClient)
	if err != nil {
		return err
	}

	err = psb.BuildCommonPodSpec(ctx, k8sClient)
	if err != nil {
		return err
	}

	return psb.DetermineRestartPolicy()
}

func (psb *PodTemplateSpecBuilder) DetermineRestartPolicy() error {
	restartPolicy := psb.MergedSpec.ContainerFields.RestartPolicy
	if restartPolicy == nil || restartPolicy.Value == nil {
		// set the restart policy to on failure if there are more than one container
		// this is hotfix for status stuck in init, as the container is not ready but not failed either
		// this workaround creates an backoff limit for the pod and restarts until it fails by backoff limit
		if len(psb.Pts.Spec.Containers) > 1 {
			psb.Pts.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
		} else {
			psb.Pts.Spec.RestartPolicy = corev1.RestartPolicyNever
		}
		return nil
	}

	specRestartPolicy, err := workloadbuilders.GetRestartPolicy(restartPolicy)
	if err != nil {
		return err
	}
	if specRestartPolicy != nil {
		psb.Pts.Spec.RestartPolicy = *specRestartPolicy
	}
	return nil
}

func (psb *PodTemplateSpecBuilder) BuildCommonPodSpec(ctx context.Context, k8sClient client.Client) error {
	ps := &psb.Pts.Spec

	if ps.Volumes == nil {
		ps.Volumes = []corev1.Volume{}
	}

	if ps.Tolerations == nil {
		ps.Tolerations = []corev1.Toleration{}
	}

	ps.HostIPC = psb.MergedSpec.HostIpc.Deref()
	ps.HostNetwork = psb.MergedSpec.HostNetwork.Deref()
	if ps.HostNetwork {
		ps.DNSPolicy = corev1.DNSClusterFirstWithHostNet
	}

	if psb.MergedSpec.NodeType.Deref() != "" {
		if ps.NodeSelector == nil {
			ps.NodeSelector = map[string]string{}
		}
		ps.NodeSelector[constants.NodeTypeKey] = psb.MergedSpec.NodeType.Deref()
	}

	ps.SchedulerName = constants.RunaiScheduler

	tolerations, err := workloadbuilders.GetTolerations(ctx, &psb.MergedSpec.ContainerFields)
	if err != nil {
		return err
	}
	ps.Tolerations = append(ps.Tolerations, tolerations...)

	namespace := psb.ObjectNaming.Namespace
	volumes, err := workloadbuilders.GetVolumes(ctx, &psb.MergedSpec.StorageFields, &psb.MergedSpec.ContainerFields, &psb.MergedSpec.NetworkingFields, k8sClient, namespace, psb.ResourceBaseName, "", psb.MergedSpec.IsJupyter.Deref(), true, false)
	if err != nil {
		return err
	}
	ps.Volumes = append(ps.Volumes, volumes...)

	err = workloadbuilders.SetPriorityClassName(ctx, k8sClient, psb.MergedSpec.PriorityClassName, ps)
	if err != nil {
		return err
	}

	workloadKindInitials, err := common.GetWorkloadInitials(psb.WorkloadKind)
	if err != nil {
		return err
	}

	err = workloadbuilders.BuildImagePullSecrets(ctx, psb.MergedSpec.ImagePullSecrets, psb.K8sClient, namespace, &ps.ImagePullSecrets, psb.ObjectNaming.Name, string(workloadKindInitials))
	if err != nil {
		return err
	}

	return nil
}

func (psb *PodTemplateSpecBuilder) BuildJobPodSpecContainers(ctx context.Context, k8sClient client.Client) error {

	ps := &psb.Pts.Spec

	initials, err := common.GetWorkloadInitials(psb.WorkloadKind)
	if err != nil {
		return err
	}

	var envs []corev1.EnvVar
	for _, name := range flds.SortedDictionaryMapKeys(psb.MergedSpec.Environment.Deref()) {
		value := psb.MergedSpec.Environment.Deref()[name]
		if !ptr.Deref(value.Deleted, false) {
			envs = append(envs, workloadbuilders.ConstructEnvVar(ctx, name, value.Value, &psb.MergedSpec.ContainerFields, psb.ObjectNaming, psb.ResourceBaseName, string(initials)))
		}
	}

	//  add home directory env var. do not do it for Jupyter, cause Jupyter has special handling for home directory
	if !psb.MergedSpec.IsJupyter.Deref() {
		if workloadbuilders.NeedToSetupRunaiHome(&psb.MergedSpec.ContainerFields, psb.MergedSpec.IsJupyter.Deref()) {
			envs = append(envs, corev1.EnvVar{
				Name:  constants.HomeDirEnvName,
				Value: constants.RunaiDefaultHomeDir,
			})
		}
	}

	//
	//   for workloads created from frontend v2 ui - add env variables containing url for jumping to the
	//   workload/training in frontend. may be useful for tools such as WANDB
	//
	annotations := psb.MergedSpec.Annotations.Deref()
	workspaceId := ptr.Deref(annotations[commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload].Value, "")
	clusterId := ptr.Deref(annotations[annotationClusterId].Value, "")

	if workspaceId != "" && clusterId != "" {
		jobName := psb.ObjectNaming.Name
		workloadUrl, err := workloadbuilders.ComposeWorkspaceUrl(ctx, psb.WorkloadKind, jobName, clusterId)
		if err != nil {
			return err
		}
		envs = append(envs, workloadbuilders.ConstructEnvVar(ctx, envVarWorkloadUrlForWandb, ptr.To(workloadUrl.String()),
			&psb.MergedSpec.ContainerFields, psb.ObjectNaming, psb.ResourceBaseName, string(initials)))
	}

	//
	//   break the command and arguments to parts.
	//
	command := workloadbuilders.PatchTemplateVariables(psb.MergedSpec.Command.Deref(), &psb.MergedSpec.ContainerFields, psb.ObjectNaming, ctx)
	args := workloadbuilders.PatchTemplateVariables(psb.MergedSpec.Arguments.Deref(), &psb.MergedSpec.ContainerFields, psb.ObjectNaming, ctx)

	commandVerbs, err := workloadbuilders.ParseCommandLine(command)
	if err != nil {
		return err
	}
	commandArgs, err := workloadbuilders.ParseCommandLine(args)
	if err != nil {
		return err
	}

	if commandArgs == nil {
		commandArgs = make([]string, 0)
	}

	if psb.IsTensorboard {
		if len(commandVerbs) == 0 {
			commandVerbs = []string{tensorboardCommand}
		}
		if psb.MergedSpec.IngressUrl.Deref() != "" {
			jobUrl, err := workloadbuilders.ComposeJupyterOrTensorboardJobUrl(ctx, psb.ResourceBaseName)
			if err != nil {
				return err
			}
			commandArgs = append(commandArgs, "--logdir")
			commandArgs = append(commandArgs, psb.MergedSpec.TensorboardLogdir.Deref())
			commandArgs = append(commandArgs, "--port")
			commandArgs = append(commandArgs, fmt.Sprintf("%d", workloadbuilders.DefaultJupyterAndTensorContainerPort))
			commandArgs = append(commandArgs, "--host")
			commandArgs = append(commandArgs, "0.0.0.0")
			if jobUrl.Path != "" && jobUrl.Path != "/" {
				commandArgs = append(commandArgs, "--path_prefix")
				commandArgs = append(commandArgs, jobUrl.Path)
			}
		}
	}

	if psb.IsJupyter {
		//
		//    default command for jupyter is to start the notebook
		//    - in userIdentityMode it goes into the command
		//    - in vanilla mode it goes to args, allowing the image to specify custom command
		//
		if workloadconfig.JupyterUserIdentity() {
			if len(commandVerbs) == 0 {
				commandVerbs = []string{jupyterCommand}
			}
		} else {
			commandArgs = append(commandArgs, jupyterCommand)
		}

		//
		//    add start notebook arguments, if needed
		//
		if psb.MergedSpec.IngressUrl.Deref() != "" {
			jobUrl, err := workloadbuilders.ComposeJupyterOrTensorboardJobUrl(ctx, psb.ObjectNaming.Name)
			if err != nil {
				return err
			}
			if jobUrl.Path != "" && jobUrl.Path != "/" {
				commandArgs = append(commandArgs, fmt.Sprintf(jupyterBaseUrl, jobUrl.Path))
			}
		}
		commandArgs = append(commandArgs, fmt.Sprintf("--NotebookApp.token='%s'", psb.MergedSpec.NotebookToken.Deref()))
		commandArgs = append(commandArgs, "--NotebookApp.password=''")

		if workloadconfig.JupyterUserIdentity() {
			chownHome := false
			if psb.MergedSpec.RunAsUser != nil && psb.MergedSpec.RunAsUser.Deref() {
				// https://github.com/jupyter/docker-stacks/blob/master/base-notebook/start.sh
				// Set UID and GID for jupyter itself
				if psb.MergedSpec.RunAsUid != nil {
					uidStr := fmt.Sprintf("%d", psb.MergedSpec.RunAsUid.Deref())
					envs = append(envs, workloadbuilders.ConstructEnvVar(ctx, constants.JupyterNbUid, ptr.To(uidStr), nil, psb.ObjectNaming, psb.ResourceBaseName, string(initials)))
					chownHome = true
				}
				if psb.MergedSpec.RunAsGid != nil {
					gidStr := fmt.Sprintf("%d", psb.MergedSpec.RunAsGid.Deref())
					envs = append(envs, workloadbuilders.ConstructEnvVar(ctx, constants.JupyterNbGid, ptr.To(gidStr), nil, nil, psb.ResourceBaseName, string(initials)))
					chownHome = true
				}
			}
			if chownHome {
				// see https://jupyter-docker-stacks.readthedocs.io/en/latest/using/common.html#docker-options
				envs = append(envs, workloadbuilders.ConstructEnvVar(ctx, "CHOWN_HOME", ptr.To("yes"), nil, nil, psb.ResourceBaseName, string(initials)))
				envs = append(envs, workloadbuilders.ConstructEnvVar(ctx, "CHOWN_HOME_OPTS", ptr.To("-R"), nil, nil, psb.ResourceBaseName, string(initials)))
			}
		}
	}

	jobResources, envVarResources, err := workloadbuilders.GetResourcesReqAndLimits(ctx, &psb.MergedSpec.ResourceAllocationFields)
	if err != nil {
		return err
	}

	envs = append(envs, envVarResources...)

	//
	//    supplemental groups, if specified, should go to the pod security context
	//
	err = security.GetPodSecurityContext(&psb.MergedSpec.ContainerFields, psb.MergedSpec.IsJupyter.Deref(), ps)
	if err != nil {
		return err
	}

	ports, err := workloadbuilders.GetPorts(ctx, &psb.MergedSpec.NetworkingFields, false)
	if err != nil {
		return err
	}

	sc := security.GetSecurityContext(&psb.MergedSpec.ContainerFields, &psb.MergedSpec.AccessControlFields, psb.MergedSpec.IsJupyter.Deref(), &envs)

	var image string
	if psb.IsJupyter {
		image = psb.MergedSpec.Image.DerefOrDefault(jupyterImage)
	} else if psb.IsTensorboard {
		image = psb.MergedSpec.Image.DerefOrDefault(tensorboardImage)
	} else {
		image = psb.MergedSpec.Image.Deref()
		if image == "" {
			return fmt.Errorf("missing value for required: image")
		}
	}

	err = workloadbuilders.IdentifyPvcsVolumeMode(ctx, psb.MergedSpec.WorkloadFields.Pvcs, psb.ObjectNaming.Namespace, k8sClient)
	if err != nil {
		return err
	}

	volumeMounts, err := workloadbuilders.GetVolumeMounts(ctx, &psb.MergedSpec.StorageFields, &psb.MergedSpec.ContainerFields, psb.MergedSpec.IsJupyter.Deref(), true)
	if err != nil {
		return err
	}

	c := corev1.Container{
		Image:           image,
		ImagePullPolicy: corev1.PullPolicy(psb.MergedSpec.ImagePullPolicy.DerefOrDefault(constants.PullPolicyIfNotPresent)),
		Name:            psb.ObjectNaming.Name,
		SecurityContext: sc,
		Command:         commandVerbs,
		Args:            commandArgs,
		WorkingDir:      psb.MergedSpec.WorkingDir.Deref(),
		Env:             envs,
		Resources:       *jobResources,
		VolumeMounts:    volumeMounts,
		VolumeDevices:   workloadbuilders.GetVolumeDevices(ctx, &psb.MergedSpec.StorageFields),
		Ports:           ports,
		Stdin:           psb.MergedSpec.Stdin.Deref(),
		TTY:             psb.MergedSpec.Tty.Deref(),
	}

	err = workloadbuilders.CreateProbes(ctx, &psb.MergedSpec.ContainerFields, &c, true)
	if err != nil {
		return err
	}

	ps.Containers = append(ps.Containers, c)

	return nil
}

func (psb *PodTemplateSpecBuilder) BuildPodSpecInitContainers(ctx context.Context) error {
	if psb.Pts.Spec.InitContainers == nil {
		psb.Pts.Spec.InitContainers = make([]corev1.Container, 0, 1)
	}

	gitSyncInitContainers, err := workloadbuilders.GetGitSyncInitContainersWithError(ctx, &psb.MergedSpec.ContainerFields, &psb.MergedSpec.StorageFields, psb.IsJupyter)
	if err != nil {
		return err
	}
	psb.Pts.Spec.InitContainers = append(psb.Pts.Spec.InitContainers, gitSyncInitContainers...)

	usernameLoader, err := workloadbuilders.GetUsernameLoaderInitContainerWithError(psb.MergedSpec.ExposedUrls.Deref(), nil)
	if err != nil {
		return err
	}

	if usernameLoader != nil {
		psb.Pts.Spec.InitContainers = append(psb.Pts.Spec.InitContainers, *usernameLoader)
	}
	return nil
}

func (psb *PodTemplateSpecBuilder) BuildPodSpecContainers(ctx context.Context, k8sClient client.Client) error {

	if psb.Pts.Spec.Containers == nil {
		psb.Pts.Spec.Containers = []corev1.Container{}
	}

	err := psb.BuildJobPodSpecContainers(ctx, k8sClient)
	if err != nil {
		return err
	}

	s3container, err := workloadbuilders.GetS3containerWithError(ctx, &psb.MergedSpec.WorkloadFields)
	if err != nil {
		return err
	}

	if s3container != nil {
		featureGateUtil := featuregate.GetFeatureGateUtilInstance()

		if featureGateUtil.FeatureGateSidecarContainersEnable {
			psb.Pts.Spec.InitContainers = append(psb.Pts.Spec.InitContainers, s3container...)
		} else {
			psb.Pts.Spec.Containers = append(psb.Pts.Spec.Containers, s3container...)
		}
	}

	// Proxy Authorization related

	if workloadconfig.ExternalAuthUrlEnabled() {
		exposedUrls := psb.MergedSpec.WorkloadFields.ExposedUrls.Deref()

		noSuffixUrlGenerated := false
		if psb.IsJupyter || psb.IsTensorboard {
			noSuffixUrlGenerated = true
		}

		for idx, nick := range commonutils.SortedKeys(exposedUrls) {
			exposedUrl := exposedUrls[nick]
			if exposedUrl.Value == nil {
				continue
			}

			jobUrl, err := workloadbuilders.ComposeOauthIngressUrl(ctx, exposedUrl, nick, &noSuffixUrlGenerated, psb.ObjectNaming.Name)
			if err != nil {
				return err
			}

			cookiePrefix := fmt.Sprintf("%s-%s-%s-%s", psb.WorkloadKind, psb.ObjectNaming.Namespace, psb.ObjectNaming.Name, nick)
			authContainer, err := workloadbuilders.GetOauth2ProxyContainer(jobUrl, *exposedUrl.Value.ContainerPort,
				exposedUrl.Value.AuthorizedUsers, exposedUrl.Value.AuthorizedGroups, idx, nick, false, false, cookiePrefix)
			if err != nil {
				return err
			}

			featureGateUtil := featuregate.GetFeatureGateUtilInstance()

			if featureGateUtil.FeatureGateSidecarContainersEnable {
				psb.Pts.Spec.InitContainers = append(psb.Pts.Spec.InitContainers, *authContainer)
			} else {
				psb.Pts.Spec.Containers = append(psb.Pts.Spec.Containers, *authContainer)
			}
		}
	}

	return nil
}
