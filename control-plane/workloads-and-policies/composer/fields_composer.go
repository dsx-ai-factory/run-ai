/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/workloadutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	trackerApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	policyConstants "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/modern-go/reflect2"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
)

const (
	AutoScalingMetric_THROUGHPUT  = "throughput"
	AutoScalingMetric_CONCURRENCY = "concurrency"
	AutoScalingMetric_LATENCY     = "latency"

	// Node Affinity
	nodeAffinityName         = "Node Affinity"
	nodeAffinityRequiredName = "nodeAffinityRequired"

	PodAffinityRequiredValue  = "requiredDuringSchedulingIgnoredDuringExecution"
	PodAffinityPreferredValue = "preferredDuringSchedulingIgnoredDuringExecution"

	WorkloadType_NIM = "NIMService"

	RunaiSchedulerName = "runai-scheduler"
)

const (
	Interactive          = "Interactive"
	Workspace            = "Workspace"
	Training             = "Training"
	Distributed          = "Distributed"
	Inference            = "Inference"
	DistributedInference = "DistributedInference"
)

type FieldsComposerOptions struct {
	// composer behavior
	IgnoreMissingValues          bool
	NormalizeServingPortAuthType bool

	// workload related
	ClusterOriginatedWorkload bool

	// CleanupCreatedPvcsOnCompletion opts the workload into deleting the PVCs the
	// composer creates for it once the WorkloadRunner reaches a terminal phase (storage plugin)
	// v1 parity: only the Workspace/Training/Distributed submitters set this.
	CleanupCreatedPvcsOnCompletion *bool

	// Cluster-version-gated options. Only capabilities introduced in cluster version >= 2.22
	// belong here; anything older is always-on (min supported cluster is 2.22).
	EmptyDirAsVolumeSupported        bool
	FlexiblePriorityClassesSupported bool
	GitSyncWithSshKeySupported       bool
	GpuComputeModeSupported          bool
	// GpuMemoryLimitAnnotationsSupported is set from checker.MinVersionGpuMemoryLimitAnnotations
	// (cluster >= 3.0.0). When true, the composer emits the per-container CUDA/KAI GPU
	// memory-limit annotations (nvidia.com/container.<name>.gpu-memory.limit and
	// kai.scheduler/container.<name>.gpu-memory.portion.limit). When false (older
	// cluster), it falls back to the legacy RUNAI_GPU_MEMORY_LIMIT / RUNAI_GPU_PORTION_LIMIT
	// container env vars.
	GpuMemoryLimitAnnotationsSupported   bool
	LauncherCreationPolicySupported      bool
	NodeAffinitySupported                bool
	OAuth2ProxyTemplatesSupported        bool
	PreemptibilitySupported              bool
	UserCredentialsSupported             bool
	VolumeDefaultModeSupported           bool
	WorkloadV2NetworkingSupported        bool
	WorkloadV2ServiceSpecLabelsSupported bool
}

func OptionsForPolicy() FieldsComposerOptions {
	return FieldsComposerOptions{
		IgnoreMissingValues:          true,
		NormalizeServingPortAuthType: false,
		ClusterOriginatedWorkload:    false,

		EmptyDirAsVolumeSupported:            true,
		FlexiblePriorityClassesSupported:     true,
		GitSyncWithSshKeySupported:           true,
		GpuMemoryLimitAnnotationsSupported:   true,
		GpuComputeModeSupported:              true,
		LauncherCreationPolicySupported:      true,
		NodeAffinitySupported:                true,
		OAuth2ProxyTemplatesSupported:        true,
		PreemptibilitySupported:              true,
		UserCredentialsSupported:             true,
		VolumeDefaultModeSupported:           true,
		WorkloadV2NetworkingSupported:        true,
		WorkloadV2ServiceSpecLabelsSupported: true,
	}
}

func OptionsForNimService() FieldsComposerOptions {
	return FieldsComposerOptions{
		UserCredentialsSupported: false,
	}
}

