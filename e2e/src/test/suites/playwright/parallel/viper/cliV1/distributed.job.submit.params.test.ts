/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    config,
    DistributedJobType,
    DistributedWorkload,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    Utils,
    waitUtils
} from '../../../../../../modules/helpers';
import { bashCli, kubeCli } from '../../../../../../modules/cli';
import {
    SubmitFlagsMerged,
    cliV2WorkloadType,
    cliV2TrainingType
} from '../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Project } from '../../../../../../modules/api/apis/projectsDepartmentsApi';
import {
    Platform,
    shouldRunNativeWorkloadCrdTests,
    supportedFrom,
    SupportedVersions,
    unsupportedPlatforms
} from '../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test } from '@playwright/test';

interface frameworkParams {
    type: string;
    wlType: string;
    image: string;
    expectedControllerName: string;
}

const xgboostParams: frameworkParams = {
    type: 'xgboost',
    wlType: DistributedJobType.XGBoostJobType,
    image: config.quickStartImage.distributed,
    expectedControllerName: 'xgboostjob-controller'
};

const tfParams: frameworkParams = {
    type: 'tf',
    wlType: DistributedJobType.TensorJobType,
    image: config.quickStartImage.distributedTf,
    expectedControllerName: 'tfjob-controller'
};

const pytorchParams: frameworkParams = {
    type: 'pytorch',
    wlType: DistributedJobType.PyTorchJobType,
    image: config.quickStartImage.distributedPytorch,
    expectedControllerName: 'pytorchjob-controller'
};

const paramsAllFrameworks: frameworkParams[] = [xgboostParams, tfParams, pytorchParams];

const paramsFrameworksWithNoMasterCapability: frameworkParams[] = [tfParams, pytorchParams];

shouldRunNativeWorkloadCrdTests(() => {
test.describe(`Submit distributed training job by CLI: ${tags.runOn.k8sFlavCertKit} ${tags.tool.api} ${tags.teams.viper}`, function () {
    unsupportedPlatforms([Platform.OPENSHIFT])(function () {
        let projectName: string, namespace: string;
        let project: Project;
        let workload: DistributedWorkload;
        let jobName: string;

        test.beforeEach(async function () {
            projectName = objectGenerator.generateNameByPrefix('project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 8,
                maxAllowedGpus: 8
            });
            jobName = objectGenerator.generateNameByPrefix('job');

            // CLI v1 resolves projects by listing namespaces with label runai/queue=<project>,
            // so the namespace must exist in the cluster before any CLI v1 command can reference the project.
            await waitUtils.waitForNamespaceToExist(namespace);
        });

        test.afterEach(async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
                logger.info(`deleted project ${projectName}`);
            } catch {
                logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
            }
        });

        test.describe('should submit distributed job and validate its data', function () {
            paramsAllFrameworks.forEach(function (params) {
                test(`${params.type} job`, async function () {
                    workload = new DistributedWorkload(params.image, namespace, jobName);
                    workload.imagePullPolicy = 'IfNotPresent';
                    workload.gpu = '1';
                    workload.workers = 2;
                    workload.jobType = params.wlType;
                    workload.username = 'asaf.eks';

                    await kubeCli.createWorkload(workload);
                    const pods = await waitUtils.waitForPods(jobName, namespace, 3);
                    const masterPod = pods.find(
                        (p) => p.metadata.labels['training.kubeflow.org/job-role'] === 'master'
                    );
                    expect(masterPod).not.to.be.undefined;
                    logger.debug(`Launcher pod annotations: ${JSON.stringify(masterPod, null, 2)}`);
                    expect(masterPod.metadata.labels['training.kubeflow.org/operator-name']).to.equal(
                        params.expectedControllerName
                    );
                    expect(masterPod.metadata.annotations).to.include.keys('totalGPUs');
                    expect(masterPod.metadata.annotations['totalGPUs']).to.equal('3'); // 1 for each pod, master and worker
                    expect(masterPod.spec.schedulerName).to.equal('runai-scheduler');

                    // need to validate that code in the keys.exist
                    const workerPods = pods.filter(
                        (p) => p.metadata.labels['training.kubeflow.org/replica-type'] === 'worker'
                    );
                    logger.info(`found ${workerPods.length} worker pods`);
                    expect(workerPods.length).to.equal(2);
                    for (const pod of workerPods) {
                        logger.debug(
                            `Worker pod annotations: ${JSON.stringify(pod.metadata.annotations, null, 2)}`
                        );
                        logger.debug(`Worker pod spec: ${JSON.stringify(pod.spec, null, 2)}`);
                        expect(pod.spec.schedulerName).to.equal('runai-scheduler');
                    }
                });
            });
        });

        test.describe('submit distributed job with environment vars', function () {
            const myEnvValue = 'myvalue';
            const myEnvUser = 'myuser';
            paramsAllFrameworks.forEach(function (params) {
                test(`${params.type} job`, async function () {
                    workload = new DistributedWorkload(params.image, namespace, jobName);
                    bashCli.exec(
                        `echo '${myEnvUser}' > username; kubectl -n ${namespace} create secret generic mysecret --from-file username`
                    );
                    workload.environment = {
                        RUNAI_ENV_VAL: myEnvValue,
                        RUNAI_ENV_SEC: 'SECRET:mysecret,username'
                    };
                    workload.image = 'ubuntu';
                    workload.command = '/bin/bash -c "env && sleep 60"';
                    workload.jobType = params.wlType;
                    await kubeCli.createWorkload(workload);
                    await waitUtils.waitForCrd(workload.getCrd(), jobName, namespace, 40, true, 2);
                    const pods = await waitUtils.waitForPods(
                        jobName,
                        namespace,
                        2,
                        60,
                        ['Running'],
                        false,
                        2
                    );
                    expect(pods.length, 'pods').to.be.above(0);
                    pods.forEach((pod) => {
                        const environmentVars: string = bashCli.exec(
                            `kubectl logs ${pod.metadata.name} -n ${namespace}`
                        );
                        expect(environmentVars).to.contain('RUNAI_ENV_VAL=' + myEnvValue);
                        expect(environmentVars).to.contain('RUNAI_ENV_SEC=' + myEnvUser);
                    });
                });
            });
        });
    });
});
});
