/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed_to_kubeflow

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/topology"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	trackerApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	mpiv2beta1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	v2 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/modern-go/reflect2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultNumWorkers = 1

	// Env var injected by Kubeflow MPI builders so user containers can size MPI runs.
	mpiNumWorkersEnvName = "RUNAI_MPI_NUM_WORKERS"
)

var (
	MpiTypeMeta = metav1.TypeMeta{
		Kind:       "MPIJob",
		APIVersion: "kubeflow.org/v2beta1",
	}
	PytorchTypeMeta = metav1.TypeMeta{
		Kind:       "PyTorchJob",
		APIVersion: "kubeflow.org/v1",
	}
	TensorflowTypeMeta = metav1.TypeMeta{
		Kind:       "TFJob",
		APIVersion: "kubeflow.org/v1",
	}
	XGBoostTypeMeta = metav1.TypeMeta{
		Kind:       "XGBoostJob",
		APIVersion: "kubeflow.org/v1",
	}
	JaxTypeMeta = metav1.TypeMeta{
		Kind:       "JAXJob",
		APIVersion: "kubeflow.org/v1",
	}
)

type DistributedTarget struct {
	Framework workloadsApi.DistributedFramework

	MpiJob     *mpiv2beta1.MPIJob
	PyTorchJob *v2.PyTorchJob
	TFJob      *v2.TFJob
	XgboostJob *v2.XGBoostJob
	JaxJob     *v2.JAXJob

	TypeMeta   *metav1.TypeMeta
	ObjectMeta *metav1.ObjectMeta

	WorkerSpec *v2.ReplicaSpec
	MasterSpec *v2.ReplicaSpec

	RunPolicy *v2.RunPolicy

	Config openapi.WorkloadV2Configuration
}

func (dt *DistributedTarget) Manifest() any {
	switch {
	case dt.MpiJob != nil:
		return dt.MpiJob
	case dt.PyTorchJob != nil:
		return dt.PyTorchJob
	case dt.TFJob != nil:
		return dt.TFJob
	case dt.XgboostJob != nil:
		return dt.XgboostJob
	case dt.JaxJob != nil:
		return dt.JaxJob
	default:
		return nil
	}
}

func NewMinimalDistributedTarget(framework workloadsApi.DistributedFramework, isMaster bool) *DistributedTarget {
	result := DistributedTarget{
		Framework: framework,
	}

	switch framework {
	case workloadsApi.DISTRIBUTEDFRAMEWORK_MPI:
		result.MpiJob = &mpiv2beta1.MPIJob{TypeMeta: MpiTypeMeta}
		result.TypeMeta = &result.MpiJob.TypeMeta
		result.ObjectMeta = &result.MpiJob.ObjectMeta
		result.RunPolicy = &result.MpiJob.Spec.RunPolicy
		result.MpiJob.Spec.MPIReplicaSpecs = make(map[mpiv2beta1.MPIReplicaType]*v2.ReplicaSpec)
		launcherType := mpiv2beta1.MPIReplicaTypeLauncher

		result.WorkerSpec, result.MasterSpec = initWorkerAndMasterReplicas(
			result.MpiJob.Spec.MPIReplicaSpecs, mpiv2beta1.MPIReplicaTypeWorker, &launcherType, isMaster)

	case workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH:
		result.PyTorchJob = &v2.PyTorchJob{TypeMeta: PytorchTypeMeta}
		result.TypeMeta = &result.PyTorchJob.TypeMeta
		result.ObjectMeta = &result.PyTorchJob.ObjectMeta
		result.RunPolicy = &result.PyTorchJob.Spec.RunPolicy
		result.PyTorchJob.Spec.PyTorchReplicaSpecs = make(map[v2.ReplicaType]*v2.ReplicaSpec)

		pytorchMasterType := v2.PyTorchJobReplicaTypeMaster
		result.WorkerSpec, result.MasterSpec = initWorkerAndMasterReplicas(
			result.PyTorchJob.Spec.PyTorchReplicaSpecs, v2.PyTorchJobReplicaTypeWorker, &pytorchMasterType, isMaster)

	case workloadsApi.DISTRIBUTEDFRAMEWORK_TF:
		result.TFJob = &v2.TFJob{TypeMeta: TensorflowTypeMeta}
		result.TypeMeta = &result.TFJob.TypeMeta
		result.ObjectMeta = &result.TFJob.ObjectMeta
		result.RunPolicy = &result.TFJob.Spec.RunPolicy
		result.TFJob.Spec.TFReplicaSpecs = make(map[v2.ReplicaType]*v2.ReplicaSpec)

		tfChiefType := v2.TFJobReplicaTypeChief
		result.WorkerSpec, result.MasterSpec = initWorkerAndMasterReplicas(
			result.TFJob.Spec.TFReplicaSpecs, v2.TFJobReplicaTypeWorker, &tfChiefType, isMaster)

	case workloadsApi.DISTRIBUTEDFRAMEWORK_JAX:
		result.JaxJob = &v2.JAXJob{TypeMeta: JaxTypeMeta}
		result.TypeMeta = &result.JaxJob.TypeMeta
		result.ObjectMeta = &result.JaxJob.ObjectMeta
		result.RunPolicy = &result.JaxJob.Spec.RunPolicy
		result.JaxJob.Spec.JAXReplicaSpecs = make(map[v2.ReplicaType]*v2.ReplicaSpec)

		result.WorkerSpec, _ = initWorkerAndMasterReplicas(
			result.JaxJob.Spec.JAXReplicaSpecs, v2.JAXJobReplicaTypeWorker, nil, isMaster)

	case workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST:
		result.XgboostJob = &v2.XGBoostJob{TypeMeta: XGBoostTypeMeta}
		result.TypeMeta = &result.XgboostJob.TypeMeta
		result.ObjectMeta = &result.XgboostJob.ObjectMeta
		result.RunPolicy = &result.XgboostJob.Spec.RunPolicy
		result.XgboostJob.Spec.XGBReplicaSpecs = make(map[v2.ReplicaType]*v2.ReplicaSpec)

		xgbMasterType := v2.XGBoostJobReplicaTypeMaster
		result.WorkerSpec, result.MasterSpec = initWorkerAndMasterReplicas(
			result.XgboostJob.Spec.XGBReplicaSpecs, v2.XGBoostJobReplicaTypeWorker, &xgbMasterType, isMaster)
	}

	if result.WorkerSpec != nil {
		result.WorkerSpec.Template.Spec.Containers = append(result.WorkerSpec.Template.Spec.Containers, corev1.Container{})
	}

	if result.MasterSpec != nil {
		result.MasterSpec.Template.Spec.Containers = append(result.MasterSpec.Template.Spec.Containers, corev1.Container{})
	}

	return &result
}

// initWorkerAndMasterReplicas populates a Kubeflow replica-spec map with a worker entry
// and, when isMaster is true and masterType is set, a master/launcher/chief entry.
func initWorkerAndMasterReplicas[K comparable](
	specs map[K]*v2.ReplicaSpec,
	workerType K,
	masterType *K,
	isMaster bool,
) (workerSpec, masterSpec *v2.ReplicaSpec) {
	worker := &v2.ReplicaSpec{}
	specs[workerType] = worker
	workerSpec = worker

	if isMaster && masterType != nil {
		masterReplicas := int32(1)
		master := &v2.ReplicaSpec{
			Replicas: &masterReplicas,
		}
		specs[*masterType] = master
		masterSpec = master
	}

	return workerSpec, masterSpec
}

type DistributedToKubeflowComposer struct {
	fc          *composer.FieldsComposer
	ReplicaType *workloadsApi.ReplicaType

	ActualNodePools []clusterApi.Nodepool

	WorkerAssetsMapping []workloadsApi.AssetsMapping
	MasterAssetsMapping []workloadsApi.AssetsMapping

	WorkloadRequirements composer.WorkloadRequirements
	DataVolumeClaimNames map[string]string // maps data volume ID to PVC claim name
}

