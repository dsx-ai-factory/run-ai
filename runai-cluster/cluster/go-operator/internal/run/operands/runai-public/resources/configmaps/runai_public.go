/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configmaps

import (
	"context"
	"fmt"
	"regexp"
	_ "regexp"
	"slices"
	_ "strconv"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/version"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/consts"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"

	kube_utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/common/utils/kubernetes"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"

	"go.uber.org/multierr"

	"github.com/aws/amazon-ssm-agent/agent/versionutil"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/knative"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/prometheus/resources"
	controller_utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	kfv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	mpiJobKind                     = "MPIJob"
	kubeflowMPIV2Beta1GroupVersion = "kubeflow.org/v2beta1"

	name      = runv1.RunaiPublicConfigMapName
	namespace = runv1.DefaultNamespace
)

const gatewayAPIGroupVersion = "gateway.networking.k8s.io/v1"

func NewRunaiPublic(operandName string, kbClient kc.HybridClient, runaiConfig runv1.RunaiConfig) *RunaiPublic {
	return &RunaiPublic{
		operandName: operandName,
		kbClient:    kbClient,
		runaiConfig: &runaiConfig,
	}
}

type RunaiPublic struct {
	operandName string
	kbClient    kc.HybridClient
	runaiConfig *runv1.RunaiConfig
	kubeVersion string
}

func (rp *RunaiPublic) OperandName() string {
	return rp.operandName
}

