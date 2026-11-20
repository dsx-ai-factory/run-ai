/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Shared, assertion-free helpers for the Personal Credentials Plugin e2e suites (API / CLI / UI).
 * Centralizes WorkloadRunner CRD access, the dest-Secret ownerRef discriminator that proves the
 * v2/personalcreds path ran (a silent native fallback owns nothing), and the credential payload builders.
 */
import { kubeCli } from '../../cli';
import { flowRootState } from '../config/config';
import { logger, objectGenerator, waitUtils } from '../index';
import {
    DockerRegistryCreationRequest,
    GenericSecretCreationRequest,
    NgcApiKeyCreationRequest,
    Scope
} from '../../api/apis/assetsApi';

// WorkloadRunner CRD coordinates — the v2-path discriminator.
export const PC_WLR = {
    group: 'run.ai',
    version: 'v1alpha1',
    plural: 'workloadrunners',
    kind: 'WorkloadRunner'
} as const;
export const PC_ANNOTATION = 'run.ai/personal-credentials';

export class PersonalCredentialsUtils {
    decodeSecret(b64: string): string {
        return Buffer.from(b64, 'base64').toString();
    }

    // A materialized dest Secret owned by the WorkloadRunner proves the plugin ran. CLI cases also pin
    // controller === true; pass requireController for that stricter check.
    isWorkloadRunnerOwned(
        ownerRefs: { kind?: string; controller?: boolean }[] | undefined,
        requireController = false
    ): boolean {
        return (ownerRefs ?? []).some(
            (ref) => ref.kind === PC_WLR.kind && (!requireController || ref.controller === true)
        );
    }

    async listWorkloadRunners(namespace: string): Promise<Record<string, any>[]> {
        const res = await kubeCli.CustomObjectsApi.listNamespacedCustomObject({
            group: PC_WLR.group,
            version: PC_WLR.version,
            namespace,
            plural: PC_WLR.plural
        });
        return ((res as any).items ?? []) as Record<string, any>[];
    }

    async findRunnerByWorkloadName(
        namespace: string,
        workloadName: string
    ): Promise<Record<string, any> | undefined> {
        return (await this.listWorkloadRunners(namespace)).find((r) =>
            (r.metadata?.name as string | undefined)?.includes(workloadName)
        );
    }

    async getRunnerByName(namespace: string, name: string): Promise<Record<string, any>> {
        return (await kubeCli.CustomObjectsApi.getNamespacedCustomObject({
            group: PC_WLR.group,
            version: PC_WLR.version,
            namespace,
            plural: PC_WLR.plural,
            name
        })) as Record<string, any>;
    }

    // No product surface forces a reconcile; bumping a label (json-patch) re-triggers the controller.
    async forceRunnerReconcile(namespace: string, runnerName: string): Promise<void> {
        await kubeCli.CustomObjectsApi.patchNamespacedCustomObject({
            group: PC_WLR.group,
            version: PC_WLR.version,
            namespace,
            plural: PC_WLR.plural,
            name: runnerName,
            body: [
                {
                    op: 'add',
                    path: '/metadata/labels/e2e-reconcile-nonce',
                    value: objectGenerator.generateShortNameByPrefix('rec', 4)
                }
            ]
        });
    }

    // personalcreds is a StatusPlugin, so `.status.plugins.personalcreds` is always present after a status
    // reconcile (phase Disabled/Ready/Failed) — assert `.phase` / `.issues`, never key-absence.
    async readPersonalCredsPluginStatus(
        namespace: string,
        runnerName: string
    ): Promise<Record<string, any> | undefined> {
        return (await this.getRunnerByName(namespace, runnerName))?.status?.plugins?.personalcreds;
    }