func NewDistributedToKubeflowComposer(
	workloadType string,
	workloadName string,
	projectName string,
	namespace string,
	requester string,
	workerAssetMapping []workloadsApi.AssetsMapping,
	masterAssetMapping []workloadsApi.AssetsMapping,
	options composer.FieldsComposerOptions,
	actualNodePools []clusterApi.Nodepool,
) *DistributedToKubeflowComposer {
	result := DistributedToKubeflowComposer{
		fc: &composer.FieldsComposer{
			WorkloadType:      workloadType,
			WorkloadName:      workloadName,
			Namespace:         namespace,
			SubmittedBy:       requester,
			ProjectName:       projectName,
			ResourcePrefix:    composer.NetworkingResourcePrefixKubeflow,
			Options:           options,
			ItemizedNumbering: make(itemized.ItemizedNumbering),
			OauthPortNumber:   composer.Oauth2ContainerDefaultPort,
		},
		WorkerAssetsMapping: workerAssetMapping,
		MasterAssetsMapping: masterAssetMapping,
		ActualNodePools:     actualNodePools,
	}

	return &result
}

func (dkc *DistributedToKubeflowComposer) WithStorageClassConfiguration(config map[string]trackerApi.StorageClassConfiguration) *DistributedToKubeflowComposer {
	dkc.fc.StorageClassConfiguration = config
	return dkc
}

func (dkc *DistributedToKubeflowComposer) WithClusterStorageConfig(config composer.ClusterStorageConfig) *DistributedToKubeflowComposer {
	dkc.fc.ClusterStorageConfig = config
	return dkc
}

// KubeflowServiceSelector builds a label selector for Kubeflow workloads.
// Pass this as the cookie parameter to UrlsToWorkloadRunner.
// replicaType is "master", "worker", "launcher", "chief", etc.
// replicaIndex is the replica number (0, 1, 2, ...).
func KubeflowServiceSelector(workloadName string, framework workloadsApi.DistributedFramework, replicaType string, replicaIndex int) map[string]string {
	selector := map[string]string{
		"training.kubeflow.org/job-name": workloadName,
	}

	if framework == workloadsApi.DISTRIBUTEDFRAMEWORK_MPI {
		selector["training.kubeflow.org/job-role"] = replicaType
		if replicaType == "worker" {
			selector["training.kubeflow.org/replica-index"] = strconv.Itoa(replicaIndex)
		}
		return selector
	}

	selector["training.kubeflow.org/replica-type"] = replicaType
	selector["training.kubeflow.org/replica-index"] = strconv.Itoa(replicaIndex)
	return selector
}

type kubeflowNetworkingRole struct {
	// resourceRole is used in resource names and CP correlation metadata.
	resourceRole string
	// selectorReplicaType is used only for Kubernetes Service selectors.
	selectorReplicaType string
}

func kubeflowNetworkingRoleForReplica(framework workloadsApi.DistributedFramework, replicaType workloadsApi.ReplicaType) kubeflowNetworkingRole {
	if replicaType != workloadsApi.REPLICATYPE_MASTER {
		return kubeflowNetworkingRole{
			resourceRole:        "worker",
			selectorReplicaType: "worker",
		}
	}

	switch framework {
	case workloadsApi.DISTRIBUTEDFRAMEWORK_MPI:
		return kubeflowNetworkingRole{
			resourceRole:        "launcher",
			selectorReplicaType: "launcher",
		}
	case workloadsApi.DISTRIBUTEDFRAMEWORK_TF:
		return kubeflowNetworkingRole{
			resourceRole:        "master",
			selectorReplicaType: "chief",
		}
	default:
		return kubeflowNetworkingRole{
			resourceRole:        "master",
			selectorReplicaType: "master",
		}
	}
}

func kubeflowPortsServiceOptions(workloadName string, framework workloadsApi.DistributedFramework, role kubeflowNetworkingRole, replicaIdx int) *composer_v2.PortsServiceOptions {
	return &composer_v2.PortsServiceOptions{
		Selector:      KubeflowServiceSelector(workloadName, framework, role.selectorReplicaType, replicaIdx),
		NameQualifier: role.resourceRole,
	}
}

func (dkc *DistributedToKubeflowComposer) WithDataVolumeClaimNames(claimNames map[string]string) *DistributedToKubeflowComposer {
	dkc.DataVolumeClaimNames = claimNames
	return dkc
}

// Compose function takes care to compose the common fields to all the workload types.
func (dkc *DistributedToKubeflowComposer) Compose(ctx context.Context, src *workloadsApi.DistributedWorkload,
	target *DistributedTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if dkc.fc.WorkloadType != string(workloadsApi.WORKLOADTYPE_DISTRIBUTED) {
			return workloadsApi.NewError(http.StatusNotImplemented, "this composer supports only distributed workloads")
		}

		srcFramework := convutils.Deref(src.Spec.DistributedFramework.Get(), "")
		if srcFramework != target.Framework {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("expecting target of type %s (not '%s')", target.Framework, srcFramework))
		}

		// Framework-level master rules. The target's MasterSpec is nil when the
		// caller built the target with isMaster=false (i.e. noMaster path).
		if target.MasterSpec == nil && frameworkRequiresMaster(target.Framework) {
			return workloadsApi.NewError(http.StatusBadRequest,
				fmt.Sprintf("noMaster is not supported for %s framework", target.Framework))
		}
		if target.MasterSpec != nil && frameworkForbidsMaster(target.Framework) {
			return workloadsApi.NewError(http.StatusBadRequest,
				fmt.Sprintf("master spec is not supported for %s framework", target.Framework))
		}

		segmentedPyTorch := target.Framework == workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH &&
			hasSegmentSizeAnnotation(src.Spec)

		var topologyName, topologyPreferred, topologyRequired *string
		if segmentedPyTorch {
			topologyPreferred = activeNonEmptyAnnotationValue(src.Spec.Annotations, topology.KAITopologyAnnotationPreferredKey)
			topologyRequired = activeNonEmptyAnnotationValue(src.Spec.Annotations, topology.KAITopologyAnnotationRequiredKey)
			// A topology name alone would create an incomplete root constraint and
			// prevent pod-group-assigner from adding its complete system default.
			if topologyPreferred != nil || topologyRequired != nil {
				topologyName = activeNonEmptyAnnotationValue(src.Spec.Annotations, topology.KAITopologyAnnotationKey)
			}
		}

		runner := &composer.StepRunner{}

		runner.Batch(
			dkc.fc.ComposeFullySaasInd(&target.ObjectMeta.Labels, composer.BoolToMapString, composer.ControlPlaneManagedLabel),

			// mark the workload as converted-from-native. some plugins may need to know this
			dkc.fc.CopyBoolean(convutils.Ref(true), &target.ObjectMeta.Labels, composer.BoolToMapString, composer.ConvertedFromNativeLabel),
			dkc.fc.CopyString(convutils.Ref(dkc.fc.SubmittedBy), &target.ObjectMeta.Annotations, composer.StringToMapString, composer.SubmittedByAnnotation),
			// keep the AJR opt-out on the job's own metadata (user annotations otherwise
			// land only on the replica pod templates) so the control plane can read it.
			dkc.fc.CopyString(activeNonEmptyAnnotationValue(src.Spec.Annotations, constants.DisableAJRAnnotation), &target.ObjectMeta.Annotations, composer.StringToMapString, constants.DisableAJRAnnotation),
			// runtime-rules AutoDeleteCompleted reads this from the WorkloadRunner
			// embedded template's top-level metadata (same as IW/TW job metadata).
			dkc.fc.CopyInt(src.Spec.AutoDeletionTimeAfterCompletionSeconds.Get(), &target.ObjectMeta.Annotations, composer.NonZeroIntToMapString, composer.AutoDeleteTimeoutAnnotation),

			dkc.fc.CopyString(&dkc.fc.Namespace, &target.ObjectMeta.Namespace, composer.StringToString, nil),
			dkc.fc.CopyString(&dkc.fc.WorkloadName, &target.ObjectMeta.Name, composer.StringToString, nil),

			dkc.fc.ComposeNameAndTypeLabels(dkc.fc.WorkloadName, dkc.fc.WorkloadType, target.TypeMeta.Kind, dkc.fc.ProjectName, &target.ObjectMeta.Labels, composer.MapToMapString, nil),

			dkc.fc.CopyString(topologyName, &target.ObjectMeta.Annotations, composer.StringToMapString, topology.KAITopologyAnnotationKey),
			dkc.fc.CopyString(topologyPreferred, &target.ObjectMeta.Annotations, composer.StringToMapString, topology.KAITopologyAnnotationPreferredKey),
			dkc.fc.CopyString(topologyRequired, &target.ObjectMeta.Annotations, composer.StringToMapString, topology.KAITopologyAnnotationRequiredKey),
		)

		if runner.Error() != nil {
			return runner.Error()
		}

		// numWorkers is sourced from the worker spec (it's a worker-only field)
		// and threaded into both replica passes so the master pass injects the
		// same value into the launcher's RUNAI_MPI_NUM_WORKERS env var.
		// Passed as a pointer so composeNumWorkersEnv can apply its own
		// "missing -> default 1" handling, keeping that policy in one place.
		workerNumWorkers := src.Spec.NumWorkers.Get()
		// Preemptibility is taken from the worker/root spec for every replica,
		// matching workload-controller BuildAnnotationsAndLabels which always
		// reads MergedWorkload.Spec.Preemptibility (not the master-only fields).
		workerPreemptibility := src.Spec.Preemptibility.Get()

		err := dkc.processReplicaFields(ctx, src.Spec, target, workloadsApi.REPLICATYPE_WORKER, workerNumWorkers, segmentedPyTorch, workerPreemptibility)
		if err != nil {
			return err
		}

		if target.MasterSpec != nil {
			var masterFromWorker *workloadsApi.DistributedSpecSpec

			if convutils.BoolIsTrue(src.MasterSpecSameAsWorker.Get()) || !src.MasterSpec.IsSet() {
				masterFromWorker = src.Spec
				// MPI master specs do not accept GPU resources, so remove fields inherited from the worker.
				if target.Framework == workloadsApi.DISTRIBUTEDFRAMEWORK_MPI {
					masterFromWorker = &workloadsApi.DistributedSpecSpec{}
					copyErr := convutils.CopySpecs(&masterFromWorker, src.Spec)
					if copyErr != nil {
						return workloadsApi.NewError(http.StatusInternalServerError, copyErr.Error())
					}
					clearMpiMasterGpuFields(masterFromWorker.Compute.Get())
				}
			} else {
				masterFromWorker = &workloadsApi.DistributedSpecSpec{}
				copyErr := convutils.CopySpecs(&masterFromWorker, src.MasterSpec)
				if copyErr != nil {
					return workloadsApi.NewError(http.StatusInternalServerError, copyErr.Error())
				}
			}

			// MPI: GPU resources are not allowed on the launcher/master pod.
			// Mirrors the cluster builder's masterResourceGPUCheck.
			if target.Framework == workloadsApi.DISTRIBUTEDFRAMEWORK_MPI && computeHasGpu(masterFromWorker.Compute.Get()) {
				return workloadsApi.NewError(http.StatusBadRequest,
					"GPU resources are not allowed on MPI master pod. Please resubmit without master GPU fields")
			}

			err := dkc.processReplicaFields(ctx, masterFromWorker, target, workloadsApi.REPLICATYPE_MASTER, workerNumWorkers, segmentedPyTorch, workerPreemptibility)
			if err != nil {
				return err
			}
		}

		// Composed after the replica passes so fc.GpuRequested reflects the
		// worker (and master) compute sections before MNNVL defaulting runs.
		runner.Run(dkc.fc.ComposeMNNVL(src.Spec.Labels, dkc.defaultMNNVLMode(), &target.Config, composer.MNNVLToWorkloadRunner, nil))

		return runner.Error()
	}
}

