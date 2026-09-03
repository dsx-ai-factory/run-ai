/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_workloads

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	trackerApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	corev1 "k8s.io/api/core/v1"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const (
	KnativeServiceAPIVersion = "serving.knative.dev/v1"
	KnativeServiceKind       = "Service"

	annotationAutoscalingClass     = "autoscaling.knative.dev/class"
	annotationAutoscalingMetric    = "autoscaling.knative.dev/metric"
	annotationAutoscalingTarget    = "autoscaling.knative.dev/target"
	annotationAutoscalingMin       = "autoscaling.knative.dev/min-scale"
	annotationAutoscalingMax       = "autoscaling.knative.dev/max-scale"
	annotationScaleToZeroRetention = "autoscaling.knative.dev/scale-to-zero-pod-retention-period"
	annotationProgressDeadline     = "serving.knative.dev/progress-deadline"
	annotationClusterLocal         = "networking.knative.dev/visibility"
	clusterLocalValue              = "cluster-local"
	autoscalingClassKPA            = "kpa.autoscaling.knative.dev"
	autoscalingClassHPA            = "hpa.autoscaling.knative.dev"
	autoscalingMetricRPS           = "rps"
	autoscalingMetricLatency       = "knative_revision_request_latency"
)

// KnativeServiceTarget — workload-manager wraps Service + Config into a WorkloadRunner downstream.
type KnativeServiceTarget struct {
	Service knservingv1.Service
	Config  openapi.WorkloadV2Configuration
}

func NewMinimalKnativeServiceTarget() *KnativeServiceTarget {
	result := &KnativeServiceTarget{}
	result.Service.APIVersion = KnativeServiceAPIVersion
	result.Service.Kind = KnativeServiceKind
	result.Service.Spec.Template.Spec.Containers = []corev1.Container{{}}
	return result
}

type InferenceToKnativeComposer struct {
	fc *composer.FieldsComposer

	AssetsMapping        []workloadsApi.AssetsMapping
	WorkloadRequirements composer.WorkloadRequirements
	DataVolumeClaimNames map[string]string // maps data volume ID to PVC claim name
}

func NewInferenceToKnativeComposer(
	workloadName string,
	projectName string,
	namespace string,
	requester string,
	assetsMapping []workloadsApi.AssetsMapping,
	options composer.FieldsComposerOptions,
) *InferenceToKnativeComposer {
	return &InferenceToKnativeComposer{
		fc: &composer.FieldsComposer{
			WorkloadType:      composer.Inference,
			WorkloadName:      workloadName,
			Namespace:         namespace,
			ProjectName:       projectName,
			ResourcePrefix:    composer.NetworkingResourcePrefixInference,
			SubmittedBy:       requester,
			Options:           options,
			ItemizedNumbering: make(itemized.ItemizedNumbering),
			OauthPortNumber:   composer.Oauth2ContainerDefaultPort,
		},
		AssetsMapping: assetsMapping,
	}
}

func (itc *InferenceToKnativeComposer) WithStorageClassConfiguration(config map[string]trackerApi.StorageClassConfiguration) *InferenceToKnativeComposer {
	itc.fc.StorageClassConfiguration = config
	return itc
}

func (itc *InferenceToKnativeComposer) WithClusterStorageConfig(config composer.ClusterStorageConfig) *InferenceToKnativeComposer {
	itc.fc.ClusterStorageConfig = config
	return itc
}

func (itc *InferenceToKnativeComposer) WithDataVolumeClaimNames(claimNames map[string]string) *InferenceToKnativeComposer {
	itc.DataVolumeClaimNames = claimNames
	return itc
}

