/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { config, settings, flowRootState } from './config/config';
import { supportedPlatforms, gpuTypes } from './config/types';
import type { NodeResources, User } from './config/types';
import { logger } from './logger/logger';
import { InfoUtils } from './utils/InfoUtils';
import { ObjectGenerator } from './ObjectGenerator';
import { ApiUtils } from './utils/ApiUtils';
import { Utils } from './utils/utils';
import { AuthUtils } from './utils/AuthUtils';
import { UpgradeUtils } from './utils/UpgradeUtils';
import { WorkloadUtils } from './utils/WorkloadUtils';
import { SchedulingUtils } from './utils/SchedulingUtils';
import { CliError, BackendApiError, UtilsError } from './error';
import { tags } from './config/tags';
import { Operations } from './Operations';
import { NodePoolUtils } from './utils/NodePoolUtils';
import { MetricUtils } from './utils/MetricUtils';
import { DepartmentUtils } from './utils/DepartmentUtils';
import { AssignedResourcesUtils } from './utils/AssignedResourcesUtils';
import { crds, CustomResourceDefinition } from './models/customResourceDefinition';
import { WaitUtils } from './utils/WaitUtils';
import { HelmReleaseUtils } from './utils/HelmReleaseUtils';
import { AiApplicationUtils } from './utils/AiApplicationUtils';
import { BaseWorkload, WorkloadConditionType } from './models/baseWorkload';
import { Preemptibility } from '../api/apis/assetsApi/generated/model/preemptibility';
import { InferenceWorkload } from './models/inferenceWorkload';
import { TrainingWorkload } from './models/trainingWorkload';
import { InteractiveWorkload } from './models/interactiveWorkload';
import { DistributedJobType, DistributedWorkload } from './models/distributedWorkload';
import { PodStatus } from './models/pods';
import { ProjectUtils } from './utils/ProjectUtils';
import { ProjectResourcesUtils } from './utils/ProjectResourcesUtils';
import { TokenObject } from './models/tokenObject';
import { InferenceClientCreator } from './utils/DeploymentUtils';
import { ProjectsAndDepartments } from './ProjectsAndDepartments';
import { UpgradeClusterUtils } from './utils/upgradeClusterUtils';
import { RunaiHealthChecker } from './RunaiHealthChecker';
import { UpgradesHealthChecker } from './UpgradesHealthChecker';
import { WorkspaceUtils } from './utils/WorkspaceUtils';
import { TemplateUtils } from './utils/TemplateUtils';
import { EnvironmentUtils, isOCP } from './utils/EnvironmentUtils';
import { ComputeResourceUtils } from './utils/ComputeResourceUtils';
import { CredentialUtils } from './utils/CredentialUtils';
import { CliUtils } from './utils/CliUtils';
import { DataSourceUtils } from './utils/DataSourceUtils';
import { Keycloak } from './utils/Keycloak';
import { ClusterUtils } from './utils/ClusterUtils';
import { AuthV2Utils } from './utils/AuthV2Utils';
import { TenantUtils } from './utils/TenantUtils';
import { BffUtils } from './utils/BffUtils';
import { AuthorizationUtils } from './utils/AuthorizationUtils';
import { AccessRulesClusterUtils } from './utils/AccessRulesClusterUtils';
import { VersionUtils } from './utils/VersionUtils';
import { CrdUtils } from './utils/CrdUtils';
import { TrainingUtils } from './utils/TrainingUtils';
import { NodeUtils } from './utils/NodeUtils';
import { WorkloadsUtilsV2 } from './utils/WorkloadsUtilsV2';
import { PolicyUtils } from './utils/policyUtils';
import { JobUtils } from './utils/jobUtils';
import { TelemetryUtils } from './utils/TelemetryUtils';
import { WorkloadsCrudApiUtils } from './utils/WorkloadsCrudUtils';
import { WorkloadsManagerUtils } from './utils/WorkloadsManagerUtils';
import { HooksUtils } from './utils/HooksUtils';
import { AllureUtils } from './utils/AllureUtils';
import { KubeConfigUtils } from './utils/KubeConfigUtils';
import { ClusterService } from './utils/ClusterServiceUtils';
import { IdpUtils } from './utils/idpUtils';
import { AssetUtils } from './utils/AssetUtils';
import { StorageUtils } from './utils/StorageUtils';
import { DatavolumeUtils } from './utils/DatavolumeUtils';
import { AssetsApiV2Utils } from './utils/AssetsApiV2Utils';
import { OrgUnitUtils } from './utils/OrgUnitUtils';
import { UnleashClient } from './utils/Unleash/Unleash';
import { UsersUtils } from './utils/UsersUtils';
import { ReportApiUtils } from './utils/ReportApiUtils';
import { AuditLogsUtils } from './utils/AuditUtils';
import { MultiClusterUtils } from './utils/MultiClusterUtils';
import { ClusterApiK8sObjectUtils } from './utils/ClusterApiK8sObjectUtils';
import { ExternalWorkloadUtils } from './utils/externalWorkloadUtils';
import { InferenceUtils } from './utils/inferenceUtils';
import { DistributedInferenceUtils } from './utils/distributedInferenceUtils';
import { DistributedUtils } from './utils/distributedUtils';
import { UsersManager } from './utils/UsersManager';
import { NetworkTopologiesUtils } from './utils/NetworkTopologiesUtils';
import { Notifications } from './utils/NotificationUtils';
import { AccessKeysUtils } from './utils/AccessKey';
import { ServiceAccountUtils } from './utils/ServiceAccountUtils';
import { NgcCatalogTestUtils } from './utils/ngcCatalogTestUtils';
import { NgcCredentialVerificationUtils } from './utils/NgcCredentialVerificationUtils';
import { MnnvlUtils } from './utils/mnnvlUtils';
import { WorkloadAuthUtils } from './utils/workloadAuthUtils';
import { PersonalCredentialsUtils, PC_WLR, PC_ANNOTATION } from './utils/PersonalCredentialsUtils';
import {
    ensureControlPlaneCredentials,
    resolveControlPlaneCredentials,
    resolveKeycloakPassword
} from './utils/CredentialsResolver';
import { resolveHelmSecretRef } from './utils/HelmReleaseUtils';