// defaultMNNVLMode derives the MNNVL default for distributed-training workloads.
// A workload that requests no GPU always defaults to unset (no MNNVL) regardless
// of the cluster flag; fc.GpuRequested is set while composing the replica compute.
func (dkc *DistributedToKubeflowComposer) defaultMNNVLMode() string {
	if !dkc.fc.GpuRequested {
		return ""
	}
	config := dkc.fc.ClusterStorageConfig
	if config == nil {
		config = composer.DefaultClusterStorageConfig()
	}
	if config.DistGPUNetworkAcceleration() {
		return string(openapi.MNNVLMODE_REQUIRED)
	}
	return ""
}

// composeContainerName sets the main container's name to the Kubeflow operator
func (dkc *DistributedToKubeflowComposer) composeContainerName(
	framework workloadsApi.DistributedFramework,
	target any, setter composer.FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		name := ""

		switch framework {
		case workloadsApi.DISTRIBUTEDFRAMEWORK_MPI:
			name = "mpi"
		case workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH:
			name = "pytorch"
		case workloadsApi.DISTRIBUTEDFRAMEWORK_TF:
			name = "tensorflow"
		case workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST:
			name = "xgboost"
		case workloadsApi.DISTRIBUTEDFRAMEWORK_JAX:
			name = "jax"
		}

		if name == "" {
			return nil
		}

		return setter(&name, target, cookie)
	}
}

// composeNumWorkersEnv injects RUNAI_MPI_NUM_WORKERS into a container's env
// list. It is framework-aware: for non-MPI workloads it is a no-op (the
// setter is never called), so callers don't need to gate the step themselves.
// When numWorkers is nil it falls back to DefaultNumWorkers, mirroring the
// behavior of composeNumWorkers (which sets the worker Replicas field).
func (dkc *DistributedToKubeflowComposer) composeNumWorkersEnv(numWorkers *int32,
	framework workloadsApi.DistributedFramework,
	target any, setter composer.FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if framework != workloadsApi.DISTRIBUTEDFRAMEWORK_MPI {
			return nil
		}
		value := int32(DefaultNumWorkers)
		if numWorkers != nil {
			value = *numWorkers
		}
		return setter(convutils.Ref(strconv.FormatInt(int64(value), 10)), target, cookie)
	}
}

// composeExposedUrls composes the role's exposed URLs and owns the validations:
//
//  1. A multi-pod role cannot expose authenticated URLs: all pods of a role
//     share a single Kubeflow pod template (and thus a single in-pod
//     oauth2-proxy configuration), so a distinct per-pod URL cannot be
//     expressed. This case is rejected so we don't emit a manifest that
//     authenticates only one of the pods.
//
//  2. The "no-suffix" URL ofer-<workload>.<domain> is owned by the PRIMARY
//     role - the master/launcher when present, otherwise the worker. The
//     non-primary role's URLs are forced to always include their nick suffix
//     (ofer-<workload>-url-worker[-N].<domain>) so the two roles' Ingresses
//     don't compete for the same host. Selectors stay role-specific (Kubeflow
//     replica labels), so each role's Services target only its own pods.
//
// When valid, it delegates to the shared FieldsComposer.ComposeExposedUrls.
func (dkc *DistributedToKubeflowComposer) composeExposedUrls(
	isPrimaryReplica bool, replicaCount int, replicaType workloadsApi.ReplicaType,
	kubeflowRole kubeflowNetworkingRole, framework workloadsApi.DistributedFramework, src []workloadsApi.ExposedUrl, target *DistributedTarget,
	jupyterContainerPort *int32, infoOut **composer.ExposedUrlInfo,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if replicaCount > 1 && len(src) > 0 {
			return workloadsApi.NewError(http.StatusBadRequest,
				"exposedUrls are not supported for a distributed workload with more than one worker: "+
					"all worker pods share a single pod template and cannot each expose a distinct authenticated URL")
		}

		// Force suffix on the non-primary role so its first URL doesn't collide
		// with the primary role's no-suffix host (forceSuffix is per-call state
		// passed through, not stored on FieldsComposer).
		urlSelector := KubeflowServiceSelector(dkc.fc.WorkloadName, framework, kubeflowRole.selectorReplicaType, 0)
		return dkc.fc.ComposeExposedUrls(replicaBasedItemPrefix(itemized.ExposedUrlPrefix, &replicaType),
			src, &target.Config, composer_v2.UrlsToWorkloadRunner, urlSelector, jupyterContainerPort, infoOut,
			!isPrimaryReplica)()
	}
}

func (dkc *DistributedToKubeflowComposer) composePorts(
	replicaCount int, replicaType workloadsApi.ReplicaType, kubeflowRole kubeflowNetworkingRole, framework workloadsApi.DistributedFramework, src []workloadsApi.Port,
	target *DistributedTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		if replicaCount > 1 {
			return workloadsApi.NewError(http.StatusBadRequest,
				"ports are not supported for a distributed workload with more than one worker: "+
					"all worker pods share a single pod template and cannot each expose a distinct port connection")
		}

		for replicaIdx := 0; replicaIdx < replicaCount; replicaIdx++ {
			portOptions := kubeflowPortsServiceOptions(dkc.fc.WorkloadName, framework, kubeflowRole, replicaIdx)

			if err := dkc.fc.ComposePorts(replicaBasedItemPrefix(itemized.PortPrefix, &replicaType),
				src, &target.Config, composer_v2.PortsToWorkloadRunner, portOptions)(); err != nil {
				return err
			}
		}

		return nil
	}
}

