/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	sm "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/system_mapping"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
)

// wipeCachesAfterSystemSeed drops the entire breakdown cache + dependency
// tables. used by the system seeding paths whenever any system-scoped
// component (mapping, field group, formatter) is created or updated: any
// cached breakdown could now be stale, so we rebuild from scratch on next read.
//
// since this runs at service startup, an error here means the cache is in an
// undefined state — fail fast rather than continue with potentially stale data.
func wipeCachesAfterSystemSeed(ctx context.Context, what string) {
	if err := dao.WipeAllEntityCaches(ctx); err != nil {
		logging.Fatalf("failed to wipe breakdown caches after seeding %s: %v", what, err)
	}
}

const (
	systemMappingActor = "run:ai"

	// NOTE! Change the version here whenever you want to update the system element
	// v4/v3/v2 -> enriched argsInfo with all $ARGS.named references
	systemPodSpecFormatterDesc             = "formatter for podSpec fields (v6)"
	systemPodSpecFgDesc                    = "podSpec fields (v4)"
	systemPodSpecContainerFgDesc           = "podSpecContainer fields (v3)"
	systemPodSpecAndContainerFgDesc        = "podSpecAndContainer fields (v4)"
	systemJobSpecFormatterDesc             = "formatter for jobSpec fields (v3)"
	systemJobSpecFgDesc                    = "jobSpec fields (v3)"
	systemSelectorFormatterDesc            = "formatter for selector fields (v4)"
	systemSelectorFgDesc                   = "selector fields (v4)"
	systemMetadataFormatterDesc            = "formatter for metadata fields (v3)"
	systemDeploymentSpecFormatterDesc      = "formatter for deploymentSpec fields (v4)"
	systemDeploymentSpecFgDesc             = "deploymentSpec fields (v4)"
	systemReplicaSetSpecFgDesc             = "replicaSetSpec fields (v2)"
	systemStatefulSetSpecFormatterDesc     = "formatter for statefulSetSpec fields (v5)"
	systemStatefulSetSpecFgDesc            = "statefulSetSpec fields (v5)"
	systemCronJobSpecFormatterDesc         = "formatter for cronJobSpec fields (v2)"
	systemCronJobSpecFgDesc                = "cronJobSpec fields (v2)"
	systemMetadataFgDesc                   = "metadata fields (v5)"
	systemNimCacheSpecFormatterDesc        = "formatter for nimCacheSpec fields (v2)"
	systemNimCacheSpecFgDesc               = "nimCacheSpec fields (v2)"
	systemNimCacheContainerFgDesc          = "nimCacheContainer fields (v2)"
	systemNimServiceSpecFormatterDesc      = "formatter for nimServiceSpec fields (v2)"
	systemNimServiceSpecFgDesc             = "nimServiceSpec fields (v2)"
	systemNimServiceContainerFgDesc        = "nimServiceContainer fields (v3)"
	systemKubeflowRunPolicyFormatterDesc   = "formatter for kubeflow runPolicy fields (v2)"
	systemKubeflowReplicaSpecFormatterDesc = "formatter for kubeflow replicaSpec fields (v2)"
	systemKubeflowRunPolicyFgDesc          = "kubeflow runPolicy fields (v2)"
	systemKubeflowReplicaSpecFgDesc        = "kubeflow replicaSpec fields (v2)"
	systemPyTorchJobSpecFormatterDesc      = "formatter for pytorchJobSpec fields (v2)"
	systemPyTorchJobSpecFgDesc             = "pytorchJobSpec fields (v2)"
	systemMPIJobSpecFgDesc                 = "mpiJobSpec fields (v2)"

	// System workload mappings: logical name for logs; bump description to re-seed spec selectors.
	systemDeploymentMappingName      = "apps/Deployment/v1"
	systemDeploymentMappingDesc      = "apps/Deployment/v1 mapping (v7)"
	systemBatchJobMappingName        = "batch/Job/v1"
	systemBatchJobMappingDesc        = "batch/Job/v1 mapping (v6)"
	systemAppsReplicaSetMappingName  = "apps/ReplicaSet/v1"
	systemAppsReplicaSetMappingDesc  = "apps/ReplicaSet/v1 mapping (v6)"
	systemAppsStatefulSetMappingName = "apps/StatefulSet/v1"
	systemAppsStatefulSetMappingDesc = "apps/StatefulSet/v1 mapping (v6)"
	systemBatchCronJobMappingName    = "batch/CronJob/v1"
	systemBatchCronJobMappingDesc    = "batch/CronJob/v1 mapping (v6)"
	systemCorePodMappingName         = "core/Pod/v1"
	systemCorePodMappingDesc         = "core/Pod/v1 mapping (v4)"
	systemNimCacheMappingName        = "apps.nvidia.com/NIMCache/v1alpha1"
	systemNimCacheMappingDesc        = "apps.nvidia.com/NIMCache/v1alpha1 mapping (v4)"
	systemNimServiceMappingName      = "apps.nvidia.com/NIMService/v1alpha1"
	systemNimServiceMappingDesc      = "apps.nvidia.com/NIMService/v1alpha1 mapping (v4)"
	systemPyTorchJobMappingName      = "kubeflow.org/PyTorchJob/v1"
	systemPyTorchJobMappingDesc      = "kubeflow.org/PyTorchJob/v1 mapping (v4)"
	systemMPIJobMappingName          = "kubeflow.org/MPIJob/v2beta1"
	systemMPIJobMappingDesc          = "kubeflow.org/MPIJob/v2beta1 mapping (v4)"
	systemTFJobMappingName           = "kubeflow.org/TFJob/v1"
	systemTFJobMappingDesc           = "kubeflow.org/TFJob/v1 mapping (v4)"
	systemXGBoostJobMappingName      = "kubeflow.org/XGBoostJob/v1"
	systemXGBoostJobMappingDesc      = "kubeflow.org/XGBoostJob/v1 mapping (v4)"
	systemJAXJobMappingName          = "kubeflow.org/JAXJob/v1"
	systemJAXJobMappingDesc          = "kubeflow.org/JAXJob/v1 mapping (v4)"

	systemLeaderWorkerSetSpecFormatterDesc = "formatter for leaderWorkerSetSpec fields (v3)"
	systemLeaderWorkerSetSpecFgDesc        = "leaderWorkerSetSpec fields (v2)"
	systemLeaderWorkerTemplateFgDesc       = "leaderWorkerTemplate fields (v2)"
	systemLeaderWorkerSetMappingName       = "leaderworkerset.x-k8s.io/LeaderWorkerSet/v1"
	systemLeaderWorkerSetMappingDesc       = "leaderworkerset.x-k8s.io/LeaderWorkerSet/v1 mapping (v4)"

	systemDynamoGraphDeploymentFormatterDesc = "formatter for dynamoGraphDeploymentSpec fields (v2)"
	systemDynamoGraphDeploymentSpecFgDesc    = "dynamoGraphDeploymentSpec fields (v2)"
	systemDynamoServiceSpecFgDesc            = "dynamoServiceSpec fields (v2)"
	systemDynamoGraphDeploymentMappingName   = "nvidia.com/DynamoGraphDeployment/v1alpha1"
	systemDynamoGraphDeploymentMappingDesc   = "nvidia.com/DynamoGraphDeployment/v1alpha1 mapping (v4)"

	systemRayClusterSpecFgDesc        = "rayClusterSpec fields (v2)"
	systemRayClusterSpecFormatterDesc = "formatter for rayClusterSpec fields (v2)"
	systemRayWorkerGroupSpecFgDesc    = "rayWorkerGroupSpec fields (v2)"
	systemRayJobSpecFgDesc            = "rayJobSpec fields (v2)"
	systemRayJobSpecFormatterDesc     = "formatter for rayJobSpec fields (v2)"
	systemRayServiceSpecFgDesc        = "rayServiceSpec fields (v2)"
	systemRayServiceSpecFormatterDesc = "formatter for rayServiceSpec fields (v2)"
	systemRayClusterMappingName       = "ray.io/RayCluster/v1"
	systemRayClusterMappingDesc       = "ray.io/RayCluster/v1 mapping (v4)"
	systemRayJobMappingName           = "ray.io/RayJob/v1"
	systemRayJobMappingDesc           = "ray.io/RayJob/v1 mapping (v4)"
	systemRayServiceMappingName       = "ray.io/RayService/v1"
	systemRayServiceMappingDesc       = "ray.io/RayService/v1 mapping (v4)"

	systemKnativeServiceMappingName = "serving.knative.dev/Service/v1"
	systemKnativeServiceMappingDesc = "serving.knative.dev/Service/v1 mapping (v3)"

	// Native policy spec mappings (fictive GVKs)
	systemNativeSpecFormatterDesc     = "formatter for nativeSpec fields (v2)"
	systemNativeSpecFgDesc            = "nativeSpec fields (v7)"
	systemNativeDistributedSpecFgDesc = "nativeDistributedSpec fields (v1)"
	systemRunAiWorkspaceMappingName   = "run.ai/Workspace/v1"
	systemRunAiWorkspaceMappingDesc   = "run.ai/Workspace/v1 mapping (v1)"
	systemRunAiDistributedMappingName = "run.ai/Distributed/v1"
	systemRunAiDistributedMappingDesc = "run.ai/Distributed/v1 mapping (v1)"
)

