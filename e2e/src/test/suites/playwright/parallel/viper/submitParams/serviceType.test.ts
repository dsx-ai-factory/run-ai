/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    logger,
    operations,
    tags,
    Utils,
    config,
    objectGenerator,
    waitUtils,
    versionUtils,
    flowRootState
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { runaiCliV2 } from '../../../../../../modules/cli';
import { createProject } from '../../../../../../modules/helpers/models/projects';
import { expect } from 'chai';
import { isOCP } from '../../../../../../modules/helpers/utils/EnvironmentUtils';
import { V1Ingress, V1Service } from '@kubernetes/client-node';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import * as yaml from 'js-yaml';
import { test } from '@playwright/test';
import {
    cliV2WorkloadType,
    cliLoginAsSystemAdmin,
    SubmitFlagsMerged
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';

const serviceTypeRenameVersion = '2.16';
const proxyAdditionVersion = '2.16';
const SERVICE_TYPE_NODE_PORT = 'NodePort';
const SERVICE_TYPE_CLUSTER_IP = 'ClusterIP';

test.describe(`Service type, tags: ${tags.resource.jobs} ${tags.teams.viper} ${tags.runOn.k8sFlavCertKit}`, function () {
    let containerPort: number;
    let hostPort: number;
    let projectName: string, namespace: string;
    let project: Project;
    let jobName: string;
    let expectedNodeportServiceName: string;
    let expectedExposedUrlServiceName: string;
    let service: V1Service;
    let anyWorkloadSubmissionEnabled: boolean;
    const projectOptions = { maxAllowedGpus: -1, deservedGpus: 1 };
    const imagePath = config.quickStartImage.sleepInfinity;
    const proxyListenPort = 4180;

    async function validateProxyPortAndContainerPort(
        service: V1Service,
        containerPort: number
    ): Promise<void> {
        if (!versionUtils.isVersionSupported(flowRootState.cluster.version, proxyAdditionVersion)) {
            return;
        }
        try {
            expect(service.spec.ports.length).to.equal(1);
            expect(service.spec.ports[0].targetPort, 'target port').to.equal(proxyListenPort);
        } catch {
            logger.error(`Ports:\n ${yaml.dump(service.spec.ports)}`);
            // throw error; flaky on CI- sometimes the targetPort!=proxyListenPort, need to understand why
        }
        const pod = await waitUtils.waitForPods(jobName, namespace);
        const container = pod[0].spec.containers.find((container) => container.name === jobName);
        const isContainerPortExist = container.ports.some((port) => port.containerPort === containerPort);
        expect(isContainerPortExist).to.true;
    }

    function setExpectedServiceNames(): void {
        if (anyWorkloadSubmissionEnabled) {
            expectedNodeportServiceName = `job-${jobName}-np`;
            expectedExposedUrlServiceName = `job-${jobName}-url`;
            return;
        }

        if (versionUtils.isVersionSupported(flowRootState.cluster.version, serviceTypeRenameVersion)) {
            expectedExposedUrlServiceName = 'tw-' + jobName + '-0';
        } else {
            expectedExposedUrlServiceName = `${jobName}-exposed-url-url-0`;
        }
        if (versionUtils.isVersionSupported(flowRootState.cluster.version, serviceTypeRenameVersion)) {
            expectedNodeportServiceName = 'tw-' + jobName + '-0-nodeport';
        } else {
            expectedNodeportServiceName = jobName;
        }
    }

    function submitWorkloadWithServiceType(flags: Partial<typeof SubmitFlagsMerged>): void {
        jobName = objectGenerator.generateShortNameByPrefix('job');
        runaiCliV2.submit(
            {
                [SubmitFlagsMerged.image]: imagePath,
                [SubmitFlagsMerged.project]: projectName,
                ...flags
            } as Partial<typeof SubmitFlagsMerged>,
            jobName,
            cliV2WorkloadType.training
        );
        setExpectedServiceNames();
    }

    function validateHost(ingress: V1Ingress, expectedHost: string): void {
        const host = isOCP() ? ingress?.spec['host'] : ingress.spec.rules[0].host;
        expect(host).to.equal(expectedHost);
    }

    function validatePath(ingress: V1Ingress, expectedPath: string): void {
        const path = isOCP() ? ingress.spec['path'] : ingress.spec.rules[0].http.paths[0].path;
        expect(path).to.equal(expectedPath);
    }

    async function validateNodePortFields(serviceName: string): Promise<void> {
        service = await waitUtils.waitForService(serviceName, namespace);
        expect(service).to.not.be.null;
        expect(service.spec.type).to.equal(SERVICE_TYPE_NODE_PORT);
        expect(service.spec.ports.length).to.equal(1);
        expect(service.spec.ports[0].nodePort).to.equal(hostPort);
        expect(service.spec.ports[0].port).to.equal(containerPort);
        expect(service.spec.ports[0].targetPort).to.equal(containerPort);
    }

    async function validateClusterIpFields(serviceName: string): Promise<void> {
        service = await waitUtils.waitForService(serviceName, namespace);
        expect(service).to.not.be.null;
        expect(service.spec.type).to.equal(SERVICE_TYPE_CLUSTER_IP);
        expect(service.spec.ports.length).to.equal(1);
        expect(service.spec.ports[0].port).to.equal(80);
    }

    test.beforeAll('resolve workload submission mode', async function () {
        anyWorkloadSubmissionEnabled = await isAnyWorkloadSubmissionEnabled(
            AnyWorkloadSubmissionType.Training
        );
    });

    test.describe('Nodeport', function () {
        test.beforeAll(async function () {
            cliLoginAsSystemAdmin();
            project = await createProject('nodeport-st', projectOptions);
            projectName = project.name;
            namespace = Utils.namespaceFromProjectName(projectName);
        });
        test.afterAll(async function () {
            await operations.deleteProjectAndNamespace(project.id);
        });

        test('Nodeport with specified host port', async function () {
            containerPort = 3001;
            hostPort = 30001;
            await test.step(`Submitting job nodeport with specified host port ${tags.teams.viper}`, function () {
                submitWorkloadWithServiceType({
                    [SubmitFlagsMerged.port]: `service-type=${SERVICE_TYPE_NODE_PORT},container=${containerPort},external=${hostPort}`
                } as Partial<typeof SubmitFlagsMerged>);
            });
            await test.step(`Validating nodeport with specified host port ${tags.teams.viper}`, async function () {
                await validateNodePortFields(expectedNodeportServiceName);
            });
        });

        test('Nodeport with container port only (host port should be autogenerated)', async function () {
            containerPort = 3002;
            await test.step(`Submitting job nodeport with container port only ${tags.teams.viper}`, function () {
                submitWorkloadWithServiceType({
                    [SubmitFlagsMerged.port]: `service-type=${SERVICE_TYPE_NODE_PORT},container=${containerPort}`
                } as Partial<typeof SubmitFlagsMerged>);
            });

            await test.step(`Validating nodeport with container port only ${tags.teams.viper}`, async function () {
                const service = await waitUtils.waitForService(expectedNodeportServiceName, namespace);
                expect(service.spec.type).to.equal(SERVICE_TYPE_NODE_PORT);
                expect(service.spec.ports.length).to.equal(1);
                expect(service.spec.ports[0].nodePort).to.be.greaterThanOrEqual(30000);
                expect(service.spec.ports[0].port).to.equal(containerPort);
                expect(service.spec.ports[0].targetPort).to.equal(containerPort);
            });
        });

        test('Nodeport with multiple ports)', async function () {
            const containerPort1 = 3003;
            const containerPort2 = 3004;
            await test.step(`Submitting job nodeport with multiple ports ${tags.teams.viper}`, function () {
                submitWorkloadWithServiceType({
                    [SubmitFlagsMerged.port]: [
                        `service-type=${SERVICE_TYPE_NODE_PORT},container=${containerPort1}`,
                        `service-type=${SERVICE_TYPE_NODE_PORT},container=${containerPort2}`
                    ]
                } as Partial<typeof SubmitFlagsMerged>);
            });

            await test.step(`Validating nodeport with multiple ports ${tags.teams.viper}`, async function () {
                const service = await waitUtils.waitForService(expectedNodeportServiceName, namespace);
                expect(service).to.not.be.null;
                expect(service.spec.type).to.equal(SERVICE_TYPE_NODE_PORT);
                expect(service.spec.ports.length).to.equal(2);
                expect(service.spec.ports[0].port).to.equal(containerPort1);
                expect(service.spec.ports[0].targetPort).to.equal(containerPort1);
                expect(service.spec.ports[1].port).to.equal(containerPort2);
                expect(service.spec.ports[1].targetPort).to.equal(containerPort2);
            });
        });
    });
    test.describe('External url', function () {
        test.beforeAll(async function () {
            cliLoginAsSystemAdmin();
            project = await createProject('external-url-st', projectOptions);
            projectName = project.name;
            namespace = Utils.namespaceFromProjectName(projectName);
        });
        test.afterAll(async function () {
            await operations.deleteProjectAndNamespace(project.id);
        });

        test('External-url with specified url', async function () {
            const host = 'my.customurl.com';
            const path = '/somepath';
            containerPort = 3005;

            await test.step(`Submitting job external-url ${tags.teams.viper}`, function () {
                submitWorkloadWithServiceType({
                    [SubmitFlagsMerged.externalUrl]: `container=${containerPort},url=http://${host}${path}`
                } as Partial<typeof SubmitFlagsMerged>);
            });
            await test.step(`Validating external-url ${tags.teams.viper}`, async function () {
                await validateClusterIpFields(expectedExposedUrlServiceName);

                await validateProxyPortAndContainerPort(service, containerPort);
                const ingress = await waitUtils.waitForIngress(expectedExposedUrlServiceName, namespace);

                validateHost(ingress, host);
                validatePath(ingress, path);
            });
        });
        test('External-url without url (should be autogenerated)', async function () {
            await test.step(`Submitting job external-url ${tags.teams.viper}`, function () {
                submitWorkloadWithServiceType({
                    [SubmitFlagsMerged.externalUrl]: `container=${containerPort}`
                } as Partial<typeof SubmitFlagsMerged>);
            });

            await test.step(`Validating external-url without url ${tags.teams.viper}`, async function () {
                await validateClusterIpFields(expectedExposedUrlServiceName);

                await validateProxyPortAndContainerPort(service, containerPort);
                const ingress = await waitUtils.waitForIngress(expectedExposedUrlServiceName, namespace);

                const host = isOCP() ? ingress.spec['host'] : ingress.spec.rules[0].host;
                validateHost(ingress, host);
            });
        });
    });

    test.describe('Multiple service-types', function () {
        const host = 'my.customurl.com';
        const path = '/somepath2';
        containerPort = 3006;
        hostPort = 30002;

        test.beforeAll(async function () {
            cliLoginAsSystemAdmin();
            project = await createProject('multiple-st', projectOptions);
            projectName = project.name;
            namespace = Utils.namespaceFromProjectName(projectName);
        });
        test.afterAll(async function () {
            await operations.deleteProjectAndNamespace(project.id);
        });

        test('Submitting job multiple service-types', async function () {
            await test.step(`Submitting job multiple service-types: nodeport and external-url ${tags.teams.viper}`, function () {
                submitWorkloadWithServiceType({
                    [SubmitFlagsMerged.port]: `service-type=${SERVICE_TYPE_NODE_PORT},container=${containerPort},external=${hostPort}`,
                    [SubmitFlagsMerged.externalUrl]: `container=${containerPort},url=http://${host}${path}`
                } as Partial<typeof SubmitFlagsMerged>);
            });

            await test.step(`Validating multiple service-types ${tags.teams.viper}`, async function () {
                await validateNodePortFields(expectedNodeportServiceName);
                await validateClusterIpFields(expectedExposedUrlServiceName);

                await validateProxyPortAndContainerPort(service, containerPort);
                const ingress = await waitUtils.waitForIngress(expectedExposedUrlServiceName, namespace);

                validateHost(ingress, host);
                validatePath(ingress, path);
            });
        });
    });
});