func (rp *RunaiPublic) Object(ctx context.Context) (client.Object, error) {
	logger := log.FromContext(ctx)
	discoveryClient, err := rp.kbClient.NewDiscoveryClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	configMap := &corev1.ConfigMap{}

	err = rp.kbClient.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, configMap)
	if err != nil && !kerrors.IsNotFound(err) {
		return nil, err
	}

	if rp.kubeVersion == "" {
		kubeVersion, err := kube_utils.GetKubernetesClusterVersion(discoveryClient)
		if err != nil {
			logger.Error(err, "Failed to get Kubernetes cluster version")
		}
		rp.kubeVersion = kubeVersion
	}

	var ingressClass string
	ingressEnabled := true
	if rp.runaiConfig.Spec.Global.Ingress != nil {
		ingressClass = rp.runaiConfig.Spec.Global.Ingress.Class
		ingressEnabled = ptr.Deref(rp.runaiConfig.Spec.Global.Ingress.Enabled, true)
	}
	gatewayAPIEnabled := false
	if gw := rp.runaiConfig.Spec.Global.GatewayAPI; gw != nil {
		gatewayAPIEnabled = ptr.Deref(gw.Enabled, false)
	}
	var oauth2ProxyResources *corev1.ResourceRequirements
	var usernameLoaderResources *corev1.ResourceRequirements
	var gitSyncResources *corev1.ResourceRequirements
	var s3FileSystemResources *corev1.ResourceRequirements
	if injected := rp.runaiConfig.Spec.Global.WorkloadInjectedContainers; injected != nil {
		if injected.OAuth2Proxy != nil {
			oauth2ProxyResources, err = injected.OAuth2Proxy.Resources.ToResourceRequirements()
			if err != nil {
				return nil, fmt.Errorf("validate oauth2Proxy resources: %w", err)
			}
		}
		if injected.UsernameLoader != nil {
			usernameLoaderResources, err = injected.UsernameLoader.Resources.ToResourceRequirements()
			if err != nil {
				return nil, fmt.Errorf("validate usernameLoader resources: %w", err)
			}
		}
		if injected.GitSync != nil {
			gitSyncResources, err = injected.GitSync.Resources.ToResourceRequirements()
			if err != nil {
				return nil, fmt.Errorf("validate gitSync resources: %w", err)
			}
		}
		if injected.S3FileSystem != nil {
			s3FileSystemResources, err = injected.S3FileSystem.Resources.ToResourceRequirements()
			if err != nil {
				return nil, fmt.Errorf("validate s3FileSystem resources: %w", err)
			}
		}
	}

	configuration := &runv1.PublicConfiguration{
		WorkloadOwnershipProtection: rp.runaiConfig.Spec.Global.EnableWorkloadOwnershipProtection,
		SubdomainEnabled:            ptr.To(ptr.Deref(rp.runaiConfig.Spec.Global.SubdomainSupport, true)),
		RunaiOauth2ProxyResources:   resourceRequirementsToPublic(oauth2ProxyResources),
		UsernameLoaderResources:     resourceRequirementsToPublic(usernameLoaderResources),
		GitSyncResources:            resourceRequirementsToPublic(gitSyncResources),
		S3FileSystemResources:       resourceRequirementsToPublic(s3FileSystemResources),
		IngressClass:                &ingressClass,
		IngressEnabled:              ptr.To(ingressEnabled),
		GatewayAPIEnabled:           ptr.To(gatewayAPIEnabled),
	}

	if gw := rp.runaiConfig.Spec.Global.GatewayAPI; gw != nil && gatewayAPIEnabled {
		configuration.GatewayName = &gw.Name
		configuration.GatewayNamespace = &gw.Namespace
	}
	if rp.inferenceEndpointIntegrationAvailable(ctx) {
		configuration.InferenceEndpointSchemaVersion = ptr.To(runv1.InferenceEndpointIntegrationV1)
	}

	if aw := rp.runaiConfig.Spec.AnyworkloadController; aw != nil {
		configuration.GitSyncImage = formatImageRef(aw.GitSyncImage)
		configuration.S3FileSystemImage = formatImageRef(aw.S3FileSystemImage)
		configuration.RunaiOauth2ProxyImage = formatImageRef(aw.RunaiOauth2ProxyImage)
		configuration.RunaiOauth2ProxyUsernameLoaderImage = formatImageRef(aw.RunaiOauth2ProxyUsernameLoaderImage)
		configuration.RunaiOauth2ProxyRedisFQDN = aw.RunaiOauth2ProxyRedisFQDN
		configuration.RunaiOauth2ProxyWorkloadAuthSecretName = aw.RunaiOauth2ProxyWorkloadAuthSecretName
		configuration.RunaiOauth2ProxySessionStoreType = aw.RunaiOauth2ProxySessionStoreType
		configuration.ExternalAuthUrlEnabled = aw.ExternalAuthUrlEnabled
		configuration.OpenshiftRouteTerminationPolicy = ptr.To(string(aw.OpenshiftRouteConfiguration.GetTerminationPolicy()))
		configuration.DistGPUNetworkAcceleration = aw.GPUNetworkAccelerationEnabled
		configuration.DinfGPUNetworkAcceleration = aw.GPUNetworkAccelerationEnabled
	}
	configuration.CustomCAGit = customCAToConfig(rp.runaiConfig.Spec.Global.CustomCAGit)
	configuration.CustomCAS3 = customCAToConfig(rp.runaiConfig.Spec.Global.CustomCAS3)
	if internal := rp.runaiConfig.Spec.Internal; internal != nil && internal.Global != nil {
		configuration.CustomCAOauth2Proxy = customCAToConfig(internal.Global.CustomCA)
		if internal.Global.FipsMode != "" {
			configuration.FipsMode = ptr.To(internal.Global.FipsMode)
		}
		if issuer := internal.Global.ControlPlane.Config.GetResearcherService().OauthProxy.GetOidcIssuerUrl(); issuer != "" {
			configuration.RunaiOauth2ProxyOidcIssuerUrl = ptr.To(issuer)
		}
	}

	runaiPublic := runv1.RunaiPublic{
		Version:           rp.runaiConfig.Spec.Internal.Global.Version,
		RunaiConfigStatus: &rp.runaiConfig.Status,
		Dependencies: &runv1.PublicDependencies{
			Required: getRequiredComponentsStatus(ctx, rp.kbClient, rp.runaiConfig),
			Optional: getOptionalComponentsStatus(ctx, rp.kbClient, rp.runaiConfig, discoveryClient),
		},
		Platform: &runv1.PublicPlatform{
			Type:        strings.ToLower(rp.runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor.String()),
			KubeVersion: rp.kubeVersion,
		},
		Configuration: configuration,
	}

	runaiPublicYaml, err := yaml.Marshal(&runaiPublic)

	if err != nil {
		return nil, err
	}

	configMap.Name = name
	configMap.Namespace = namespace
	configMap.Data = map[string]string{}
	configMap.Data["cluster-version"] = rp.runaiConfig.Spec.Internal.Global.Version
	configMap.Data["runai-public"] = string(runaiPublicYaml)

	return configMap, nil
}

