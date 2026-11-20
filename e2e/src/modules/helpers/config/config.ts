/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import * as dotenv from 'dotenv';
import {
    Settings,
    gpuTypes,
    supportedPlatforms,
    Config,
    StateData,
    FeatureFlags,
    McStateData
} from './types';
import { Utils } from '../utils/utils';
import os from 'os';
import path from 'path';
import { fileURLToPath } from 'url';
import { getSharedConfig } from '../../../../playwright.config';

// Resolve `.env` relative to this file (the e2e/ directory) instead of process.cwd(),
// so the module loads correctly when imported from any cwd (e.g. the Playwright
// Test extension in VS Code/Cursor, which spawns from the workspace root).
const __dirname = path.dirname(fileURLToPath(import.meta.url));
dotenv.config({ path: path.resolve(__dirname, '../../../../.env') });

process.env.NODE_TLS_REJECT_UNAUTHORIZED = '0';

export const isAirgapped = process.env.IS_AIRGAPPED === 'true';

function hostnameFromUrl(url: string | undefined): string {
    if (!url) {
        return '';
    }
    try {
        return new URL(url).hostname;
    } catch {
        return '';
    }
}

const appHostname = hostnameFromUrl(process.env.APP_URL);
export const isStagingEnv = appHostname.split('.')[0] === 'staging';
export const isTestEnv = appHostname.split('.')[0] === 'test';
export const saasTenantName = process.env.CONTROL_PLANE_URL
    ? new URL(process.env.CONTROL_PLANE_URL).hostname.split('.')[0]
    : '';
export const tenantClientSecret = isStagingEnv
    ? process.env.E2E_TENANT_CLIENT_SECRET_STAGING
    : process.env.E2E_TENANT_CLIENT_SECRET_TEST;
export const saasConsoleUrl = appHostname ? `https://console.${appHostname}` : '';
const clusterPlatform = process.env.KUBERNETES_PLATFORM || '';
const baseDomain = process.env.E2E_BASE_DOMAIN;

const defineRegistry = (): string => {
    if (isAirgapped) {
        if (clusterPlatform === 'openshift') {
            return `${process.env.CLUSTER_NAME_PREFIX}-harbor.apps.${process.env.CLUSTER_NAME_PREFIX}.${baseDomain}/runai`;
        } else if (clusterPlatform === 'k8s' || clusterPlatform === 'op') {
            return `${process.env.CLUSTER_NAME_PREFIX}-harbor.${baseDomain}/runai`;
        }
    }
    return 'gcr.io';
};

const registry = defineRegistry();

export const openshiftApiUrl = (
    process.env.OPENSHIFT_API_URL ||
    (process.env.CLUSTER_NAME_PREFIX && baseDomain
        ? `https://api.${process.env.CLUSTER_NAME_PREFIX}.${baseDomain}:6443`
        : '')
).replace(/\/$/, '');
export const preservedUsers = (process.env.E2E_PRESERVED_USERS || '')
    .split(',')
    .map((user) => user.trim())
    .filter(Boolean);