// Compose translates an inference spec into a Knative Service and workload configuration.
func (itc *InferenceToKnativeComposer) Compose(
	ctx context.Context,
	src *workloadsApi.InferenceSpecSpec,
	target *KnativeServiceTarget,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if target == nil {
			return workloadsApi.NewError(http.StatusInternalServerError, "Compose: target is nil")
		}
		if target.Service.APIVersion == "" {
			target.Service.APIVersion = KnativeServiceAPIVersion
		}
		if target.Service.Kind == "" {
			target.Service.Kind = KnativeServiceKind
		}
		if len(target.Service.Spec.Template.Spec.Containers) == 0 {
			target.Service.Spec.Template.Spec.Containers = []corev1.Container{{}}
		}

		if src == nil {
			return nil
		}

		engine, engineErr := detectInferenceEngine(src.Labels, convutils.Deref(src.Image.Get(), ""))
		if engineErr != nil {
			return engineErr
		}
		svcMeta := &target.Service.ObjectMeta
		podMeta := &target.Service.Spec.Template.ObjectMeta
		container := &target.Service.Spec.Template.Spec.Containers[0]
		podSpec := &target.Service.Spec.Template.Spec.PodSpec

		svcMeta.Name = itc.fc.WorkloadName
		svcMeta.Namespace = itc.fc.Namespace

		runner := &composer.StepRunner{}

		runner.Batch(
			itc.fc.ComposeFullySaasInd(&svcMeta.Labels, composer.BoolToMapString, composer.ControlPlaneManagedLabel),
			itc.fc.CopyBoolean(convutils.Ref(true), &svcMeta.Labels, composer.BoolToMapString, composer.ConvertedFromNativeLabel),
			itc.fc.CopyString(convutils.Ref(itc.fc.SubmittedBy), &svcMeta.Annotations, composer.StringToMapString, composer.SubmittedByAnnotation),
			itc.fc.ComposeNameAndTypeLabels(itc.fc.WorkloadName, itc.fc.WorkloadType, KnativeServiceKind, itc.fc.ProjectName, &svcMeta.Labels, composer.MapToMapString, nil),
			itc.fc.ComposeLabels(src.Labels, &svcMeta.Labels, composer.MapToMapString, nil),
			itc.fc.ComposeAnnotations(src.Annotations, &svcMeta.Annotations, composer.MapToMapString, nil),
			itc.fc.CopyString(src.Category.Get(), &svcMeta.Labels, composer.StringToMapString, composer.WorkloadCategoryLabel),
		)

		runner.Batch(
			itc.fc.ComposeNameAndTypeLabels(itc.fc.WorkloadName, itc.fc.WorkloadType, KnativeServiceKind, itc.fc.ProjectName, &podMeta.Labels, composer.MapToMapString, nil),
			itc.fc.ComposeLabels(src.Labels, &podMeta.Labels, composer.MapToMapString, nil),
			itc.fc.ComposeAnnotations(src.Annotations, &podMeta.Annotations, composer.MapToMapString, nil),
		)
		if engineLabel := engine.identityLabel(); engineLabel != "" {
			runner.Batch(
				itc.fc.CopyBoolean(convutils.Ref(true), &svcMeta.Labels, composer.BoolToMapString, engineLabel),
				itc.fc.CopyBoolean(convutils.Ref(true), &podMeta.Labels, composer.BoolToMapString, engineLabel),
			)
		}

		runner.Batch(
			itc.fc.CopyString(convutils.Ref(composer.RunaiScheduler), &podSpec.SchedulerName, composer.StringToString, nil),
			itc.fc.CopyString(convutils.Ref(itc.fc.WorkloadName), &container.Name, composer.StringToString, nil),
			itc.fc.ComposeImage(src.Image.Get(), &container.Image, composer.StringToString, nil),
			itc.fc.CopyString(composer.EnumToString(src.ImagePullPolicy.Get()), &container.ImagePullPolicy, composer.StringToString, nil),
			itc.fc.CopyString(src.Command.Get(), &container.Command, composer.UnixStrToStringArray, nil),
			itc.fc.CopyString(src.Args.Get(), &container.Args, composer.UnixStrToStringArray, nil),
			itc.fc.CopyString(src.WorkingDir.Get(), &container.WorkingDir, composer.StringToString, nil),
		)

		envVarsUsingCreds := make([]composer.EnvVarInfo, 0)
		runner.Run(
			itc.fc.ComposeEnvironmentVariables(itc.AssetsMapping, src.EnvironmentVariables, &container.Env, composer.EnvVarsToK8sEnv, nil, &envVarsUsingCreds),
		)

		targetForEnvVarsUsingCreds := composer_v2.EnvVarUsingCredTarget{
			Config:  &target.Config,
			EnvVars: &container.Env,
		}
		cookieForEnvVarsUsingCreds := composer_v2.PersonalCredsCookie{
			WorkloadName: itc.fc.WorkloadName,
		}

		runner.Run(itc.fc.ComposeEnvVarsUsingCreds(envVarsUsingCreds, &targetForEnvVarsUsingCreds, composer_v2.EnvVarsUsingCredSetter, &cookieForEnvVarsUsingCreds))

		credTarget := composer_v2.PersonalCredsTarget{
			Config:  &target.Config,
			PodSpec: podSpec,
		}
		credCookie := composer_v2.PersonalCredsCookie{
			WorkloadName: itc.fc.WorkloadName,
		}
		runner.Run(itc.fc.ComposeImagePullSecrets(itc.AssetsMapping, src.ImagePullSecrets, &credTarget, composer_v2.ArrayToImagePullSecretsSetter, &credCookie))

		runner.Run(itc.processComputeFields(ctx, src.Compute.Get(), target))
		runner.Run(itc.processSecurityFields(ctx, src.Security.Get(), target))

		rc := composer_v2.ExpandTemplateVarsInCommandAndArgs(container, itc.fc.ProjectName, itc.fc.WorkloadName)
		if rc != nil {
			return rc
		}

		runner.Run(itc.processStorageFields(ctx, src.Storage.Get(), target))

		runner.Run(itc.processPlacementFields(ctx, src, target))
		if convutils.HasValue(src.Probes) {
			runner.Run(itc.fc.ComposeReadinessProbe(src.Probes.Get().Readiness.Get(), &container.ReadinessProbe, composer.ProbeToK8s, nil))
		}

		if as := src.Autoscaling.Get(); as != nil {
			runner.Run(itc.fc.ComposeAutoScaling(as, &target.Service, autoScalingToKnative, nil))
			runner.Run(itc.fc.CopyInt(as.ConcurrencyHardLimit.Get(), &target.Service.Spec.Template.Spec.ContainerConcurrency, composer.IntToInt[int64], nil))
		}

		runner.Run(itc.processServingConfigFields(src.ServingConfiguration.Get(), target))

		if sp := src.ServingPort.Get(); sp != nil && convutils.BoolIsTrue(sp.ClusterLocalAccessOnly.Get()) {
			runner.Run(itc.fc.CopyString(convutils.Ref(clusterLocalValue), &svcMeta.Labels, composer.StringToMapString, annotationClusterLocal))
		}

		// Resolve serving-port authorization before composing exposed URLs.
		servingPortAuth, authErr := ResolveServingPortAuth(src.ServingPort.Get())
		if authErr != nil {
			return authErr
		}
		servingPortProtected := !servingPortAuth.IsPublic()
		oauthPort := composer.Oauth2ContainerDefaultPort
		if servingPortProtected {
			// Shift exposed-URL OAuth proxies to port 4181 because the serving proxy uses 4180.
			oauthPort++
		}
		itc.fc.ResetOAuthPortNumbering(oauthPort)

		var jupyterContainerPort int32
		var exposedUrlInfo *composer.ExposedUrlInfo
		// Collect exposed URLs before creating combined OAuth resources.
		runner.Run(itc.fc.ComposeExposedUrls(itemizedPrefix(itemized.ExposedUrlPrefix), src.ExposedUrls,
			nil, func(*composer.ExposedUrlInfo, any, any) *workloadsApi.Error { return nil }, nil,
			&jupyterContainerPort, &exposedUrlInfo, servingPortProtected))

		// Create serving and exposed-URL OAuth resources.
		runner.Run(itc.processNetworkingFields(src.ServingPort.Get(), servingPortAuth, engine, exposedUrlInfo, target))
		// Apply exposed-URL Service and Ingress configuration.
		runner.Run(itc.fc.ApplyExposedUrls(&exposedUrlInfo, &target.Config, composer_v2.UrlsToWorkloadRunner, nil))
		// Use standalone exposed-URL resources only when the serving port is not protected.
		if !servingPortProtected {
			runner.Batch(
				itc.fc.ApplyExposedUrls(&exposedUrlInfo, podSpec, composer_v2.UrlsToK8sInitContainers, nil),
				itc.fc.ApplyExposedUrls(&exposedUrlInfo, podSpec, composer.UrlsToK8sContainers, nil),
			)
		}

		return runner.Error()
	}
}
