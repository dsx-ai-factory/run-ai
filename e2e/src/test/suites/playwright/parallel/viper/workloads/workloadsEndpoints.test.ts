/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    crds,
    flowRootState,
    isOCP,
    logger,
    objectGenerator,
    operations,
    projectUtils,
    tags,
    Utils,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../modules/helpers';
import { createProject } from '../../../../../../modules/helpers/models/projects';
import { workloadsCrudService } from '../../../../../../modules/api';
import { expect } from 'chai';
import { bashCli, kubeCli } from '../../../../../../modules/cli';
import fs from 'fs';
import { test } from '@playwright/test';
import {
    createClusterIPServiceYaml,
    createDeployment,
    createGatewayYaml,
    createGRPCRouteYaml,
    createHTTPRouteYaml,
    createInferencePoolYaml,
    createIngressYaml,
    createLoadBalancerEndpointYaml,
    createNodePortEndpointYaml,
    createTLSRouteYaml,
    createKnativeServiceYaml,
    createOpenshiftRouteYaml,
    createDgdYaml,
    createNimServiceYaml,
    createAiAppHelmChart,
    getLoadBalancerIngress
} from '../../../../../../modules/helpers/utils/WorkloadEndpointsUtils';
import {
    isKGatewayInstalled,
    isInferencePoolCRDInstalled,
    isKnativeInstalled,
    isNimOperatorInstalled,
    isDynamoV1Alpha1Served
} from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { WorkloadType } from '../../../../../../modules/helpers/utils/WorkloadsUtilsV2';
import { WorkloadDetailed } from '../../../../../../modules/api/apis/workloadsApi/generated/model/workloadDetailed';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { Endpoint } from '../../../../../../modules/api/apis/workloadsApi/generated/model/endpoint';
import { createInference } from '../../viper/inference/inferenceUtils';
import { AIApplicationVerbosity } from '../../../../../../modules/api/apis/workloadsApi/generated/model/aIApplicationVerbosity';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