func (dkc *DistributedToKubeflowComposer) processReplicaFields(ctx context.Context, src *workloadsApi.DistributedSpecSpec,
	target *DistributedTarget, replicaType workloadsApi.ReplicaType, numWorkers *int32, segmentedPyTorch bool,
	rootPreemptibility *workloadsApi.Preemptibility,
) *workloadsApi.Error {
	var jupyterContainerPort int32
	var exposedUrlInfo *composer.ExposedUrlInfo
	var obsoletePremptibility bool

	var replicaSpec *corev1.PodTemplateSpec
	var assetMapping []workloadsApi.AssetsMapping

	// for backward compatibility each replica starts with port 4180 for the oauth port.
	// it is not mandatory to do so (we could continue from where it stopped) but this
	// keeps the behavior the same as it was
	dkc.fc.ResetOAuthPortNumbering(composer.Oauth2ContainerDefaultPort)

	if replicaType == workloadsApi.REPLICATYPE_MASTER {
		replicaSpec = &target.MasterSpec.Template
		assetMapping = dkc.MasterAssetsMapping
	} else {
		replicaSpec = &target.WorkerSpec.Template
		assetMapping = dkc.WorkerAssetsMapping
	}

	podMeta := &replicaSpec.ObjectMeta
	podSpec := &replicaSpec.Spec
	container := &podSpec.Containers[0]

	envVarsUsingCreds := make([]composer.EnvVarInfo, 0)
	replicaAnnotationsCookie := &kubeflowReplicaAnnotationsCookie{
		framework:        target.Framework,
		replicaType:      replicaType,
		segmentedPyTorch: segmentedPyTorch,
	}

	// Match workload-controller NeedToSetupRunaiHome: createHomeDir OR run-as-user
	// mode implies a managed HOME unless the user already set a custom HOME env.
	createHomeDir := effectiveCreateHomeDir(src.CreateHomeDir.Get(), src.Security.Get(), src.EnvironmentVariables)

	runner := &composer.StepRunner{}

	runner.Batch(
		dkc.fc.ComposeNameAndTypeLabels(dkc.fc.WorkloadName, dkc.fc.WorkloadType, target.TypeMeta.Kind, dkc.fc.ProjectName, &podMeta.Labels, composer.MapToMapString, nil),

		dkc.composeContainerName(target.Framework, &container.Name, composer.StringToString, nil),

		dkc.fc.ComposeLabels(src.Labels, &podMeta.Labels, composer.MapToMapString, nil),
		dkc.fc.ComposeAnnotations(src.Annotations, &podMeta.Annotations, kubeflowReplicaAnnotationsToMapString, replicaAnnotationsCookie),

		// preserve the native pod "user" annotation contract (AnnotationUserName).
		dkc.fc.CopyString(convutils.Ref(dkc.fc.SubmittedBy), &podMeta.Annotations, composer.StringToMapString, constants.AnnotationUserName),

		// assign the workload to the runai scheduler
		dkc.fc.CopyString(convutils.Ref(composer.RunaiScheduler), &podSpec.SchedulerName, composer.StringToString, nil),

		dkc.fc.ComposeImage(src.Image.Get(), &container.Image, composer.StringToString, nil),
		dkc.fc.CopyString(composer.EnumToString(src.ImagePullPolicy.Get()), &container.ImagePullPolicy, composer.StringToString, nil),
		dkc.fc.CopyString(src.Command.Get(), &container.Command, composer.UnixStrToStringArray, nil),
		dkc.fc.CopyString(src.Args.Get(), &container.Args, composer.UnixStrToStringArray, nil),
		dkc.fc.ComposeHomeDirEnvVar(createHomeDir, &container.Env, composer.EnvVarsToK8sEnv, nil),
		dkc.fc.CopyBoolean(src.Tty.Get(), &container.TTY, composer.BoolToBool, nil),
		dkc.fc.CopyBoolean(src.Stdin.Get(), &container.Stdin, composer.BoolToBool, nil),

		// Project user-supplied restartPolicy onto the pod template (no default),
		// matching workload-controller GetRestartPolicy.
		dkc.composeRestartPolicy(src.RestartPolicy.Get(), &podSpec.RestartPolicy, composer.StringToString, nil),
	)

	// Determine replica count and Kubeflow networking role.
	// Master is always a single replica (index 0). Workers can have multiple replicas.
	kubeflowRole := kubeflowNetworkingRoleForReplica(target.Framework, replicaType)
	replicaCount := 1
	if replicaType == workloadsApi.REPLICATYPE_WORKER && numWorkers != nil && *numWorkers > 1 {
		replicaCount = int(*numWorkers)
	}

	// The "primary" replica owns the no-suffix user-facing URL
	// (ofer-<workload>.<domain>): master/launcher when the workload has one,
	// otherwise the worker (e.g. JAX, or noMaster=true). The non-primary role
	// still composes URLs but its hosts always include the nick suffix
	// (ofer-<workload>-url-worker[-N].<domain>) so the two roles' Ingresses
	// don't compete for the same host.
	isPrimaryReplica := replicaType == workloadsApi.REPLICATYPE_MASTER ||
		(replicaType == workloadsApi.REPLICATYPE_WORKER && target.MasterSpec == nil)

	// Exposed URLs and ports are composed once per role. The helpers own the
	// multi-worker validation so we do not emit resources that require
	// per-replica connection correlation in the control plane.
	runner.Batch(
		dkc.composeExposedUrls(isPrimaryReplica, replicaCount, replicaType, kubeflowRole,
			target.Framework, src.ExposedUrls, target, &jupyterContainerPort, &exposedUrlInfo),
		dkc.fc.ApplyExposedUrls(&exposedUrlInfo, podSpec, composer_v2.UrlsToK8sInitContainers, nil),
		dkc.fc.ApplyExposedUrls(&exposedUrlInfo, podSpec, composer.UrlsToK8sContainers, nil),
	)

	runner.Batch(dkc.composePorts(replicaCount, replicaType, kubeflowRole, target.Framework, src.Ports, target))

	var uidGid composer.UidGid
	runner.Run(dkc.processSecurityFields(ctx, src.Security.Get(), replicaSpec, &uidGid))

	rc := composer_v2.ExpandTemplateVarsInCommandAndArgs(container, dkc.fc.ProjectName, dkc.fc.WorkloadName)
	if rc != nil {
		return rc
	}

	runaiVariables := composer.RunaiEnvVars(dkc.fc.ProjectName, dkc.fc.WorkloadName, int(jupyterContainerPort), &uidGid)

	runner.Batch(
		dkc.fc.ComposeEnvironmentVariables(assetMapping, runaiVariables, &container.Env, composer.EnvVarsToK8sEnv, nil, nil),
		dkc.fc.ComposeEnvironmentVariables(assetMapping, src.EnvironmentVariables, &container.Env, composer.EnvVarsToK8sEnv, nil, &envVarsUsingCreds),

		// MPI: inject RUNAI_MPI_NUM_WORKERS into the main container so the
		// launcher and worker know the workload's worker count. The compose
		// function is a no-op for non-MPI frameworks.
		dkc.composeNumWorkersEnv(numWorkers, target.Framework, &container.Env, composer.StringToEnvVar, mpiNumWorkersEnvName),

		dkc.fc.CopyString(src.WorkingDir.Get(), &container.WorkingDir, composer.StringToString, nil),
	)

	runner.Batch(
		dkc.fc.CopyString(src.NodeType.Get(), &podSpec.NodeSelector, composer.StringToMapString, "run.ai/node-type"), //nolint:staticcheck // NodeType is deprecated but still supported for backward compatibility

		dkc.fc.CopyString(src.Category.Get(), &podMeta.Labels, composer.StringToMapString, composer.WorkloadCategoryLabel),

		dkc.fc.ComposePriorityClass(ctx, src.PriorityClass.Get(), &podSpec.PriorityClassName, composer.StringToString,
			&obsoletePremptibility, composer.BoolToBool, nil),
		// Always use the worker/root preemptibility for every replica.
		dkc.fc.ComposePreemptibility(ctx, rootPreemptibility, &podMeta.Labels, composer.PreemptibilityToK8sLabel, nil),

		dkc.processDistributedFields(ctx, src, target, replicaType),
	)

	ipsTarget := composer_v2.PersonalCredsTarget{
		Config:  &target.Config,
		PodSpec: podSpec,
	}
	ipsCookie := composer_v2.PersonalCredsCookie{
		WorkloadName: dkc.fc.WorkloadName,
	}
	runner.Run(dkc.fc.ComposeImagePullSecrets(assetMapping, src.ImagePullSecrets, &ipsTarget, composer_v2.ArrayToImagePullSecretsSetter, &ipsCookie))

	// sizingInfo is populated by processComputeFields and read immediately after by
	// ComposeGpuDerivedFrameworkFields (for MPI SlotsPerWorker / PyTorch
	// NprocPerNode). It also feeds annotations and env vars from within
	// processComputeFields itself.
	var sizingInfo composer.GpuSizingInfo

	runner.Batch(
		dkc.fc.ComposeTerminationGracePeriod(ctx, src.TerminationGracePeriodSeconds.Get(), &podSpec.TerminationGracePeriodSeconds, composer.NonZeroIntToInt, nil),

		dkc.processComputeFields(ctx, src.Compute.Get(), replicaSpec, &sizingInfo),
		dkc.processStorageFields(ctx, src.Storage.Get(), uidGid, createHomeDir, getLargeShmRequest(src.Compute.Get()), replicaSpec, &target.Config, replicaType),
		dkc.processAffinityFields(ctx, src, replicaSpec),
	)

	if composer.IsMainNode(&replicaType) && target.MpiJob != nil {
		runner.Batch(
			dkc.composeSlotsPerWorker(src.SlotsPerWorker.Get(), &sizingInfo, &target.MpiJob.Spec.SlotsPerWorker, composer.IntToInt, nil),
		)
	}

	if composer.IsMainNode(&replicaType) && target.PyTorchJob != nil {
		runner.Batch(
			dkc.composeNprocPerNode(&sizingInfo, &target.PyTorchJob.Spec.NprocPerNode, composer.StringToString, nil),
		)
	}

	targetForEnvVarsUsingCreds := composer_v2.EnvVarUsingCredTarget{
		Config:  &target.Config,
		EnvVars: &container.Env,
	}
	cookieForEnvVarsUsingCreds := composer_v2.PersonalCredsCookie{
		WorkloadName: dkc.fc.WorkloadName,
	}

	runner.Run(dkc.fc.ComposeEnvVarsUsingCreds(envVarsUsingCreds, &targetForEnvVarsUsingCreds, composer_v2.EnvVarsUsingCredSetter, &cookieForEnvVarsUsingCreds))

	runner.Run(dkc.fc.ComposeTolerations(replicaBasedItemPrefix(itemized.TolerationPrefix, nil), src.Tolerations, &podSpec.Tolerations, composer.MapToK8sTolerations, nil))

	if convutils.HasValue(src.Probes) {
		runner.Run(dkc.fc.ComposeReadinessProbe(src.Probes.Get().Readiness.Get(), &container.ReadinessProbe, composer.ProbeToK8s, nil))
	}

	return runner.Error()
}

