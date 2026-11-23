/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    config,
    distributedInferenceUtils,
    flowRootState,
    inferenceUtils,
    logger,
    objectGenerator,
    orgUnitUtils,
    tags,
    waitUtils,
    workloadsCrudUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';
import {
    isKnativeInstalled,
    isLwsInstalled
} from '../../../../../../../modules/helpers/utils/IsInstalledUtils';
import {
    supportedFrom,
    SupportedVersions
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import { DistributedInferenceSpecSpec } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/distributedInferenceSpecSpec';
import { DistributedInferenceLeaderWorkerSpecCompute } from '../../../../../../../modules/api/apis/workloadsApi/generated/api';

test.describe(`real-gpu inference workload vllm tests: ${tags.resource.inferenceWorkload} ${tags.tool.api} ${tags.teams.viper} ${tags.runOn.k8sFlavWorkloadsKit}`, function () {
    supportedFrom(SupportedVersions.V2_24)(function () {
        // Pin both tests to a validated vLLM release. `:latest` moved the
        // multi-node helper from examples/online_serving to examples/ray_serving
        // in v0.21.0 and dropped Ray from the openai image in v0.19.0, which
        // crashed the LWS worker shortly after startup.
        // Digest is the v0.29.0 multi-arch manifest (amd64/arm64).
        const VLLM_IMAGE =
            'vllm/vllm-openai:v0.29.0@sha256:c2914767605584b6d8f45686b82de173ecc99e781897aa3d0a66dacd72c51ae1';
        const VLLM_MODEL = 'Qwen/Qwen2.5-0.5B-Instruct';
        const VLLM_MULTI_NODE_SCRIPT = '/vllm-workspace/examples/ray_serving/multi-node-serving.sh';
        // Ray is optional in the openai image; 2.56.0 is the v0.29.0 CI pin.
        const VLLM_RAY_PACKAGE = 'ray[cgraph,default]==2.56.0';
        // Install Ray into a venv under /tmp that inherits the system site-packages
        // (vllm, torch) instead of writing into the system site-packages. The default
        // image runs as root on vanilla k8s, but OpenShift runs containers as an
        // arbitrary non-root UID, where `uv pip install --system` cannot write.
        const VLLM_RAY_VENV = '/tmp/ray-venv';
        const VLLM_RAY_INSTALL =
            'export UV_CACHE_DIR=/tmp/uv-cache && ' +
            `uv venv --system-site-packages --python python3 ${VLLM_RAY_VENV} && ` +
            `export VIRTUAL_ENV=${VLLM_RAY_VENV} && export PATH=${VLLM_RAY_VENV}/bin:$PATH && ` +
            `uv pip install '${VLLM_RAY_PACKAGE}'`;

        let project: Project;
        let workloadId: string;
        let workloadName: string;

        test.beforeAll(async function () {
            if (process.env.GPU_TYPE !== 'real') {
                logger.warn('Not a real GPU environment, skipping test');
                test.skip();
            }

            if (!(await isKnativeInstalled())) {
                logger.warn('knative is not installed on the cluster, skipping test');
                test.skip();
            }

            if (!(await isLwsInstalled())) {
                logger.warn('LeaderWorkerSet (LWS) is not installed on the cluster, skipping test');
                test.skip();
            }

            project = await orgUnitUtils.createProjectAndValidate({
                name: objectGenerator.generateShortNameByPrefix('proj'),
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 2, limit: 2 } }]
            });
        });

        test.beforeEach(function () {
            workloadName = objectGenerator.generateNameByPrefix('wl');
            workloadsCrudUtils.accessToken = flowRootState.users.system_administrator.accessToken;
        });

        test.afterAll(async function () {
            if (project === undefined) {
                return;
            }

            try {
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            } catch (error) {
                logger.error(
                    `Failed to delete workloads and project ${project.name} namespace ${project.status.namespace}`
                );
            }
        });

        test.afterEach(async function () {
            if (project === undefined) {
                return;
            }

            try {
                await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
            } catch (error) {
                logger.error(
                    `Failed to delete workloads of project ${project.status.namespace}; error: ${error}`
                );
            }
        });

        test('creates a vllm inference workload with single gpu', async function () {
            test.setTimeout(7 * 60 * 1000);

            const workloadSpec = {
                image: VLLM_IMAGE,
                args: `--model ${VLLM_MODEL}`,
                compute: {
                    gpuDevicesRequest: 1
                },
                servingPort: {
                    container: 8000
                }
            };

            logger.info(
                `creating vllm inference workload ${workloadName} in namespace ${project.status.namespace}`
            );

            [workloadId, workloadName] = await inferenceUtils.createInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            logger.info(`waiting for workload ${workloadId} to reach Running phase`);
            await waitUtils.waitForBackendWorkloadByIdAndPhase(
                workloadId,
                Phase.Running,
                6 * 60, // image pull may take time
                1
            );

            logger.info(`vllm inference workload ${workloadName} created successfully`);

            // Get the workload details to extract the internal URL
            const workloadDetails = await workloadsUtilsV2.getWorkloadByWorkloadId(workloadId);
            const internalUrl = inferenceUtils.getInternalUrlForInferenceWorkload(workloadDetails.body);

            // Send a chat completion request to the workload
            logger.info(`sending chat completion request to workload ${workloadName} at ${internalUrl}`);
            await inferenceUtils.sendModelPromptRequest(project.status.namespace, internalUrl, VLLM_MODEL);
        });

        test('creates a vllm distributed inference workload with multi-node setup', async function () {
            test.setTimeout(7 * 60 * 1000);

            const leaderCommand = `sh -c "${VLLM_RAY_INSTALL} && bash ${VLLM_MULTI_NODE_SCRIPT} leader --ray_cluster_size=$(LWS_GROUP_SIZE) && python3 -m vllm.entrypoints.openai.api_server --port 8000 --model ${VLLM_MODEL} --tensor-parallel-size 1 --pipeline-parallel-size 2 --distributed-executor-backend ray"`;
            const workerCommand = `sh -c "${VLLM_RAY_INSTALL} && bash ${VLLM_MULTI_NODE_SCRIPT} worker --ray_address=$(LWS_LEADER_ADDRESS)"`;

            const leaderCompute = new DistributedInferenceLeaderWorkerSpecCompute();
            leaderCompute.gpuDevicesRequest = 1;
            leaderCompute.largeShmRequest = true;
            const workerCompute = new DistributedInferenceLeaderWorkerSpecCompute();
            workerCompute.gpuDevicesRequest = 1;
            workerCompute.largeShmRequest = true;

            const workloadSpec: DistributedInferenceSpecSpec = {
                workers: 1,
                replicas: 1,
                servingPort: {
                    port: 8000,
                    exposeExternally: false
                },
                leader: {
                    image: VLLM_IMAGE,
                    command: leaderCommand,
                    compute: leaderCompute
                },
                worker: {
                    image: VLLM_IMAGE,
                    command: workerCommand,
                    compute: workerCompute
                }
            };

            logger.info(
                `creating vllm distributed inference workload ${workloadName} in namespace ${project.status.namespace} with image ${VLLM_IMAGE}`
            );
            logger.info(`multi-node helper ${VLLM_MULTI_NODE_SCRIPT}; ray install: ${VLLM_RAY_INSTALL}`);

            [workloadId, workloadName] = await distributedInferenceUtils.createDistributedInference(
                workloadName,
                workloadSpec,
                String(project.id)
            );

            const expectedPods = (1 + workloadSpec.workers) * workloadSpec.replicas;
            logger.info(
                `waiting for workload ${workloadId} to reach Running phase with ${expectedPods} pods`
            );

            try {
                await waitUtils.waitForBackendWorkloadByIdAndPhase(
                    workloadId,
                    Phase.Running,
                    6 * 60, // image pull may take time
                    expectedPods
                );
            } catch (error) {
                logger.error(
                    `workload ${workloadName} did not reach Running with ${expectedPods} pods, collecting leader/worker pod diagnostics`
                );
                await distributedInferenceUtils.logWorkloadPodsDiagnostics(
                    workloadName,
                    project.status.namespace
                );
                throw error;
            }

            logger.info(
                `vllm distributed inference workload ${workloadName} created successfully with ${expectedPods} pods`
            );

            // Get the internal URL and send a chat completion request
            const internalUrl = await inferenceUtils.getServingUrl(workloadId);
            logger.info(`sending chat completion request to workload ${workloadName} at ${internalUrl}`);
            await inferenceUtils.sendModelPromptRequest(project.status.namespace, internalUrl, VLLM_MODEL);
        });
    });
});