type FieldsComposer struct {
	WorkloadName string
	WorkloadType string
	Namespace    string
	SubmittedBy  string

	ProjectName string

	// ResourcePrefix is prepended to every K8s resource name produced by the
	// networking setters (Service, Ingress, Route, HTTPRoute) for an exposed URL.
	// It must be set per workload-kind to prevent collisions when multiple
	// workloads share the same name across kinds. Conventional values:
	//   - "job" for BatchJob-based workloads (interactive / training).
	//   - "kf"  for Kubeflow-based workloads (distributed).
	// Equivalent to the V1 cluster builder's "workloadInitials" (iw, tw, dw, ...).
	ResourcePrefix string

	Options FieldsComposerOptions

	StorageClassConfiguration map[string]trackerApi.StorageClassConfiguration

	ClusterStorageConfig ClusterStorageConfig

	ItemizedNumbering itemized.ItemizedNumbering
	OauthPortNumber   int32

	// GpuRequested records whether any replica composed so far requested a GPU.
	// It is set by ComposeGpuRequestsAndLimits (reusing the resolved GpuSizingInfo)
	// and consumed by MNNVL defaulting so a GPU-less workload never inherits a
	// cluster-wide GPU-network-acceleration default.
	GpuRequested bool
}

func (fc *FieldsComposer) WithStorageClassConfiguration(config map[string]trackerApi.StorageClassConfiguration) *FieldsComposer {
	fc.StorageClassConfiguration = config
	return fc
}

func (fc *FieldsComposer) WithClusterStorageConfig(config ClusterStorageConfig) *FieldsComposer {
	fc.ClusterStorageConfig = config
	return fc
}

func (fc *FieldsComposer) ResetOAuthPortNumbering(startingPort int32) {
	fc.OauthPortNumber = startingPort
}

var ThirtyDays = 30 * 24 * time.Hour
var DefaultAutoDeleteAfterCompletionSeconds = int64(ThirtyDays.Seconds())

////////////////////////////////////////////////////////////////////////////////////////////////////////
// General purpose
////////////////////////////////////////////////////////////////////////////////////////////////////////

func (fc *FieldsComposer) CopyString(src *string, target any, setter FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) CopyArray(src []string, target any, setter FnArraySetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) CopyBoolean(src *bool, target any, setter FnBooleanSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) CopyInt(src any, target any, setter FnIntSetter[int64], cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if reflect2.IsNil(src) {
			return nil
		}

		var val int64

		switch t := src.(type) {
		case *int64:
			val = convutils.Deref(t, 0)
		case *int32:
			val = int64(convutils.Deref(t, 0))
		case *int:
			val = int64(convutils.Deref(t, 0))
		default:
			return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("unsupported type %T for int field", t))
		}

		return setter(&val, target, cookie)
	}
}

func (fc *FieldsComposer) CopyIntWithDefault(defaultValue int64, src any, target any, setter FnIntSetter[int64], cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if reflect2.IsNil(src) {
			return fc.CopyInt(&defaultValue, target, setter, cookie)()
		} else {
			return fc.CopyInt(src, target, setter, cookie)()
		}
	}
}

func (fc *FieldsComposer) CopyStringToMap(src *string, target any, setter FnMapToMapSetter[string], itemName string) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if convutils.IsStringHasValue(src) {
			result := map[string]*string{
				itemName: src,
			}

			return setter(result, target, nil)
		}

		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////
// Workload Fields
////////////////////////////////////////////////////////////////////////////////////////////////////////

