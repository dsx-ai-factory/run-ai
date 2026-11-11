/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Cluster } from '../../api/apis/backendClusterApi';
import { AccessRule } from '../../api/apis/authorizationApi/generated/model/accessRule';
import { Role } from '../../api/apis/authorizationApi/generated/model/role';
import { Tenant } from '../../api/apis/tenantsManagerApi/generated/api';
import {
    DistributedInferencePolicyDefaultsAndRulesV2Defaults,
    DistributedInferencePolicyDefaultsAndRulesV2Rules,
    DistributedPolicyDefaultsV2,
    DistributedPolicyRulesV2,
    InferencePolicyDefaultsAndRulesV2Defaults,
    InferencePolicyDefaultsAndRulesV2Rules,
    MatchExpressionOperator,
    TrainingPolicyDefaultsAndRulesV2Defaults,
    TrainingPolicyDefaultsAndRulesV2Rules,
    WorkspacePolicyDefaultsAndRulesV2Defaults,
    WorkspacePolicyDefaultsAndRulesV2Rules
} from '../../api/apis/policyApi';

export const supportedPlatforms = {
    k8s: 'k8s',
    ocp: 'ocp'
};

export const gpuTypes = {
    fake: 'fake',
    real: 'real'
};

interface Images {
    ubuntu: string;
    training: string;
    interactive: string;
    inference: string;
    emptyImage: string;
    distributed: string;
    distributedMpi: string;
    distributedTf: string;
    distributedPytorch: string;
    inferenceHelloWorld: string;
    inferenceAutoScale: string;
    inferenceAutoScaleClient: string;
    nim: string;
    metricsGpu: string;
    sleepInfinity: string;
    networkTools: string;
    nonExistingImage: string;
    idleGpu: string;
    nvidiaTritonServer: string;
    nvidiaTritonClient: string;
    failedImage: string;
}

export interface User {
    email: string;
    password: string;
    userId: string;
    role: Role;
    idToken?: string;
    accessToken?: string;
    refreshToken?: string;
    accessRules?: AccessRule[];
}

export interface Users {
    system_administrator?: User;
    department_admin?: User;
    editor?: User;
    viewer?: User;
    research_manager?: User;
    ml_engineer?: User;
    data_volume_administrator?: User;
    department_viewer?: User;
    template_administrator?: User;
    compute_resource_administrator?: User;
    data_source_administrator?: User;
    credentials_administrator?: User;
    environment_administrator?: User;
    l1_researcher?: User;
    l2_researcher?: User;
    ai_practitioner?: User;
    application_administrator?: User;
    project_administrator?: User;
}

export interface Settings {
    logLevel: string;
    browserVideosPath: string;
    browserHarsPath: string;
    headlessBrowser: boolean;
    taskTimeoutMS: number;
    browserVideosFileType: string;
    gpuType: string;
    // ocp / k8s / etc...
    platform: string;
    enableSSO: boolean;
    enableConsoleLogs: boolean;
    enableAuth: boolean;
    isSelfHosted: boolean;
    isOpNightlyEnv: boolean;
    isRke: boolean;
    isRke2: boolean;
    isArm: boolean;
    isAws: boolean;
    isOke: boolean;
    isLocalRun: boolean;
    isClusterDeletionWorkflow: boolean;
    isGatewayApiWorkflow: boolean;
    isMultiTenancyCluster: boolean;
    isUpgradesTests: boolean;
    isInstallRunaiConfigValues: boolean;
    tls_key: string;
    tls_cert: string;
    registry: string;
    // tenant
    tenantClientSecret: string;
    smtpHost: string;
    smtpPort: string;
    smtpUsername: string;
    smtpPassword: string;
}

export interface UnleashConfig {
    url: string;
    apiToken: string;
}

export interface Config {
    appUrl: string;
    tenantUrl: string;
    clusterUrl: string;
    controlPlaneAdminUsername: string;
    controlPlaneAdminPassword: string | undefined;
    KEYCLOAKADMINPASSWORD: string;
    KEYCLOAKADMINPASSWORDTEST: string;
    KEYCLOAKADMINPASSWORDSTAGING: string;
    kubeconfigPath: string;
    kubeconfigClusterName: string;
    clusterNamePrefix: string;
    runaiCliPath: string;
    runaiV2CliPath: string;
    quickStartImage: Images;
    unleash: UnleashConfig;
}

export interface NodeResources {
    name: string;
    cpu: number;
    memoryMB: number;
    gpu: number;
}

export interface ReportIds {
    consumptionReportIdByNodePool?: string;
    consumptionReportIdByDepartment?: string;
    consumptionReportIdByProject?: string;
    // Lives here because only flowRootState is rehydrated in the workers.
    generationError?: string;
}

// Uses to pass data from root hook to test suites
export interface StateData {
    users: Users;
    usersSSO: Users;
    cluster: Cluster;
    cliVersion: string;
    currentFlowUuid: string;
    numOfGpus: number;
    nodesResources: NodeResources[];
    controlPlaneVersion: string;
    currentTenant: Tenant;
    reportIds: ReportIds;
    /**
     * The NodePool CRD group/version this cluster serves, resolved once in global setup so that
     * workers read it from shared state instead of probing the cluster. Only an upgrade can change
     * which CRD is installed. See NodePoolCrdFields.resolveNodePoolCrd.
     */
    nodePoolCrd?: { group: string; version: string };
    /**
     * Whether this cluster serves the KAI org-unit CRDs (projects.kai.resources,
     * departments.kai.resources) vs the legacy ones. Resolved once in global setup.
     * See OrgUnitCrdFields.resolveOrgUnitCrds.
     */
    orgUnitCrdsKai?: boolean;
}

export interface McStateData {
    clusters: Map<string, Cluster>;
}

export interface FeatureFlags {
    enableCpuResources: boolean;
    enableOverQuotaPriority: boolean;
}

export enum WorkloadType {
    workspace = 'Workspace',
    training = 'Training',
    deployment = 'Deployment',
    distributed = 'Distributed'
}

export interface NodeSelector {
    nodeAffinityRequired: {
        nodeSelectorTerms: Array<{
            matchExpressions: Array<{
                key: string;
                operator: MatchExpressionOperator;
                values: string[];
            }>;
        }>;
    };
}

export interface Compute {
    gpuDevicesRequest: number;
    cpuCoreRequest: number;
    cpuMemoryRequest: string;
}

export type PolicyRuleType =
    | TrainingPolicyDefaultsAndRulesV2Rules
    | WorkspacePolicyDefaultsAndRulesV2Rules
    | DistributedPolicyRulesV2
    | InferencePolicyDefaultsAndRulesV2Rules
    | DistributedInferencePolicyDefaultsAndRulesV2Rules
    | null;

export type PolicyDefaultsType =
    | TrainingPolicyDefaultsAndRulesV2Defaults
    | DistributedPolicyDefaultsV2
    | WorkspacePolicyDefaultsAndRulesV2Defaults
    | InferencePolicyDefaultsAndRulesV2Defaults
    | DistributedInferencePolicyDefaultsAndRulesV2Defaults
    | null;
