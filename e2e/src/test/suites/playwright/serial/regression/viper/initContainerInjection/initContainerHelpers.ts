/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { V1Container, V1ContainerStatus, V1OwnerReference, V1Pod } from '@kubernetes/client-node';
import { expect } from '@playwright/test';

import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { WorkspaceSpecSpec } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/workspaceSpecSpec';
import { WorkspaceSpecSpecAllOfCompute } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/workspaceSpecSpecAllOfCompute';
import { kubeCli } from '../../../../../../../modules/cli';
import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    projectUtils,
    Utils,
    waitUtils,
    workloadsCrudUtils
} from '../../../../../../../modules/helpers';
import { enableAnyWorkloadSubmission, RestoreAnyWorkload } from '../anyWorkloadSubmission';

export const WLR_KIND = 'WorkloadRunner';
export const WLR_API_VERSION_PREFIX = 'run.ai/';

export const initContainerNames = {
    gitSyncPrefix: 'git-sync-',
    s3FsPrefix: 's3-fs-',
    oauth2ProxyPrefix: 'runai-oauth2-proxy-',
    usernameLoader: 'runai-auth-user-loader'
} as const;

export const authVolumeName = 'runai-auth';
export const authVolumePath = '/etc/runai/auth/users';

const AUTHENTICATED_EMAILS_FILE_FLAG = '--authenticated-emails-file=';
const LOADER_ECHO_PREFIX = 'echo -e "';

const adminToken = (): string => flowRootState.users.system_administrator.accessToken;

const SUBMIT_IMAGE = config.quickStartImage.training;
const CPU_ONLY_COMPUTE: WorkspaceSpecSpecAllOfCompute = {
    cpuCoreRequest: 0.1,
    cpuMemoryRequest: '100M',
    gpuDevicesRequest: 0
};

export interface InitContainerSuiteContext {
    project: Project;
    projectName: string;
    namespace: string;
    authorizedEmail: string;
    restore: RestoreAnyWorkload;
}

export async function setupInitContainerSuite(projectPrefix: string): Promise<InitContainerSuiteContext> {
    const restore = await enableAnyWorkloadSubmission();
    try {
        const projectName = objectGenerator.generateShortNameByPrefix(projectPrefix);
        const namespace = Utils.namespaceFromProjectName(projectName);
        const project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 0,
            maxAllowedGpus: 0
        });
        await projectUtils.waitForRoleBindingForUserInProject(
            namespace,
            flowRootState.users.system_administrator.email,
            projectUtils.getSystemAdministratorRoleBindingName(),
            true
        );
        return {
            project,
            projectName,
            namespace,
            authorizedEmail: flowRootState.users.l1_researcher.email,
            restore
        };
    } catch (error) {
        await restore();
        throw error;
    }
}

export async function teardownInitContainerSuite(ctx: InitContainerSuiteContext | undefined): Promise<void> {
    if (!ctx) {
        return;
    }
    try {
        await operations.deleteProjectAndNamespace(ctx.project.id);
        logger.info(`deleted project ${ctx.projectName} and namespace ${ctx.namespace}`);
    } catch (error) {
        logger.error(`failed to delete project ${ctx.projectName}: ${error}`);
    }
    await ctx.restore();
}

export async function submitWorkspaceAndWaitRunning(
    namespace: string,
    projectId: string,
    workspaceName: string,
    specOverrides: Partial<WorkspaceSpecSpec>,
    timeoutSeconds = 180
): Promise<string> {
    const spec: WorkspaceSpecSpec = {
        image: SUBMIT_IMAGE,
        command: 'sleep infinity',
        compute: CPU_ONLY_COMPUTE,
        ...specOverrides
    };
    await workloadsCrudUtils.createWorkspaceAndWaitForPodsRunning(
        workspaceName,
        false,
        flowRootState.cluster.uuid,
        projectId,
        spec,
        namespace,
        adminToken(),
        timeoutSeconds
    );
    return workspaceName;
}