// General environment settings
export const settings: Settings = {
    logLevel: process.env.LOG_LEVEL || 'debug',
    gpuType: gpuTypes[process.env.GPU_TYPE] || 'fake',
    taskTimeoutMS: parseInt(process.env.TASK_TIMEOUT_MS) || 60000,
    browserVideosPath: process.env.BROWSER_VIDEOS_PATH || 'videos/',
    browserHarsPath: 'network-activity/',
    browserVideosFileType: process.env.BROWSER_VIDEOS_FILE_TYPE || '.webm',
    registry,
    headlessBrowser: process.env.HEADLESS_BROWSER
        ? process.env.HEADLESS_BROWSER === 'true'
        : process.env.LOG_LEVEL !== '' && process.env.LOG_LEVEL !== 'debug',
    platform: supportedPlatforms[process.env.KUBERNETES_PLATFORM] || 'k8s',
    enableSSO: process.env.ENABLE_SSO === 'true',
    enableConsoleLogs: process.env.ENABLE_CONSOLE_LOGS ? process.env.ENABLE_CONSOLE_LOGS === 'true' : true,
    enableAuth: process.env.ENABLE_AUTH === 'true',
    isSelfHosted:
        process.env.IS_SELF_HOSTED && process.env.IS_SELF_HOSTED !== ''
            ? process.env.IS_SELF_HOSTED === 'true'
            : process.env.CLUSTER_NAME_PREFIX
              ? !process.env.CLUSTER_NAME_PREFIX.includes('saas-')
              : false,
    smtpHost: process.env.SMTP_HOST || 'live.smtp.mailtrap.io',
    smtpPort: process.env.SMTP_PORT || '587',
    smtpUsername: process.env.SMTP_USERNAME,
    smtpPassword: process.env.SMTP_PASSWORD,
    isOpNightlyEnv: process.env.CLUSTER_NAME_PREFIX
        ? process.env.CLUSTER_NAME_PREFIX.includes('-op-')
        : false,
    isRke2: process.env.CLUSTER_NAME_PREFIX ? process.env.CLUSTER_NAME_PREFIX.includes('-rke2-') : false,
    isRke: process.env.CLUSTER_NAME_PREFIX ? process.env.CLUSTER_NAME_PREFIX.includes('-rke-') : false,
    isArm: process.env.CLUSTER_NAME_PREFIX ? process.env.CLUSTER_NAME_PREFIX.includes('-arm-') : false,
    isAws: process.env.CLUSTER_NAME_PREFIX ? process.env.CLUSTER_NAME_PREFIX.includes('-aws-') : false,
    isOke: process.env.CLUSTER_NAME_PREFIX ? process.env.CLUSTER_NAME_PREFIX.includes('-oke-') : false,
    isUpgradesTests: process.env.IS_UPGRADE_TEST === 'true' ||
        (process.env.CLUSTER_NAME_PREFIX
            ? /up.*to/.test(process.env.CLUSTER_NAME_PREFIX) ||
              process.env.CLUSTER_NAME_PREFIX.startsWith('workload-upgrade')
            : false),
    isInstallRunaiConfigValues: process.env.CLUSTER_NAME_PREFIX
        ? process.env.CLUSTER_NAME_PREFIX.includes('helm-runaiconfig-')
        : false,
    isClusterDeletionWorkflow: process.env.CLUSTER_NAME_PREFIX
        ? process.env.CLUSTER_NAME_PREFIX.includes('cluster-deletion-')
        : false,
    isGatewayApiWorkflow: process.env.GATEWAY_API_ENABLED === 'true'
        || (process.env.CLUSTER_NAME_PREFIX
            ? process.env.CLUSTER_NAME_PREFIX.includes('gateway-api')
            : false),
    isMultiTenancyCluster: process.env.CLUSTER_NAME_PREFIX
        ? process.env.CLUSTER_NAME_PREFIX.includes('multi-tenant-')
        : false,
    isLocalRun: process.env.LOCAL_RUN ? process.env.LOCAL_RUN === 'true' : false,
    tls_key: process.env.TLS_KEY,
    tls_cert: process.env.TLS_CERT,
    tenantClientSecret
};

let _flowRootState: StateData | null = null;

export function resetFlowRootState(): void {
    _flowRootState = null;
}

function getFlowRootState(): StateData {
    if (!_flowRootState) {
        const shared = getSharedConfig();
        _flowRootState = shared?.flowRootState || {
            currentFlowUuid: Utils.generateShortUuid(),
            users: {},
            usersSSO: {},
            cluster: undefined,
            controlPlaneVersion: undefined,
            cliVersion: undefined,
            numOfGpus: 0,
            nodesResources: [],
            currentTenant: undefined,
            reportIds: {}
        };
    }
    return _flowRootState;
}

const ADMIN_TOKEN_MIN_MINUTES_LEFT = 8;
const TOKEN_SYNC_THROTTLE_MS = 15_000;
let _lastTokenSyncMs = 0;

interface TokenBearer {
    idToken?: string;
    accessToken?: string;
    refreshToken?: string;
}

function adminTokenMinutesLeft(state: StateData): number {
    const admin = state.users?.system_administrator;
    const token = admin?.idToken ?? admin?.accessToken;
    if (!token) {
        return Number.POSITIVE_INFINITY;
    }
    try {
        const payload = JSON.parse(Buffer.from(token.split('.')[1], 'base64url').toString('utf8'));
        if (typeof payload.exp !== 'number') {
            return Number.POSITIVE_INFINITY;
        }
        return (payload.exp - Date.now() / 1000) / 60;
    } catch {
        return Number.POSITIVE_INFINITY;
    }
}

function mergeUserTokens(target: object | undefined, source: object | undefined): void {
    if (!target || !source) {
        return;
    }
    const current = target as Record<string, TokenBearer>;
    const fresh = source as Record<string, TokenBearer>;
    for (const key of Object.keys(current)) {
        const currentUser = current[key];
        const freshUser = fresh[key];
        if (!currentUser || !freshUser) {
            continue;
        }
        if (freshUser.idToken) {
            currentUser.idToken = freshUser.idToken;
        }
        if (freshUser.accessToken) {
            currentUser.accessToken = freshUser.accessToken;
        }
        if (freshUser.refreshToken) {
            currentUser.refreshToken = freshUser.refreshToken;
        }
    }
}

