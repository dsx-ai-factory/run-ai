/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect, test } from '@playwright/test';
import type { V1CustomResourceDefinition } from '@kubernetes/client-node';
import { tags } from '../../modules/helpers/config/tags';
import {
    isDynamoTopologyBindingInSync,
    resolveDynamoMockRuntimeVersion
} from '../../modules/helpers/utils/DynamoFixtureUtils';

const definition = (supported: boolean, served = true): V1CustomResourceDefinition => ({
    spec: {
        group: 'nvidia.com',
        names: { kind: 'DynamoGraphDeployment', plural: 'dynamographdeployments' },
        scope: 'Namespaced',
        versions: ['v1alpha1', 'v1beta1'].map((name) => ({
            name,
            served,
            storage: name === 'v1beta1',
            schema: {
                openAPIV3Schema: {
                    type: 'object',
                    properties: {
                        spec: {
                            type: 'object',
                            properties: {
                                services: {
                                    type: 'object',
                                    additionalProperties: {
                                        type: 'object',
                                        properties: supported
                                            ? { runtimeVersionOverride: { type: 'string' } }
                                            : {}
                                    }
                                },
                                components: {
                                    type: 'array',
                                    items: {
                                        type: 'object',
                                        properties: supported
                                            ? { runtimeVersionOverride: { type: 'string' } }
                                            : {}
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }))
    }
});

test.describe(`Dynamo mock runtime selection, tags: ${tags.teams.viper}`, () => {
    for (const version of ['v1alpha1', 'v1beta1']) {
        test(`${version}: omits unsupported fields on older CRDs`, () => {
            expect(resolveDynamoMockRuntimeVersion(definition(false), version, [])).toBeUndefined();
            expect(resolveDynamoMockRuntimeVersion(definition(true, false), version, [])).toBeUndefined();
        });

        test(`${version}: uses the installed operator release instead of a fixed version`, () => {
            expect(
                resolveDynamoMockRuntimeVersion(definition(true), version, [
                    'nvcr.io/nvidia/ai-dynamo/kubernetes-operator:1.4.2'
                ])
            ).toBe('1.4.2');
            expect(
                resolveDynamoMockRuntimeVersion(definition(true), version, [
                    'mirror.example/operator:v1.5.0-rc.1@sha256:abcdef'
                ])
            ).toBe('1.5.0-rc.1');
        });
    }

    test('rejects absent, unversioned and conflicting operator releases', () => {
        for (const images of [[], ['operator:latest'], ['operator:1.4.2', 'operator:1.5.0']]) {
            expect(() => resolveDynamoMockRuntimeVersion(definition(true), 'v1beta1', images)).toThrow(
                'Cannot resolve one Dynamo operator release'
            );
        }
    });
});

test.describe('Grove topology binding readiness', () => {
    const ready = (): { binding: any; topology: any } => ({
        binding: {
            metadata: { name: 'fixture', generation: 2 },
            status: {
                observedGeneration: 2,
                conditions: [{ type: 'SchedulerTopologyDrift', status: 'False' }],
                schedulerTopologyStatuses: [
                    {
                        schedulerName: 'kai-scheduler',
                        topologyReference: 'fixture',
                        inSync: true,
                        schedulerBackendTopologyObservedGeneration: 3
                    }
                ]
            }
        },
        topology: {
            metadata: {
                name: 'fixture',
                generation: 3,
                annotations: { 'run.ai/grove-topology-name': 'fixture' }
            }
        }
    });

    test('requires current Grove and KAI generations and admission annotation', () => {
        const { binding, topology } = ready();
        expect(isDynamoTopologyBindingInSync(binding, topology)).toBe(true);
        binding.status.observedGeneration = 1;
        expect(isDynamoTopologyBindingInSync(binding, topology)).toBe(false);
        binding.status.observedGeneration = 2;
        topology.metadata.generation = 4;
        expect(isDynamoTopologyBindingInSync(binding, topology)).toBe(false);
    });

    test('does not treat missing annotation or drift as ready', () => {
        const { binding, topology } = ready();
        delete topology.metadata.annotations['run.ai/grove-topology-name'];
        expect(isDynamoTopologyBindingInSync(binding, topology)).toBe(false);
        topology.metadata.annotations['run.ai/grove-topology-name'] = 'fixture';
        binding.status.conditions[0].status = 'True';
        expect(isDynamoTopologyBindingInSync(binding, topology)).toBe(false);
    });
});