// Type aliases for the system_mapping package types.
func HandleSystemMapping() {
	ctx := auth.MockAuthContext(systemMappingActor, 0)
	db := runaidb.DB().WithContext(ctx)
	ms := MappingApiService{}
	cv := NewMappingConsistencyValidator(ctx, api.MAPPINGSCOPE_SYSTEM)

	handleSystemFormatters(ctx, ms, cv)

	handleSystemFieldGroups(ctx, db, ms, cv)

	handleSystemMappings(ctx, db, ms, cv)
}

func handleSystemFormatters(ctx context.Context, ms MappingApiService, _ *MappingConsistencyValidator) {
	logging.Infof("handle system mapping formatters")

	handleSystemFormatter(ctx, ms, sm.SystemPodSpecFormatterName, systemPodSpecFormatterDesc, sm.BuildSystemPodSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemJobSpecFormatterName, systemJobSpecFormatterDesc, sm.BuildSystemJobSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemSelectorFormatterName, systemSelectorFormatterDesc, sm.BuildSystemSelectorFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemMetadataFormatterName, systemMetadataFormatterDesc, sm.BuildSystemMetadataFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemDeploymentSpecFormatterName, systemDeploymentSpecFormatterDesc, sm.BuildSystemDeploymentSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemStatefulSetSpecFormatterName, systemStatefulSetSpecFormatterDesc, sm.BuildSystemStatefulSetSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemCronJobSpecFormatterName, systemCronJobSpecFormatterDesc, sm.BuildSystemCronJobSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemNimCacheSpecFormatterName, systemNimCacheSpecFormatterDesc, sm.BuildSystemNimCacheSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemNimServiceSpecFormatterName, systemNimServiceSpecFormatterDesc, sm.BuildSystemNimServiceSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemKubeflowRunPolicyFormatterName, systemKubeflowRunPolicyFormatterDesc, sm.BuildSystemKubeflowRunPolicyFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemKubeflowReplicaSpecFormatterName, systemKubeflowReplicaSpecFormatterDesc, sm.BuildSystemKubeflowReplicaSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemPyTorchJobSpecFormatterName, systemPyTorchJobSpecFormatterDesc, sm.BuildSystemPyTorchJobSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemLeaderWorkerSetFormatterName, systemLeaderWorkerSetSpecFormatterDesc, sm.BuildSystemLeaderWorkerSetSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemDynamoGraphDeploymentFormatterName, systemDynamoGraphDeploymentFormatterDesc, sm.BuildSystemDynamoGraphDeploymentSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemRayClusterSpecFormatterName, systemRayClusterSpecFormatterDesc, sm.BuildSystemRayClusterSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemRayJobSpecFormatterName, systemRayJobSpecFormatterDesc, sm.BuildSystemRayJobSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemRayServiceSpecFormatterName, systemRayServiceSpecFormatterDesc, sm.BuildSystemRayServiceSpecFormatterFields)
	handleSystemFormatter(ctx, ms, sm.SystemNativeSpecFormatterName, systemNativeSpecFormatterDesc, sm.BuildSystemNativeSpecFormatterFields)
}