type kubeflowReplicaAnnotationsCookie struct {
	framework        workloadsApi.DistributedFramework
	replicaType      workloadsApi.ReplicaType
	segmentedPyTorch bool
}

// kubeflowReplicaAnnotationsToMapString keeps the topology name and segment
// constraints on the Worker PodTemplate. Explicit workload-level placement is
// projected separately onto the PyTorchJob. ComposeAnnotations has already
// resolved duplicate and excluded annotations before this setter is called, so
// this setter only owns the Kubeflow target-specific projection.
func kubeflowReplicaAnnotationsToMapString(src map[string]*string, target any, rawCookie any) *workloadsApi.Error {
	cookie, ok := rawCookie.(*kubeflowReplicaAnnotationsCookie)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("kubeflowReplicaAnnotationsToMapString: cookie type %T is not *kubeflowReplicaAnnotationsCookie", rawCookie))
	}

	if !cookie.segmentedPyTorch || cookie.framework != workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH {
		return composer.MapToMapString(src, target, nil)
	}

	filtered := make(map[string]*string, len(src))
	for name, value := range src {
		switch name {
		case topology.KAITopologyAnnotationPreferredKey,
			topology.KAITopologyAnnotationRequiredKey:
			continue
		case topology.KAITopologyAnnotationKey,
			topology.KAISegmentSizeAnnotationKey,
			topology.KAISegmentTopologyPreferredAnnotationKey,
			topology.KAISegmentTopologyRequiredAnnotationKey:
			if cookie.replicaType == workloadsApi.REPLICATYPE_MASTER {
				continue
			}
		}
		filtered[name] = value
	}

	return composer.MapToMapString(filtered, target, nil)
}

func hasSegmentSizeAnnotation(src *workloadsApi.DistributedSpecSpec) bool {
	if src == nil {
		return false
	}
	segmentSize, found := topology.ActiveAnnotationValue(src.Annotations, topology.KAISegmentSizeAnnotationKey)
	return found && segmentSize != ""
}

func activeNonEmptyAnnotationValue(annotations []workloadsApi.Annotation, name string) *string {
	value, found := topology.ActiveAnnotationValue(annotations, name)
	if !found || value == "" {
		return nil
	}
	return &value
}

func (dkc *DistributedToKubeflowComposer) processComputeFields(ctx context.Context, src *workloadsApi.WorkspaceSpecSpecAllOfCompute,
	replicaSpec *corev1.PodTemplateSpec, sizingInfo *composer.GpuSizingInfo,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}
		canonical, err := gpucompat.NormalizeLegacyWorkloadCompute(src)
		if err != nil {
			return workloadsApi.NewError(http.StatusBadRequest, err.Error())
		}
		src = canonical

		podMeta := &replicaSpec.ObjectMeta
		podSpec := &replicaSpec.Spec
		container := &podSpec.Containers[0]

		runner := &composer.StepRunner{}

		runner.Batch(
			dkc.fc.ComposeCpuCoreRequest(src.CpuCoreRequest.Get(), &container.Resources.Requests, composer.FloatToK8sResourceList, composer.ResourceCpu),
			dkc.fc.ComposeCpuCoreLimit(src.CpuCoreLimit.Get(), src.CpuCoreRequest.Get(), &container.Resources.Limits, composer.FloatToK8sResourceList, composer.ResourceCpu),
			dkc.fc.ComposeCpuMemoryRequest(src.CpuMemoryRequest.Get(), &container.Resources.Requests, composer.StringToK8sResourceList, composer.ResourceMemory),
			dkc.fc.ComposeCpuMemoryLimit(src.CpuMemoryLimit.Get(), src.CpuMemoryRequest.Get(), &container.Resources.Limits, composer.StringToK8sResourceList, composer.ResourceMemory),

			dkc.fc.ComposeGpuRequestsAndLimits(ctx, src.GpuDevicesRequest.Get(), src.GpuRequestType.Get(),
				src.GpuMemoryPortionRequest.Get(), src.GpuMemoryPortionLimit.Get(),
				src.GpuMemoryRequest.Get(), src.GpuMemoryLimit.Get(), composer.WholeGpusToK8sResources,
				&container.Resources, nil, sizingInfo),

			dkc.fc.ComposeExtendedResources(replicaBasedItemPrefix(itemized.ExtendedResourceNickPrefix, nil), src.ExtendedResources,
				&container.Resources, composer.ExtendedResourcesToK8sResources, nil),
		)

		if sizingInfo.MemoryLimit != nil || sizingInfo.PortionLimit != nil {
			dkc.WorkloadRequirements.GpuOptimization = true
		}

		runner.Batch(
			// GPU memory/portion limit is written either as the legacy env var (< 3.0
			// clusters) or as the per-container CUDA/KAI annotations (>= 3.0). The
			// two pairs are mutually exclusive at the compose helper level.
			dkc.fc.ComposeGpuComputeMode(composer.EnumToString(src.GpuComputeMode.Get()), sizingInfo, container.Name, &podMeta.Annotations, composer.StringToMapString),
			dkc.fc.ComposeMemoryLimitEnv(sizingInfo, &container.Env, composer.StringToEnvVar, composer.GpuMemoryLimitEnvVar),
			dkc.fc.ComposePortionLimitEnv(sizingInfo, &container.Env, composer.StringToEnvVar, composer.GpuPortionLimitEnvVar),
			dkc.fc.ComposeMemoryLimitAnnotation(sizingInfo, container.Name, &podMeta.Annotations, composer.StringToMapString),
			dkc.fc.ComposePortionLimitAnnotation(sizingInfo, container.Name, &podMeta.Annotations, composer.StringToMapString),
			dkc.fc.ComposeMemoryReqAnnotation(sizingInfo, &podMeta.Annotations, composer.StringToMapString, composer.GpuMemoryAnnotation),
			dkc.fc.ComposePortionReqAnnotation(sizingInfo, &podMeta.Annotations, composer.StringToMapString, composer.GpuFractionAnnotation),
			dkc.fc.ComposeNumDevicesAnnotation(sizingInfo, &podMeta.Annotations, composer.StringToMapString, composer.GpuFractionDevicesAnnotation),
		)

		return runner.Error()
	}
}