func (fc *FieldsComposer) ComposeNameAndTypeLabels(workloadName string, workloadType string, jobType string, projectName string, target any, setter FnMapToMapSetter[string], cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		labels := map[string]*string{
			constants.LabelApp:          convutils.Ref(strings.ToLower(jobType)),
			constants.LabelCreatedBy:    convutils.Ref(jobType),
			constants.RunaiProjectLabel: convutils.Ref(projectName),
			constants.LabelRelease:      convutils.Ref(workloadName),
			constants.LabelWorkloadKind: convutils.Ref(workloadType),
			constants.LabelWorkloadName: convutils.Ref(workloadName),
		}

		return setter(labels, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeImage(src *string, target any, setter FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if !fc.Options.IgnoreMissingValues && convutils.IsEmptyOrNull(src) {
			return workloadsApi.NewError(http.StatusBadRequest, "value for 'image' must be provided")
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) CopyMandatoryString(fieldName string, src *string, target any, setter FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if convutils.IsEmptyOrNull(src) {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("value for '%s' must be provided", fieldName))
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeEnvironmentVariables(assetsMapping []workloadsApi.AssetsMapping, src []workloadsApi.EnvironmentVariable,
	target any, setter FnEnvVarSetter, cookie any, envVarsUsingCred *[]EnvVarInfo,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		result := make([]EnvVarInfo, 0, len(src))

		var assetsMap map[string]workloadsApi.AssetsMapping
		if len(assetsMapping) != 0 {
			assetsMap = CreateAssetMappingMap(assetsMapping)
		}

		for _, envVar := range src {
			envVarName := convutils.Deref(envVar.Name.Get(), "")
			if envVarName == "" {
				return workloadsApi.NewError(http.StatusBadRequest, "environmentVariables: name is required")
			}

			if convutils.BoolIsTrue(envVar.Exclude.Get()) {
				result = append(result, EnvVarInfo{
					Name:    envVarName,
					Deleted: true,
				})
				continue
			}

			if convutils.HasValue(envVar.UserCredential) && !fc.Options.UserCredentialsSupported {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("environmentVariables '%s': userCredential is not supported", envVarName))
			}

			var valType string

			// count how many of the mutually exclusive fields are set
			fieldCount := 0
			if envVar.Value.IsSet() { // nil value for env var is allowed
				fieldCount++
				valType = EnvValueText
			}
			if convutils.HasValue(envVar.UserCredential) {
				fieldCount++
				valType = EnvValueUserCredentials
			}
			if convutils.HasValue(envVar.Secret) {
				fieldCount++
				valType = EnvValueSecret
			}
			if convutils.HasValue(envVar.ConfigMap) {
				fieldCount++
				valType = EnvValueConfigMap
			}
			if convutils.HasValue(envVar.PodFieldRef) {
				fieldCount++
				valType = EnvValuePodFieldRef
			}

			// check mutual exclusivity
			if fieldCount != 1 {
				errMsg := fmt.Sprintf("environmentVariables '%s': value", envVarName)
				if fc.Options.UserCredentialsSupported {
					errMsg += ", userCredential"
				}
				errMsg += ", configMap, podFieldRef"

				if fieldCount > 1 {
					errMsg += " and secret are mutually exclusive"
					return workloadsApi.NewError(http.StatusBadRequest, errMsg)
				} else {
					errMsg += " or secret must be set"
					return workloadsApi.NewError(http.StatusBadRequest, errMsg)
				}
			}

			item := EnvVarInfo{
				Name:      envVarName,
				ValueType: valType,
			}

			// handle each type properly
			if convutils.HasValue(envVar.Value) {
				item.ValueFrom = *envVar.Value.Get()
			} else if convutils.HasValue(envVar.Secret) {
				if !convutils.HasValue(envVar.Secret.Get().Name) {
					return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("environmentVariables '%s': required secret name is not provided", envVarName))
				}
				if !convutils.HasValue(envVar.Secret.Get().Key) {
					return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("environmentVariables '%s': required secret key is not provided", envVarName))
				}

				item.ValueFrom = *envVar.Secret.Get().Name.Get()
				item.Key = *envVar.Secret.Get().Key.Get()
			} else if convutils.HasValue(envVar.UserCredential) {
				if !convutils.HasValue(envVar.UserCredential.Get().Name) {
					return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("environmentVariables '%s': required user credential name is not provided", envVarName))
				}
				if !convutils.HasValue(envVar.UserCredential.Get().Key) {
					return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("environmentVariables '%s': required user credential key is not provided", envVarName))
				}

				if len(assetsMap) != 0 {
					credentialName := convutils.Deref(envVar.UserCredential.Get().Name.Get(), "")

					assetMapping, exists := (assetsMap)[credentialName]
					if !exists {
						return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("asset mapping not found for credential name: %s (env var: %s)", credentialName, envVarName))
					}

					if assetMapping.Id == nil {
						return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("asset mapping ID is nil for credential name: %s (env var: %s)", credentialName, envVarName))
					}

					credentialName, buildErr := BuildUserCredentialName(
						credentialName,
						assetMapping.Id,
						envVar.UserCredential.Get().Key.Get(),
						*assetMapping.Kind, NgcApiKeyGenericSecretPrefix,
					)
					if buildErr != nil {
						return buildErr
					}

					item.ValueFrom = credentialName
					item.Key = *envVar.UserCredential.Get().Key.Get()

					if envVarsUsingCred != nil {
						*envVarsUsingCred = append(*envVarsUsingCred, item)
					}
				} else {
					continue
				}
			} else if convutils.HasValue(envVar.ConfigMap) {
				if !convutils.HasValue(envVar.ConfigMap.Get().Name) {
					return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("environmentVariables '%s': required configMap name is not provided", envVarName))
				}
				if !convutils.HasValue(envVar.ConfigMap.Get().Key) {
					return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("environmentVariables '%s': required configMap key is not provided", envVarName))
				}

				item.ValueFrom = *envVar.ConfigMap.Get().Name.Get()
				item.Key = *envVar.ConfigMap.Get().Key.Get()
			} else if convutils.HasValue(envVar.PodFieldRef) {
				if !convutils.HasValue(envVar.PodFieldRef.Get().Path) {
					return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("environmentVariables '%s': required filedRef path is not provided", envVarName))
				}

				item.ValueFrom = *envVar.PodFieldRef.Get().Path.Get()
			}

			result = append(result, item)
		}

		return setter(result, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeLabels(src []workloadsApi.Label,
	target any, setter FnMapToMapSetter[string], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		result := make(map[string]*string)

		for _, label := range src {
			labelName := convutils.Deref(label.Name.Get(), "")
			if labelName == "" {
				return workloadsApi.NewError(http.StatusBadRequest, "labels: name must be provided")
			}

			if convutils.BoolIsTrue(label.Exclude.Get()) {
				result[labelName] = nil
			} else if convutils.HasValue(label.Value) {
				result[labelName] = label.Value.Get()
			} else {
				result[labelName] = convutils.Ref("")
			}
		}

		return setter(result, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeAnnotations(src []workloadsApi.Annotation,
	target any, setter FnMapToMapSetter[string], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		result := make(map[string]*string)

		for _, annotation := range src {
			annName := convutils.Deref(annotation.Name.Get(), "")
			if annName == "" {
				return workloadsApi.NewError(http.StatusBadRequest, "annotations: name must be provided")
			}

			if convutils.BoolIsTrue(annotation.Exclude.Get()) {
				result[annName] = nil
			} else if convutils.HasValue(annotation.Value) {
				result[annName] = annotation.Value.Get()
			} else {
				result[annName] = convutils.Ref("")
			}
		}

		return setter(result, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeFullySaasInd(target any, setter FnBooleanSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		saasInd := convutils.Ref(!fc.Options.ClusterOriginatedWorkload)

		return setter(saasInd, target, cookie)
	}
}

func (fc *FieldsComposer) ComposePriorityClass(ctx context.Context, src *string,
	targetPriorityClass any, setterPriorityClass FnStringSetter,
	targetPremptibility any, setterPremptibility FnBooleanSetter,
	cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if convutils.IsStringHasValue(src) {
			var setErr *workloadsApi.Error

			// priority class is already populated (on submission/policy)
			priorityClass := src
			if !fc.Options.FlexiblePriorityClassesSupported {
				// for old priority classes - validate and set preemptible
				err := ValidateStaticPriorityClass(*priorityClass, fc.WorkloadType)
				if err != nil {
					return err
				}
				// Preemptible is deprecated for new cluster, but still is set for old clusters
				// nolint:staticcheck
				if setterPremptibility != nil && *priorityClass == policyConstants.PriorityClassInteractivePreemptible {
					setErr = setterPremptibility(convutils.Ref(true), targetPremptibility, cookie)
				}
			} else {
				// for new priority feature - put the priority class on spec
				setErr = setterPriorityClass(priorityClass, targetPriorityClass, cookie)
			}
			if setErr != nil {
				return setErr
			}
			return nil
		}

		if fc.Options.FlexiblePriorityClassesSupported {
			// priority class is not populated + using new priority feature - add defaults according to system
			defaultTypeConfig, found := workloadutils.DefaultRunaiWorkloadTypesConfig[fc.WorkloadType]
			if found && defaultTypeConfig.Priority != "" {
				defaultPriorityClass := defaultTypeConfig.Priority
				setErr := setterPriorityClass(&defaultPriorityClass, targetPriorityClass, cookie)
				if setErr != nil {
					return setErr
				}
			} else {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Cannot find default priority class name for workload type %s", fc.WorkloadType)
			}
		}

		return nil
	}
}

func (fc *FieldsComposer) ComposePreemptibility(_ context.Context, src *workloadsApi.Preemptibility,
	target any, setter FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		if !fc.Options.PreemptibilitySupported {
			return workloadsApi.NewError(http.StatusBadRequest, "preemptibility is not supported in this cluster")
		}

		valueToSet := convutils.Deref(src, "")

		return setter(convutils.Ref(string(valueToSet)), target, cookie)
	}
}

func (fc *FieldsComposer) ComposeAutoDeletionAfterTermination(src *int64,
	target any, setter FnIntSetter[int64], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			src = convutils.Ref(DefaultAutoDeleteAfterCompletionSeconds)
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeParallelism(src *int32, completions *int32,
	target any, setter FnIntSetter[int32], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		var val int32

		if src != nil {
			val = *src
		} else if convutils.Deref(completions, 0) > 0 { // Set default Parallelism to 1 only if Parallelism isn't set and Completions is set to a positive number
			val = 1
		}

		return setter(&val, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeTerminationGracePeriod(_ context.Context, src *int64,
	target any, setter FnIntSetter[int64], cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		return setter(convutils.Ref(*src), target, cookie)
	}
}

func (fc *FieldsComposer) ComposeRestartPolicy(allowedValues datastructs.Set[string], src *string,
	target any, setter FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		err := ValidateSetOfValues(src, allowedValues, "restartPolicy")
		if err != nil {
			return err
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeAutoScaling(src *workloadsApi.AutoScaling,
	target any, setter FnAutoScalingSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			src = &workloadsApi.AutoScaling{}
		}

		// if min is not set, use 1 as default
		minReplicas := convutils.Deref(src.MinReplicas.Get(), 1)

		// if max is not set, use min as default unless min is 0 and then use 1
		maxReplicas := max(convutils.Deref(src.MaxReplicas.Get(), minReplicas), 1)

		// if min is set to 0 and metric is set, metric must be either throughput or concurrency
		metric := src.Metric.Get()
		isKpaMetric := metric == nil || *metric == AutoScalingMetric_THROUGHPUT || *metric == AutoScalingMetric_CONCURRENCY

		if src.MinReplicas.Get() == nil || minReplicas == 0 {
			if !isKpaMetric {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("autoscaling: metric '%s' is not supported when minReplicas is 0", *metric))
			}
		}

		// store back the min max that were determined
		src.MinReplicas.Set(&minReplicas)
		src.MaxReplicas.Set(&maxReplicas)

		if maxReplicas < minReplicas {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("autoscaling: maxReplicas (%d) must be no less than minReplicas (%d)", maxReplicas, minReplicas))
		}

		// if min < max, metric must be set, as it means that autoscaling is enabled.
		// exception: min = 0 and max = 1 which is a special autoscaling case that
		// exception: min = 0 and max = 1 which is a special autoscaling case that
		// doesn't involve metrics, but is derived purely by network activity
		if metric == nil && minReplicas < maxReplicas && (minReplicas != 0 || maxReplicas != 1) && !fc.Options.IgnoreMissingValues {
			return workloadsApi.NewError(http.StatusBadRequest, "autoscaling: metric is mandatory when minReplicas < maxReplicas")
		}

		if convutils.HasValue(src.InitialReplicas) {
			initialReplicas := *src.InitialReplicas.Get()

			if convutils.HasValue(src.MinReplicas) && initialReplicas < minReplicas {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("autoscaling: initialReplicas (%d) must be no less than minReplicas (%d)", initialReplicas, minReplicas))
			}

			if convutils.HasValue(src.MinReplicas) && initialReplicas > maxReplicas {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("autoscaling: initialReplicas (%d) must be no greater than maxReplicas (%d)", initialReplicas, maxReplicas))
			}
		}

		if convutils.HasValue(src.ActivationReplicas) {
			activationReplicas := *src.ActivationReplicas.Get()

			if convutils.HasValue(src.MinReplicas) && activationReplicas < minReplicas {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("autoscaling: activationReplicas (%d) must be no less than minReplicas (%d)", activationReplicas, minReplicas))
			}

			if convutils.HasValue(src.MinReplicas) && activationReplicas > maxReplicas {
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("autoscaling: activationReplicas (%d) must be no greater than maxReplicas (%d)", activationReplicas, maxReplicas))
			}
		}

		if convutils.HasValue(src.Metric) {
			// metric is set, so metricThreshold must be set as well
			if !convutils.HasValue(src.MetricThreshold) && !fc.Options.IgnoreMissingValues {
				return workloadsApi.NewError(http.StatusBadRequest, "autoscaling: metricThreshold is mandatory when metric is provided")
			}
		}

		if convutils.HasValue(src.MetricThreshold) {
			// metricThreshold is set, so metric must be set as well
			if !convutils.HasValue(src.Metric) && !fc.Options.IgnoreMissingValues {
				return workloadsApi.NewError(http.StatusBadRequest, "autoscaling: metric is mandatory when metricThreshold is provided")
			}
		}

		if convutils.HasValue(src.MetricThresholdPercentage) {
			if !isKpaMetric {
				return workloadsApi.NewError(http.StatusBadRequest, "autoscaling: metricThresholdPercentage is only supported when metric is either throughput or concurrency")
			}
		}

		if convutils.HasValue(src.ScaleToZeroRetentionSeconds) && minReplicas > 0 {
			return workloadsApi.NewError(http.StatusBadRequest, "autoscaling: scaleToZeroRetentionSeconds is only supported when minReplicas is 0")
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeNodeAffinity(src *workloadsApi.NodeAffinityRequired,
	target any, setter FnNodeAffinitySetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		if !fc.Options.NodeAffinitySupported {
			return workloadsApi.NewError(http.StatusBadRequest, "Node affinity is not supported in this cluster")
		}

		// convert each term
		convertedNodeAffinityRequiredTerms := make([][]corev1.NodeSelectorRequirement, 0, len(src.NodeSelectorTerms))
		for _, term := range src.NodeSelectorTerms {

			//convert each expression in evert term
			convertedMatchExpressions := make([]corev1.NodeSelectorRequirement, 0, len(term.MatchExpressions))
			for _, expr := range term.MatchExpressions {

				if validationErr := ValidateMatchExpression(expr); validationErr != nil {
					return validationErr
				}

				convertedMatchExpressions = append(convertedMatchExpressions, corev1.NodeSelectorRequirement{
					Key:      expr.Key,
					Operator: corev1.NodeSelectorOperator(expr.Operator),
					Values:   expr.Values,
				})
			}

			convertedNodeAffinityRequiredTerms = append(convertedNodeAffinityRequiredTerms, convertedMatchExpressions)
		}

		return setter(convertedNodeAffinityRequiredTerms, target, cookie)
	}
}