type systemFieldGroupSeed struct {
	name        string
	description string
	builder     sm.SystemFieldGroupBuilder
}

var systemFieldGroupSeeds = []systemFieldGroupSeed{
	{sm.SystemPodSpecFgName, systemPodSpecFgDesc, sm.BuildSystemPodSpecFg},
	{sm.SystemPodSpecContainerFgName, systemPodSpecContainerFgDesc, sm.BuildSystemPodSpecContainerFg},
	{sm.SystemPodSpecAndContainerFgName, systemPodSpecAndContainerFgDesc, sm.BuildSystemPodSpecAndContainerFg},
	{sm.SystemJobSpecFgName, systemJobSpecFgDesc, sm.BuildSystemJobSpecFg},
	{sm.SystemSelectorFgName, systemSelectorFgDesc, sm.BuildSystemSelectorFg},
	{sm.SystemDeploymentSpecFgName, systemDeploymentSpecFgDesc, sm.BuildSystemDeploymentSpecFg},
	{sm.SystemReplicaSetSpecFgName, systemReplicaSetSpecFgDesc, sm.BuildSystemReplicaSetSpecFg},
	{sm.SystemStatefulSetSpecFgName, systemStatefulSetSpecFgDesc, sm.BuildSystemStatefulSetSpecFg},
	{sm.SystemCronJobSpecFgName, systemCronJobSpecFgDesc, sm.BuildSystemCronJobSpecFg},
	{sm.SystemMetadataFgName, systemMetadataFgDesc, sm.BuildSystemMetadataFg},
	{sm.SystemNimCacheSpecFgName, systemNimCacheSpecFgDesc, sm.BuildSystemNimCacheSpecFg},
	{sm.SystemNimCacheContainerFgName, systemNimCacheContainerFgDesc, sm.BuildSystemNimCacheContainerFg},
	{sm.SystemNimServiceSpecFgName, systemNimServiceSpecFgDesc, sm.BuildSystemNimServiceSpecFg},
	{sm.SystemNimServiceContainerFgName, systemNimServiceContainerFgDesc, sm.BuildSystemNimServiceContainerFg},
	{sm.SystemKubeflowRunPolicyFgName, systemKubeflowRunPolicyFgDesc, sm.BuildSystemKubeflowRunPolicyFg},
	{sm.SystemKubeflowReplicaSpecFgName, systemKubeflowReplicaSpecFgDesc, sm.BuildSystemKubeflowReplicaSpecFg},
	{sm.SystemPyTorchJobSpecFgName, systemPyTorchJobSpecFgDesc, sm.BuildSystemPyTorchJobSpecFg},
	{sm.SystemMPIJobSpecFgName, systemMPIJobSpecFgDesc, sm.BuildSystemMPIJobSpecFg},
	{sm.SystemLeaderWorkerSetSpecFgName, systemLeaderWorkerSetSpecFgDesc, sm.BuildSystemLeaderWorkerSetSpecFg},
	{sm.SystemLeaderWorkerTemplateFgName, systemLeaderWorkerTemplateFgDesc, sm.BuildSystemLeaderWorkerTemplateFg},
	{sm.SystemDynamoGraphDeploymentSpecFgName, systemDynamoGraphDeploymentSpecFgDesc, sm.BuildSystemDynamoGraphDeploymentSpecFg},
	{sm.SystemDynamoServiceSpecFgName, systemDynamoServiceSpecFgDesc, sm.BuildSystemDynamoServiceSpecFg},
	{sm.SystemRayClusterSpecFgName, systemRayClusterSpecFgDesc, sm.BuildSystemRayClusterSpecFg},
	{sm.SystemRayWorkerGroupSpecFgName, systemRayWorkerGroupSpecFgDesc, sm.BuildSystemRayWorkerGroupSpecFg},
	{sm.SystemRayJobSpecFgName, systemRayJobSpecFgDesc, sm.BuildSystemRayJobSpecFg},
	{sm.SystemRayServiceSpecFgName, systemRayServiceSpecFgDesc, sm.BuildSystemRayServiceSpecFg},
	{sm.SystemNativeSpecFgName, systemNativeSpecFgDesc, sm.BuildSystemNativeSpecFg},
	{sm.SystemNativeDistributedSpecFgName, systemNativeDistributedSpecFgDesc, sm.BuildSystemNativeDistributedSpecFg},
}