// composeSlotsPerWorker sets MPIJob.spec.SlotsPerWorker, using the explicit
// value when provided and otherwise falling back to the post-process GPU sizing
// (whole-GPU portion >=1 first, then gpuDevices >=1). Mirrors the cluster MPI
// builder's behavior. No-op for non-MPI workloads or when no value applies.
func (dkc *DistributedToKubeflowComposer) composeSlotsPerWorker(
	src *int32, sizingInfo *composer.GpuSizingInfo, target **int32, setter composer.FnIntSetter[int32], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		value := src
		if value == nil && sizingInfo != nil {
			switch {
			case sizingInfo.PortionRequest != nil && *sizingInfo.PortionRequest >= 1:
				value = convutils.Ref(int32(*sizingInfo.PortionRequest))
			case sizingInfo.NumDevices != nil && *sizingInfo.NumDevices >= 1:
				value = sizingInfo.NumDevices
			}
		}
		if value == nil {
			// No explicit value and no whole-GPU fallback; leave the field unset.
			// (IntToInt would otherwise dereference to 0 and write SlotsPerWorker: 0.)
			return nil
		}
		return setter(value, target, cookie)
	}
}

// composeNprocPerNode sets PyTorchJob.spec.NprocPerNode from the post-process
// GPU sizing (whole-GPU portion >=1 first, then gpuDevices >=1). Mirrors the
// cluster PyTorch builder. No-op for non-PyTorch workloads or when no whole-GPU
// request is present on the worker.
func (dkc *DistributedToKubeflowComposer) composeNprocPerNode(
	sizingInfo *composer.GpuSizingInfo, target **string, setter composer.FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if sizingInfo == nil {
			return nil
		}

		var value *string
		switch {
		case sizingInfo.PortionRequest != nil && *sizingInfo.PortionRequest >= 1:
			value = convutils.Ref(strconv.FormatFloat(*sizingInfo.PortionRequest, 'f', -1, 64))
		case sizingInfo.NumDevices != nil && *sizingInfo.NumDevices >= 1:
			value = convutils.Ref(strconv.FormatInt(int64(*sizingInfo.NumDevices), 10))
		}

		return setter(value, target, cookie)
	}
}

func (dkc *DistributedToKubeflowComposer) processSecurityFields(_ context.Context, src *workloadsApi.WorkspaceSpecSpecAllOfSecurity,
	replicaSpec *corev1.PodTemplateSpec, uidGid *composer.UidGid,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		podSpec := &replicaSpec.Spec
		container := &podSpec.Containers[0]

		if container.SecurityContext == nil {
			container.SecurityContext = &corev1.SecurityContext{}
		}

		var seccompSrc *workloadsApi.SeccompProfileType
		if src != nil {
			seccompSrc = src.SeccompProfileType.Get()
		}

		runner := &composer.StepRunner{}

		// Always set seccomp (default RuntimeDefault) even when security is unset.
		runner.Batch(
			dkc.fc.ComposeSeccompProfile(seccompSrc, convutils.Ref(constants.SeccompProfileTypeRuntimeDefault),
				&container.SecurityContext.SeccompProfile, composer.StringToSeccompProfile, nil),
		)

		if src == nil {
			return runner.Error()
		}

		if podSpec.SecurityContext == nil {
			podSpec.SecurityContext = &corev1.PodSecurityContext{}
		}

		runner.Batch(
			dkc.fc.CopyBoolean(src.HostIpc.Get(), &podSpec.HostIPC, composer.BoolToBool, nil),
			dkc.fc.CopyBoolean(src.HostNetwork.Get(), &podSpec.HostNetwork, composer.BoolToBool, nil),
			dkc.fc.ComposeDnsPolicy(src.HostNetwork.Get(), &podSpec.DNSPolicy, composer.StringToString, nil),
			dkc.fc.CopyBoolean(src.Privileged.Get(), &container.SecurityContext.Privileged, composer.BoolToBool, nil),
			dkc.fc.CopyBoolean(src.ReadOnlyRootFilesystem.Get(), &container.SecurityContext.ReadOnlyRootFilesystem, composer.BoolToBool, nil),
			dkc.fc.ComposeAllowPrivilegeEscalation(src.AllowPrivilegeEscalation.Get(), src.Privileged.Get(), &container.SecurityContext.AllowPrivilegeEscalation, composer.BoolToBool, nil),
			dkc.fc.ComposeRunAsNonRoot(src.RunAsNonRoot.Get(), src.Privileged.Get(), &container.SecurityContext.RunAsNonRoot, composer.BoolToBool, nil),
			dkc.fc.ComposeCapabilities(src.Capabilities, &container.SecurityContext.Capabilities, composer.MapToK8sCapabilities, nil),
		)

		var runAsUser bool
		runner.Run(
			dkc.fc.ComposeRunAsUser(composer.EnumToString(src.UidGidSource.Get()), &runAsUser, composer.BoolToBool, nil),
		)

		if runAsUser {
			runner.Batch(
				dkc.fc.CopyInt(src.RunAsUid.Get(), &container.SecurityContext.RunAsUser, composer.IntToInt, nil),
				dkc.fc.CopyInt(src.RunAsGid.Get(), &container.SecurityContext.RunAsGroup, composer.IntToInt, nil),
				dkc.fc.ComposeSupplementalGroups(src.SupplementalGroups.Get(), &podSpec.SecurityContext.SupplementalGroups, composer.StringToSupplementalGroups, nil),
				dkc.fc.CopyInt(src.RunAsGid.Get(), &podSpec.SecurityContext.FSGroup, composer.IntToInt, nil),
			)

			uidGid.Uid = src.RunAsUid.Get()
			uidGid.Gid = src.RunAsGid.Get()
		}

		if composer.IsEmpty(*podSpec.SecurityContext) {
			podSpec.SecurityContext = nil
		}

		return runner.Error()
	}
}