const policyUtils = new PolicyUtils();
const infoUtils = new InfoUtils();
const apiUtils = new ApiUtils();
const workloadUtils = new WorkloadUtils();
const waitUtils = new WaitUtils();
const helmReleaseUtils = new HelmReleaseUtils();
const aiApplicationUtils = new AiApplicationUtils();
const objectGenerator = new ObjectGenerator();
const operations = new Operations();
const nodePoolUtils = new NodePoolUtils();
const metricUtils = new MetricUtils();
const storageUtils = new StorageUtils();
const datavolumeUtils = new DatavolumeUtils();
const departmentUtils = new DepartmentUtils();
const assignedResourcesUtils = new AssignedResourcesUtils();
const projectUtils = new ProjectUtils();
const projectResourcesUtils = new ProjectResourcesUtils();
const authUtils = new AuthUtils();
const upgradeUtils = new UpgradeUtils();
const authorizationUtils = new AuthorizationUtils();
const authV2Utils = new AuthV2Utils();
const inferenceClientCreator = new InferenceClientCreator();
const projectsAndDepartments = new ProjectsAndDepartments();
const upgradeClusterUtils = new UpgradeClusterUtils();
const runaiHealthChecker = new RunaiHealthChecker();
const upgradesHealthChecker = new UpgradesHealthChecker();
const workspaceUtils = new WorkspaceUtils();
const workspaceTemplateUtils = new TemplateUtils();
const credentialUtils = new CredentialUtils();
const environmentUtils = new EnvironmentUtils();
const computeResourceUtils = new ComputeResourceUtils();
const dataSourcesUtils = new DataSourceUtils();
const trainingUtils = new TrainingUtils();
const keycloak = new Keycloak();
const clusterUtils = new ClusterUtils();
const cliUtil = new CliUtils();
const accessRulesClusterUtils = new AccessRulesClusterUtils();
const tenantUtils = new TenantUtils();
const bffUtils = new BffUtils();
const versionUtils = new VersionUtils();
const crdUtils = new CrdUtils();
const nodeUtils = new NodeUtils();
const jobUtils = new JobUtils();
const workloadsUtilsV2 = new WorkloadsUtilsV2();
const telemetryUtils = new TelemetryUtils();
const workloadsCrudUtils = new WorkloadsCrudApiUtils();
const workloadsManagerUtils = new WorkloadsManagerUtils();
const hooksUtils = new HooksUtils();
const allureUtils = new AllureUtils();
const kubeConfigUtils = new KubeConfigUtils();
const clusterService = new ClusterService();
const idpUtils = new IdpUtils();
const assetUtils = new AssetUtils();
const assetsApiV2Utils = new AssetsApiV2Utils();
const orgUnitUtils = new OrgUnitUtils();
const unleashClient = new UnleashClient();
const usersUtils = new UsersUtils();
const reportApiUtils = new ReportApiUtils();
const auditLogsUtils = new AuditLogsUtils();
const multiClusterUtils = new MultiClusterUtils();
const externalWorkloadUtils = new ExternalWorkloadUtils();
const clusterApiK8sObjectUtils = new ClusterApiK8sObjectUtils();
const distributedUtils = new DistributedUtils();
const inferenceUtils = new InferenceUtils();
const distributedInferenceUtils = new DistributedInferenceUtils();
const ngcCatalogUtils = new NgcCatalogTestUtils();
const templateUtils = new TemplateUtils();
const usersManager = new UsersManager();
const networkTopologiesUtils = new NetworkTopologiesUtils();
const notificationsUtils = new Notifications();
const accessKeyUtils = new AccessKeysUtils();
const serviceAccountUtils = new ServiceAccountUtils();
const ngcCredentialVerificationUtils = new NgcCredentialVerificationUtils();
const mnnvlUtils = new MnnvlUtils();
const workloadAuthUtils = new WorkloadAuthUtils();
const personalCredentialsUtils = new PersonalCredentialsUtils();
// Tools
export {
    cliUtil,
    infoUtils,
    objectGenerator,
    apiUtils,
    workloadUtils,
    waitUtils,
    helmReleaseUtils,
    aiApplicationUtils,
    operations,
    nodePoolUtils,
    metricUtils,
    departmentUtils,
    assignedResourcesUtils,
    authUtils,
    upgradeUtils,
    authV2Utils,
    authorizationUtils,
    projectUtils,
    projectResourcesUtils,
    SchedulingUtils,
    Utils,
    projectsAndDepartments,
    inferenceClientCreator,
    upgradeClusterUtils,
    runaiHealthChecker,
    upgradesHealthChecker,
    workspaceUtils,
    workspaceTemplateUtils,
    credentialUtils,
    environmentUtils,
    computeResourceUtils,
    dataSourcesUtils,
    keycloak,
    clusterUtils,
    accessRulesClusterUtils,
    tenantUtils,
    bffUtils,
    versionUtils,
    crdUtils,
    trainingUtils,
    nodeUtils,
    policyUtils,
    jobUtils,
    workloadsUtilsV2,
    telemetryUtils,
    workloadsCrudUtils,
    workloadsManagerUtils,
    hooksUtils,
    allureUtils,
    kubeConfigUtils,
    clusterService,
    idpUtils,
    assetUtils,
    storageUtils,
    datavolumeUtils,
    orgUnitUtils,
    assetsApiV2Utils,
    unleashClient,
    usersUtils,
    reportApiUtils,
    auditLogsUtils,
    multiClusterUtils,
    externalWorkloadUtils,
    clusterApiK8sObjectUtils,
    distributedUtils,
    inferenceUtils,
    distributedInferenceUtils,
    ngcCatalogUtils,
    templateUtils,
    usersManager,
    networkTopologiesUtils,
    notificationsUtils,
    accessKeyUtils,
    serviceAccountUtils,
    ngcCredentialVerificationUtils,
    mnnvlUtils,
    workloadAuthUtils,
    personalCredentialsUtils
};