const fetchOwnerReferences = async (
    namespace: string,
    owner: V1OwnerReference
): Promise<V1OwnerReference[]> => {
    const apiVersion = owner.apiVersion ?? '';
    if (!apiVersion.includes('/')) {
        return [];
    }
    const [group, version] = apiVersion.split('/');
    const plural = `${owner.kind.toLowerCase()}s`;
    try {
        const object = await kubeCli.CustomObjectsApi.getNamespacedCustomObject({
            group,
            version,
            plural,
            namespace,
            name: owner.name
        });
        const metadata = (object as { metadata?: { ownerReferences?: V1OwnerReference[] } }).metadata;
        return metadata?.ownerReferences ?? [];
    } catch (error) {
        logger.warn(
            `failed to fetch ${owner.kind}.${group}/${owner.name} in ${namespace} while walking owner chain: ${error}`
        );
        return [];
    }
};

const chainContainsRunaiOwner = async (
    namespace: string,
    owners: V1OwnerReference[],
    acceptedKinds: string[],
    visited: string[]
): Promise<boolean> => {
    for (const owner of owners) {
        const isAccepted =
            acceptedKinds.includes(owner.kind) && (owner.apiVersion ?? '').startsWith(WLR_API_VERSION_PREFIX);
        visited.push(`${owner.apiVersion}/${owner.kind}/${owner.name}`);
        if (isAccepted) {
            return true;
        }
        const parentOwners = await fetchOwnerReferences(namespace, owner);
        if (parentOwners.length === 0) {
            continue;
        }
        if (await chainContainsRunaiOwner(namespace, parentOwners, acceptedKinds, visited)) {
            return true;
        }
    }
    return false;
};

export async function assertOwnedByRunaiWorkload(
    pod: V1Pod,
    acceptedKinds: string[] = [WLR_KIND]
): Promise<void> {
    const namespace = pod.metadata?.namespace;
    if (!namespace) {
        throw new Error(`Pod ${pod.metadata?.name} has no namespace; cannot resolve owner chain.`);
    }
    const visited: string[] = [`Pod/${pod.metadata?.name}`];
    const owners = pod.metadata?.ownerReferences ?? [];
    if (await chainContainsRunaiOwner(namespace, owners, acceptedKinds, visited)) {
        return;
    }
    throw new Error(
        `Pod ${namespace}/${pod.metadata?.name} is not owned by any of [${acceptedKinds.join(', ')}] ` +
            `(direct or transitive). Visited owner chain: [${visited.join(' -> ')}]. This usually means the ` +
            'anyworkload submission master switch was not on when the workload was created.'
    );
}

export async function assertOwnedByWorkloadRunner(pod: V1Pod): Promise<void> {
    await assertOwnedByRunaiWorkload(pod, [WLR_KIND]);
}

export function findInitContainer(
    pod: V1Pod,
    namePredicate: (name: string) => boolean
): V1Container | undefined {
    return (pod.spec?.initContainers ?? []).find((c) => namePredicate(c.name));
}

export function listInitContainerNames(pod: V1Pod): string[] {
    return (pod.spec?.initContainers ?? []).map((c) => c.name);
}

const oneShotCompleted = (pod: V1Pod, namePredicate: (name: string) => boolean): void => {
    const container = findInitContainer(pod, namePredicate);
    if (!container) {
        const observed = listInitContainerNames(pod).join(', ');
        throw new Error(
            `Expected one-shot init container not found on pod ${pod.metadata?.name}. ` +
                `Observed init containers: [${observed}]`
        );
    }
    const status = (pod.status?.initContainerStatuses ?? []).find((s) => s.name === container.name);
    if (!status) {
        throw new Error(`Init container ${container.name} has no status on pod ${pod.metadata?.name}.`);
    }
    const terminated = status.state?.terminated;
    if (!terminated) {
        throw new Error(
            `Init container ${container.name} did not terminate on pod ${pod.metadata?.name}. ` +
                `Current state: ${JSON.stringify(status.state)}`
        );
    }
    expect(terminated.exitCode).toBe(0);
    expect(terminated.reason).toBe('Completed');
};

const sidecarRunning = (pod: V1Pod, namePredicate: (name: string) => boolean): void => {
    const container = findInitContainer(pod, namePredicate);
    if (!container) {
        const observed = listInitContainerNames(pod).join(', ');
        throw new Error(
            `Expected sidecar init container not found on pod ${pod.metadata?.name}. ` +
                `Observed init containers: [${observed}]`
        );
    }
    expect(container.restartPolicy).toBe('Always');
    const status = (pod.status?.initContainerStatuses ?? []).find((s) => s.name === container.name);
    if (!status) {
        throw new Error(`Sidecar ${container.name} has no status on pod ${pod.metadata?.name}.`);
    }
    const isRunning = status.state?.running !== undefined || status.started === true;
    if (!isRunning) {
        throw new Error(
            `Sidecar ${container.name} is not running on pod ${pod.metadata?.name}. ` +
                `State: ${JSON.stringify(status.state)} started=${status.started}`
        );
    }
};

