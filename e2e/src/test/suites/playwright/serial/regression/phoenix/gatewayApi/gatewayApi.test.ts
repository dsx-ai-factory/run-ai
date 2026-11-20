/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import {
    crds,
    flowRootState,
    logger,
    settings,
    tags,
    waitUtils
} from '../../../../../../../modules/helpers';
import { kubeCli } from '../../../../../../../modules/cli';
import { shouldRunTests } from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import yaml from 'js-yaml';


test.describe(`Gateway API Validation: ${tags.teams.phoenix} ${tags.resource.gatewayApi} ${tags.runOn.serialCI}`, () => {

    shouldRunTests(settings.isGatewayApiWorkflow)(() => {
    type RunaiConfigLike = {
        spec: any;
        status?: {
            conditions?: any[];
            operands?: Record<string, { ready: boolean; reasons?: string[] }>;
        };
    };
    
    // Expected backend ports for HTTPRoute validation
    const CLUSTER_API_PORT = 4190;
    const TRAEFIK_PORT = 8080;

    test.beforeAll(() => {
        expect(flowRootState.cluster).toBeDefined();
        expect(flowRootState.users).toBeDefined();
        expect(flowRootState.users.system_administrator).toBeDefined();

        logger.info(`Using cluster: ${flowRootState.cluster.name} (${flowRootState.cluster.uuid})`);
    });

    test('runaipublic configmap shows gateway-api available', async () => {
        test.setTimeout(120000);

        await waitUtils.waitForCondition(
            async () => {
                try {
                    const cm = await kubeCli.getConfigMap('runai-public', 'runai');
                    const runaiPublicYaml = cm.data?.['runai-public'];
                    if (!runaiPublicYaml) return false;
                    const runaiPublic = yaml.load(runaiPublicYaml) as any;
                    return runaiPublic?.dependencies?.optional?.['gateway-api']?.available === true;
                } catch {
                    return false;
                }
            },
            'Wait for gateway-api status in runai-public configmap',
            true,
            60
        );

        const cm = await kubeCli.getConfigMap('runai-public', 'runai');
        const runaiPublic = yaml.load(cm.data['runai-public']) as any;
        const gatewayStatus = runaiPublic.dependencies.optional['gateway-api'];

        expect(gatewayStatus.available).toBe(true);
        expect(gatewayStatus.components['httproute-crd'].available).toBe(true);
        expect(gatewayStatus.components.gateway.available).toBe(true);
        logger.info('runaipublic configmap gateway-api status validated');
    });

    test('RunaiConfig operand statuses show Ready for gateway-api operands', async () => {
        test.setTimeout(120000);

        const rc = (await kubeCli.getCrd(
            crds.runaiConfigCrd,
            'runai',
            'runai'
        )) as RunaiConfigLike;

        const operands = rc.status?.operands;
        expect(operands, 'operands should be defined in RunaiConfig status').toBeDefined();
        if (!operands) return;

        const gatewayOperands = ['cluster-api'];
        for (const operand of gatewayOperands) {
            const status = operands[operand];
            expect(status, `operand ${operand} should exist in status`).toBeDefined();
            expect(status.ready, `operand ${operand} should be ready, reasons: ${status.reasons?.join(', ')}`).toBe(true);
            logger.info(`Operand ${operand}: ready=${status.ready}`);
        }
    });

    test('workload-controller has gateway API env vars configured', async () => {
        test.setTimeout(60000);

        const deploy = await kubeCli.getDeployment('workload-controller', 'runai');
        const envVars = deploy.spec.template.spec.containers[0].env;

        const getEnv = (name: string): string | undefined =>
            envVars.find((e: any) => e.name === name)?.value;

        expect(getEnv('GATEWAY_API_ENABLED'), 'GATEWAY_API_ENABLED should be true').toBe('true');
        expect(getEnv('GATEWAY_NAME'), 'GATEWAY_NAME should be set').toBeDefined();
        expect(getEnv('GATEWAY_NAMESPACE'), 'GATEWAY_NAMESPACE should be set').toBeDefined();
        logger.info(`workload-controller env: GATEWAY_API_ENABLED=${getEnv('GATEWAY_API_ENABLED')}, GATEWAY_NAME=${getEnv('GATEWAY_NAME')}, GATEWAY_NAMESPACE=${getEnv('GATEWAY_NAMESPACE')}`);
    });

    test('cluster-side HTTPRoute resources exist and are accepted', async () => {
        test.setTimeout(120000);

        const httpRoutes = (await kubeCli.getCrds(crds.httpRouteCrd, 'runai')) as any[];
        logger.info(`Found ${httpRoutes.length} HTTPRoutes in runai namespace`);

        // Helper to find route by name substring
        const findRoute = (substring: string): any =>
            httpRoutes.find((r: any) => r.metadata.name.includes(substring));

        // Helper to assert HTTPRoute is accepted by gateway
        const assertAccepted = (route: any, name: string): void => {
            const parents = route.status?.parents || [];
            const accepted = parents.some((p: any) =>
                p.conditions?.some(
                    (c: any) => c.type === 'Accepted' && (c.status === 'True' || c.status === true)
                )
            );
            expect(accepted, `${name} HTTPRoute should be accepted by gateway`).toBe(true);
        };

        // Helper to assert parentRefs point to configured gateway
        const assertParentRefs = (route: any, name: string): void => {
            expect(route.spec.parentRefs?.length, `${name} should have parentRefs`).toBeGreaterThanOrEqual(1);
        };

        // Cluster-api HTTPRoute
        const clusterApiRoute = findRoute('cluster-api');
        expect(clusterApiRoute, 'cluster-api HTTPRoute should exist').toBeDefined();
        assertParentRefs(clusterApiRoute, 'cluster-api');
        expect(clusterApiRoute.spec.hostnames?.length, 'cluster-api should have hostnames').toBeGreaterThanOrEqual(1);
        const clusterApiPaths = clusterApiRoute.spec.rules.flatMap((r: any) =>
            r.matches?.map((m: any) => m.path?.value) || []
        );
        expect(clusterApiPaths).toContain('/cluster-api');
        const clusterApiBackend = clusterApiRoute.spec.rules[0].backendRefs?.[0];
        expect(clusterApiBackend?.port).toBe(CLUSTER_API_PORT);
        assertAccepted(clusterApiRoute, 'cluster-api');
        logger.info('cluster-api HTTPRoute validated');

    });

    test('control plane HTTPRoute exists and routes to traefik', async () => {
        test.setTimeout(120000);

        // Control plane HTTPRoutes are in the control plane namespace (same as where traefik runs)
        // List across all namespaces and find the one routing to traefik
        const allHttpRoutes = (await kubeCli.getCrds(crds.httpRouteCrd)) as any[];
        const cpRoute = allHttpRoutes.find((r: any) =>
            r.spec?.rules?.some((rule: any) =>
                rule.backendRefs?.some((ref: any) => ref.name?.includes('traefik'))
            )
        );
        expect(cpRoute, 'control plane HTTPRoute routing to traefik should exist').toBeDefined();
        logger.info(`Found control plane HTTPRoute: ${cpRoute.metadata.name} in ${cpRoute.metadata.namespace}`);

        // Validate parentRefs exist
        expect(cpRoute.spec.parentRefs?.length).toBeGreaterThanOrEqual(1);

        // Validate hostname is set
        expect(cpRoute.spec.hostnames?.length).toBeGreaterThanOrEqual(1);

        // Validate backend ref points to traefik on port 8080
        const traefikBackend = cpRoute.spec.rules
            .flatMap((r: any) => r.backendRefs || [])
            .find((ref: any) => ref.name?.includes('traefik'));
        expect(traefikBackend).toBeDefined();
        expect(traefikBackend.port).toBe(TRAEFIK_PORT);

        // Validate security headers in ResponseHeaderModifier filter
        const filters = cpRoute.spec.rules.flatMap((r: any) => r.filters || []);
        const headerFilter = filters.find(
            (f: any) => f.type === 'ResponseHeaderModifier'
        );
        expect(headerFilter, 'ResponseHeaderModifier filter should exist').toBeDefined();

        const headerNames = headerFilter.responseHeaderModifier.set.map(
            (h: any) => h.name.toLowerCase()
        );
        expect(headerNames).toContain('x-content-type-options');
        expect(headerNames).toContain('x-frame-options');
        expect(headerNames).toContain('referrer-policy');
        logger.info('Control plane HTTPRoute validated');
    });

});
});