func (fc *FieldsComposer) ComposePodAffinityKey(src *string, target any, setter FnStringSetter, cookie any) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {

		if src == nil {
			if !fc.Options.IgnoreMissingValues {
				return workloadsApi.NewError(http.StatusBadRequest, "podAffinity: key must be provided")
			}
			return nil
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) ComposePodAffinity(src *workloadsApi.PodAffinity,
	target any, setter FnPodAffinitySetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		if !fc.Options.IgnoreMissingValues {
			if !src.Key.IsSet() || src.Key.Get() == nil || *src.Key.Get() == "" {
				return workloadsApi.NewError(http.StatusBadRequest, "podAffinity: key must be provided")
			}
			if !src.Type.IsSet() || src.Type.Get() == nil {
				return workloadsApi.NewError(http.StatusBadRequest, "podAffinity: type must be provided")
			}
		}

		return setter(src, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeNodePools(input *NodePoolsInfo,
	target any, setter FnNodePoolsSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if input == nil || len(input.RequestedNodePools) == 0 || (len(input.RequestedNodePools) == 1 && input.RequestedNodePools[0] == "default") {
			return nil
		}

		return setter(input, target, cookie)
	}
}

func ValidateMatchExpression(expr workloadsApi.MatchExpression) *workloadsApi.Error {
	if expr.Key == "" {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: Key must be set", nodeAffinityRequiredName))
	}
	if expr.Operator == "" {
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: Operator must be set", nodeAffinityRequiredName))
	}

	switch expr.Operator {
	case workloadsApi.MATCHEXPRESSIONOPERATOR_EXISTS, workloadsApi.MATCHEXPRESSIONOPERATOR_DOES_NOT_EXIST:
		if len(expr.Values) > 0 {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: Values must be empty when Operator is %s", nodeAffinityRequiredName, expr.Operator))
		}
	case workloadsApi.MATCHEXPRESSIONOPERATOR_IN, workloadsApi.MATCHEXPRESSIONOPERATOR_NOT_IN:
		if len(expr.Values) == 0 {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: Values must be non-empty when Operator is %s", nodeAffinityRequiredName, expr.Operator))
		}
	case workloadsApi.MATCHEXPRESSIONOPERATOR_LT, workloadsApi.MATCHEXPRESSIONOPERATOR_GT:
		if len(expr.Values) != 1 {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: Values must have a single element when Operator is %s", nodeAffinityRequiredName, expr.Operator))
		}
	}

	return nil
}

func (fc *FieldsComposer) ComposePodAffinityType(src *workloadsApi.PodAffinityType,
	target any, setter FnStringSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			if !fc.Options.IgnoreMissingValues {
				return workloadsApi.NewError(http.StatusBadRequest, "podAffinity: type must be provided")
			}
			return nil
		}

		var outputTypeValue *string
		switch *src {
		case workloadsApi.PODAFFINITYTYPE_REQUIRED:
			outputTypeValue = convutils.Ref(PodAffinityRequiredValue)
		case workloadsApi.PODAFFINITYTYPE_PREFERRED:
			outputTypeValue = convutils.Ref(PodAffinityPreferredValue)
		}

		return setter(outputTypeValue, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeTolerations(prefix string, src []workloadsApi.Toleration,
	target any, setter FnMapToTolerationSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}

		entries := make(map[string]*workloadsApi.Toleration)

		for _, toleration := range src {
			itemId := fc.ItemizedNumbering.Compose(prefix, toleration.Name.Get(), nil, false)

			if convutils.BoolIsTrue(toleration.Exclude.Get()) {
				entries[itemId] = nil
				continue
			}

			if !fc.Options.IgnoreMissingValues {
				if !convutils.HasValue(toleration.Key) {
					return workloadsApi.NewError(http.StatusBadRequest, "tolerations: key must be provided")
				}

				if !convutils.HasValue(toleration.Operator) {
					return workloadsApi.NewError(http.StatusBadRequest, "tolerations: operator must be provided")
				}

				if !convutils.HasValue(toleration.Effect) {
					return workloadsApi.NewError(http.StatusBadRequest, "tolerations: effect must be provided")
				}

				if *toleration.Operator.Get() == workloadsApi.TOLERATIONOPERATOR_EQUAL && !convutils.HasValue(toleration.Value) {
					return workloadsApi.NewError(http.StatusBadRequest, "tolerations: value must be provided when operator is set to Equal")
				}

				if *toleration.Operator.Get() == workloadsApi.TOLERATIONOPERATOR_EXISTS && convutils.HasValue(toleration.Value) {
					return workloadsApi.NewError(http.StatusBadRequest, "tolerations: value is not allowed when operator is set to Exists")
				}

				if *toleration.Effect.Get() != workloadsApi.TOLERATIONEFFECT_NO_EXECUTE && convutils.HasValue(toleration.Seconds) {
					return workloadsApi.NewError(http.StatusBadRequest, "tolerations: seconds is not allowed when effect is not set to NoExecute")
				}
			}

			entries[itemId] = &toleration
		}

		return setter(entries, target, cookie)
	}
}

func (fc *FieldsComposer) ComposeReadinessProbe(src *workloadsApi.Probe,
	target any, setter FnProbeSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if src == nil {
			return nil
		}

		if !convutils.HasValue(src.Handler) {
			if !fc.Options.IgnoreMissingValues {
				return workloadsApi.NewError(http.StatusBadRequest, "probes.readiness: handler must be provided")
			}
		}

		httpGetHandler := src.Handler.Get().HttpGet

		if !fc.Options.IgnoreMissingValues && !convutils.HasValue(httpGetHandler.Port) {
			return workloadsApi.NewError(http.StatusBadRequest, "probes.readiness.handler.httpGet: port must be provided")
		}

		return setter(src, target, cookie)
	}
}

