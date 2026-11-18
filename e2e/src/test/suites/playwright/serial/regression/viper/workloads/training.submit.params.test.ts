/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    config,
    crds,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    PodStatus,
    projectResourcesUtils,
    settings,
    tags,
    TrainingWorkload,
    Utils,
    versionUtils,
    waitUtils,
    workloadsUtilsV2
} from '../../../../../../../modules/helpers';
import { bashCli, kubeCli, runaiCliV2 } from '../../../../../../../modules/cli';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { ExtendedResource } from '../../../../../../../modules/helpers/models/baseWorkload';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import {
    Platform,
    shouldRunNativeWorkloadCrdTests,
    unsupportedPlatformsAsync
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, TestInfo } from '@playwright/test';
import {
    cliV2WorkloadType,
    cliLoginAsSystemAdmin,
    execFlags,
    SubmitFlagsMerged,
    DeleteFlagsMerged,
    describeFlags
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';

async function skipRedirectedGap(testInfo: TestInfo, reason: string): Promise<void> {
    if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Training)) {
        skipPlaywrightTest(test, testInfo, reason);
    }
}

test.describe(`Submit training jobs by CLI: ${tags.runOn.k8sFlavCertKit} ${tags.runOn.k8sFlavCertKitTemp} ${tags.resource.jobs} ${tags.tool.api} ${tags.teams.viper}`, function () {
    const gpuQuantity = '1';
    const publicUbuntuImage = 'docker.io/library/ubuntu';
    const NvidiaGPUS = 'nvidia.com/gpu';
    const sleepInfinity = 'sleep infinity';
    const IpcCapability = 'IPC_LOCK';
    const environmentSecretName = 'mysecret';
    const gpuExtendedResource: ExtendedResource = { name: NvidiaGPUS, quantity: gpuQuantity };

    test.describe('runai job with dirs and security context', function () {
        let projectName: string, namespace: string;
        let project: Project;
        let workload: TrainingWorkload;
        let jobName: string;
        let environmentSecretCreated: boolean;
        let submitFlags = {} as Partial<typeof SubmitFlagsMerged>;

        test.beforeAll(async function () {
            cliLoginAsSystemAdmin();
            projectName = objectGenerator.generateNameByPrefix('project');
            namespace = Utils.namespaceFromProjectName(projectName);
            project = await projectResourcesUtils.createProject(namespace, {
                name: projectName,
                deservedGpus: 4,
                maxAllowedGpus: 4
            });
        });

        test.beforeEach(function () {
            jobName = objectGenerator.generateNameByPrefix('cli-train');
            workload = new TrainingWorkload(config.quickStartImage.emptyImage, namespace, jobName);
            workload.image = config.quickStartImage.inferenceHelloWorld;
            environmentSecretCreated = false;
            submitFlags = {
                [SubmitFlagsMerged.image]: workload.image,
                [SubmitFlagsMerged.project]: projectName
            } as Partial<typeof SubmitFlagsMerged>;
        });

        test.afterEach(async function () {
            test.setTimeout(300 * 1000);
            try {
                await workloadsUtilsV2.deleteWorkloadsByProjectId(project.id.toString(), 120);
                await waitUtils.waitForCondition(
                    async () => (await kubeCli.listPodsByNamepsace(namespace)).length === 0,
                    `all workload pods in namespace ${namespace} to be deleted`,
                    true,
                    120,
                    2
                );
            } finally {
                if (environmentSecretCreated) {
                    try {
                        await kubeCli.deleteSecret(environmentSecretName, namespace);
                    } catch (error) {
                        logger.warn(
                            `Failed to delete secret ${environmentSecretName} in namespace ${namespace}: ${error}`
                        );
                    }
                }
            }
        });

        test.afterAll(async function () {
            try {
                await operations.deleteProjectAndNamespace(project.id);
            } catch {
                logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
            }
        });

        test.describe('submission', function () {
            shouldRunNativeWorkloadCrdTests(() =>
            test('submit job with environment vars', async function () {
                bashCli.exec(
                    `echo 'myuser' > username; kubectl -n ${namespace} create secret generic ${environmentSecretName} --from-file username`
                );
                environmentSecretCreated = true;
                workload.environment = {
                    MYVAR: 'myvalue',
                    MYVAR2: `SECRET:${environmentSecretName},username`
                };
                workload.image = settings.isArm
                    ? config.quickStartImage.ubuntu
                    : config.quickStartImage.sleepInfinity;
                workload.command = 'env';
                await kubeCli.createWorkload(workload);
                    const podName = (
                        await waitUtils.waitForPods(jobName, namespace, 1, 60, ['Succeeded'])
                    ).pop().metadata.name;
                const environmentVars: string = bashCli.exec(`kubectl logs ${podName} -n ${namespace}`);
                expect(environmentVars).to.contain('MYVAR=myvalue');
                expect(environmentVars).to.contain('MYVAR2=myuser');
                })
            );

            shouldRunNativeWorkloadCrdTests(() =>
            test('privilege escalation created successfully in the pod', async function () {
                workload.allowPrivilegeEscalation = true;

                const output = await kubeCli.createWorkload(workload);
                logger.debug(`command output: ${JSON.stringify(output)}`);

                const pods = await waitUtils.waitForPods(jobName, namespace, 1, 60, ['Running']);
                expect(pods.length).to.equal(1);
                    expect(pods[0].spec.containers[0].securityContext.allowPrivilegeEscalation).to.equal(
                        true
                    );
                })
            );

            test('home directory created successfully', async function ({}, testInfo: TestInfo) {
                await skipRedirectedGap(
                    testInfo,
                    'CLI exec cannot locate the redirected training workload'
                );
                const workingDir = isOCP() ? '/tmp' : '/home/mydir';
                workload.workingDir = workingDir;

                submitFlags[SubmitFlagsMerged.createHomeDir] = '';
                submitFlags[SubmitFlagsMerged.workingDir] = workingDir;
                submitFlags[SubmitFlagsMerged.command] = '-- sleep 240';
                runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.training);

                logger.debug(`command output: submitted workload ${jobName}`);

                const pods = await waitUtils.waitForPods(jobName, namespace, 1, 60, ['Running']);
                expect(pods.length).to.equal(1);

                await waitUtils.waitForBackendWorkloadPhase(jobName, project.id, Phase.Running);

                const resultPwd = runaiCliV2.exec(
                    { [execFlags.project]: projectName, [execFlags.command]: 'pwd' } as Partial<
                        typeof execFlags
                    >,
                    workload.name,
                    cliV2WorkloadType.training
                );
                expect(resultPwd.stdout).to.equal(`${workingDir}\n`);
            });
        });

        test.describe('describe job', function () {
            test('shows correct job info', async function ({}, testInfo: TestInfo) {
                await skipRedirectedGap(
                    testInfo,
                    'describe shows Requested 0.00 - redirected trainings get an empty compute spec on the cp record'
                );
                workload.gpu = '0.5';
                submitFlags[SubmitFlagsMerged.gpuPortionRequest] = workload.gpu;
                runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.training);
                logger.debug(`command output: submitted workload ${jobName}`);

                const pods = await waitUtils.waitForPods(jobName, namespace, 1, 120, ['Running']);
                expect(pods.length).to.equal(1);

                // describe reads control-plane data, which lags the cluster; wait for the
                // backend workload to reach Running before asserting on its resources
                await waitUtils.waitForBackendWorkloadPhase(jobName, project.id, Phase.Running);

                const describePayload = runaiCliV2.describe(
                    { [describeFlags.project]: projectName },
                    jobName,
                    cliV2WorkloadType.training
                );
                expect(describePayload).to.include('GPU Compute           0.50');
            });
        });

        test.describe('memory', function () {
            test('should have memory', async function () {
                workload.memory = '1G';
                submitFlags[SubmitFlagsMerged.cpuMemoryRequest] = workload.memory;
                runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.training);

                const pods = await waitUtils.waitForPods(jobName, namespace, 1);
                expect(pods.length).to.equal(1);

                expect(pods[0].spec.containers[0].resources.requests['memory']).to.equal(workload.memory);
            });

            test('should have memory-limit', async function () {
                workload.memoryLimit = '1G';
                submitFlags[SubmitFlagsMerged.cpuMemoryLimit] = workload.memoryLimit;
                runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.training);

                const pods = await waitUtils.waitForPods(jobName, namespace, 1);
                expect(pods.length).to.equal(1);

                expect(pods[0].spec.containers[0].resources.limits['memory']).to.equal(workload.memoryLimit);
            });
        });

        test.describe('volume (training job)', function () {
            // if this case fails, it might be because the nfs server is not available
            // check with devops if the nfs server is up (3.235.197.16/sgw-2023-03-08-f99b5)
            test('should be able to mount the volume content to file system', async function ({}, testInfo: TestInfo) {
                await skipRedirectedGap(
                    testInfo,
                    'CLI exec cannot locate the redirected training workload'
                );
                await unsupportedPlatformsAsync([Platform.RKE2])(async function () {
                    submitFlags[SubmitFlagsMerged.image] = publicUbuntuImage;
                    submitFlags[SubmitFlagsMerged.nfs] =
                        'server=3.235.197.16,path=/sgw-2023-03-08-f99b5,mountpath=/volume1path';
                    submitFlags[SubmitFlagsMerged.command] = `-- ${sleepInfinity}`;
                    runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.training);
                    const pods = await waitUtils.waitForPods(jobName, namespace, 1, 120, ['Running']);
                    expect(pods.length).to.equal(1);
                    await waitUtils.waitForBackendWorkloadPhase(jobName, project.id, Phase.Running);

                    const execFlag = {
                        [execFlags.project]: projectName,
                        [execFlags.command]: 'ls /volume1path'
                    } as Partial<typeof execFlags>;
                    const resultLs = runaiCliV2.exec(execFlag, workload.name, cliV2WorkloadType.training);
                    logger.debug(`result of ls /volume1path: ${resultLs.stdout}`);
                    expect(resultLs.stdout.trim()).to.include('succeeded.txt');
                });
            });
        });

        test.describe('cpu', function () {
            test('should have cpu', async function () {
                workload.cpu = gpuQuantity;
                submitFlags[SubmitFlagsMerged.cpuCoreRequest] = workload.cpu;
                runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.training);

                const pods = await waitUtils.waitForPods(jobName, namespace, 1);
                expect(pods.length).to.equal(1);

                expect(pods[0].spec.containers[0].resources.requests['cpu']).to.equal(workload.cpu);
            });

            test('should have cpu-limit', async function () {
                workload.cpuLimit = '3';
                submitFlags[SubmitFlagsMerged.cpuCoreLimit] = workload.cpuLimit;
                runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.training);

                const pods = await waitUtils.waitForPods(jobName, namespace, 1);
                expect(pods.length).to.equal(1);

                expect(pods[0].spec.containers[0].resources.limits['cpu']).to.equal(workload.cpuLimit);
            });
        });

        test.describe('gpu', function () {
            test('should have gpu', async function () {
                workload.gpu = gpuQuantity;
                const gpuSubmitFlags = {
                    [SubmitFlagsMerged.image]: workload.image,
                    [SubmitFlagsMerged.project]: projectName,
                    [SubmitFlagsMerged.gpuRequest]: workload.gpu
                } as Partial<typeof SubmitFlagsMerged>;
                runaiCliV2.submit(gpuSubmitFlags, workload.name, cliV2WorkloadType.training);

                const pods = await waitUtils.waitForPods(jobName, namespace, 1);
                expect(pods.length).to.equal(1);

                expect(pods[0].spec.containers[0].resources.requests[NvidiaGPUS]).to.equal(workload.gpu);
            });
        });

        test.describe('parallelism', function () {
            test('should run multiple pods', async function () {
                const parallelSubmitFlags = {
                    ...submitFlags,
                    [SubmitFlagsMerged.parallelism]: '2',
                    [SubmitFlagsMerged.command]: `-- ${sleepInfinity}`
                } as Partial<typeof SubmitFlagsMerged>;
                runaiCliV2.submit(parallelSubmitFlags, workload.name, cliV2WorkloadType.training);

                if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Training)) {
                    await waitUtils.waitForBatchJobToExist(jobName, namespace);
                } else {
                    await waitUtils.waitForCrd(crds.runaiJobCrd, jobName, namespace);
                }

                const describePayload: string = runaiCliV2.describe(
                    { [describeFlags.project]: projectName },
                    jobName,
                    cliV2WorkloadType.training
                );
                expect(describePayload).to.include('parallelism 2');
            });
        });

        test.describe('completions', function () {
            shouldRunNativeWorkloadCrdTests(() =>
            test('should have completions', async function () {
                workload.completions = 1;
                await kubeCli.createWorkload(workload);
                await waitUtils.waitForCrd(crds.runaiJobCrd, jobName, namespace);
                await waitUtils.waitForBackendWorkloadPhase(jobName, project.id, Phase.Running);

                const pods = await waitUtils.waitForPods(jobName, namespace, workload.completions);
                expect(pods.length).to.equal(workload.completions);
                })
            );
        });

        test.describe('backoffLimit', function () {
            shouldRunNativeWorkloadCrdTests(() =>
            test('should have backoffLimit', async function () {
                workload.backoffLimit = 1;
                await kubeCli.createWorkload(workload);

                await waitUtils.waitForCrd(crds.runaiJobCrd, jobName, namespace);

                const job = await kubeCli.getCrd(crds.runaiJobCrd, jobName, namespace);
                expect(job['spec']['backoffLimit']).to.equal(workload.backoffLimit);
                })
            );
        });

        test.describe('test pvcs (training job)', function () {
            test('test pvcs (ephemeral + regular)', async function ({}, testInfo: TestInfo) {
                await skipRedirectedGap(
                    testInfo,
                    'Omitted PVC access mode is translated to an invalid empty list'
                );
                const regularClaimName = 'regularclaimname';
                const ephemeralClaimName = 'ephemeralclaimname';
                const pathPrefix = 'xxxx';
                const pvc1Path = `/${pathPrefix}-regular`;
                const pvc2Path = `/${pathPrefix}-ephemeral`;
                workload.pvcs = {
                    pvc1: {
                        claimName: regularClaimName,
                        path: pvc1Path,
                        size: '1Gi',
                        readOnly: true,
                        ephemeral: false
                    },
                    pvc2: {
                        claimName: ephemeralClaimName,
                        path: pvc2Path,
                        size: '1Gi',
                        readOnly: true,
                        ephemeral: true
                    }
                };

                submitFlags[SubmitFlagsMerged.command] = `-- ${sleepInfinity}`;
                submitFlags[SubmitFlagsMerged.newPvc] = [
                    `claimname=${regularClaimName},path=${pvc1Path},size=1Gi,ro`,
                    `claimname=${ephemeralClaimName},path=${pvc2Path},size=1Gi,ro,ephemeral`
                ];

                logger.debug('create workload with ephemeral + regular pvc');
                runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.training);
                await waitUtils.waitForPvcs(namespace, 1, 120, workload.name);
                await waitUtils.waitForPvcs(namespace, 1, 120, regularClaimName);

                const podName = (await waitUtils.waitForPods(workload.name, namespace, 1, 60, ['Running']))[0]
                    .metadata.name;
                const directories = bashCli
                    .exec(
                        `kubectl exec -it -n ${namespace} ${podName} -- /bin/sh -c "cd ..; ls | grep ${pathPrefix}"`
                    )
                    .split('\n');
                expect(directories.includes(pvc1Path.slice(1)), `directory ${pvc1Path} exist`).to.true;
                expect(directories.includes(pvc2Path.slice(1)), `directory ${pvc2Path} exist`).to.true;

                logger.debug('delete workload with ephemeral + regular pvc');
                runaiCliV2.delete(
                    { [DeleteFlagsMerged.project]: projectName } as Partial<typeof DeleteFlagsMerged>,
                    workload.name,
                    cliV2WorkloadType.training
                );
                const numOfExpectedPvcs = versionUtils.isVersionSupported(
                    flowRootState.cluster.version,
                    '2.14'
                )
                    ? 0
                    : 1;
                await waitUtils.waitForPvcs(namespace, numOfExpectedPvcs, 120, regularClaimName);
            });

            test('should have pvcs after create and automatically deleted after delete (ephemeral)', async function () {
                const claimName = 'ephemeralclaimname';
                const pvcsSubmitFlags = {
                    [SubmitFlagsMerged.image]: publicUbuntuImage,
                    [SubmitFlagsMerged.project]: projectName,
                    [SubmitFlagsMerged.newPvc]: `claimname=${claimName},path=/path,size=2Gi,accessmode-rwo,ephemeral`,
                    [SubmitFlagsMerged.command]: `-- ${sleepInfinity}`
                } as Partial<typeof SubmitFlagsMerged>;
                runaiCliV2.submit(pvcsSubmitFlags, workload.name, cliV2WorkloadType.training);
                const pvcList = await waitUtils.waitForPvcs(namespace, 1, 120, workload.name);
                expect(pvcList.items.length).to.equal(1);
                expect(pvcList.items[0].spec.accessModes).to.be.deep.eq(['ReadWriteOnce']);
                expect(pvcList.items[0].spec.resources.requests).to.be.deep.eq({ storage: '2Gi' });

                logger.debug('delete workload with ephemeral pvc');
                runaiCliV2.delete(
                    { [DeleteFlagsMerged.project]: projectName } as Partial<typeof DeleteFlagsMerged>,
                    workload.name,
                    cliV2WorkloadType.training
                );
                await waitUtils.waitForPvcs(namespace, 0, 120, workload.name);
            });

            test('should have pvcs after create and automatically deleted after delete (regular pvc : pvcExist + pvcNew)', async function () {
                const claimName = 'regularclaimname';
                const jobNameForPvcExist = objectGenerator.generateNameByPrefix('job');
                const pvcNewSubmitFlags = {
                    [SubmitFlagsMerged.image]: publicUbuntuImage,
                    [SubmitFlagsMerged.project]: projectName,
                    [SubmitFlagsMerged.newPvc]: `claimname=${claimName},path=/path,size=1Gi,accessmode-rwm,ro`,
                    [SubmitFlagsMerged.command]: `-- ${sleepInfinity}`
                } as Partial<typeof SubmitFlagsMerged>;
                runaiCliV2.submit(pvcNewSubmitFlags, workload.name, cliV2WorkloadType.training);
                const pvcList = await waitUtils.waitForPvcs(namespace, 1, 120, claimName);
                expect(pvcList.items[0].spec.accessModes).to.be.deep.eq(['ReadWriteMany']);
                expect(pvcList.items[0].spec.resources.requests).to.be.deep.eq({ storage: '1Gi' });

                logger.debug('create workload with exists regular pvc');
                const pvcExistSubmitFlags = {
                    [SubmitFlagsMerged.image]: publicUbuntuImage,
                    [SubmitFlagsMerged.project]: projectName,
                    [SubmitFlagsMerged.existingPvc]: `claimname=${claimName},path=/path`,
                    [SubmitFlagsMerged.command]: `-- ${sleepInfinity}`
                } as Partial<typeof SubmitFlagsMerged>;
                runaiCliV2.submit(pvcExistSubmitFlags, jobNameForPvcExist, cliV2WorkloadType.training);
                await waitUtils.waitForPods(jobNameForPvcExist, namespace, 1, 120, ['Pending', 'Running']);

                logger.debug('delete workload consuming the regular pvc');
                runaiCliV2.delete(
                    { [DeleteFlagsMerged.project]: projectName } as Partial<typeof DeleteFlagsMerged>,
                    jobNameForPvcExist,
                    cliV2WorkloadType.training
                );
                await waitUtils.waitForPodsDeletion(jobNameForPvcExist, namespace, 120);

                logger.debug('delete workload that created the regular pvc');
                runaiCliV2.delete(
                    { [DeleteFlagsMerged.project]: projectName } as Partial<typeof DeleteFlagsMerged>,
                    workload.name,
                    cliV2WorkloadType.training
                );
                await waitUtils.waitForPvcs(namespace, 0, 120, claimName);
            });
        });

        test.describe('extended resources workload', function () {
            shouldRunNativeWorkloadCrdTests(() =>
            test('should have pod with 1 gpu requested', async function () {
                workload.command = sleepInfinity;
                workload.extendedResources = { extended: gpuExtendedResource };
                await kubeCli.createWorkload(workload);
                const pods = await waitUtils.waitForPods(workload.name, namespace, 1, 120, ['Running']);
                expect(pods).length(1);
                const pod = pods[0];
                expect(pod.spec.containers[0].resources.requests[NvidiaGPUS]).to.equal(gpuQuantity);
                })
            );
        });

        test.describe('security add linux capabilities workload', function () {
            shouldRunNativeWorkloadCrdTests(() =>
            test('should have one capabilities added to container using workload', async function () {
                workload.capabilities = { cap: { capability: IpcCapability } };
                workload.command = sleepInfinity;
                await kubeCli.createWorkload(workload);
                const pods = await waitUtils.waitForPods(workload.name, namespace, 1, 120, ['Running']);
                expect(pods).length(1);
                const pod = pods[0];
                expect(pod.spec.containers[0].securityContext.capabilities.add).length(1);
                expect(pod.spec.containers[0].securityContext.capabilities.add).contain.all.members([
                    IpcCapability
                ]);
                })
            );
        });

        test.describe('git-sync (training job)', function () {
            test(
                'should be able to mount the git repository to file system',
                { tag: [tags.services.gitSync, tags.runOn.serialCI] },
                async function ({}, testInfo: TestInfo) {
                    await skipRedirectedGap(
                        testInfo,
                        'CLI exec cannot locate the redirected training workload'
                    );
                    submitFlags[SubmitFlagsMerged.command] = `-- ${sleepInfinity}`;
                    submitFlags[SubmitFlagsMerged.gitSync] =
                        'name=git-mount,repository=https://github.com/octocat/Hello-World,path=/gitmount';

                    runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.training);
                    logger.debug(`job name: ${jobName} namespace: ${namespace} was created`);
                    const pods = await waitUtils.waitForPods(jobName, namespace, 1, 120, ['Running']);
                    logger.debug(`pod for job name: ${jobName} namespace: ${namespace} was created`);
                    expect(pods.length).to.equal(1);

                    const execFlag = {
                        [execFlags.project]: projectName,
                        [execFlags.command]: 'ls /gitmount'
                    } as Partial<typeof execFlags>;
                    const resultLs = runaiCliV2.exec(execFlag, workload.name, cliV2WorkloadType.training);
                    logger.debug(`result of ls /gitmount: ${resultLs.stdout}`);
                    expect(resultLs.stdout.trim()).to.include('Hello-World');
                }
            );
        });

        test.describe('s3', function () {
            let serviceIp: string;

            async function setupMinioServiceAndCreateS3Bucket(): Promise<void> {
                await kubeCli.deployMinio(namespace);
                await kubeCli.createMinioService(namespace);
                serviceIp = await kubeCli.getMinioServiceIP(namespace);
                await kubeCli.createS3Bucket('my-python-pod', namespace);
            }

            const s3SecretName = 'minio-s3-credentials';

            async function cleanup(): Promise<void> {
                const results = await Promise.allSettled([
                    kubeCli.deleteSecret(s3SecretName, namespace),
                    kubeCli.deletePod('my-python-pod', namespace),
                    kubeCli.deleteMinioService(namespace),
                    kubeCli.deleteMinio(namespace)
                ]);
                for (const result of results) {
                    if (result.status === 'rejected') {
                        logger.warn(`S3 cleanup failed: ${result.reason}`);
                    }
                }
            }

            test('should be able to mount the s3 file system', async function ({}, testInfo: TestInfo) {
                await skipRedirectedGap(
                    testInfo,
                    'CLI exec cannot locate the redirected training workload'
                );
                await unsupportedPlatformsAsync([Platform.OPENSHIFT])(async function () {
                    test.setTimeout(240000);
                    try {
                        await setupMinioServiceAndCreateS3Bucket();
                        // --s3 accesskey= is a K8s Secret name, not the raw credential.
                        // The secret must contain fields "AccessKeyId" and "SecretKey".
                        const base64 = (s: string): string => Buffer.from(s).toString('base64');
                        await kubeCli.createSecret(s3SecretName, namespace, {
                            AccessKeyId: base64('minioadmin'),
                            SecretKey: base64('minioadmin')
                        });

                        submitFlags[SubmitFlagsMerged.image] = publicUbuntuImage;
                        submitFlags[SubmitFlagsMerged.imagePullPolicy] = 'IfNotPresent';
                        submitFlags[SubmitFlagsMerged.createHomeDir] = '';
                        submitFlags[SubmitFlagsMerged.runAsUid] = 501;
                        submitFlags[SubmitFlagsMerged.runAsGid] = 100;
                        submitFlags[SubmitFlagsMerged.hostNetwork] = '';
                        submitFlags[SubmitFlagsMerged.s3] =
                            `name=s3-0,bucket=my-new-bucket,path=/tmp/e2e,accesskey=${s3SecretName},url=http://${serviceIp}:9000`;
                        submitFlags[SubmitFlagsMerged.command] = `-- ${sleepInfinity}`;

                        runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.training);
                        await waitUtils.waitForPods(jobName, namespace, 1, 120, ['Running']);

                        const execFlag = {
                            [execFlags.project]: projectName,
                            [execFlags.command]: 'ls -la /tmp/e2e'
                        } as Partial<typeof execFlags>;
                        // Running does not mean that the Goofys mount has propagated to this container.
                        await waitUtils.waitForCondition(
                            () => {
                                try {
                                    const mounts = runaiCliV2.exec(
                                        { ...execFlag, [execFlags.command]: 'cat /proc/mounts' },
                                        workload.name,
                                        cliV2WorkloadType.training
                                    );
                                    const mounted = mounts.stdout.split('\n').some((line) => {
                                        const [, path, type] = line.split(' ');
                                        return path === '/tmp/e2e' && type?.startsWith('fuse');
                                    });
                                    if (!mounted) return false;
                                    const files = runaiCliV2.exec(
                                        execFlag,
                                        workload.name,
                                        cliV2WorkloadType.training
                                    );
                                    return files.stdout.includes('e2e-test-file');
                                } catch (error) {
                                    logger.debug(`S3 mount is not ready: ${error.message}`);
                                    return false;
                                }
                            },
                            'S3 FUSE mount at /tmp/e2e containing e2e-test-file',
                            true,
                            30
                        );
                    } catch (error) {
                        await kubeCli.logS3Diagnostics(namespace);
                        throw error;
                    } finally {
                        await cleanup();
                    }
                });
            });
        });

        // if this case fails, it might be because the nfs server is not available
        // check with devops if the nfs server is up (3.235.197.16/sgw-2023-03-08-f99b5)
        shouldRunNativeWorkloadCrdTests(() =>
        test('submit mount-propagation', async function () {
            workload.command = 'sleep infinity';
            workload.mountPropagation = true;
            workload.nfs = {
                'nfs-0': {
                    sourcePath: '/sgw-2023-03-08-f99b5',
                    targetPath: '/volume1path',
                    readOnly: true,
                    nfsServer: '3.235.197.16'
                }
            };
            await kubeCli.createWorkload(workload);
            const pods = await waitUtils.waitForPods(jobName, namespace, 1, 120, [PodStatus.RUNNING]);
                expect(pods[0].spec.containers[0].volumeMounts[0].mountPropagation).to.contain(
                    'HostToContainer'
                );
            })
        );

        shouldRunNativeWorkloadCrdTests(() =>
        test('use exact GPU memory', async function () {
            workload.gpuMemory = '1G';
            await kubeCli.createWorkload(workload);
            const pods = await waitUtils.waitForPods(jobName, namespace, 1, waitUtils.waitForPodTimeout, [
                'Running'
            ]);
            expect(pods[0].metadata.annotations['gpu-memory']).to.contain('1000');
            })
        );
    });
});