func (rp *RunaiPublic) Status(ctx context.Context) error {
	configMap := &corev1.ConfigMap{}

	err := rp.kbClient.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, configMap)
	if err != nil && !kerrors.IsNotFound(err) {
		return err
	} else if kerrors.IsNotFound(err) {
		return fmt.Errorf("configmap [%s] is missing", name)
	}

	return nil
}

func (rp *RunaiPublic) Enabled() bool {
	return true
}

func getRequiredComponentsStatus(ctx context.Context, kbClient kc.HybridClient, config *runv1.RunaiConfig) map[string]runv1.PublicStatus {
	components := map[string]runv1.PublicStatus{}

	var gpuStackStatus, prometheusStatus runv1.PublicStatus

	if config.ShouldSkipDependencyChecks() {
		gpuStackStatus = runv1.PublicStatus{
			Available: runv1.AvailableStatus(true),
			Reason:    "Managed Environment - Dependency checks skipped",
		}

		prometheusStatus = runv1.PublicStatus{
			Available: runv1.AvailableStatus(true),
			Reason:    "Managed Environment - Dependency checks skipped",
		}
	} else {
		gpuStackStatus = getGpuStackStatus(ctx, kbClient)
		prometheusStatus = getPrometheusStatus(config)
	}

	components["gpu-stack"] = gpuStackStatus
	components["prometheus"] = prometheusStatus

	return components
}

func getPrometheusStatus(config *runv1.RunaiConfig) runv1.PublicStatus {
	// Skip version check for OpenShift built-in prometheus
	if config.Spec.Internal.Global.Platform.KubernetesFlavor == runv1_internal.KubernetesFlavorOpenShift &&
		*config.Spec.Prometheus.Openshift.UseBuiltInPrometheus {
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(true),
			Reason:    "",
			Components: map[string]runv1.PublicStatus{
				"Oauth2": {
					Available: runv1.AvailableStatus(true),
					Reason:    "Running on OpenShift using built-in Thanos which always supports OAuth2",
				},
			},
		}
	}

	version, err := resources.PrometheusVersion(config)
	if err != nil {
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    errToString(err),
		}
	}

	versionError := isVersionSupportsOauth2(version)
	return runv1.PublicStatus{
		Available: runv1.AvailableStatus(true),
		Reason:    "",
		Components: map[string]runv1.PublicStatus{
			"Oauth2": {
				Available: runv1.AvailableStatus(versionError == nil),
				Reason:    errToString(versionError),
			},
		},
	}
}

func isVersionSupportsOauth2(version string) error {
	var MIN_VERSION = "2.27.0"
	res, err := versionutil.VersionCompare(version, MIN_VERSION)
	if err != nil {
		return err
	}

	if res < 0 {
		return fmt.Errorf("%s", "Prometheus version: "+version+" - not support oauth2, oauth2 valid in Prometheus versions 2.27.0 and newer")
	}

	return nil
}

func getGpuStackStatus(ctx context.Context, kbClient client.Client) runv1.PublicStatus {
	nvidiaStatus := getNvidiaStatus(ctx, kbClient)
	gpuFractioningStatus := controller_utils.CheckGpuFractioningReadiness(ctx, kbClient)

	available := *nvidiaStatus.Available && *gpuFractioningStatus.Available

	var reason string
	if !*gpuFractioningStatus.Available {
		reason = gpuFractioningStatus.Reason
	} else if !*nvidiaStatus.Available {
		reason = nvidiaStatus.Reason
	}

	return runv1.PublicStatus{
		Available: runv1.AvailableStatus(available),
		Reason:    reason,
		Components: map[string]runv1.PublicStatus{
			"nvidia":          nvidiaStatus,
			"gpu-fractioning": gpuFractioningStatus,
		},
	}
}

func getNvidiaStatus(ctx context.Context, kbClient client.Client) runv1.PublicStatus {
	svc, nvidiaStackVerificationErr := controller_utils.GetDCGMExporterService(ctx, kbClient)
	if svc == nil || nvidiaStackVerificationErr != nil {
		return getUnavailableNvidiaStatus(nvidiaStackVerificationErr)
	}

	nvidiaExporterVerificationErr, missingMetrics := controller_utils.VerifyNvidiaStackExistsInRightState(svc)
	return getAvailableNvidiaStatus(nvidiaExporterVerificationErr, missingMetrics)
}

