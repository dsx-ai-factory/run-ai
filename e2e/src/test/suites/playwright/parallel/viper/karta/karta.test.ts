/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import { fileURLToPath } from 'url';
import { operations, projectResourcesUtils, tags, workloadUtils } from '../../../../../../modules/helpers';
import { crds } from '../../../../../../modules/helpers';
import { kartaUtils } from '../../../../../../modules/helpers/utils/kartaUtils';
import { kubeCli, bashCli } from '../../../../../../modules/cli';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Karta operator, ${tags.teams.viper}`, () => {
    supportedFrom(SupportedVersions.V3_0)(function () {
        const __filename = fileURLToPath(import.meta.url);
        const __dirname = path.dirname(__filename);
        let project: Project;
        let categoryId: string;
        let priorityId: string;

        test.beforeAll(async () => {
            project = await projectResourcesUtils.createProjectAndNamespace('karta');
            const [categories, priorities] = await Promise.all([
                workloadUtils.getWorkloadCategories(),
                workloadUtils.getWorkloadPriorities()
            ]);
            categoryId = categories[0].id;
            priorityId = priorities[0].id;
        });

        test.afterAll(async () => {
            if (project?.id) {
                await operations.deleteProjectAndNamespace(project.id);
            }
        });

        test('native Kartas are Ready', async () => {
            const nativeKartas = [
                { group: '', kind: 'Pod', version: 'v1' },
                { group: 'apps', kind: 'Deployment', version: 'v1' },
                { group: 'apps', kind: 'ReplicaSet', version: 'v1' },
                { group: 'apps', kind: 'StatefulSet', version: 'v1' },
                { group: 'batch', kind: 'CronJob', version: 'v1' },
                { group: 'batch', kind: 'Job', version: 'v1' }
            ];

            for (const { group, kind, version } of nativeKartas) {
                const label = group ? `${group}/${kind}` : kind;
                await test.step(`${label}/v1 Karta is Ready`, async () => {
                    await kartaUtils.waitForKartaReady(group, kind, version, 60);
                    await kartaUtils.expectKartaReady(group, kind, version);
                });
            }
        });

        test('API rejects workload type without statusDefinition', async () => {
            await expect(kartaUtils.createWorkloadType(false, categoryId, priorityId)).rejects.toThrow(/400/);
        });

        test('Karta CRDExists condition goes True when CRD is installed', async () => {
            const fakeCrdFile = path.resolve(__dirname, 'testdata/e2e.karta.test_e2efakeresources.yaml');
            let workloadTypeId: string | undefined;

            try {
                workloadTypeId = await kartaUtils.createWorkloadType(true, categoryId, priorityId);
                await kartaUtils.waitForKartaConditionByWorkloadTypeId(
                    workloadTypeId,
                    'Validated',
                    'True',
                    60
                );

                const karta = await kartaUtils.getKartaByWorkloadTypeId(workloadTypeId);
                expect(kartaUtils.getCondition(karta, 'CRDExists')?.status).toBe('False');

                bashCli.exec(`kubectl apply -f ${fakeCrdFile}`);

                await kartaUtils.waitForKartaConditionByWorkloadTypeId(
                    workloadTypeId,
                    'CRDExists',
                    'True',
                    240
                );
                await kartaUtils.waitForKartaConditionByWorkloadTypeId(workloadTypeId, 'Ready', 'True', 60);
                const kartaWithCrd = await kartaUtils.getKartaByWorkloadTypeId(workloadTypeId);
                expect(kartaUtils.getCondition(kartaWithCrd, 'CRDExists')?.status).toBe('True');
                expect(kartaUtils.getCondition(kartaWithCrd, 'Ready')?.status).toBe('True');
            } finally {
                if (workloadTypeId) {
                    try {
                        await kartaUtils.deleteWorkloadType(workloadTypeId);
                    } catch {
                        /* noop */
                    }
                }
                bashCli.exec('kubectl delete crd e2efakeresources.e2e.karta.test --ignore-not-found');
            }
        });

        test('deleted Karta is removed from the cluster', async () => {
            let workloadTypeId: string | undefined;

            try {
                workloadTypeId = await kartaUtils.createWorkloadType(true, categoryId, priorityId);
                await kartaUtils.waitForKartaConditionByWorkloadTypeId(
                    workloadTypeId,
                    'Validated',
                    'True',
                    60
                );

                const deletedId = workloadTypeId;
                await kartaUtils.deleteWorkloadType(deletedId);
                workloadTypeId = undefined;

                await kartaUtils.waitForKartaGoneByWorkloadTypeId(deletedId, 120);
            } finally {
                if (workloadTypeId) {
                    try {
                        await kartaUtils.deleteWorkloadType(workloadTypeId);
                    } catch {
                        /* noop */
                    }
                }
            }
        });

        test('all Kartas have required conditions and ownership invariants hold', async () => {
            const allKartas = (await kubeCli.getCrds(crds.kartaCrd)) as any[];
            expect(allKartas.length, 'Expected at least one Karta to exist on the cluster').toBeGreaterThan(
                0
            );

            const requiredConditions = [
                'Validated',
                'CRDExists',
                'Ready',
                'RBACReady',
                'SyncEnabled',
                'RunaiReady'
            ];

            for (const karta of allKartas) {
                const name = karta.metadata?.name ?? 'unknown';
                const conditions: any[] = karta?.status?.conditions ?? [];

                for (const condType of requiredConditions) {
                    const cond = conditions.find((c: any) => c.type === condType);
                    expect(cond, `Karta ${name}: condition ${condType} not present`).toBeDefined();
                }

                const isTrue = (type: string): boolean =>
                    (conditions.find((c: any) => c.type === type)?.status ?? 'False') === 'True';

                expect(
                    isTrue('Ready'),
                    `Karta ${name}: invariant violated — Ready must equal Validated && CRDExists`
                ).toBe(isTrue('Validated') && isTrue('CRDExists'));

                expect(
                    isTrue('RunaiReady'),
                    `Karta ${name}: invariant violated — RunaiReady must equal Ready && RBACReady`
                ).toBe(isTrue('Ready') && isTrue('RBACReady'));
            }
        });
    });
});