func handleSystemFieldGroups(ctx context.Context, db orm.DB, ms MappingApiService, cv *MappingConsistencyValidator) {
	logging.Infof("handle system mapping field groups")

	// Every outdated group is pruned before any group is merged: a key renamed or moved between groups that
	// share a spec selector would otherwise exist in both and fail the consistency check.
	for _, seed := range systemFieldGroupSeeds {
		pruneSystemFieldGroup(ctx, db, ms, seed.name, seed.description, seed.builder)
	}

	for _, seed := range systemFieldGroupSeeds {
		handleSystemFieldGroup(ctx, db, ms, seed.name, seed.description, seed.builder, cv)
	}
}

func handleSystemMappings(ctx context.Context, db orm.DB, ms MappingApiService, cv *MappingConsistencyValidator) {
	logging.Infof("handle system mapping and spec selectors")

	handleSystemWorkloadMapping(ctx, db, ms, systemDeploymentMappingName, systemDeploymentMappingDesc, sm.BuildSystemAppsDeploymentV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemBatchJobMappingName, systemBatchJobMappingDesc, sm.BuildSystemBatchJobV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemAppsReplicaSetMappingName, systemAppsReplicaSetMappingDesc, sm.BuildSystemAppsReplicaSetV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemAppsStatefulSetMappingName, systemAppsStatefulSetMappingDesc, sm.BuildSystemAppsStatefulSetV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemBatchCronJobMappingName, systemBatchCronJobMappingDesc, sm.BuildSystemBatchCronJobV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemCorePodMappingName, systemCorePodMappingDesc, sm.BuildSystemCorePodV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemNimCacheMappingName, systemNimCacheMappingDesc, sm.BuildSystemAppsNvidiaComNimCacheV1alpha1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemNimServiceMappingName, systemNimServiceMappingDesc, sm.BuildSystemAppsNvidiaComNimServiceV1alpha1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemPyTorchJobMappingName, systemPyTorchJobMappingDesc, sm.BuildSystemKubeflowOrgPyTorchJobV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemMPIJobMappingName, systemMPIJobMappingDesc, sm.BuildSystemKubeflowOrgMPIJobV2beta1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemTFJobMappingName, systemTFJobMappingDesc, sm.BuildSystemKubeflowOrgTFJobV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemXGBoostJobMappingName, systemXGBoostJobMappingDesc, sm.BuildSystemKubeflowOrgXGBoostJobV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemJAXJobMappingName, systemJAXJobMappingDesc, sm.BuildSystemKubeflowOrgJAXJobV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemLeaderWorkerSetMappingName, systemLeaderWorkerSetMappingDesc, sm.BuildSystemLeaderWorkerSetXK8sIoV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemDynamoGraphDeploymentMappingName, systemDynamoGraphDeploymentMappingDesc, sm.BuildSystemNvidiaComDynamoGraphDeploymentV1alpha1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemRayClusterMappingName, systemRayClusterMappingDesc, sm.BuildSystemRayIoRayClusterV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemRayJobMappingName, systemRayJobMappingDesc, sm.BuildSystemRayIoRayJobV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemRayServiceMappingName, systemRayServiceMappingDesc, sm.BuildSystemRayIoRayServiceV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemKnativeServiceMappingName, systemKnativeServiceMappingDesc, sm.BuildSystemServingKnativeDevServiceV1Mapping, cv)

	// Native policy spec mappings (fictive GVKs)
	handleSystemWorkloadMapping(ctx, db, ms, systemRunAiWorkspaceMappingName, systemRunAiWorkspaceMappingDesc, sm.BuildSystemRunAiWorkspaceV1Mapping, cv)
	handleSystemWorkloadMapping(ctx, db, ms, systemRunAiDistributedMappingName, systemRunAiDistributedMappingDesc, sm.BuildSystemRunAiDistributedV1Mapping, cv)
}