func getUnavailableNvidiaStatus(nvidiaStackVerificationErr error) runv1.PublicStatus {
	return runv1.PublicStatus{
		Available: runv1.AvailableStatus(false),
		Reason:    errToString(nvidiaStackVerificationErr),
		Components: map[string]runv1.PublicStatus{
			"nvidia-metrics": {
				Available: runv1.AvailableStatus(false),
			},
		},
	}
}

func getAvailableNvidiaStatus(nvidiaExporterVerificationErr error, missingMetrics string) runv1.PublicStatus {
	return runv1.PublicStatus{
		Available: runv1.AvailableStatus(true),
		Reason:    "",
		Components: map[string]runv1.PublicStatus{
			"nvidia-metrics": {
				Available: runv1.AvailableStatus(nvidiaExporterVerificationErr == nil && len(missingMetrics) == 0),
				Reason:    getMissingMetricsErrorStr(nvidiaExporterVerificationErr, missingMetrics),
			},
		},
	}
}

func getMissingMetricsErrorStr(nvidiaExporterVerificationErr error, missingMetrics string) string {
	if nvidiaExporterVerificationErr != nil {
		return nvidiaExporterVerificationErr.Error()
	}

	if len(missingMetrics) == 0 {
		return ""
	}

	return "Metrics: " + missingMetrics + " are missing in the DCGM-exporter. " +
		"Make sure they are mentioned in the file specified under the DCGM_EXPORTER_COLLECTORS environment variable"
}

func getOptionalComponentsStatus(ctx context.Context, kbClient kc.HybridClient, runaiConfig *runv1.RunaiConfig, discoveryClient discovery.DiscoveryInterface) map[string]runv1.PublicStatus {
	components := map[string]runv1.PublicStatus{}

	components["knative"] = knative.GetKnativeStatus(ctx, kbClient)
	components["lws"] = getLWSStatus(ctx, kbClient, runaiConfig)
	components[runv1.RunaiPublicMPI] = getMpiStatus(ctx, kbClient, discoveryClient)

	// handle xgboost,tensorflow,pytorch
	trainingOperatorComponentsStatus := getTrainingOperatorComponentsStatus(ctx, kbClient, discoveryClient)
	for framework, status := range trainingOperatorComponentsStatus {
		components[framework] = status
	}

	if gwStatus := getGatewayAPIStatus(ctx, kbClient, runaiConfig, discoveryClient); gwStatus != nil {
		components[runv1.RunaiPublicGatewayAPI] = *gwStatus
	}

	components[runv1.RunaiPublicAdvice] = getAdviceCRDStatus(discoveryClient)

	return components
}

func getAdviceCRDStatus(discoveryClient discovery.DiscoveryInterface) runv1.PublicStatus {
	installed, err := isCrdInstalled(discoveryClient, "Advice", "heimdall.nvidia.com/v1alpha1")
	if err != nil {
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    fmt.Sprintf("CRD check failed: %v", err),
		}
	}
	return runv1.PublicStatus{
		Available: runv1.AvailableStatus(installed),
	}
}

func getGatewayAPIStatus(ctx context.Context, kbClient kc.HybridClient, runaiConfig *runv1.RunaiConfig, discoveryClient discovery.DiscoveryInterface) *runv1.PublicStatus {
	gw := runaiConfig.Spec.Global.GatewayAPI
	if gw == nil || gw.Enabled == nil || !*gw.Enabled {
		return nil
	}

	// Check HTTPRoute CRD via discovery API
	crdInstalled, crdErr := isCrdInstalled(discoveryClient, "HTTPRoute", gatewayAPIGroupVersion)
	crdStatus := runv1.PublicStatus{}
	if crdErr != nil {
		crdStatus.Available = runv1.AvailableStatus(false)
		crdStatus.Reason = fmt.Sprintf("Gateway API CRD check failed: %v", crdErr)
	} else {
		crdStatus.Available = runv1.AvailableStatus(crdInstalled)
		if !crdInstalled {
			crdStatus.Reason = "Gateway API CRDs not installed"
		}
	}

	// Check referenced Gateway resource exists
	gatewayStatus := runv1.PublicStatus{Available: runv1.AvailableStatus(false)}
	if crdInstalled {
		gateway := &gwv1.Gateway{}
		err := kbClient.Uncached().Get(ctx, types.NamespacedName{Name: gw.Name, Namespace: gw.Namespace}, gateway)
		if err != nil {
			if kerrors.IsNotFound(err) {
				gatewayStatus.Reason = fmt.Sprintf("Gateway %s/%s not found", gw.Namespace, gw.Name)
			} else {
				gatewayStatus.Reason = fmt.Sprintf("failed to get Gateway %s/%s: %v", gw.Namespace, gw.Name, err)
			}
		} else {
			gatewayStatus.Available = runv1.AvailableStatus(true)
		}
	} else {
		gatewayStatus.Reason = "Gateway API CRDs not installed"
	}

	topAvailable := ptr.Deref(crdStatus.Available, false) && ptr.Deref(gatewayStatus.Available, false)
	return &runv1.PublicStatus{
		Available: runv1.AvailableStatus(topAvailable),
		Components: map[string]runv1.PublicStatus{
			"httproute-crd": crdStatus,
			"gateway":       gatewayStatus,
		},
	}
}

