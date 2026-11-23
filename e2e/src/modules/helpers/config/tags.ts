/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

const resources = {
    department: '@department',
    nodepools: '@nodepools',
    metrics: '@metrics',
    deployments: '@deployments',
    projects: '@projects',
    jobs: '@jobs',
    workloads: '@workloads',
    workloadsV2: '@workloads-v2',
    nodes: '@nodes',
    inferenceWorkload: '@inference-workload',
    distInferenceWorkload: '@distributed-inference-workload',
    mimService: '@nim-service',
    patchRunaiConfig: '@patchRunaiConfig',
    applications: '@applications',
    users: '@users',
    authentication: '@authentication',
    ssoConfiguration: '@ssoConfiguration',
    tenant: '@tenant',
    scheduler: '@scheduler',
    workspace: '@workspace',
    workspaceTemplate: '@workspace-template',
    credential: '@credential',
    datasource: '@datasource',
    datavolume: '@datavolume',
    environment: '@environment',
    computeResource: '@compute-resource',
    cluster: '@cluster',
    accessRules: '@accessrules',
    roles: '@roles',
    orgTree: '@orgTree',
    policy: '@policy',
    overview: '@overview',
    audit: '@audit',
    templateV2: '@templateV2',
    networkTopologies: '@networkTopologies',
    trainingWorkload: '@training-workload',
    distributedTrainingWorkload: '@distributed-training-workload',
    gatewayApi: '@gateway-api',
    aiApplications: '@ai-applications'
};

// Execution cohorts stay separate from team ownership tags.
const suites = {
    legacySphinx: '@legacy-sphinx'
};

const operations = {
    submit: '@submit'
};

const support = {
    fairness: '@fairness'
};

const tools = {
    ui: '@ui',
    api: '@api',
    kubeCli: '@kubeCli',
    runaiCli: '@runai-cli'
};

const services = {
    datavolume: '@datavolumes-service',
    k8sObjectsTracker: '@k8s-objects-tracker',
    policy: '@policy-service',
    assets: '@assets-service',
    gitSync: '@git-sync',
    workloads: '@workloads',

    identityManager: '@identity-manager',
    keycloak: '@keycloak',
    tenantsManager: '@tenants-manager',
    tempBackendSettings: '@temp-backend-settings',
    clusterService: '@cluster-service',
    clusterInstaller: '@cluster-installer',
    redoc: '@redoc',

    auditService: '@audit-service',
    metricsService: '@metrics-service',
    notificationsService: '@notifications-service',
    notificationsProxy: '@notifications-proxy',
    runaiCli: '@runai-cli',
    runaiExposer: '@runai-exposer',
    frontend: '@frontend',

    orgUnitService: '@org-unit-service',
    authorization: '@authorization',
    authorizationMigrator: '@authorization-migrator',

    agent: '@agent',
    clusterSync: '@cluster-sync',
    projectController: '@project-controller',
    nodepoolController: '@nodepool-controller',
    accessruleController: '@accessrule-controller',
    externalWorkloadIntegrator: '@external-workload-integrator',
    podGroupAssigner: '@pod-group-assigner',

    clusterApi: '@cluster-api',
    workloadOverseer: '@workload-overseer'
};

// Isn't supported yet
const runOn = {
    nightly: '@nightly',
    parallelCI: '@parallel-ci',
    serialCI: '@serial-ci',
    clusterDeletion: '@cluster-deletion',
    preUpgrade: '@pre-upgrade',
    postUpgrade: '@post-upgrade',
    workloadPreUpgrade: '@workload-pre-upgrade',
    workloadPostUpgrade: '@workload-post-upgrade',
    upgradeHealthCheck: '@upgrade-health-check',
    multiTenancyUpgrade: '@multi-tenancy-upgrade',
    dgxc: '@dgxc',
    k8sFlavCertKit: '@k8s-flav-cert-kit-o',
    k8sFlavCertKitTemp: '@k8s-flav-cert-kit-temp',
    k8sFlavWorkloadsKit: '@k8s-flav-workloads-kit',
    networkCert: '@network-cert',
    basicCI: '@basic-ci'
};

export const teams = {
    viper: '@viper',
    phoenix: '@phoenix',
    loong: '@loong',
    pegasus: '@pegasus'
};

export const tags = {
    resource: resources,
    support: support,
    tool: tools,
    operations: operations,
    runOn: runOn,
    teams: teams,
    services: services,
    suites: suites
};