// handleSystemWorkloadMapping ensures a system-scoped mapping row exists for the workload GVK returned by builder.
// mappingName is used only in log lines (DB identity is GVK).
func handleSystemWorkloadMapping(ctx context.Context, db orm.DB, ms MappingApiService, mappingName, wantDescription string,
	builder sm.SystemWorkloadMappingBuilder, cv *MappingConsistencyValidator,
) {
	if builder == nil {
		logging.Errorf("nil builder for system workload mapping %s", mappingName)
		return
	}

	gvk, specSelectorsAPI := builder()
	logging.Infof("handle system workload mapping %s (%s/%s/%s)", mappingName, convutils.Deref(gvk.Group, ""), gvk.Kind, gvk.Version)

	systemRow, err := dao.GetMappingByGvkAndScope(ctx, gvk, api.MAPPINGSCOPE_SYSTEM)
	if err != nil {
		if err.Code != http.StatusNotFound {
			logging.Errorf("failed to load system workload mapping %s: %v", mappingName, err)
			return
		}

		built := addScopeToSpecSelectors(specSelectorsAPI, api.MAPPINGSCOPE_SYSTEM)
		if vErr := validateSpecSelectorMapKeys(built); vErr != nil {
			logging.Errorf("system workload mapping %s seed validate keys: %v", mappingName, vErr)
			return
		}
		if vErr := validateSpecSelectorJQExprs(built); vErr != nil {
			logging.Errorf("system workload mapping %s seed validate JQ: %v", mappingName, vErr)
			return
		}

		now := time.Now().UTC()
		desc := wantDescription
		row := &dbmodel.MappingSelectors{
			ID:            uuid.New(),
			GvkGroup:      convutils.Deref(gvk.Group, ""),
			GvkKind:       gvk.Kind,
			GvkVersion:    gvk.Version,
			Description:   &desc,
			ScopeType:     api.MAPPINGSCOPE_SYSTEM,
			TenantId:      0,
			SpecSelectors: built,
			CreatedAt:     now,
			CreatedBy:     systemMappingActor,
			UpdatedAt:     now,
			UpdatedBy:     systemMappingActor,
		}

		if _, cerr := ms.createMapping(ctx, row, cv); cerr != nil {
			if cerr.Code != http.StatusConflict {
				logging.Errorf("failed to create system workload mapping %s: %v", mappingName, cerr)
				return
			}
			systemRow, err = dao.GetMappingByGvkAndScope(ctx, gvk, api.MAPPINGSCOPE_SYSTEM)
			if err != nil {
				logging.Errorf("failed to reload system workload mapping %s after conflict: %v", mappingName, err)
				return
			}
		} else {
			logging.Infof("system workload mapping %s created", mappingName)
			return
		}
	}

	if systemRow.Description != nil && *systemRow.Description == wantDescription {
		logging.Infof("system workload mapping %s: description unchanged (%q), skipping seed", mappingName, wantDescription)
		return
	}

	built := addScopeToSpecSelectors(specSelectorsAPI, api.MAPPINGSCOPE_SYSTEM)
	if vErr := validateSpecSelectorMapKeys(built); vErr != nil {
		logging.Errorf("system workload mapping %s seed validate keys: %v", mappingName, vErr)
		return
	}
	if vErr := validateSpecSelectorJQExprs(built); vErr != nil {
		logging.Errorf("system workload mapping %s seed validate JQ: %v", mappingName, vErr)
		return
	}
	if vErr := cv.ValidateMappingChange(ctx, db, built); vErr != nil {
		logging.Errorf("system workload mapping %s consistency: %v", mappingName, vErr)
		return
	}

	desc := wantDescription
	systemRow.Description = &desc
	systemRow.SpecSelectors = built
	systemRow.UpdatedAt = time.Now().UTC()
	systemRow.UpdatedBy = systemMappingActor

	if sErr := dao.SaveMapping(ctx, runaidb.DB().WithContext(ctx), systemRow); sErr != nil {
		logging.Errorf("seed system workload mapping %s: %v", mappingName, sErr)
		return
	}

	wipeCachesAfterSystemSeed(ctx, "system workload mapping "+mappingName)

	logging.Infof("system workload mapping %s updated with %s", mappingName, wantDescription)
}