test.describe(
    `Workloads endpoints tests, tags: ${tags.teams.viper} ${tags.resource.workloads} ${tags.runOn.parallelCI}`,
    { tag: [tags.services.workloads, tags.services.externalWorkloadIntegrator, tags.runOn.parallelCI] },
    function () {
        supportedFrom(SupportedVersions.V2_25)(function () {
            let project;
            let namespace;
            let workload;
            let servicesToCleanup: string[] = [];

            function getIngressHost(): string {
                const baseDomain = process.env.E2E_BASE_DOMAIN;
                if (!baseDomain) {
                    throw new Error('E2E_BASE_DOMAIN must be set for workload endpoint tests');
                }
                return `test.${baseDomain}`;
            }

            test.beforeAll(async function () {
                const options = { maxAllowedGpus: -1, deservedGpus: 1 };
                project = await createProject('project1', options);
                namespace = Utils.namespaceFromProjectName(project.name);
            });

            test.afterAll(async function () {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch (error) {
                    logger.error(`failed to delete project: ${project.name}`, error);
                }
                await projectUtils.waitForProjectCrdToBeDeleted(project.name);
            });

            test.afterEach(async function () {
                if (workload?.name) {
                    try {
                        await kubeCli.deleteCrd(crds.runaiTrainingWorkloadCrd, workload.name, namespace);
                    } catch (error) {
                        logger.debug(`failed to delete workload ${workload.name}: ${error}`);
                    }
                }
                workload = null;

                // Delete any service created by the test, even when the test body failed,
                // so a leftover does not keep its (node)port allocated for the next run.
                for (const serviceName of servicesToCleanup) {
                    await kubeCli.deleteService(serviceName, namespace);
                }
                servicesToCleanup = [];
            });

            function validateEndpoint(
                endpoint: Endpoint,
                host: string,
                port: number,
                protocol: string,
                path: string,
                tls: boolean,
                url: string
            ): void {
                expect(endpoint.host).to.equal(host);
                expect(endpoint.port).to.equal(port);
                expect(endpoint.path).to.equal(path);
                expect(endpoint.protocol.toLowerCase()).to.equal(protocol.toLowerCase());
                expect(endpoint.tlsEnabled).to.equal(tls);
                expect(endpoint.url.replace(/\/$/, '')).to.equal(url.replace(/\/$/, ''));
            }

            test.describe('Sanity discovery tests', function () {
                test('test nodePort endpoints', async function () {
                    const app = 'ai-inference';
                    const serviceName = `${app}-svc`;

                    const deploymentName = objectGenerator.generateNameByPrefix('d');
                    const deployment = await kubeCli.createDeployment(
                        createDeployment(deploymentName, namespace, project.name, app)
                    );

                    // Register for cleanup first, so afterEach tears the service down even if
                    // the pre-delete/wait below throws.
                    servicesToCleanup.push(serviceName);
                    // Remove any leftover service from a prior aborted run so its nodePort is
                    // freed before we re-create it.
                    await kubeCli.deleteService(serviceName, namespace);
                    await waitUtils.waitForCondition(
                        async () => (await kubeCli.getServiceByName(namespace, serviceName)) === undefined,
                        `service ${serviceName} to be deleted before re-create`,
                        true,
                        30
                    );
                    await kubeCli.createAllFromYaml(createNodePortEndpointYaml(app, namespace));
                    const workloadId = deployment.metadata.uid;

                    await workloadsUtilsV2.waitForWorkloadWithEndpoints(workloadId);
                    workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workload.body.endpoints.length).to.equal(1);
                    let endpoint = workload.body.endpoints[0];

                    const nodeAddresses = (await kubeCli.getNodes()).items
                        .map((node) => node.status.addresses)
                        .flat()
                        .map((address) => address.address);

                    expect(nodeAddresses).to.contain(endpoint.host);
                    validateEndpoint(
                        endpoint,
                        endpoint.host,
                        32057,
                        'TCP',
                        '/',
                        undefined,
                        endpoint.host + ':32057'
                    );
                    const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId)).body;
                    expect(workloadEndpoints.endpoints.length).to.equal(1);
                    endpoint = workloadEndpoints.endpoints[0];
                    validateEndpoint(
                        endpoint,
                        endpoint.host,
                        32057,
                        'TCP',
                        '/',
                        undefined,
                        endpoint.host + ':32057'
                    );
                });

                test('test loadBalancer endpoints', async function () {
                    const app = 'ai-inference';
                    const lbServiceName = `${app}-lb-service`;

                    const deploymentName = objectGenerator.generateNameByPrefix('d');
                    const deployment = await kubeCli.createDeployment(
                        createDeployment(deploymentName, namespace, project.name, app)
                    );
                    await kubeCli.createAllFromYaml(createLoadBalancerEndpointYaml(app, namespace));

                    // Wait for a LB controller to assign an IP, or patch manually if none exists
                    const lbIp = await getLoadBalancerIngress(namespace, lbServiceName);

                    const workloadId = deployment.metadata.uid;

                    await workloadsUtilsV2.waitForWorkloadWithEndpoints(workloadId);
                    workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workload.body.endpoints.length).to.equal(1);
                    let endpoint = workload.body.endpoints[0];
                    validateEndpoint(endpoint, lbIp, 443, 'TCP', '/', true, lbIp + ':443');
                    const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId)).body;
                    expect(workloadEndpoints.endpoints.length).to.equal(1);
                    endpoint = workloadEndpoints.endpoints[0];
                    validateEndpoint(endpoint, lbIp, 443, 'TCP', '/', true, lbIp + ':443');
                });

                test('test ingress endpoints', async function () {
                    const app = objectGenerator.generateNameByPrefix('ingress-app');
                    const serviceName = `${app}-clusterip-svc`;
                    const ingressHost = `${app}.example.com`;

                    const deploymentName = objectGenerator.generateNameByPrefix('d');
                    const deployment = await kubeCli.createDeployment(
                        createDeployment(deploymentName, namespace, project.name, app)
                    );

                    await kubeCli.createAllFromYaml(createClusterIPServiceYaml(app, namespace, serviceName));
                    await kubeCli.createAllFromYaml(createIngressYaml(namespace, serviceName, ingressHost));
                    const workloadId = deployment.metadata.uid;

                    await workloadsUtilsV2.waitForWorkloadWithEndpoints(workloadId);
                    workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workload.body.endpoints.length).to.equal(1);

                    let endpoint = workload.body.endpoints[0];
                    validateEndpoint(
                        endpoint,
                        ingressHost,
                        80,
                        'http',
                        '/predict',
                        false,
                        'http://' + endpoint.host + endpoint.path
                    );
                    const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId)).body;
                    expect(workloadEndpoints.endpoints.length).to.equal(1);
                    endpoint = workloadEndpoints.endpoints[0];
                    validateEndpoint(
                        endpoint,
                        ingressHost,
                        80,
                        'http',
                        '/predict',
                        false,
                        'http://' + endpoint.host + endpoint.path
                    );
                });

                test.describe('Gateway API endpoints', function () {
                    test.beforeAll(async function () {
                        if (!(await isKGatewayInstalled())) {
                            logger.warn('gateway api is not installed on the cluster, skipping test');
                            test.skip();
                        }
                    });

                    test('test httpRoute endpoints', async function () {
                        const app = objectGenerator.generateNameByPrefix('httproute-app');
                        const serviceName = `${app}-clusterip-svc`;
                        const gatewayName = `${app}-gw`;
                        const routeName = `${app}-httproute`;
                        const hostname = `${app}.example.com`;
                        const listenerPort = 8080;

                        const deploymentName = objectGenerator.generateNameByPrefix('d');
                        const deployment = await kubeCli.createDeployment(
                            createDeployment(deploymentName, namespace, project.name, app)
                        );

                        await kubeCli.createAllFromYaml(
                            createClusterIPServiceYaml(app, namespace, serviceName)
                        );
                        await kubeCli.createAllFromYaml(
                            createGatewayYaml(namespace, gatewayName, listenerPort, 'HTTP')
                        );
                        await kubeCli.createAllFromYaml(
                            createHTTPRouteYaml(
                                namespace,
                                routeName,
                                gatewayName,
                                serviceName,
                                hostname,
                                8000,
                                '/predict'
                            )
                        );
                        const workloadId = deployment.metadata.uid;

                        await workloadsUtilsV2.waitForWorkloadWithEndpoints(workloadId);
                        workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                        expect(workload.body.endpoints.length).to.equal(1);

                        let endpoint = workload.body.endpoints[0];
                        validateEndpoint(
                            endpoint,
                            hostname,
                            listenerPort,
                            'http',
                            '/predict',
                            false,
                            'http://' + endpoint.host + ':' + endpoint.port + endpoint.path
                        );
                        const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId))
                            .body;
                        expect(workloadEndpoints.endpoints.length).to.equal(1);
                        endpoint = workloadEndpoints.endpoints[0];
                        validateEndpoint(
                            endpoint,
                            hostname,
                            listenerPort,
                            'http',
                            '/predict',
                            false,
                            'http://' + endpoint.host + ':' + endpoint.port + endpoint.path
                        );
                    });

                    test.skip('test tlsRoute endpoints', async function () {
                        // TLSRoute is not supported by kgateway
                        const app = objectGenerator.generateNameByPrefix('tlsroute-app');
                        const serviceName = `${app}-clusterip-svc`;
                        const gatewayName = `${app}-gw`;
                        const routeName = `${app}-tlsroute`;
                        const hostname = `${app}.example.com`;
                        const listenerPort = 8443;

                        const deploymentName = objectGenerator.generateNameByPrefix('d');
                        const deployment = await kubeCli.createDeployment(
                            createDeployment(deploymentName, namespace, project.name, app)
                        );

                        await kubeCli.createAllFromYaml(
                            createClusterIPServiceYaml(app, namespace, serviceName)
                        );
                        await kubeCli.createAllFromYaml(
                            createGatewayYaml(namespace, gatewayName, listenerPort, 'TLS', 'tls-listener')
                        );
                        await kubeCli.createAllFromYaml(
                            createTLSRouteYaml(namespace, routeName, gatewayName, serviceName, hostname, 8000)
                        );
                        const workloadId = deployment.metadata.uid;

                        await workloadsUtilsV2.waitForWorkloadWithEndpoints(workloadId);
                        workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                        expect(workload.body.endpoints.length).to.equal(1);

                        let endpoint = workload.body.endpoints[0];
                        validateEndpoint(
                            endpoint,
                            hostname,
                            listenerPort,
                            'https',
                            '/',
                            true,
                            'https://' + endpoint.host + ':' + endpoint.port + endpoint.path
                        );
                        const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId))
                            .body;
                        expect(workloadEndpoints.endpoints.length).to.equal(1);
                        endpoint = workloadEndpoints.endpoints[0];
                        validateEndpoint(
                            endpoint,
                            hostname,
                            listenerPort,
                            'https',
                            '/',
                            true,
                            'https://' + endpoint.host + ':' + endpoint.port + endpoint.path
                        );
                    });

                    test('test grpcRoute endpoints', async function () {
                        const app = objectGenerator.generateNameByPrefix('grpcroute-app');
                        const serviceName = `${app}-clusterip-svc`;
                        const gatewayName = `${app}-gw`;
                        const routeName = `${app}-grpcroute`;
                        const hostname = `${app}.example.com`;
                        const listenerPort = 8080;

                        const deploymentName = objectGenerator.generateNameByPrefix('d');
                        const deployment = await kubeCli.createDeployment(
                            createDeployment(deploymentName, namespace, project.name, app)
                        );

                        await kubeCli.createAllFromYaml(
                            createClusterIPServiceYaml(app, namespace, serviceName)
                        );
                        await kubeCli.createAllFromYaml(
                            createGatewayYaml(namespace, gatewayName, listenerPort, 'HTTP')
                        );
                        await kubeCli.createAllFromYaml(
                            createGRPCRouteYaml(
                                namespace,
                                routeName,
                                gatewayName,
                                serviceName,
                                hostname,
                                8000
                            )
                        );
                        const workloadId = deployment.metadata.uid;

                        await workloadsUtilsV2.waitForWorkloadWithEndpoints(workloadId);
                        workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                        expect(workload.body.endpoints.length).to.equal(1);

                        let endpoint = workload.body.endpoints[0];
                        validateEndpoint(
                            endpoint,
                            hostname,
                            listenerPort,
                            'http',
                            '/',
                            false,
                            'http://' + endpoint.host + ':' + endpoint.port + endpoint.path
                        );
                        const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId))
                            .body;
                        expect(workloadEndpoints.endpoints.length).to.equal(1);
                        endpoint = workloadEndpoints.endpoints[0];
                        validateEndpoint(
                            endpoint,
                            hostname,
                            listenerPort,
                            'http',
                            '/',
                            false,
                            'http://' + endpoint.host + ':' + endpoint.port + endpoint.path
                        );
                    });

                    test('test inferencePool endpoints', async function () {
                        if (!(await isInferencePoolCRDInstalled())) {
                            logger.warn('InferencePool CRD is not installed on the cluster, skipping test');
                            test.skip();
                        }
                        const app = objectGenerator.generateNameByPrefix('infpool-app');
                        const poolName = `${app}-pool`;
                        const gatewayName = `${app}-gw`;
                        const routeName = `${app}-httproute`;
                        const hostname = `${app}.example.com`;
                        const listenerPort = 8080;

                        const deploymentName = objectGenerator.generateNameByPrefix('d');
                        const deployment = await kubeCli.createDeployment(
                            createDeployment(deploymentName, namespace, project.name, app)
                        );

                        await kubeCli.createAllFromYaml(
                            createInferencePoolYaml(namespace, poolName, app, 8000)
                        );
                        await kubeCli.createAllFromYaml(
                            createGatewayYaml(namespace, gatewayName, listenerPort, 'HTTP')
                        );
                        await kubeCli.createAllFromYaml(
                            createHTTPRouteYaml(
                                namespace,
                                routeName,
                                gatewayName,
                                poolName,
                                hostname,
                                8000,
                                '/inf',
                                'InferencePool',
                                'inference.networking.k8s.io'
                            )
                        );
                        const workloadId = deployment.metadata.uid;

                        await workloadsUtilsV2.waitForWorkloadWithEndpoints(workloadId);
                        workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                        expect(workload.body.endpoints.length).to.equal(1);

                        let endpoint = workload.body.endpoints[0];
                        validateEndpoint(
                            endpoint,
                            hostname,
                            listenerPort,
                            'http',
                            '/inf',
                            false,
                            'http://' + endpoint.host + ':' + endpoint.port + endpoint.path
                        );
                        const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId))
                            .body;
                        expect(workloadEndpoints.endpoints.length).to.equal(1);
                        endpoint = workloadEndpoints.endpoints[0];
                        validateEndpoint(
                            endpoint,
                            hostname,
                            listenerPort,
                            'http',
                            '/inf',
                            false,
                            'http://' + endpoint.host + ':' + endpoint.port + endpoint.path
                        );
                    });
                });
            });

            test.describe('openshift route endpoints', function () {
                test.beforeAll(function () {
                    if (!isOCP()) {
                        logger.warn('not openshift, skipping test');
                        test.skip();
                    }
                });

                test('test route endpoints', async function () {
                    const app = objectGenerator.generateNameByPrefix('route-app');
                    const serviceName = `${app}-clusterip-svc`;
                    const routeName = `${app}-route`;
                    const routeHost = `${app}.apps.example.com`;

                    const deploymentName = objectGenerator.generateNameByPrefix('d');
                    const deployment = await kubeCli.createDeployment(
                        createDeployment(deploymentName, namespace, project.name, app)
                    );

                    await kubeCli.createAllFromYaml(createClusterIPServiceYaml(app, namespace, serviceName));
                    await kubeCli.createAllFromYaml(
                        createOpenshiftRouteYaml(namespace, routeName, serviceName, routeHost, '/api')
                    );
                    const workloadId = deployment.metadata.uid;

                    await workloadsUtilsV2.waitForWorkloadWithEndpoints(workloadId);
                    workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workload.body.endpoints.length).to.equal(1);

                    let endpoint = workload.body.endpoints[0];
                    validateEndpoint(
                        endpoint,
                        routeHost,
                        443,
                        'https',
                        '/api',
                        true,
                        'https://' + routeHost + '/api'
                    );
                    const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId)).body;
                    expect(workloadEndpoints.endpoints.length).to.equal(1);
                    endpoint = workloadEndpoints.endpoints[0];
                    validateEndpoint(
                        endpoint,
                        routeHost,
                        443,
                        'https',
                        '/api',
                        true,
                        'https://' + routeHost + '/api'
                    );
                });
            });

            test.describe('Test handlers with specific urls', function () {
                test('test knative workloads', async function () {
                    if (!(await isKnativeInstalled())) {
                        logger.warn('knative is not installed on the cluster, skipping test');
                        test.skip();
                    }

                    const app = objectGenerator.generateNameByPrefix('ksvc');
                    const ksvcName = app;

                    const ksvcYaml = createKnativeServiceYaml(
                        namespace,
                        ksvcName,
                        app,
                        'https://app.run.ai/jobs/123'
                    );
                    const tmpFile = `/tmp/ksvc-${ksvcName}.yaml`;
                    fs.writeFileSync(tmpFile, ksvcYaml);
                    bashCli.exec(`kubectl apply -n ${namespace} -f ${tmpFile}`);

                    // Get the knative Service's UID - the EWI uses this as the workload ID
                    const ksvc = await kubeCli.getKnativeService(ksvcName, namespace);
                    const workloadId = ksvc.metadata.uid;

                    await workloadsUtilsV2.waitForWorkloadWithEndpoints(workloadId);
                    workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workload.body.endpoints.length).to.equal(2);

                    let endpoint1 = workload.body.endpoints[0];
                    let endpoint2 = workload.body.endpoints[1];
                    if (endpoint1.host === 'app.run.ai') {
                        endpoint1 = workload.body.endpoints[1];
                        endpoint2 = workload.body.endpoints[0];
                    }
                    expect(endpoint1.host).to.contain(ksvcName);
                    const url = new URL(workload.body.urls[1]);
                    const ksvcScheme = url.protocol.replace(/:$/, '');
                    const ksvcPort = Number(url.port || (ksvcScheme === 'https' ? 443 : 80));
                    const ksvcTls = ksvcScheme === 'https';
                    validateEndpoint(
                        endpoint1,
                        url.hostname,
                        ksvcPort,
                        ksvcScheme,
                        '/',
                        ksvcTls,
                        workload.body.urls[1]
                    );
                    validateEndpoint(
                        endpoint2,
                        'app.run.ai',
                        443,
                        'https',
                        '/jobs/123',
                        true,
                        'https://app.run.ai/jobs/123'
                    );

                    const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId)).body;
                    expect(workloadEndpoints.endpoints.length).to.equal(2);
                    endpoint1 = workloadEndpoints.endpoints[0];
                    endpoint2 = workloadEndpoints.endpoints[1];
                    if (endpoint1.host === 'app.run.ai') {
                        endpoint1 = workloadEndpoints.endpoints[1];
                        endpoint2 = workloadEndpoints.endpoints[0];
                    }
                    expect(endpoint1.host).to.contain(ksvcName);
                    validateEndpoint(
                        endpoint1,
                        url.hostname,
                        ksvcPort,
                        ksvcScheme,
                        '/',
                        ksvcTls,
                        workload.body.urls[1]
                    );
                    validateEndpoint(
                        endpoint2,
                        'app.run.ai',
                        443,
                        'https',
                        '/jobs/123',
                        true,
                        'https://app.run.ai/jobs/123'
                    );
                });

                test('test dgd workloads', async function () {
                    // The DGD submitted below is v1alpha1.
                    if (!(await isDynamoV1Alpha1Served())) {
                        logger.warn('dynamo does not serve v1alpha1 on the cluster, skipping test');
                        test.skip();
                    }

                    const app = objectGenerator.generateNameByPrefix('dgd');
                    const dgdName = app;
                    const ingressHost = getIngressHost();
                    const expectedUrl = `https://${app}-api.${ingressHost}`;

                    const dgdYaml = createDgdYaml(namespace, dgdName, app, ingressHost);
                    const tmpFile = `/tmp/dgd-${dgdName}.yaml`;
                    fs.writeFileSync(tmpFile, dgdYaml);
                    bashCli.exec(`kubectl apply -n ${namespace} -f ${tmpFile}`);

                    // Wait for the ExternalWorkload to be created by the EWI
                    await waitUtils.waitForWorkload(
                        crds.externalWorkloadCrd,
                        `ew-dynamographdeployment-${dgdName}`,
                        namespace
                    );
                    workload = await kubeCli.getWorkload(
                        `ew-dynamographdeployment-${dgdName}`,
                        namespace,
                        crds.externalWorkloadCrd
                    );
                    const workloadId = workload.spec.originalWorkloadUID;

                    await workloadsUtilsV2.waitForWorkloadWithEndpointUrl(workloadId, expectedUrl);
                    workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);

                    expect(workload.body.endpoints.length).to.equal(1);
                    let endpoint = workload.body.endpoints[0];
                    validateEndpoint(
                        endpoint,
                        `${app}-api.${ingressHost}`,
                        443,
                        'https',
                        '/',
                        true,
                        expectedUrl
                    );

                    const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId)).body;
                    expect(workloadEndpoints.endpoints.length).to.equal(1);
                    endpoint = workloadEndpoints.endpoints[0];
                    validateEndpoint(
                        endpoint,
                        `${app}-api.${ingressHost}`,
                        443,
                        'https',
                        '/',
                        true,
                        expectedUrl
                    );
                });

                test('test nim-service workloads', async function () {
                    if (!(await isNimOperatorInstalled())) {
                        logger.warn('nim is not installed on the cluster, skipping test');
                        test.skip();
                    }

                    const app = objectGenerator.generateNameByPrefix('nim');
                    const nimName = app;
                    const ingressHost = getIngressHost();
                    const expectedUrl = `https://${nimName}.${ingressHost}`;

                    const nimYaml = createNimServiceYaml(namespace, nimName, app, ingressHost);
                    const tmpFile = `/tmp/nim-${nimName}.yaml`;
                    fs.writeFileSync(tmpFile, nimYaml);
                    bashCli.exec(`kubectl apply -n ${namespace} -f ${tmpFile}`);

                    // Wait for the ExternalWorkload to be created by the EWI
                    await waitUtils.waitForWorkload(
                        crds.externalWorkloadCrd,
                        `ew-nimservice-${nimName}`,
                        namespace
                    );
                    workload = await kubeCli.getWorkload(
                        `ew-nimservice-${nimName}`,
                        namespace,
                        crds.externalWorkloadCrd
                    );
                    const workloadId = workload.spec.originalWorkloadUID;

                    await workloadsUtilsV2.waitForWorkloadWithEndpointUrl(workloadId, expectedUrl);
                    workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);

                    expect(workload.body.endpoints.length).to.equal(1);
                    let endpoint = workload.body.endpoints[0];
                    validateEndpoint(
                        endpoint,
                        `${nimName}.${ingressHost}`,
                        443,
                        'https',
                        '/',
                        true,
                        expectedUrl
                    );

                    const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId)).body;
                    expect(workloadEndpoints.endpoints.length).to.equal(1);
                    endpoint = workloadEndpoints.endpoints[0];
                    validateEndpoint(
                        endpoint,
                        `${nimName}.${ingressHost}`,
                        443,
                        'https',
                        '/',
                        true,
                        expectedUrl
                    );
                });
            });

            test.describe('Test runai workload', function () {
                let trainingWorkload: WorkloadDetailed;

                test.beforeAll(async function () {
                    if (!(await isKnativeInstalled())) {
                        logger.warn('knative is not installed on the cluster, skipping test');
                        test.skip();
                    }
                });

                test.afterAll(async function () {
                    if (trainingWorkload) {
                        await workloadsUtilsV2.deleteWorkload(trainingWorkload.id, WorkloadType.TRAINING);
                    }
                    trainingWorkload = null;
                });

                test('test inference workload', async ({}, testInfo) => {
                    skipPlaywrightTest(
                        test,
                        testInfo,
                        'control-plane returns endpoints: null for inference workloads (only the raw urls field is populated)'
                    );
                    workloadsCrudService.inferencesApi.accessToken =
                        flowRootState.users.system_administrator.accessToken;
                    const inferenceName = objectGenerator.generateShortNameByPrefix('inference');

                    logger.info('Creating inference workload with serving port');
                    const inferSpec = {
                        image: 'jupyter/base-notebook:latest',
                        compute: { gpuDevicesRequest: 0 },
                        servingPort: {
                            container: 8888
                        },
                        nodePools: ['default']
                    };

                    const [workloadId] = await createInference(inferenceName, inferSpec, project);

                    await workloadsUtilsV2.waitForWorkloadWithEndpoints(workloadId);
                    workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workload.body.environments.length).to.equal(1);
                    expect(workload.body.environments[0].connections.length).to.equal(1);

                    expect(workload.body.endpoints.length).to.equal(1);
                    let endpoint = workload.body.endpoints[0];
                    const url = new URL(workload.body.environments[0].connections[0].url);
                    const hostName = url.hostname;
                    const ksvcScheme = url.protocol.replace(/:$/, '');
                    const ksvcTls = ksvcScheme.toLowerCase() === 'https';

                    validateEndpoint(
                        endpoint,
                        hostName,
                        8888,
                        ksvcScheme,
                        undefined,
                        ksvcTls,
                        workload.body.environments[0].connections[0].url
                    );
                    const workloadEndpoints = (await workloadsUtilsV2.getWorkloadEndpoints(workloadId)).body;
                    expect(workloadEndpoints.endpoints.length).to.equal(1);
                    endpoint = workloadEndpoints.endpoints[0];
                    validateEndpoint(
                        endpoint,
                        hostName,
                        8888,
                        ksvcScheme,
                        undefined,
                        ksvcTls,
                        workload.body.environments[0].connections[0].url
                    );
                });
            });

            test.describe('AI applications', function () {
                const aiAppReleaseName = objectGenerator.generateNameByPrefix('ai-app');

                test.beforeAll(function () {
                    if (isOCP()) {
                        logger.warn('openshift env, skipping test');
                        test.skip();
                    }
                });

                test.afterAll(function () {
                    try {
                        bashCli.execWithOutput(`helm uninstall ${aiAppReleaseName} -n ${namespace}`, false);
                    } catch (error) {
                        logger.error(`Failed to uninstall AI app helm release: ${error}`);
                    }
                });

                test('test AI application endpoints', async () => {
                    const ingressHost = getIngressHost();
                    // Under parallel CI load, AI application endpoint population can approach the
                    // 180s wait below, which exceeds the 150s self-hosted default test timeout.
                    // Give this test enough headroom for the wait plus helm install and assertions.
                    test.setTimeout(300_000);
                    const chartDir = createAiAppHelmChart(aiAppReleaseName, ingressHost, project.name);
                    bashCli.exec(`helm install ${aiAppReleaseName} -n ${namespace} ${chartDir}`);

                    // Wait for AI application to appear with endpoints
                    let aiApp;
                    await waitUtils.waitForCondition(
                        async () => {
                            const res = await workloadsUtilsV2.listApplications({
                                verbosity: AIApplicationVerbosity.Full
                            });
                            const apps = res.body.applications ?? [];
                            aiApp = apps.find((a) => a.name === aiAppReleaseName);
                            return aiApp && aiApp.endpoints?.length >= 2;
                        },
                        'Wait for AI application with 2 endpoints',
                        true,
                        180,
                        5
                    );

                    expect(aiApp.endpoints.length).to.equal(2);

                    // Verify the two endpoints come from different workloads
                    const workloadIds = [...new Set(aiApp.endpoints.map((ep) => ep.workloadId))];
                    expect(workloadIds.length).to.equal(2);

                    // Find the nodeport and ingress endpoints from the application
                    const nodePortEndpoint = aiApp.endpoints.find((ep) => ep.protocol === 'TCP');
                    const ingressEndpoint = aiApp.endpoints.find((ep) => ep.protocol === 'http');

                    expect(nodePortEndpoint).to.not.be.undefined;
                    expect(ingressEndpoint).to.not.be.undefined;

                    // Validate nodeport endpoint
                    expect(nodePortEndpoint.port).to.not.be.undefined;
                    expect(nodePortEndpoint.protocol).to.equal('TCP');
                    expect(nodePortEndpoint.path).to.equal('/');

                    // Validate ingress endpoint
                    expect(ingressEndpoint.host).to.equal(`${aiAppReleaseName}.${ingressHost}`);
                    expect(ingressEndpoint.port).to.equal(80);
                    expect(ingressEndpoint.path).to.equal('/');
                    expect(ingressEndpoint.protocol).to.equal('http');
                    expect(ingressEndpoint.tlsEnabled).to.equal(false);

                    // Verify each workload has 1 endpoint matching the application endpoint
                    for (const appEndpoint of aiApp.endpoints) {
                        const workloadEndpoints = (
                            await workloadsUtilsV2.getWorkloadEndpoints(appEndpoint.workloadId)
                        ).body;
                        expect(workloadEndpoints.endpoints.length).to.equal(1);

                        const wlEndpoint = workloadEndpoints.endpoints[0];
                        expect(wlEndpoint.host).to.equal(appEndpoint.host);
                        expect(wlEndpoint.port).to.equal(appEndpoint.port);
                        expect(wlEndpoint.protocol).to.equal(appEndpoint.protocol);
                        expect(wlEndpoint.path).to.equal(appEndpoint.path);
                        expect(wlEndpoint.tlsEnabled).to.equal(appEndpoint.tlsEnabled);
                    }
                });
            });
        });
    }
);