    // Diagnostic (assertion-free): dump the WorkloadRunner's personalcreds status + every pod's ownerRefs so a
    // timed-out "wait for pod" run can distinguish (A) runner Degraded because USERCREDENTIAL: was never resolved
    // → no Job → no pod, from (B) a pod that IS running but whose ownerRef.name != workloadName so the exact-match
    // listPodsOfJob missed it. Every line is prefixed [pc-diag] for grep-ability in logs/testLogger.log.
    async logSubmissionDiagnostics(namespace: string, workloadName: string): Promise<void> {
        try {
            const runner = await this.findRunnerByWorkloadName(namespace, workloadName);
            if (!runner) {
                logger.info(`[pc-diag] no WorkloadRunner matching "${workloadName}" in ${namespace}`);
            } else {
                const status = (runner.status ?? {}) as Record<string, any>;
                // ABSENT annotation + personalcreds phase=Disabled is the fingerprint of the CP composer
                // never resolving the USERCREDENTIAL: references (plugin is annotation-gated, so it no-ops
                // and the raw placeholder reaches the Job where RFC-1123 validation rejects it).
                const pcAnnotation = (runner.metadata?.annotations as Record<string, string> | undefined)?.[
                    PC_ANNOTATION
                ];
                logger.info(
                    `[pc-diag] runner=${runner.metadata?.name} phase=${status.phase} ` +
                        `personalcreds=${JSON.stringify(status.plugins?.personalcreds)} ` +
                        `annotation[${PC_ANNOTATION}]=${
                            pcAnnotation === undefined ? 'ABSENT' : JSON.stringify(pcAnnotation)
                        }`
                );
                logger.info(
                    `[pc-diag] runner issues=${JSON.stringify(status.issues ?? status.conditions ?? [])}`
                );
            }
            const pods = await kubeCli.listPodsByNamepsace(namespace);
            logger.info(`[pc-diag] ${pods.length} pod(s) in ${namespace}`);
            for (const p of pods) {
                const owners = (p.metadata?.ownerReferences ?? [])
                    .map((o) => `${o.kind}/${o.name}`)
                    .join(',');
                logger.info(
                    `[pc-diag]   pod=${p.metadata?.name} phase=${p.status?.phase} owners=[${owners}]`
                );
            }
        } catch (e) {
            logger.error('[pc-diag] diagnostics failed', e);
        }
    }

    // Poll until a Secret no longer exists (dest-Secret GC after workload delete).
    async waitForSecretGone(
        namespace: string,
        secretName: string,
        description = `dest Secret ${secretName} garbage-collected`,
        timeout = 120,
        interval = 5
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                try {
                    await kubeCli.getSecret(secretName, namespace);
                    return false;
                } catch {
                    return true;
                }
            },
            description,
            true,
            timeout,
            interval
        );
    }

    // The typed workloads client throws on non-2xx; the thrown shape exposes the HTTP code variously.
    httpStatusOf(err: any): number | undefined {
        return err?.statusCode ?? err?.response?.statusCode ?? err?.response?.status;
    }

    systemAdminAccessToken(): string {
        return flowRootState.users.system_administrator.accessToken;
    }

    // Private personal-credential creation payloads (scope Tenant, _private).
    genericSecretPayload(
        name: string,
        keyValuePairs: { key: string; value: string }[],
        description = 'Personal generic secret'
    ): GenericSecretCreationRequest {
        return {
            meta: { name, description, scope: Scope.Tenant, _private: true },
            spec: { keyValuePairs }
        };
    }

    ngcApiKeyPayload(
        name: string,
        ngcApiKey: string,
        description = 'Personal ngc-api-key credential'
    ): NgcApiKeyCreationRequest {
        return {
            meta: { name, description, scope: Scope.Tenant, _private: true },
            spec: { ngcApiKey }
        };
    }

    dockerRegistryPayload(
        name: string,
        user: string,
        password: string,
        url: string,
        description = 'Personal docker-registry credential'
    ): DockerRegistryCreationRequest {
        return {
            meta: { name, description, scope: Scope.Tenant, _private: true },
            spec: { existingSecretName: null, user, password, url }
        };
    }
}