func getMpiDependencyStatus(
	ctx context.Context,
	kbClient client.Client,
	discoveryClient discovery.DiscoveryInterface,
) runv1.PublicStatus {
	// Validate the MPI component namespace
	err := controller_utils.ValidateComponentNamespace(ctx, kbClient, "mpi", "operator", "")
	if err != nil {
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    fmt.Sprintf("Namespace validation failed: %v", err),
		}
	}

	// Check if the MPIJob CRD is installed
	installed, crdErr := isCrdInstalled(discoveryClient, mpiJobKind, kubeflowMPIV2Beta1GroupVersion)
	if crdErr != nil {
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    fmt.Sprintf("CRD check failed: %v", crdErr),
		}
	}

	return runv1.PublicStatus{
		Available: runv1.AvailableStatus(installed),
		Reason:    "",
	}
}

func getMpiStatus(ctx context.Context, kbClient client.Client, discoveryClient discovery.DiscoveryInterface) runv1.PublicStatus {
	mpiStatus := getMpiDependencyStatus(ctx, kbClient, discoveryClient)

	suspendStatus := runv1.PublicStatus{
		Available: runv1.AvailableStatus(false),
	}
	if ptr.Deref(mpiStatus.Available, false) {
		suspendStatus = getMpiOperatorSuspendStatus(ctx, kbClient)
	}

	mpiStatus.Components = map[string]runv1.PublicStatus{
		runv1.DistributedFrameworkSuspend: suspendStatus,
	}

	return mpiStatus
}

func getTrainingOperatorComponentsStatus(ctx context.Context, kbClient client.Client, discoveryClient discovery.DiscoveryInterface) map[string]runv1.PublicStatus {
	var errs error

	components := map[string]runv1.PublicStatus{}
	kubeflowDeployment, err := getTrainingOperatorDeployment(ctx, kbClient)
	if err != nil {
		errs = multierr.Append(errs, err)
	}

	suspendStatus := getTrainingOperatorSuspendStatus(ctx, kbClient)
	kubeflowDeploymentFrameworkArgs, deploymentError := getKubeFlowEnabledFrameworks(kubeflowDeployment)
	for framework, kind := range consts.TrainingOperatorFrameworksToKindMap {

		installed, crdErr := isCrdInstalled(discoveryClient, kind, kfv1.SchemeGroupVersion.String())
		dependency := runv1.PublicStatus{}
		allErrs := multierr.Append(crdErr, errs)

		if deploymentError != nil {
			setFrameworkUnavailable(&dependency, deploymentError)
		} else if allErrs != nil {
			setFrameworkUnavailable(&dependency, allErrs)
		} else
		//In case no enable-scheme arguments, all the distributed jobs are enabled, if installed.
		//If there is at least one enable scheme argument, only those are effectively enabled.
		if !(len(kubeflowDeploymentFrameworkArgs) == 0 || slices.Contains(kubeflowDeploymentFrameworkArgs, consts.TrainingOperatorFrameworksToSingularMap[framework])) {
			setFrameworkUnavailable(&dependency, fmt.Errorf("%s scheme is not enabled in kubeflow deployment", framework))
		} else {
			dependency.Available = runv1.AvailableStatus(installed)
			suspendable := suspendStatus
			if !installed {
				suspendable = runv1.PublicStatus{
					Available: runv1.AvailableStatus(false),
				}
			}
			dependency.Components = map[string]runv1.PublicStatus{
				runv1.DistributedFrameworkSuspend: suspendable,
			}
		}
		components[framework] = dependency
	}
	return components
}