const TopologyAnnotationName = 'kai.scheduler/topology';
const TopologyRequiredAnnotationName = 'kai.scheduler/topology-required-placement';
const TopologyPreferredAnnotationName = 'kai.scheduler/topology-preferred-placement';

// Constants
export { TopologyAnnotationName, TopologyRequiredAnnotationName, TopologyPreferredAnnotationName };
export { PC_WLR, PC_ANNOTATION };
export {
    GPU_FRACTIONING_READY_CONDITION,
    GPU_PRODUCT_LABEL,
    NVIDIA_GPU_LABEL,
    SIMULATED_GPU_FRACTIONING_REASON
} from './utils/NodeUtils';
export type { StagedNodeCondition } from './utils/NodeUtils';

// Configutration
export {
    logger,
    config,
    settings,
    flowRootState,
    tags,
    supportedPlatforms,
    gpuTypes,
    isOCP,
    resolveControlPlaneCredentials,
    ensureControlPlaneCredentials,
    resolveKeycloakPassword,
    resolveHelmSecretRef
};
export type { NodeResources, User };
// Errors
export { CliError, BackendApiError, UtilsError };
// Models
export {
    crds,
    DistributedJobType,
    BaseWorkload,
    WorkloadConditionType,
    TrainingWorkload,
    InteractiveWorkload,
    PodStatus,
    TokenObject,
    InferenceWorkload,
    DistributedWorkload,
    Preemptibility
};
export type { CustomResourceDefinition };
