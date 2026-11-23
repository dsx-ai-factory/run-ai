/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { V1EnvVar } from '@kubernetes/client-node';

import { kubeCli } from '../../../../../../modules/cli';
import { logger, waitUtils } from '../../../../../../modules/helpers';
import {
    ANY_WORKLOAD_SUBMISSION_ENV,
    resolveWorkloadsContainerIndex,
    setAnyWorkloadSubmissionEnabledCache,
    WORKLOADS_DEPLOYMENT,
    WORKLOADS_NAMESPACE
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

const ROLLOUT_TIMEOUT_SECONDS = 240;

export type RestoreAnyWorkload = () => Promise<void>;

async function setContainerEnvValue(
    containerIndex: number,
    originalEnv: V1EnvVar[] | undefined,
    name: string,
    value: string
): Promise<void> {
    const nextEnv = [...(originalEnv ?? []).filter((entry) => entry.name !== name), { name, value }];
    await kubeCli.patchDeployment(WORKLOADS_DEPLOYMENT, WORKLOADS_NAMESPACE, [
        {
            op: originalEnv === undefined ? 'add' : 'replace',
            path: `/spec/template/spec/containers/${containerIndex}/env`,
            value: nextEnv
        }
    ]);
}

async function waitForWorkloadsRollout(): Promise<void> {
    await waitUtils.waitForCondition(
        async () => {
            const deployment = await kubeCli.getDeployment(WORKLOADS_DEPLOYMENT, WORKLOADS_NAMESPACE);
            const desired = deployment.spec?.replicas ?? 1;
            const status = deployment.status ?? {};
            const generationSettled =
                (status.observedGeneration ?? 0) >= (deployment.metadata?.generation ?? 0);
            return (
                generationSettled &&
                (status.updatedReplicas ?? 0) >= desired &&
                (status.readyReplicas ?? 0) >= desired &&
                (status.availableReplicas ?? 0) >= desired &&
                (status.unavailableReplicas ?? 0) === 0
            );
        },
        `wait for ${WORKLOADS_NAMESPACE}/${WORKLOADS_DEPLOYMENT} rollout to complete`,
        true,
        ROLLOUT_TIMEOUT_SECONDS,
        3
    );
}

export async function enableAnyWorkloadSubmission(): Promise<RestoreAnyWorkload> {
    const deployment = await kubeCli.getDeployment(WORKLOADS_DEPLOYMENT, WORKLOADS_NAMESPACE);
    const containers = deployment.spec?.template?.spec?.containers ?? [];
    const containerIndex = resolveWorkloadsContainerIndex(containers);
    const originalEnv = containers[containerIndex].env;
    const currentValue = originalEnv?.find((env) => env.name === ANY_WORKLOAD_SUBMISSION_ENV)?.value;

    if (currentValue === 'true') {
        setAnyWorkloadSubmissionEnabledCache(true);
        logger.info(
            `${ANY_WORKLOAD_SUBMISSION_ENV} already on for ${WORKLOADS_NAMESPACE}/${WORKLOADS_DEPLOYMENT}; leaving untouched`
        );
        return async () => {};
    }

    await setContainerEnvValue(containerIndex, originalEnv, ANY_WORKLOAD_SUBMISSION_ENV, 'true');
    await waitForWorkloadsRollout();
    setAnyWorkloadSubmissionEnabledCache(true);
    logger.info(`enabled ${ANY_WORKLOAD_SUBMISSION_ENV} on ${WORKLOADS_NAMESPACE}/${WORKLOADS_DEPLOYMENT}`);

    return async () => {
        try {
            const envPath = `/spec/template/spec/containers/${containerIndex}/env`;
            const restorePatch =
                originalEnv === undefined
                    ? { op: 'remove', path: envPath }
                    : { op: 'replace', path: envPath, value: originalEnv };
            await kubeCli.patchDeployment(WORKLOADS_DEPLOYMENT, WORKLOADS_NAMESPACE, [restorePatch]);
            await waitForWorkloadsRollout();
            setAnyWorkloadSubmissionEnabledCache(undefined);
            logger.info(
                `restored ${ANY_WORKLOAD_SUBMISSION_ENV} on ${WORKLOADS_NAMESPACE}/${WORKLOADS_DEPLOYMENT}`
            );
        } catch (error) {
            setAnyWorkloadSubmissionEnabledCache(undefined);
            logger.error(`failed to restore ${ANY_WORKLOAD_SUBMISSION_ENV}: ${error}`);
        }
    };
}