func setFrameworkUnavailable(dependency *runv1.PublicStatus, err error) {
	dependency.Reason = err.Error()
	dependency.Available = runv1.AvailableStatus(false)
	dependency.Components = map[string]runv1.PublicStatus{
		runv1.DistributedFrameworkSuspend: {
			Available: runv1.AvailableStatus(false),
		},
	}
}

func getKubeFlowEnabledFrameworks(kubeflowDeployment *appsv1.Deployment) ([]string, error) {
	if kubeflowDeployment == nil {
		return nil, fmt.Errorf("failed to get enabled frameworks args. kubeflow deployment is nil")
	}
	args, err := utils.GetDeploymentArgs(*kubeflowDeployment)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled frameworks args. %w", err)
	}
	var enabledSchemeArgs []string
	for _, arg := range args {
		if strings.Contains(arg, consts.EnableSchemeFlag) {
			enabledSchemeArgs = append(enabledSchemeArgs, strings.TrimPrefix(arg, consts.EnableSchemeFlag)[1:])
		}
	}
	return enabledSchemeArgs, nil
}

func getTrainingOperatorSuspendStatus(ctx context.Context, kbClient client.Client) runv1.PublicStatus {
	logger := log.FromContext(ctx)

	// get the cm that holds the mapping between training-operator image tags to version numbers
	trainingOperatorTagsCm := &corev1.ConfigMap{}
	err := kbClient.Get(ctx, types.NamespacedName{
		Name:      consts.TrainingOperatorTagsConfigMapName,
		Namespace: runv1.DefaultNamespace,
	}, trainingOperatorTagsCm)
	if err != nil {
		logger.Error(err, "Failed to get training-operator tags ConfigMap")
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    err.Error(),
		}
	}

	tagToVersionFn := func(tag string) (string, bool) {
		imageVersion, ok := trainingOperatorTagsCm.Data[tag] // training-operator image tags are hash-based, need to convert to human-readable
		return imageVersion, ok
	}
	return getSuspendStatus(ctx, kbClient, consts.TrainingOperatorDeploymentName, consts.TrainingOperatorNamespace, tagToVersionFn, consts.TrainingOperatorMinSuspendVersion)
}

func getMpiOperatorSuspendStatus(ctx context.Context, kbClient client.Client) runv1.PublicStatus {
	tagToVersionFn := func(tag string) (string, bool) {
		return tag, true // mpi-operator image tags are human-readable
	}
	return getSuspendStatus(ctx, kbClient, consts.MpiOperatorDeploymentName, consts.MpiOperatorNamespace, tagToVersionFn, consts.MpiOperatorMinSuspendVersion)
}

// getSuspendStatus determines if a distributed framework supports suspending based on the operator version
func getSuspendStatus(ctx context.Context, kbClient client.Client, deploymentName, deploymentNamespace string, tagToVersionFn func(string) (string, bool), minVersion string) runv1.PublicStatus {
	suspendStatus := runv1.PublicStatus{
		Available: runv1.AvailableStatus(false),
	}

	image, err := utils.GetDeploymentImage(ctx, kbClient, deploymentName, deploymentNamespace)
	if err != nil {
		suspendStatus.Reason = err.Error()
		return suspendStatus
	}

	imageTag, err := utils.GetImageTag(image)
	if err != nil {
		suspendStatus.Reason = err.Error()
		return suspendStatus
	}

	imageVersion, ok := tagToVersionFn(imageTag)
	if !ok {
		suspendStatus.Reason = fmt.Sprintf("deployment %s has unknown image tag %s", deploymentName, imageTag)
		return suspendStatus
	}

	supported, err := version.VersionGreaterEqualThan(imageVersion, minVersion)
	if err != nil {
		suspendStatus.Reason = err.Error()
		return suspendStatus
	}

	suspendStatus.Available = runv1.AvailableStatus(supported)
	if !supported {
		suspendStatus.Reason = fmt.Sprintf("%s version does not support suspend. minimal required: %s, actual: %s", deploymentName, minVersion, imageVersion)
	}

	return suspendStatus
}