var workspacePriorityClasses = datastructs.NewSet[string](
	policyConstants.PriorityClassBuild,
	policyConstants.PriorityClassInteractivePreemptible)

var trainingPriorityClasses = datastructs.NewSet[string](
	policyConstants.PriorityClassTrain,
	policyConstants.PriorityClassBuild,
	policyConstants.PriorityClassInteractivePreemptible)

var supportedPriorityClass = map[string]datastructs.Set[string]{ // working with strings because it is easier to format to error message
	Interactive: workspacePriorityClasses,
	Workspace:   workspacePriorityClasses,
	Training:    trainingPriorityClasses,
	Distributed: {}, // the field itself does not appear in the spec of distributed, put this here just to clarify it
	Inference:   {},

	constants.KindInteractivePolicy: workspacePriorityClasses,
	constants.KindTrainingPolicy:    trainingPriorityClasses,
	constants.KindDistributedPolicy: {},
	constants.KindInferencePolicy:   {},

	constants.KindInteractiveWorkload: workspacePriorityClasses,
	constants.KindTrainingWorkload:    trainingPriorityClasses,
	constants.KindDistributedWorkload: {},
	constants.KindInferenceWorkload:   {},
}

func ValidateStaticPriorityClass(priorityClass string, workloadType string) *workloadsApi.Error {

	supportedPriorityClasses := supportedPriorityClass[workloadType]

	if !supportedPriorityClasses.Has(priorityClass) {
		supportedClasses := maps.Keys(supportedPriorityClasses)
		if len(supportedClasses) == 0 {
			return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("priority class is not supported for %s", workloadType))
		}

		sort.Strings(supportedClasses)
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("priority class for %s is limited to: %s",
			workloadType, strings.Join(supportedClasses, ", ")))
	}

	return nil
}

func (itc *FieldsComposer) ComposeEnvVarsUsingCreds(src []EnvVarInfo,
	target any, setter FnEnvVarSetter, cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		if len(src) == 0 {
			return nil
		}
		return setter(src, target, cookie)
	}
}