// handleSystemFormatter ensures a system formatter row exists. If the stored description already matches
// wantDescription, it returns without building or merging fields. Otherwise it replaces the formatter fields
// with buildFields(), sets description to wantDescription, and saves.
func handleSystemFormatter(ctx context.Context, ms MappingApiService, formatterName, wantDescription string, buildFields sm.SystemFormatterFieldsBuilder) {
	if buildFields == nil {
		logging.Errorf("handle update of system formatter %s", formatterName)
		return
	}

	formatter, err := ms.getFormatter(ctx, formatterName)
	if err != nil {
		if err.Code == http.StatusNotFound {
			formatter, err = ms.createFormatter(ctx, newEmptySystemFormatterRow(formatterName))
			if err != nil {
				// conflict right after get-not-found means that there is another policy service pod running
				// in parallel who already created this as part of first-time-flow, and in this case we should
				// let the other policy service pod fill the formatter it created
				if err.Code == http.StatusConflict {
					formatter, err = ms.getFormatter(ctx, formatterName)
					if err != nil {
						logging.Errorf("failed to reload system formatter %s after conflict: %v", formatterName, err)
						return
					}
				} else {
					logging.Errorf("failed to create system formatter %s: %v", formatterName, err)
					return
				}
			}
		} else {
			logging.Errorf("failed to load system formatter %s: %v", formatterName, err)
			return
		}
	}

	if formatter.Description != nil && *formatter.Description == wantDescription {
		logging.Infof("system formatter %s: description unchanged (%q), skipping seed", formatterName, wantDescription)
		return
	}

	fields := buildFields()

	if stale := keysMissingFrom(formatter.Fields, fields); len(stale) > 0 {
		if err := ms.deleteFormatterFields(ctx, formatterName, stale); err != nil {
			logging.Errorf("prune system formatter %s: %v", formatterName, err)
			return
		}
		logging.Infof("system formatter %s: removed fields %v", formatterName, stale)
	}

	updateReq := api.FormatterUpdateRequest{
		Description: &wantDescription,
		Fields:      fields,
	}

	if _, err := ms.patchFormatter(ctx, formatterName, &updateReq); err != nil {
		logging.Errorf("seed system formatter %s: %v", formatterName, err)
		return
	}

	wipeCachesAfterSystemSeed(ctx, "system formatter "+formatterName)

	logging.Infof("system formatter %s: updated with %s", formatterName, wantDescription)
}

