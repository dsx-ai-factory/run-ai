/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'allure-playwright';
import { kubeCli } from '../../cli';
import { flowRootState } from '../config/config';
import { crds, logger, versionUtils, waitUtils } from '../index';

const MiB_TO_MB_MULTIPLY_FACTOR = 1.048576; // (2^20 / 10^6)

export class CrdUtils {
    roundMemoryQuotaToThreeDigitsAfterDecimal(queuespec: any): any {
        if (!queuespec || !queuespec['memory']) {
            return queuespec;
        }

        const fields = ['quota', 'overQuotaWeight', 'limit', 'deserved'];
        for (const field of fields) {
            if (queuespec['memory'][field]) {
                queuespec['memory'][field] = this.roundNumberToThreeDigitsAfterDecimal(
                    queuespec['memory'][field]
                );
            }
        }
        return queuespec;
    }

    convertQueueSpecToCrdVersion(queuespec: any, isOverQuotaFFOn = false): any {
        if (
            queuespec &&
            queuespec['memory'] &&
            versionUtils.isVersionSupported(flowRootState.cluster.version, '2.13')
        ) {
            const quota = queuespec['memory']['quota'];
            const overQuotaWeight = queuespec['memory']['overQuotaWeight'];
            const limit = queuespec['memory']['limit'];
            const convertedQuota = this.convertNumberFromMibToMB(quota);

            if (!isOverQuotaFFOn && convertedQuota >= 0) {
                // if they are not equal - probably means over quota weight FF is on, so keep the original value
                queuespec['memory']['overQuotaWeight'] = convertedQuota;
            } else {
                queuespec['memory']['overQuotaWeight'] = overQuotaWeight;
            }
            queuespec['memory']['quota'] = convertedQuota;
            queuespec['memory']['limit'] = this.convertNumberFromMibToMB(limit);
        }

        const currentVersion = flowRootState.cluster.version;

        if (!versionUtils.isVersionSupported(currentVersion, '2.14')) {
            const stack = [queuespec]; // Start with the original object in the stack

            while (stack.length > 0) {
                const current = stack.pop();
                if (typeof current === 'object') {
                    if (current.quota !== undefined) {
                        current.deserved = current.quota;
                        delete current.quota;
                    }
                    for (const key in current) {
                        stack.push(current[key]);
                    }
                }
            }
            return queuespec;
        }
        return queuespec;
    }

    convertNumberFromMibToMB(num: number): number {
        if (num === null || num <= 0) {
            return num;
        }
        return Math.round(num * MiB_TO_MB_MULTIPLY_FACTOR * 1000) / 1000;
    }

    roundNumberToThreeDigitsAfterDecimal(num: number): number {
        return Math.round(num * 1000) / 1000;
    }

    async validateDepartmentQueueCrdResources(
        deptId: number,
        expectedResources: {
            gpu?: { quota: number; overQuotaWeight: number; limit: number };
            cpu?: { quota: number; overQuotaWeight: number; limit: number };
            memory?: { quota: number; overQuotaWeight: number; limit: number };
        },
        nodepoolName: string = 'default'
    ): Promise<void> {
        const queueName = nodepoolName === 'default' ? `q-${deptId}` : `q-${deptId}-${nodepoolName}`;

        await waitUtils.waitForCondition(
            async () => {
                try {
                    const queueCrd = await kubeCli.getCrd(crds.queueCrd, queueName);
                    const queueResources = queueCrd['spec']['resources'];
                    logger.info(`Queue CRD resources config: ${JSON.stringify(queueResources)}`);

                    this.validateResources(expectedResources, queueResources);

                    return true;
                } catch (err) {
                    logger.warn(`Failed to validate the expected queue CRD config: ${err}`);
                    return false;
                }
            },
            'wait for the queue CRD to have the expected resource quota config',
            true,
            20,
            2
        );
    }

    async validateProjectQueueCrdResources(
        projectName: string,
        expectedResources: {
            gpu?: { quota: number; overQuotaWeight: number; limit: number };
            cpu?: { quota: number; overQuotaWeight: number; limit: number };
            memory?: { quota: number; overQuotaWeight: number; limit: number };
        },
        nodepoolName: string = 'default'
    ): Promise<void> {
        const queueName = nodepoolName === 'default' ? projectName : `${projectName}-${nodepoolName}`;

        await waitUtils.waitForCondition(
            async () => {
                try {
                    const queueCrd = await kubeCli.getCrd(crds.queueCrd, queueName);
                    const queueResources = queueCrd['spec']['resources'];
                    logger.info(`Queue CRD resources config: ${JSON.stringify(queueResources)}`);

                    this.validateResources(expectedResources, queueResources);

                    return true;
                } catch (err) {
                    logger.warn(`Failed to validate the expected queue CRD config: ${err}`);
                    return false;
                }
            },
            'wait for the queue CRD to have the expected resource quota config',
            true,
            20,
            2
        );
    }

    validateResources(
        expectedResources: {
            gpu?: { quota: number; overQuotaWeight: number; limit: number };
            cpu?: { quota: number; overQuotaWeight: number; limit: number };
            memory?: { quota: number; overQuotaWeight: number; limit: number };
        },
        actualResources: {
            gpu?: { quota: number; overQuotaWeight: number; limit: number };
            cpu?: { quota: number; overQuotaWeight: number; limit: number };
            memory?: { quota: number; overQuotaWeight: number; limit: number };
        }
    ): void {
        // validate gpu resources
        if (expectedResources.gpu) {
            expect(actualResources.gpu.quota, 'actualResources.gpu.quota check failed').toBe(
                expectedResources.gpu.quota
            );
            expect(
                actualResources.gpu.overQuotaWeight,
                'actualResources.gpu.overQuotaWeight check failed'
            ).toBe(expectedResources.gpu.overQuotaWeight);
            expect(actualResources.gpu.limit, 'actualResources.gpu.limit check failed').toBe(
                expectedResources.gpu.limit
            );
        }

        // validate cpu resources
        if (expectedResources.cpu) {
            expect(actualResources.cpu.quota, 'actualResources.cpu.quota check failed').toBe(
                expectedResources.cpu.quota
            );
            expect(
                actualResources.cpu.overQuotaWeight,
                'actualResources.cpu.overQuotaWeight check failed'
            ).toBe(expectedResources.cpu.overQuotaWeight);
            expect(actualResources.cpu.limit, 'actualResources.cpu.limit check failed').toBe(
                expectedResources.cpu.limit
            );
        }

        // validate cpu memory resources
        if (expectedResources.memory) {
            expect(actualResources.memory.quota, 'actualResources.memory.quota check failed').toBe(
                expectedResources.memory.quota
            );
            expect(
                actualResources.memory.overQuotaWeight,
                'actualResources.memory.overQuotaWeight check failed'
            ).toBe(expectedResources.memory.overQuotaWeight);
            expect(actualResources.memory.limit, 'actualResources.memory.limit check failed').toBe(
                expectedResources.memory.limit
            );
        }
    }
}