function syncAdminTokensFromSharedState(): void {
    const now = Date.now();
    if (now - _lastTokenSyncMs < TOKEN_SYNC_THROTTLE_MS) {
        return;
    }
    _lastTokenSyncMs = now;

    if (!_flowRootState || adminTokenMinutesLeft(_flowRootState) > ADMIN_TOKEN_MIN_MINUTES_LEFT) {
        return;
    }
    try {
        const fresh = getSharedConfig()?.flowRootState as StateData | undefined;
        if (!fresh) {
            return;
        }
        mergeUserTokens(_flowRootState.users, fresh.users);
    } catch {
        return;
    }
}

export const flowRootState: StateData = new Proxy({} as StateData, {
    get(_target: StateData, prop: string | symbol): any {
        const state = getFlowRootState();
        if (prop === 'users') {
            syncAdminTokensFromSharedState();
        }
        return state[prop as keyof StateData];
    },
    set(_target: StateData, prop: string | symbol, value: any): boolean {
        (getFlowRootState() as any)[prop] = value;
        return true;
    },
    has(_target: StateData, prop: string | symbol): boolean {
        return prop in getFlowRootState();
    },
    ownKeys(_target: StateData): ArrayLike<string | symbol> {
        return Reflect.ownKeys(getFlowRootState());
    },
    getOwnPropertyDescriptor(_target: StateData, prop: string | symbol): PropertyDescriptor | undefined {
        return Reflect.getOwnPropertyDescriptor(getFlowRootState(), prop);
    }
});

export const mcFlowRootState: McStateData = {
    clusters: undefined
};
export const DEFAULT_RUNAI_CLI_PATH = './runaiCli';
export const DEFAULT_RUNAI_V2_CLI_PATH = os.homedir() + '/.runai/bin/runai';