func newEmptySystemFormatterRow(name string) *dbmodel.MappingFormatters {
	now := time.Now().UTC()

	return &dbmodel.MappingFormatters{
		ID:        uuid.New(),
		Name:      name,
		ScopeType: api.MAPPINGSCOPE_SYSTEM,
		Fields:    map[string]api.FormatterField{},

		CreatedAt: now,
		CreatedBy: systemMappingActor,
		UpdatedAt: now,
		UpdatedBy: systemMappingActor,
	}
}

// pruneSystemFieldGroup removes from an outdated system field group the fields and item selectors that builder no
// longer defines. Kept fields that reference a removed item selector are removed too; handleSystemFieldGroup
// re-adds them from builder.
func pruneSystemFieldGroup(ctx context.Context, db orm.DB, ms MappingApiService, fieldGroupName, wantDescription string,
	builder sm.SystemFieldGroupBuilder,
) {
	if builder == nil {
		return
	}

	systemRow, err := dao.GetFieldGroupByNameAndScope(ctx, db, fieldGroupName, api.MAPPINGSCOPE_SYSTEM)
	if err != nil {
		if err.Code != http.StatusNotFound {
			logging.Errorf("failed to load system field group %s for pruning: %v", fieldGroupName, err)
		}
		return
	}

	if systemRow.Description != nil && *systemRow.Description == wantDescription {
		return
	}

	fields, itemSelectors := builder()
	staleItemSelectors := keysMissingFrom(systemRow.ItemSelectors, itemSelectors)
	staleFields := keysMissingFrom(systemRow.Fields, fields)
	for name, field := range systemRow.Fields {
		if _, kept := fields[name]; kept && slices.ContainsFunc(field.ItemSelectors, func(sel string) bool {
			return slices.Contains(staleItemSelectors, sel)
		}) {
			staleFields = append(staleFields, name)
		}
	}
	slices.Sort(staleFields)

	if len(staleFields) == 0 && len(staleItemSelectors) == 0 {
		return
	}

	if len(staleFields) > 0 {
		req := api.FieldGroupDeleteFieldsRequest{Fields: staleFields}
		if _, err := ms.deleteFieldGroupFields(ctx, fieldGroupName, api.MAPPINGSCOPE_SYSTEM, &req); err != nil {
			logging.Errorf("prune system field group %s fields: %v", fieldGroupName, err)
			return
		}
	}

	if len(staleItemSelectors) > 0 {
		req := api.FieldGroupDeleteItemSelectorsRequest{ItemSelectors: staleItemSelectors}
		if _, err := ms.deleteFieldGroupItemSelectors(ctx, fieldGroupName, api.MAPPINGSCOPE_SYSTEM, &req); err != nil {
			logging.Errorf("prune system field group %s item selectors: %v", fieldGroupName, err)
			return
		}
	}

	wipeCachesAfterSystemSeed(ctx, "system field group "+fieldGroupName)

	logging.Infof("system field group %s: removed fields %v and item selectors %v", fieldGroupName, staleFields, staleItemSelectors)
}

