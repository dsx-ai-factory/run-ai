/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import { kubeCli } from '../../cli';
import { workloadsService } from '../../api';
import { crds } from '../models/customResourceDefinition';
import { flowRootState, logger, waitUtils } from '..';
import { kartaE2EFakeGroup, kartaE2EFakeKind, kartaE2EFakeVersion } from '../config/const';

const KARTA_CONDITIONS = ['Validated', 'CRDExists', 'Ready', 'RBACReady', 'SyncEnabled', 'RunaiReady'] as const;
type KartaConditionType = (typeof KARTA_CONDITIONS)[number];

export class KartaUtils {
    async getKartaForGvk(group: string, kind: string, version: string): Promise<any | undefined> {
        const kartas = (await kubeCli.getCrds(crds.kartaCrd)) as any[];
        return kartas.find((k) => {
            const labels = k.metadata?.labels ?? {};
            return (
                labels['run.ai/karta-group'] === group &&
                labels['run.ai/karta-kind'] === kind &&
                labels['run.ai/karta-version'] === version
            );
        });
    }

    async getKartaByWorkloadTypeId(workloadTypeId: string): Promise<any | undefined> {
        workloadsService.workloadPropertiesApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const { body: wt } = await workloadsService.workloadPropertiesApi.getWorkloadType(workloadTypeId);
        const version = wt.kartas?.[0]?.version;
        if (!wt.name || !wt.group || !version) return undefined;
        return this.getKartaForGvk(wt.group, wt.name, version);
    }

    getConditions(karta: any): any[] {
        return karta?.status?.conditions ?? [];
    }

    getCondition(karta: any, type: KartaConditionType): any | undefined {
        return this.getConditions(karta).find((c: any) => c.type === type);
    }

    async expectKartaReady(group: string, kind: string, version: string): Promise<void> {
        const karta = await this.getKartaForGvk(group, kind, version);
        expect(karta, `Karta for ${group}/${kind}/${version} not found`).to.not.be.undefined;

        for (const condType of KARTA_CONDITIONS) {
            const cond = this.getCondition(karta, condType);
            expect(cond, `Karta ${group}/${kind}/${version}: condition ${condType} not present`).to.not.be.undefined;
            expect(cond.status, `Karta ${group}/${kind}/${version}: condition ${condType} is not True`).to.equal('True');
        }

        const isTrue = (type: KartaConditionType): boolean => this.getCondition(karta, type)?.status === 'True';
        expect(
            isTrue('Ready'),
            `Karta ${group}/${kind}/${version}: invariant violated — Ready != Validated && CRDExists`
        ).to.equal(isTrue('Validated') && isTrue('CRDExists'));
        expect(
            isTrue('RunaiReady'),
            `Karta ${group}/${kind}/${version}: invariant violated — RunaiReady != Ready && RBACReady`
        ).to.equal(isTrue('Ready') && isTrue('RBACReady'));
    }

    async getKartaByName(name: string): Promise<any> {
        return kubeCli.getCrd(crds.kartaCrd, name);
    }

    async waitForKartaConditionByName(
        name: string,
        type: string,
        expectedStatus: 'True' | 'False',
        timeoutSeconds = 60
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                try {
                    const karta = await this.getKartaByName(name);
                    const cond = this.getConditions(karta).find((c: any) => c.type === type);
                    return cond?.status === expectedStatus;
                } catch {
                    return false;
                }
            },
            `Karta ${name}: condition ${type}=${expectedStatus}`,
            true,
            timeoutSeconds
        );
    }

    async waitForKartaConditionByWorkloadTypeId(
        workloadTypeId: string,
        type: string,
        expectedStatus: 'True' | 'False',
        timeoutSeconds = 60
    ): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                try {
                    const karta = await this.getKartaByWorkloadTypeId(workloadTypeId);
                    if (!karta) return false;
                    const cond = this.getConditions(karta).find((c: any) => c.type === type);
                    return cond?.status === expectedStatus;
                } catch {
                    return false;
                }
            },
            `Karta for workload type ${workloadTypeId}: condition ${type}=${expectedStatus}`,
            true,
            timeoutSeconds
        );
    }

    async waitForKartaGoneByWorkloadTypeId(workloadTypeId: string, timeoutSeconds = 60): Promise<void> {
        // The workload type is already deleted at this point, so we check the cluster directly by label
        await waitUtils.waitForCondition(
            async () => {
                try {
                    const kartas = (await kubeCli.getCrds(crds.kartaCrd)) as any[];
                    return kartas.find((k) => k.metadata?.labels?.['run.ai/workload-type-id'] === workloadTypeId) === undefined;
                } catch {
                    return false;
                }
            },
            `Karta for workload type ${workloadTypeId} to be removed`,
            true,
            timeoutSeconds
        );
    }

    async waitForKartaReady(
        group: string,
        kind: string,
        version: string,
        timeoutSeconds = 120
    ): Promise<void> {
        logger.info(`Waiting for Karta ${group}/${kind}/${version} to be Ready (timeout: ${timeoutSeconds}s)`);
        await waitUtils.waitForCondition(
            async () => {
                try {
                    const karta = await this.getKartaForGvk(group, kind, version);
                    if (!karta) return false;
                    return this.getCondition(karta, 'Ready')?.status === 'True';
                } catch {
                    return false;
                }
            },
            `Karta ${group}/${kind}/${version} Ready`,
            true,
            timeoutSeconds
        );
    }

    async createWorkloadType(
        withStatusDefinition: boolean,
        categoryId: string,
        priorityId: string,
        group = kartaE2EFakeGroup,
        kind = kartaE2EFakeKind
    ): Promise<string> {
        const rootComponent: Record<string, unknown> = {
            name: 'root',
            kind: { group, version: kartaE2EFakeVersion, kind },
            specDefinition: { podTemplateSpecPath: '.spec.template' }
        };
        if (withStatusDefinition) {
            rootComponent.statusDefinition = { statusMappings: {} };
        }
        const kartaSpec = { spec: { structureDefinition: { rootComponent } } };

        workloadsService.workloadPropertiesApi.accessToken = flowRootState.users.system_administrator.accessToken;
        const response = await workloadsService.workloadPropertiesApi.createWorkloadType({
            name: kind,
            group,
            categoryId,
            priorityId,
            kartas: [kartaSpec]
        });
        return response.body.id;
    }

    async deleteWorkloadType(workloadTypeId: string): Promise<void> {
        workloadsService.workloadPropertiesApi.accessToken = flowRootState.users.system_administrator.accessToken;
        await workloadsService.workloadPropertiesApi.deleteWorkloadType(workloadTypeId);
    }
}

export const kartaUtils = new KartaUtils();