func (dkc *DistributedToKubeflowComposer) processStorageFields(_ context.Context, src *workloadsApi.WorkspaceSpecSpecAllOfStorage,
	uidGid composer.UidGid, createHomeDir *bool, largeShmRequest *bool,
	replicaSpec *corev1.PodTemplateSpec, config *openapi.WorkloadV2Configuration, replicaType workloadsApi.ReplicaType,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil && convutils.BoolIsFalse(createHomeDir) && convutils.BoolIsFalse(largeShmRequest) {
			return nil
		}

		var mountPropagation workloadsApi.HostPathMountPropagation

		podSpec := &replicaSpec.Spec
		container := &podSpec.Containers[0]

		volumesAndMounts := composer.VolumesAndMounts{
			Config: config,
		}

		runner := &composer.StepRunner{}

		if src != nil {
			runner.Batch(
				dkc.fc.ComposeHostPath(replicaBasedItemPrefix(itemized.HostPathPrefix, &replicaType), src.HostPath, &volumesAndMounts, composer.MapHostPathToK8sVol, nil, &mountPropagation),
				dkc.fc.ComposeNfs(replicaBasedItemPrefix(itemized.NfsPrefix, &replicaType), src.Nfs, &volumesAndMounts, composer.MapNfsToK8sVol, nil),
				dkc.fc.ComposeConfigMap(replicaBasedItemPrefix(itemized.ConfigMapPrefix, &replicaType), src.ConfigMapVolume, &volumesAndMounts, composer.MapConfigMapToK8sVol, nil),
				dkc.fc.ComposeEmptyDir(replicaBasedItemPrefix(itemized.EmptyDirPrefix, &replicaType), src.EmptyDirVolume, &volumesAndMounts, composer.MapEmptyDirToK8sVol, nil),
				dkc.fc.ComposeSecretVolume(replicaBasedItemPrefix(itemized.SecretStoragePrefix, &replicaType), src.SecretVolume, &volumesAndMounts, composer.MapSecretToK8sVol, nil),

				dkc.fc.ComposeGit(replicaBasedItemPrefix(itemized.GitPrefix, &replicaType), src.Git, podSpec, composer.MapGitToWorkloadK8s, nil),
				dkc.fc.ComposePvc(replicaBasedItemPrefix(itemized.PvcPrefix, &replicaType), src.Pvc, &volumesAndMounts, composer_v2.MapPvcToWorkloadRunner, nil),
				dkc.fc.ComposeS3(replicaBasedItemPrefix(itemized.S3Prefix, &replicaType), src.S3, uidGid, podSpec, composer_v2.MapS3ToWorkloadK8s, nil),
				dkc.fc.ComposeDataVolume(replicaBasedItemPrefix(itemized.DataVolumePrefix, &replicaType), src.DataVolume, &volumesAndMounts, composer.MapDataVolumesToK8s, dkc.DataVolumeClaimNames),
			)
		}

		runner.Batch(
			dkc.fc.ComposeHomeDirectory(createHomeDir, &volumesAndMounts, composer.MapEmptyDirToK8sVol, nil),
			dkc.fc.ComposeLargeShm(largeShmRequest, &volumesAndMounts, composer.MapEmptyDirToK8sVol, nil),
			dkc.fc.CopyBoolean(dkc.fc.Options.CleanupCreatedPvcsOnCompletion, volumesAndMounts.Config, composer_v2.BoolToPvcsCleanupConfig, nil),
		)

		if len(volumesAndMounts.VolumeMount) != 0 {
			container.VolumeMounts = append(container.VolumeMounts, volumesAndMounts.VolumeMount...)
		}

		if len(volumesAndMounts.Volume) != 0 {
			podSpec.Volumes = append(podSpec.Volumes, volumesAndMounts.Volume...)
		}

		return runner.Error()
	}
}

func (dkc *DistributedToKubeflowComposer) processAffinityFields(_ context.Context, src *workloadsApi.DistributedSpecSpec,
	replicaSpec *corev1.PodTemplateSpec,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		runner := &composer.StepRunner{}

		podSpec := &replicaSpec.Spec

		if podSpec.Affinity == nil {
			podSpec.Affinity = &corev1.Affinity{}
		}

		nodePoolsInfo := &composer.NodePoolsInfo{
			RequestedNodePools: src.NodePools,
			ActualNodePools:    dkc.ActualNodePools,
		}

		runner.Batch(
			dkc.fc.ComposeNodeAffinity(src.NodeAffinityRequired.Get(), &podSpec.Affinity.NodeAffinity, composer.NodeAffinityRequiredToK8s, nil),
			dkc.fc.ComposePodAffinity(src.PodAffinity.Get(), &podSpec.Affinity.PodAffinity, composer.PodAffinityToK8s, dkc.fc.WorkloadName),
			dkc.fc.ComposeNodePools(nodePoolsInfo, &podSpec.Affinity.NodeAffinity, composer.NodePoolsToK8s, nil),
			// Match workload-controller BuildAnnotationsAndLabels: stamp runai-nodepools
			// on the pod template whenever node pools were requested.
			dkc.fc.CopyString(nodePoolsAnnotationValue(src.NodePools), &replicaSpec.Annotations, composer.StringToMapString, commonconstants.AnnotationNodepools),
		)

		if composer.IsEmpty(*podSpec.Affinity) {
			podSpec.Affinity = nil
		}

		return runner.Error()
	}
}

func (dkc *DistributedToKubeflowComposer) processDistributedFields(_ context.Context, src *workloadsApi.DistributedSpecSpec,
	target *DistributedTarget, replicaType workloadsApi.ReplicaType,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		// All distributed-only fields are worker-only (see the design).
		// Skip the entire block on the master pass so we don't overwrite
		// worker-set values with the master's (often empty) input.
		if !composer.IsMainNode(&replicaType) {
			return nil
		}

		if target.RunPolicy == nil {
			return workloadsApi.NewError(http.StatusInternalServerError, "distributed target has no kubeflow job")
		}

		runner := &composer.StepRunner{}
		runner.Batch(
			dkc.composeNumWorkers(src.NumWorkers.Get(), &target.WorkerSpec.Replicas, composer.IntToInt, nil),
			dkc.composeBackoffLimit(src.BackoffLimit.Get(), src.NumWorkers.Get(), &target.RunPolicy.BackoffLimit, composer.IntToInt, nil),
			dkc.composeCleanPodPolicy(src.CleanPodPolicy.Get(), target.Framework, target.RunPolicy),
			dkc.fc.CopyInt(src.AutoDeletionTimeAfterCompletionSeconds.Get(), &target.RunPolicy.TTLSecondsAfterFinished, composer.IntToInt, nil),
		)

		if target.MpiJob != nil {
			runner.Batch(
				dkc.composeMpiLauncherCreationPolicy(src.MpiLauncherCreationPolicy.Get(), &target.MpiJob.Spec.LauncherCreationPolicy, mpiLauncherCreationPolicySetter, nil),
				dkc.fc.CopyString(src.SshAuthMountPath.Get(), &target.MpiJob.Spec.SSHAuthMountPath, composer.StringToString, nil),
			)
		} else {
			runner.Batch(
				dkc.frameworkSpecific(src.SlotsPerWorker.Get(), "slotsPerWorker", target.Framework, workloadsApi.DISTRIBUTEDFRAMEWORK_MPI),
				dkc.frameworkSpecific(src.MpiLauncherCreationPolicy.Get(), "mpiLauncherCreationPolicy", target.Framework, workloadsApi.DISTRIBUTEDFRAMEWORK_MPI),
				dkc.frameworkSpecific(src.SshAuthMountPath.Get(), "sshAuthMountPath", target.Framework, workloadsApi.DISTRIBUTEDFRAMEWORK_MPI),
			)
		}

		if target.PyTorchJob != nil {
			runner.Run(dkc.composeElasticPolicy(src.MinReplicas.Get(), src.MaxReplicas.Get(), &target.PyTorchJob.Spec.ElasticPolicy, elasticMinMaxSetter, nil))
		} else {
			runner.Batch(
				dkc.frameworkSpecific(src.MinReplicas.Get(), "minReplicas", target.Framework, workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH),
				dkc.frameworkSpecific(src.MaxReplicas.Get(), "maxReplicas", target.Framework, workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH),
			)
		}

		return runner.Error()
	}
}

// composeBackoffLimit sets runPolicy.backoffLimit. When unset, defaults to
// numWorkers * DefaultBackoffLimit (6), matching workload-controller
// BuildRunPolicy. Unset numWorkers is treated as 0 (same as Workers.Deref32()).
func (dkc *DistributedToKubeflowComposer) composeBackoffLimit(backoffLimit *int64, numWorkers *int32,
	target any, setter composer.FnIntSetter[int64], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if backoffLimit != nil {
			return setter(backoffLimit, target, cookie)
		}

		var workers int64
		if numWorkers != nil {
			workers = int64(*numWorkers)
		}
		val := workers * int64(constants.DefaultBackoffLimit)
		return setter(&val, target, cookie)
	}
}

var validDistributedRestartPolicies = datastructs.NewSet(
	string(workloadsApi.RESTARTPOLICY_ALWAYS),
	string(workloadsApi.RESTARTPOLICY_NEVER),
	string(workloadsApi.RESTARTPOLICY_ON_FAILURE),
)

// composeRestartPolicy projects a user-supplied restartPolicy onto the pod
// template. Unlike IW/TW jobs, distributed leaves the field unset when omitted
// so Kubernetes applies its default (Always) — matching workload-controller
// GetRestartPolicy.
func (dkc *DistributedToKubeflowComposer) composeRestartPolicy(src *workloadsApi.RestartPolicy,
	target any, setter composer.FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}
		val := string(*src)
		return dkc.fc.ComposeRestartPolicy(validDistributedRestartPolicies, &val, target, setter, cookie)()
	}
}