// internal functionality config
export const config: Config = {
    // Unique url per Environment, in Self-Hosted environment it is the url of the control plane
    appUrl:
        Utils.trimLastChar(process.env.APP_URL, '/') ||
        Utils.trimLastChar(process.env.CONTROL_PLANE_URL, '/'),
    //  Tenant login url
    tenantUrl: Utils.trimLastChar(process.env.CONTROL_PLANE_URL, '/'),
    clusterUrl:
        process.env.CLUSTER_URL ||
        (clusterPlatform === 'ocp'
            ? `https://runai.apps.${process.env.CLUSTER_NAME_PREFIX}.${baseDomain}`
            : `https://${process.env.CLUSTER_NAME_PREFIX}.${baseDomain}`),
    // Read synchronously at import so every Playwright worker process gets a value.
    // globalSetup resolves the real credentials from the cluster and writes them
    // to RESOLVED_CONTROL_PLANE_ADMIN_* (see ensureControlPlaneCredentials); workers spawned
    // after globalSetup pick up the resolved values via the RESOLVED_ vars without
    // overwriting the original CI-supplied CONTROL_PLANE_ADMIN_* env vars.
    controlPlaneAdminUsername:
        process.env.RESOLVED_CONTROL_PLANE_ADMIN_USERNAME || process.env.CONTROL_PLANE_ADMIN_USERNAME,
    controlPlaneAdminPassword:
        process.env.RESOLVED_CONTROL_PLANE_ADMIN_PASSWORD || process.env.CONTROL_PLANE_ADMIN_PASSWORD,
    KEYCLOAKADMINPASSWORD: process.env.KEYCLOAK_ADMIN_PASSWORD,
    KEYCLOAKADMINPASSWORDTEST: process.env.KEYCLOAK_ADMIN_PASSWORD_TEST,
    KEYCLOAKADMINPASSWORDSTAGING: process.env.KEYCLOAK_ADMIN_PASSWORD_STAGING,
    clusterNamePrefix: process.env.CLUSTER_NAME_PREFIX || 'runai-cluster',
    kubeconfigPath: process.env.KUBECONFIG_PATH || process.env.KUBECONFIG,
    kubeconfigClusterName: process.env.KUBECONFIG_CLUSTER_NAME,
    runaiCliPath: process.env.RUNAI_CLI_PATH || DEFAULT_RUNAI_CLI_PATH,
    runaiV2CliPath: DEFAULT_RUNAI_V2_CLI_PATH,

    unleash: {
        url: process.env.UNLEASH_URL,
        apiToken: process.env.UNLEASH_API_TOKEN
    },
    quickStartImage:
        process.env.GPU_TYPE === 'real'
            ? {
                  ubuntu: isAirgapped ? `${registry}/library/ubuntu` : 'ubuntu',
                  interactive: `${registry}/run-ai-demo/sleep`,
                  training: `${registry}/run-ai-demo/quickstart`, // TODO: Use slimmer image
                  inference: `${registry}/run-ai-lab/inference-demo:v2`, // TODO: Use slimmer image
                  emptyImage: `${registry}/run-ai-demo/quickstart-fake`,
                  distributed: `${registry}/run-ai-demo/quickstart-distributed:v0.3.0`,
                  distributedMpi: `${registry}/run-ai-demo/quickstart-distributed:v0.3.0`,
                  distributedTf: `${registry}/run-ai-lab/tensor-sample:0.1`,
                  // kubeflow-examples/pytorch-dist-mnist on gcr.io returns 403 for anonymous pull (use same as distributed/MPI)
                  distributedPytorch: `${registry}/run-ai-demo/quickstart-distributed:v0.3.0`,
                  inferenceHelloWorld: `${registry}/run-ai-lab/knative-helloworld-go`,
                  inferenceAutoScale: `${registry}/run-ai-lab/knative-autoscale-go`,
                  inferenceAutoScaleClient: `${registry}/run-ai-lab/knative-autoscale-go:hey`,
                  nim: `${registry}/run-ai-lab/nim-cpu`,
                  metricsGpu: `${registry}/run-ai-demo/gpu-burn`,
                  sleepInfinity: `${registry}/run-ai-lab/ubuntu:loop`,
                  networkTools: `${registry}/run-ai-lab/ubuntu:network`, // an ubuntu-based image with curl and hey installed
                  nonExistingImage: `${registry}/run-ai-demo/quickstart-non-existing-image`,
                  idleGpu: `${registry}/run-ai-demo/sleep`,
                  nvidiaTritonServer: `${registry}/run-ai-demo/example-triton-server`,
                  failedImage: `${registry}/run-ai-demo/failed-job:latest`,
                  nvidiaTritonClient: 'nvcr.io/nvidia/tritonserver:24.07-py3-sdk'
              }
            : {
                  ubuntu: isAirgapped ? `${registry}/library/ubuntu` : 'ubuntu',
                  interactive: settings.isArm
                      ? `${registry}/run-ai-demo/quickstart-fake:arm`
                      : `${registry}/run-ai-demo/quickstart-fake`,
                  training: settings.isArm
                      ? `${registry}/run-ai-demo/quickstart-fake:arm`
                      : `${registry}/run-ai-demo/quickstart-fake`,
                  inference: settings.isArm
                      ? `${registry}/run-ai-demo/quickstart-fake:arm`
                      : `${registry}/run-ai-demo/quickstart-fake`,
                  emptyImage: settings.isArm
                      ? `${registry}/run-ai-demo/quickstart-fake:arm`
                      : `${registry}/run-ai-demo/quickstart-fake`,
                  distributed: `${registry}/run-ai-demo/quickstart-distributed:v0.3.0`,
                  distributedMpi: `${registry}/run-ai-demo/quickstart-distributed:v0.3.0`,
                  distributedTf: `${registry}/run-ai-demo/quickstart-distributed:v0.3.0`,
                  distributedPytorch: `${registry}/run-ai-demo/quickstart-distributed:v0.3.0`,
                  inferenceHelloWorld: `${registry}/run-ai-lab/knative-helloworld-go`,
                  inferenceAutoScale: `${registry}/run-ai-lab/knative-autoscale-go`,
                  inferenceAutoScaleClient: `${registry}/run-ai-lab/knative-autoscale-go:hey`,
                  nim: `${registry}/run-ai-lab/nim-cpu`,
                  metricsGpu: settings.isArm
                      ? `${registry}/run-ai-demo/quickstart-fake:arm`
                      : `${registry}/run-ai-demo/quickstart-fake`,
                  sleepInfinity: `${registry}/run-ai-lab/ubuntu:loop`,
                  networkTools: `${registry}/run-ai-lab/ubuntu:network`, // an ubuntu-based image with curl and hey installed
                  nonExistingImage: `${registry}/run-ai-demo/quickstart-non-existing-image`,
                  idleGpu: `${registry}/run-ai-demo/sleep`,
                  nvidiaTritonServer: `${registry}/run-ai-demo/example-triton-server`,
                  nvidiaTritonClient: 'nvcr.io/nvidia/tritonserver:24.07-py3-sdk',
                  failedImage: `${registry}/run-ai-demo/failed-job:latest`
              }
};

export const featureFlags: FeatureFlags = {
    enableCpuResources: process.env.ENABLE_CPU_RESOURCES === 'true',
    enableOverQuotaPriority: process.env.ENABLE_OVER_QUOTA_PRIORITY === 'true'
};