// keysMissingFrom returns the keys of stored that built does not define, sorted.
func keysMissingFrom[S, B any](stored map[string]S, built map[string]B) []string {
	var missing []string
	for key := range stored {
		if _, ok := built[key]; !ok {
			missing = append(missing, key)
		}
	}
	slices.Sort(missing)

	return missing
}

func handleSystemFieldGroup(ctx context.Context, db orm.DB, ms MappingApiService, fieldGroupName, wantDescription string,
	builder sm.SystemFieldGroupBuilder, cv *MappingConsistencyValidator,
) {
	if builder == nil {
		logging.Errorf("nil builder for system field group %s", fieldGroupName)
		return
	}

	systemRow, err := dao.GetFieldGroupByNameAndScope(ctx, db, fieldGroupName, api.MAPPINGSCOPE_SYSTEM)
	if err != nil {
		if err.Code == http.StatusNotFound {
			if _, cerr := ms.createFieldGroup(ctx, newEmptySystemFieldGroupRow(fieldGroupName)); cerr != nil {
				if cerr.Code == http.StatusConflict {
					systemRow, err = dao.GetFieldGroupByNameAndScope(ctx, db, fieldGroupName, api.MAPPINGSCOPE_SYSTEM)
					if err != nil {
						logging.Errorf("failed to reload system field group %s after conflict: %v", fieldGroupName, err)
						return
					}
				} else {
					logging.Errorf("failed to create system field group %s: %v", fieldGroupName, cerr)
					return
				}
			} else {
				systemRow, err = dao.GetFieldGroupByNameAndScope(ctx, db, fieldGroupName, api.MAPPINGSCOPE_SYSTEM)
				if err != nil {
					logging.Errorf("failed to reload system field group %s after create: %v", fieldGroupName, err)
					return
				}
			}
		} else {
			logging.Errorf("failed to load system field group %s: %v", fieldGroupName, err)
			return
		}
	}

	if systemRow.Description != nil && *systemRow.Description == wantDescription {
		logging.Infof("system field group %s: description unchanged (%q), skipping seed", fieldGroupName, wantDescription)
		return
	}

	fields, itemSelectors := builder()

	updateReq := api.FieldGroupUpdateRequest{
		Description:   &wantDescription,
		Fields:        fields,
		ItemSelectors: itemSelectors,
	}

	if _, err := ms.patchFieldGroup(ctx, fieldGroupName, api.MAPPINGSCOPE_SYSTEM, &updateReq, cv); err != nil {
		logging.Errorf("seed system field group %s: %v", fieldGroupName, err)
		return
	}

	wipeCachesAfterSystemSeed(ctx, "system field group "+fieldGroupName)

	logging.Infof("system field group %s: updated with %s", fieldGroupName, wantDescription)
}

func newEmptySystemFieldGroupRow(name string) *dbmodel.MappingFieldGroups {
	now := time.Now().UTC()

	return &dbmodel.MappingFieldGroups{
		ID:            uuid.New(),
		Name:          name,
		ScopeType:     api.MAPPINGSCOPE_SYSTEM,
		Fields:        map[string]api.FieldGroupFieldsWithScopeValue{},
		ItemSelectors: map[string]api.ItemSelectorsWithScopeValue{},

		CreatedAt: now,
		CreatedBy: systemMappingActor,
		UpdatedAt: now,
		UpdatedBy: systemMappingActor,
	}
}