func (dkc *DistributedToKubeflowComposer) composeCleanPodPolicy(src *workloadsApi.CleanPodPolicy,
	framework workloadsApi.DistributedFramework, runPolicy *v2.RunPolicy,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		// When the user does not specify a cleanPodPolicy, fall back to the
		// per-framework default. Mirrors the cluster builder's
		// GetDefaultCleanPodPolicy so the operator sees the same value
		// regardless of which path produced the CRD. The target itself is
		// only ever touched by cleanPodPolicySetter.
		value := composer.EnumToString(src)
		if value == nil {
			value = convutils.Ref(string(defaultCleanPodPolicy(framework)))
		}
		return cleanPodPolicySetter(value, &runPolicy.CleanPodPolicy, nil)
	}
}

func (dkc *DistributedToKubeflowComposer) composeMpiLauncherCreationPolicy(src *workloadsApi.MpiLauncherCreationPolicy,
	target any, setter composer.FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		if !dkc.fc.Options.LauncherCreationPolicySupported {
			return workloadsApi.NewError(http.StatusBadRequest, "mpiLauncherCreationPolicy is not supported in this cluster")
		}

		return setter(composer.EnumToString(src), target, cookie)
	}
}

func (dkc *DistributedToKubeflowComposer) composeElasticPolicy(minReplicas, maxReplicas *int32,
	target **v2.ElasticPolicy, setter composer.FnMinMaxIntSetter[int32], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if minReplicas == nil && maxReplicas == nil {
			return nil
		}

		if minReplicas != nil && maxReplicas != nil && *maxReplicas < *minReplicas {
			return workloadsApi.NewError(http.StatusBadRequest, "maxReplicas should be no less than minReplicas")
		}

		pair := &composer.MinMaxPair[int32]{
			Min: minReplicas,
			Max: maxReplicas,
		}
		return setter(pair, target, cookie)
	}
}

func elasticMinMaxSetter(pair *composer.MinMaxPair[int32], target any, _ any) *workloadsApi.Error {
	if pair == nil || (pair.Min == nil && pair.Max == nil) {
		return nil
	}

	targetPtr, ok := target.(**v2.ElasticPolicy)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("elasticMinMaxSetter: target type %T is not **ElasticPolicy", target))
	}

	if *targetPtr == nil {
		*targetPtr = &v2.ElasticPolicy{}
	}

	if pair.Min != nil {
		(*targetPtr).MinReplicas = pair.Min
	}
	if pair.Max != nil {
		(*targetPtr).MaxReplicas = pair.Max
	}

	return nil
}

func cleanPodPolicySetter(src *string, target any, _ any) *workloadsApi.Error {
	if src == nil || *src == "" {
		return nil
	}

	targetPtr, ok := target.(**v2.CleanPodPolicy)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("cleanPodPolicySetter: target type %T is not **CleanPodPolicy", target))
	}

	policy := v2.CleanPodPolicy(*src)
	*targetPtr = &policy

	return nil
}

func mpiLauncherCreationPolicySetter(src *string, target any, _ any) *workloadsApi.Error {
	if src == nil || *src == "" {
		return nil
	}

	targetPtr, ok := target.(*mpiv2beta1.LauncherCreationPolicy)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("mpiLauncherCreationPolicySetter: target type %T is not *LauncherCreationPolicy", target))
	}

	*targetPtr = mpiv2beta1.LauncherCreationPolicy(*src)

	return nil
}

func (dkc *DistributedToKubeflowComposer) composeNumWorkers(src *int32,
	target any, setter composer.FnIntSetter[int32], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			defaultWorkers := int32(DefaultNumWorkers)
			src = &defaultWorkers
		}
		return setter(src, target, cookie)
	}
}

// validate that a framework-specific value is given only when composing the correct framework
func (dkc *DistributedToKubeflowComposer) frameworkSpecific(src any,
	field string, framework workloadsApi.DistributedFramework, expectedFramework workloadsApi.DistributedFramework,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if !reflect2.IsNil(src) && framework != expectedFramework {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s is not supported by %s framework", field, framework))
		}

		return nil
	}
}

// For interactive and training we do not have replicas, so just leave the prefix as is.
func replicaBasedItemPrefix(prefix string, replicaType *workloadsApi.ReplicaType) string {
	if replicaType == nil {
		return prefix
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ToLower(string(*replicaType)))
}

// defaultCleanPodPolicy mirrors the cluster builder's GetDefaultCleanPodPolicy
// behavior: MPI and TF default to Running, all other frameworks default to None.
func defaultCleanPodPolicy(framework workloadsApi.DistributedFramework) v2.CleanPodPolicy {
	switch framework {
	case workloadsApi.DISTRIBUTEDFRAMEWORK_MPI, workloadsApi.DISTRIBUTEDFRAMEWORK_TF:
		return v2.CleanPodPolicyRunning
	default:
		return v2.CleanPodPolicyNone
	}
}

// frameworkRequiresMaster reports whether the framework rejects a workers-only
// (noMaster) configuration. MPI and XGBoost always need a master/launcher.
func frameworkRequiresMaster(framework workloadsApi.DistributedFramework) bool {
	return framework == workloadsApi.DISTRIBUTEDFRAMEWORK_MPI ||
		framework == workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST
}

// frameworkForbidsMaster reports whether the framework rejects a master spec
// entirely. Currently only JAX (workers-only).
func frameworkForbidsMaster(framework workloadsApi.DistributedFramework) bool {
	return framework == workloadsApi.DISTRIBUTEDFRAMEWORK_JAX
}

func effectiveCreateHomeDir(createHomeDir *bool, security *workloadsApi.WorkspaceSpecSpecAllOfSecurity,
	envs []workloadsApi.EnvironmentVariable,
) *bool {
	runAsUser := false
	if security != nil {
		uidGid := security.UidGidSource.Get()
		if uidGid != nil && *uidGid != workloadsApi.UIDGIDSOURCE_FROM_THE_IMAGE {
			runAsUser = true
		}
	}

	shouldCreateHomeDir := runAsUser
	// RunAsUser historically implies a managed home directory.
	// Fall back to RunAsUser only when CreateHomeDir is unset;
	// an explicit true or false value takes precedence.
	if createHomeDir != nil {
		shouldCreateHomeDir = *createHomeDir
	}

	if !shouldCreateHomeDir {
		return createHomeDir
	}

	for _, env := range envs {
		name := env.GetName()
		if name == composer.HomeEnvVarName {
			if env.GetValue() == composer.RunaiDefaultHomeDir {
				return convutils.Ref(true)
			}
			return convutils.Ref(false)
		}
	}

	return convutils.Ref(true)
}

// nodePoolsAnnotationValue joins requested node pools the same way
// field_groups.GetNodepoolsValue does for the runai-nodepools annotation.
func nodePoolsAnnotationValue(pools []string) *string {
	if len(pools) == 0 {
		return nil
	}
	return convutils.Ref(strings.Join(pools, " "))
}

func getLargeShmRequest(compute *workloadsApi.WorkspaceSpecSpecAllOfCompute) *bool {
	if compute == nil {
		return nil
	}
	return compute.LargeShmRequest.Get()
}

func clearMpiMasterGpuFields(compute *workloadsApi.WorkspaceSpecSpecAllOfCompute) {
	if compute == nil {
		return
	}
	compute.GpuDevicesRequest.Unset()
	compute.GpuRequestType.Unset()
	compute.GpuPortionRequest.Unset()
	compute.GpuPortionLimit.Unset()
	compute.GpuMemoryPortionRequest.Unset()
	compute.GpuMemoryPortionLimit.Unset()
	compute.GpuMemoryRequest.Unset()
	compute.GpuMemoryLimit.Unset()
}

// computeHasGpu reports whether any GPU is requested on the given compute
// fields. Mirrors the cluster MPI builder's masterResourceGPUCheck. Reads the
// raw input fields because it is used as pre-compose validation (before the
// per-replica compute composition runs).
func computeHasGpu(compute *workloadsApi.WorkspaceSpecSpecAllOfCompute) bool {
	if compute == nil {
		return false
	}
	if devices := compute.GpuDevicesRequest.Get(); devices != nil && *devices > 0 {
		return true
	}
	if portion := compute.GpuPortionRequest.Get(); portion != nil && *portion > 0 {
		return true
	}
	if portionLimit := compute.GpuPortionLimit.Get(); portionLimit != nil && *portionLimit > 0 {
		return true
	}
	if portion := compute.GpuMemoryPortionRequest.Get(); portion != nil && *portion > 0 {
		return true
	}
	if limit := compute.GpuMemoryPortionLimit.Get(); limit != nil && *limit > 0 {
		return true
	}
	return false
}