func getTrainingOperatorDeployment(ctx context.Context, kbClient client.Client) (*appsv1.Deployment, error) {
	component := "training"
	componentBaseName := "operator"
	componentNameRegex := regexp.MustCompile(fmt.Sprintf("%s.*%s", component, componentBaseName))
	componentDisplayName := fmt.Sprintf("%s %s", component, componentBaseName)
	ns, err := controller_utils.GetComponentNamespace(ctx, kbClient, componentNameRegex, componentDisplayName, "")
	if err != nil {
		return nil, err
	}
	deployment := &appsv1.Deployment{}

	err = kbClient.Get(ctx, types.NamespacedName{
		Name:      consts.TrainingOperatorDeploymentName,
		Namespace: ns,
	}, deployment)
	if err != nil {
		return nil, err
	}
	if deployment.Status.AvailableReplicas == 0 {
		return nil, fmt.Errorf("no available replica for deployemnt %s", deployment.Name)
	}
	return deployment, nil
}

func errToString(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

func formatImageRef(image *common.Image) *string {
	if image == nil || image.Registry == nil || image.Name == nil || image.Tag == nil {
		return nil
	}
	ref := fmt.Sprintf("%s/%s:%s", *image.Registry, *image.Name, *image.Tag)
	return &ref
}

func resourceRequirementsToPublic(resources *corev1.ResourceRequirements) *runv1.PublicResourceRequirements {
	if resources == nil {
		return nil
	}
	return &runv1.PublicResourceRequirements{
		Limits:   resourceListToStrings(resources.Limits),
		Requests: resourceListToStrings(resources.Requests),
	}
}

func resourceListToStrings(resources corev1.ResourceList) map[string]string {
	if resources == nil {
		return nil
	}
	result := make(map[string]string, len(resources))
	for name, quantity := range resources {
		result[string(name)] = quantity.String()
	}
	return result
}

func customCAToConfig(ca *common.CustomCA) *runv1.CustomCAConfig {
	if ca == nil {
		return nil
	}
	out := &runv1.CustomCAConfig{Enabled: ca.Enabled}
	if ca.Secret != nil {
		out.SecretName = ptr.To(ca.Secret.Name)
		out.SecretKey = ptr.To(ca.Secret.Key)
		out.Optional = ca.Secret.Optional
	}
	return out
}

func NewCrdClientSetFn() (kubernetes.Interface, error) {
	cfg, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func isCrdInstalled(discoveryClient discovery.DiscoveryInterface, kind, resourceGroupVersion string) (bool, error) {
	apiList, err := discoveryClient.ServerResourcesForGroupVersion(resourceGroupVersion)
	if err != nil {
		if kerrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	for _, api := range apiList.APIResources {
		if api.Kind == kind {
			return true, nil
		}
	}
	return false, nil
}

func getLWSStatus(ctx context.Context, kbClient client.Client, runaiConfig *runv1.RunaiConfig) runv1.PublicStatus {
	// Check if the LWS CRD is installed by trying to create a test resource
	lwsGVK := schema.GroupVersionKind{
		Group:   "leaderworkerset.x-k8s.io",
		Version: "v1",
		Kind:    "LeaderWorkerSet",
	}

	// First check if the CRD is installed
	err := kbClient.Create(ctx, &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": lwsGVK.GroupVersion().String(),
			"kind":       lwsGVK.Kind,
			"metadata": map[string]interface{}{
				"name":      "test",
				"namespace": corev1.NamespaceDefault,
			},
		},
	}, &client.CreateOptions{DryRun: []string{metav1.DryRunAll}})
	if err != nil {
		if meta.IsNoMatchError(err) || strings.Contains(err.Error(), "the server could not find the requested resource") {
			return runv1.PublicStatus{
				Available: runv1.AvailableStatus(false),
				Reason:    "LeaderWorkerSet CRD is not installed in the cluster",
			}
		}
	}

	// Get the LWS namespace from the config (uses default if not set)
	lwsNamespace := consts.DefaultLWSNamespace
	if runaiConfig.Spec.LWS != nil && runaiConfig.Spec.LWS.Namespace != nil {
		lwsNamespace = *runaiConfig.Spec.LWS.Namespace
	}

	// Then check if the LWS operator is running
	deployment := &appsv1.Deployment{}
	err = kbClient.Get(ctx, types.NamespacedName{
		Name:      consts.LWSDeploymentName,
		Namespace: lwsNamespace,
	}, deployment)
	if err != nil {
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    fmt.Sprintf("Failed to find LeaderWorkerSet operator deployment: %v", err),
		}
	}

	if deployment.Status.AvailableReplicas == 0 {
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    "LeaderWorkerSet operator has no available replicas",
		}
	}

	return runv1.PublicStatus{
		Available: runv1.AvailableStatus(true),
	}
}
