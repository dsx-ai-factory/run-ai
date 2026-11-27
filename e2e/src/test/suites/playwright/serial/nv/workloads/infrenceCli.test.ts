/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    flowRootState,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    waitUtils,
    PodStatus,
    workloadsCrudUtils,
    workloadsUtilsV2,
    workloadUtils
} from '../../../../../../modules/helpers';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import { isKnativeInstalled } from '../../../../../../modules/helpers/utils/IsInstalledUtils';
import { Phase } from '../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { kubeCli } from '../../../../../../modules/cli';
import { createInference, sendInferenceRequest } from '../../../parallel/viper/inference/inferenceUtils.js';
import { InferenceSpecSpec } from '../../../../../../modules/api/apis/workloadsApi/generated/model/inferenceSpecSpec';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { ServingPortProtocol } from '../../../../../../modules/api/apis/workloadsApi/generated/model/servingPortProtocol';
import { skipPlaywrightTest } from '../../../../../../modules/helpers/utils/SkippedTestUtils';

let project: Project;
let workloadId: string;
let workloadName: string;
let workloadSpec: InferenceSpecSpec;

test.describe(
    `inference workload different timeout tests, tags: ${tags.resource.inferenceWorkload} ${tags.teams.viper} ${tags.teams.loong}`,
    { tag: [tags.resource.inferenceWorkload] },
    function () {
        test.setTimeout(300 * 1000);
        const pythonScript = `import http.server
import socketserver
import time
import os

class DelayedHTTPRequestHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        serv_delay = int(os.environ.get('SERV_DELAY', 0))
        if serv_delay > 0:
            time.sleep(serv_delay)
        self.send_response(200)
        self.send_header('Content-type', 'text/plain')
        self.end_headers()
        self.wfile.write(b'OK\\n')

    def do_POST(self):
        self.do_GET()

    def do_PUT(self):
        self.do_GET()

    def do_DELETE(self):
        self.do_GET()

    def log_message(self, format, *args):
        print(f"[{time.strftime('%Y-%m-%d %H:%M:%S')}] {format % args}")

def start_server():
    initial_delay = int(os.environ.get('INITIAL_START_DELAY', 0))
    print(f"Waiting {initial_delay} seconds before starting HTTP server...")
    time.sleep(initial_delay)

    PORT = 8000
    Handler = DelayedHTTPRequestHandler

    with socketserver.TCPServer(("", PORT), Handler) as httpd:
        print(f"HTTP server started on port {PORT}")
        print(f"Service delay: {os.environ.get('SERV_DELAY', 0)} seconds")
        httpd.serve_forever()

if __name__ == "__main__":
    start_server()`;
        supportedFrom(SupportedVersions.V2_23)(function () {
            test.beforeAll(async function () {
                if (!(await isKnativeInstalled())) {
                    logger.warn('knative is not installed on the cluster, skipping test');
                    test.skip();
                }

                project = await projectResourcesUtils.createProjectAndNamespace('project', null, 8, 8);

                workloadSpec = {
                    image: 'python:3.11-slim',
                    imagePullPolicy: 'IfNotPresent' as any,
                    nodePools: ['default'],
                    environmentVariables: [
                        {
                            name: 'INITIAL_START_DELAY',
                            value: '10'
                        },
                        {
                            name: 'SERV_DELAY',
                            value: '2'
                        },
                        {
                            name: 'PYTHONUNBUFFERED',
                            value: '1'
                        }
                    ],
                    annotations: [
                        {
                            name: 'http-delay-testing',
                            value: 'True'
                        }
                    ],
                    labels: [
                        {
                            name: 'app',
                            value: 'http-delay-server'
                        }
                    ],
                    servingPort: {
                        container: 8000,
                        protocol: ServingPortProtocol.Http
                    },
                    compute: {
                        cpuCoreRequest: 0.2,
                        cpuMemoryRequest: '128Mi',
                        gpuDevicesRequest: 0
                    },
                    autoscaling: {
                        minReplicas: 1,
                        maxReplicas: 1,
                        metric: 'throughput',
                        metricThreshold: 1
                    }
                };
            });

            test.beforeEach(function () {
                workloadName = objectGenerator.generateNameByPrefix('w');

                const encodedScript = Buffer.from(pythonScript).toString('base64');
                workloadSpec.command = '/bin/sh';
                workloadSpec.args =
                    `-c 'echo "${encodedScript}" | base64 -d > /tmp/server.py && python /tmp/server.py'` as any;
                workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
            });

            test.afterAll(async function () {
                try {
                    await operations.deleteProjectNamespaceAndWorkloads(project, 120);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and project ${project.name} namespace ${project.namespace} with error ${error}`
                    );
                }
            });

            test.afterEach(async function () {
                try {
                    await workloadUtils.clearProjectWorkloadsAndPods(project.name, 120);
                } catch (error) {
                    logger.error(
                        `Failed to delete workloads and pods of project ${project.name} namespace ${project.namespace} with error ${error}`
                    );
                }
            });

            test('T5193710:large initializationTimeoutSeconds then the workload initialize and become ready', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                await test.step('Create inference workload with INITIAL_START_DELAY=80 and initializationTimeoutSeconds=120', async () => {
                    workloadSpec.environmentVariables!.find(
                        (env) => env.name === 'INITIAL_START_DELAY'
                    )!.value = '80';
                    workloadSpec.servingConfiguration = {
                        initializationTimeoutSeconds: 120
                    };
                    [workloadId, workloadName] = await createInference(workloadName, workloadSpec, project);
                });

                await test.step('Wait for inference workload to be in Ready status', async () => {
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        project.id,
                        Phase.Running,
                        150000
                    );
                    const clusterWorkload = await kubeCli.getInferenceWorkload(
                        workloadName,
                        project.namespace
                    );
                    expect(clusterWorkload).to.have.nested.property('status.phase', 'Ready');
                    logger.info(`Inference workload ${workloadName} is now in Ready status`);
                });

                await test.step('Verify workload is in Running status', async () => {
                    const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workload.body.phase).to.equal('Running');
                });
            });

            test('T5193711:small initializationTimeoutSeconds then the workload become failed', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                await test.step('Create inference workload with INITIAL_START_DELAY=80 and initializationTimeoutSeconds=20', async () => {
                    workloadSpec.environmentVariables!.find(
                        (env) => env.name === 'INITIAL_START_DELAY'
                    )!.value = '80';
                    workloadSpec.servingConfiguration = {
                        initializationTimeoutSeconds: 20
                    };
                    [workloadId, workloadName] = await createInference(workloadName, workloadSpec, project);
                });

                await test.step('Wait for inference workload to be in Failed status', async () => {
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        project.id,
                        Phase.Failed,
                        150000
                    );
                    const clusterWorkload = await kubeCli.getInferenceWorkload(
                        workloadName,
                        project.namespace
                    );
                    expect(clusterWorkload).to.have.nested.property('status.phase', 'Failed');
                    logger.info(`Inference workload ${workloadName} is now in Failed status`);
                });

                await test.step('Verify workload is in Failed status', async () => {
                    const workload = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
                    expect(workload.body.phase).to.equal('Failed');
                });
            });

            test('T5193712:large requestTimeoutSeconds then the workload accept and answer to request', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                await test.step('Create inference workload with SERV_DELAY=50 and requestTimeoutSeconds=60', async () => {
                    workloadSpec.environmentVariables!.find((env) => env.name === 'SERV_DELAY')!.value = '50';
                    workloadSpec.servingConfiguration = {
                        requestTimeoutSeconds: 60
                    };
                    [workloadId, workloadName] = await createInference(workloadName, workloadSpec, project);
                });

                await test.step('Wait for inference workload to be in Ready status', async () => {
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        project.id,
                        Phase.Running,
                        150000
                    );
                    const clusterWorkload = await kubeCli.getInferenceWorkload(
                        workloadName,
                        project.namespace
                    );
                    expect(clusterWorkload).to.have.nested.property('status.phase', 'Ready');
                    logger.info(`Inference workload ${workloadName} is now in Ready status`);
                });

                await test.step('Send inference call to the URL', async () => {
                    const knativeService = await waitUtils.waitForKnativeService(
                        workloadName,
                        project.namespace
                    );
                    const url = knativeService['status']['address']['url'];
                    const command = `-w "Time: %{time_total}s\nResponse Code: %{http_code}\n" -o /dev/null -s ${url}`;
                    const clientWorkloadName = await sendInferenceRequest(project.namespace, command);
                    const logs = await kubeCli.getLogs(clientWorkloadName, project.namespace);

                    const timeMatch = logs.match(/Time: ([\d.]+)s/);
                    const codeMatch = logs.match(/Response Code: (\d+)/);

                    logger.info('Time match:', timeMatch);
                    logger.info('Code match:', codeMatch);

                    const responseTime = parseFloat(timeMatch[1]);
                    const responseCode = parseInt(codeMatch[1]);

                    expect(responseTime).to.be.greaterThan(47);
                    expect(responseTime).to.be.lessThan(53);
                    expect(responseCode).to.equal(200);
                });
            });

            test('T5193713:small requestTimeoutSeconds then the workload drops long requests', async function ({}, testInfo) {
                skipPlaywrightTest(
                    test,
                    testInfo,
                    'test depends on the legacy InferenceWorkload CR, which the any-workload path no longer creates (it produces a WorkloadRunner instead)'
                );
                await test.step('Create inference workload with SERV_DELAY=50 and requestTimeoutSeconds=30', async () => {
                    workloadSpec.environmentVariables!.find((env) => env.name === 'SERV_DELAY')!.value = '50';
                    workloadSpec.servingConfiguration = {
                        requestTimeoutSeconds: 30
                    };
                    [workloadId, workloadName] = await createInference(workloadName, workloadSpec, project);
                });

                await test.step('Wait for inference workload to be in Ready status', async () => {
                    await waitUtils.waitForBackendWorkloadPhase(
                        workloadName,
                        project.id,
                        Phase.Running,
                        150000
                    );
                    const clusterWorkload = await kubeCli.getInferenceWorkload(
                        workloadName,
                        project.namespace
                    );
                    expect(clusterWorkload).to.have.nested.property('status.phase', 'Ready');
                    logger.info(`Inference workload ${workloadName} is now in Ready status`);
                });

                await test.step('Send inference call to the URL and verify timeout', async () => {
                    const knativeService = await waitUtils.waitForKnativeService(
                        workloadName,
                        project.namespace
                    );
                    const url = knativeService['status']['address']['url'];

                    const command = `-w "Time: %{time_total}s\nResponse Code: %{http_code}\n" -o /dev/null -s ${url}`;
                    const clientWorkloadName = await sendInferenceRequest(
                        project.namespace,
                        command,
                        '',
                        false
                    );
                    const logs = await kubeCli.getLogs(clientWorkloadName, project.namespace);
                    const timeMatch = logs.match(/Time: ([\d.]+)s/);
                    const codeMatch = logs.match(/Response Code: (\d+)/);

                    logger.info('Time match:', timeMatch);
                    logger.info('Code match:', codeMatch);

                    const responseTime = parseFloat(timeMatch[1]);
                    const responseCode = parseInt(codeMatch[1]);

                    expect(responseTime).to.be.greaterThan(27);
                    expect(responseTime).to.be.lessThan(33);
                    expect(responseCode).to.equal(504);
                });
            });
        });
    }
);