const isInitStatusSettled = (status: V1ContainerStatus, isSidecar: boolean): boolean => {
    if (isSidecar) {
        return status.state?.running !== undefined || status.started === true;
    }
    return status.state?.terminated !== undefined;
};

const waitForInitContainerSettled = async (
    namespace: string,
    podName: string,
    containerName: string,
    isSidecar: boolean,
    timeoutSeconds = 180
): Promise<V1Pod> => {
    let currentPod: V1Pod | undefined;
    await waitUtils.waitForCondition(
        async () => {
            currentPod = await kubeCli.CoreV1Api.readNamespacedPod({ name: podName, namespace });
            const status = (currentPod.status?.initContainerStatuses ?? []).find(
                (s) => s.name === containerName
            );
            if (!status) {
                return false;
            }
            return isInitStatusSettled(status, isSidecar);
        },
        `init container ${containerName} ${isSidecar ? 'running' : 'terminated'} on pod ${namespace}/${podName}`,
        true,
        timeoutSeconds,
        2
    );
    return currentPod as V1Pod;
};

export type InitContainerKind = 'sidecar' | 'oneShot';

export async function assertInitContainerHealthy(
    pod: V1Pod,
    containerName: string,
    kind: InitContainerKind
): Promise<void> {
    const isSidecar = kind === 'sidecar';
    const refreshedPod = await waitForInitContainerSettled(
        pod.metadata?.namespace as string,
        pod.metadata?.name as string,
        containerName,
        isSidecar
    );
    if (isSidecar) {
        sidecarRunning(refreshedPod, (n) => n === containerName);
    } else {
        oneShotCompleted(refreshedPod, (n) => n === containerName);
    }
}

export async function assertGitSyncInitContainerHealthy(pod: V1Pod): Promise<void> {
    await assertOwnedByWorkloadRunner(pod);
    const gitSync = findInitContainer(pod, (n) => n.startsWith(initContainerNames.gitSyncPrefix));
    if (!gitSync) {
        throw new Error(
            `Expected an init container starting with "${initContainerNames.gitSyncPrefix}" on pod ${pod.metadata?.name}.`
        );
    }
    const kind: InitContainerKind = gitSync.restartPolicy === 'Always' ? 'sidecar' : 'oneShot';
    await assertInitContainerHealthy(pod, gitSync.name, kind);
}

export function proxyArgs(pod: V1Pod, idx = 0): string[] {
    const name = `${initContainerNames.oauth2ProxyPrefix}${idx}`;
    const container = findInitContainer(pod, (n) => n === name);
    if (!container) {
        const observed = listInitContainerNames(pod).join(', ');
        throw new Error(
            `Expected oauth2 proxy container ${name} not found on pod ${pod.metadata?.name}. ` +
                `Observed init containers: [${observed}]`
        );
    }
    return container.args ?? [];
}

export function authenticatedEmailsFileFromProxyArgs(pod: V1Pod, idx = 0): string {
    const args = proxyArgs(pod, idx);
    const flag = args.find((arg) => arg.startsWith(AUTHENTICATED_EMAILS_FILE_FLAG));
    if (!flag) {
        throw new Error(
            `oauth2 proxy on pod ${pod.metadata?.name} has no ${AUTHENTICATED_EMAILS_FILE_FLAG} arg; ` +
                `args: [${args.join(' ')}]`
        );
    }
    return flag.slice(AUTHENTICATED_EMAILS_FILE_FLAG.length);
}

export function decodeLoaderPayloads(command: string): string {
    const decoded: string[] = [];
    const segments = command.split(LOADER_ECHO_PREFIX).slice(1);
    for (const segment of segments) {
        const closingQuote = segment.indexOf('"');
        if (closingQuote === -1) {
            continue;
        }
        const base64 = segment.slice(0, closingQuote);
        decoded.push(Buffer.from(base64, 'base64').toString('utf8'));
    }
    return decoded.join('\n');
}
