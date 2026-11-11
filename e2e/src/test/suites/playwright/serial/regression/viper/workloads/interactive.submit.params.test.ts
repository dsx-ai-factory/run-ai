/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    config,
    crds,
    flowRootState,
    InteractiveWorkload,
    logger,
    objectGenerator,
    operations,
    Preemptibility,
    projectResourcesUtils,
    tags,
    Utils,
    versionUtils,
    waitUtils
} from '../../../../../../../modules/helpers';
import { bashCli, kubeCli, runaiCliV2 } from '../../../../../../../modules/cli';
import { Project } from '../../../../../../../modules/api/apis/projectsDepartmentsApi';
import { isOCP } from '../../../../../../../modules/helpers/utils/EnvironmentUtils';
import {
    Platform,
    shouldRunTests,
    supportedFrom,
    SupportedVersions,
    unsupportedPlatformsAsync
} from '../../../../../../../modules/helpers/utils/SupporetedEnvsUtils';
import { test, TestInfo } from '@playwright/test';
import {
    cliV2WorkloadType,
    cliLoginAsSystemAdmin,
    execFlags,
    SubmitFlagsMerged,
    DeleteFlagsMerged
} from '../../../../../../../modules/cli/tools/RunaiCliV2.constants';
import { Phase } from '../../../../../../../modules/api/apis/workloadsApi/generated/model/phase';
import {
    AnyWorkloadSubmissionType,
    isAnyWorkloadSubmissionEnabled
} from '../../../../../../../modules/helpers/utils/AnyWorkloadSubmissionUtils';
import { skipPlaywrightTest } from '../../../../../../../modules/helpers/utils/SkippedTestUtils';

async function skipRedirectedGap(testInfo: TestInfo, reason: string): Promise<void> {
    if (await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Workspace)) {
        skipPlaywrightTest(test, testInfo, reason);
    }
}

test.describe(`Submit interactive jobs by CLI: ${tags.tool.api} ${tags.resource.jobs} ${tags.teams.viper}`, function () {
    shouldRunTests(!isOCP())(function () {
        test.describe('runai job parameters - general ', function () {
            let projectName: string, namespace: string;
            let project: Project;
            let workload: InteractiveWorkload;
            let jobName: string;
            let submitFlags = {} as Partial<typeof SubmitFlagsMerged>;


            test.beforeAll(async function () {
                cliLoginAsSystemAdmin();
                projectName = objectGenerator.generateNameByPrefix('project');
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2
                });
            });

            test.beforeEach(function () {
                jobName = objectGenerator.generateNameByPrefix('job');

                workload = new InteractiveWorkload(config.quickStartImage.interactive, namespace, jobName);
                workload.cpu = '1';
                workload.createHomeDir = true;
                workload.gpu = '0.5';
                workload.imagePullPolicy = 'IfNotPresent';
                workload.memory = '100M';
                workload.runAsGid = 20;
                workload.runAsUid = 501;
                workload.runAsUser = true;
                workload.stdin = true;
                workload.supplementalGroups = '20,12,61,79,80,81,98,33,100,204,250,395,101,399,400,701';
                workload.tty = true;
                workload.username = 'ofer.lahav';
                workload.workingDir = '/home/test';

                submitFlags = {
                    [SubmitFlagsMerged.image]: workload.image,
                    [SubmitFlagsMerged.project]: projectName,
                    [SubmitFlagsMerged.cpuCoreRequest]: workload.cpu,
                    [SubmitFlagsMerged.gpuPortionRequest]: workload.gpu,
                    [SubmitFlagsMerged.cpuMemoryRequest]: workload.memory,
                    [SubmitFlagsMerged.imagePullPolicy]: workload.imagePullPolicy,
                    [SubmitFlagsMerged.runAsGid]: workload.runAsGid,
                    [SubmitFlagsMerged.runAsUid]: workload.runAsUid,
                    [SubmitFlagsMerged.createHomeDir]: '',
                    [SubmitFlagsMerged.stdin]: '',
                    [SubmitFlagsMerged.tty]: '',
                    [SubmitFlagsMerged.supplementalGroups]: workload.supplementalGroups,
                    [SubmitFlagsMerged.workingDir]: workload.workingDir,
                } as Partial<typeof SubmitFlagsMerged>;
            });

            test.afterEach(async function () {
                try {
                    await operations.deleteWorkloads([workload]);
                } catch {
                    logger.error(`Failed to delete workload ${jobName} in namespace ${namespace}`);
                }
            });

            test.afterAll(async function () {
                try {
                    await operations.deleteProjectAndNamespace(project.id);
                } catch {
                    logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
                }
            });

            test('Should create a valid pod and validate its data', async function () {
                runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);

                const pods = await waitUtils.waitForPods(jobName, namespace, 1);
                expect(pods.length).to.equal(1);

                const podMeta = pods[0].metadata;
                expect(podMeta.annotations).to.include.keys('gpu-fraction');
                expect(podMeta.annotations['gpu-fraction']).to.equal('0.5');
                expect(podMeta.annotations).to.include.keys('user');

                const podSpec = pods[0].spec;
                expect(podSpec.schedulerName).to.equal('runai-scheduler');

                const isRedirectedSubmission = await isAnyWorkloadSubmissionEnabled(
                    AnyWorkloadSubmissionType.Workspace,
                );

                // redirected IW/TW Jobs do not yet preserve all native custom UID/GID fields.
                if (!isRedirectedSubmission) {
                    expect(podSpec.securityContext.fsGroup).to.equal(20);
                }
                expect(podSpec.securityContext.supplementalGroups.length).to.greaterThanOrEqual(10);

                expect(podSpec.containers.length).to.equal(1);
                const container = podSpec.containers[0];
                expect(container.image).to.equal(workload.image);
                expect(container.imagePullPolicy).to.equal(workload.imagePullPolicy);
                expect(container.tty).to.equal(true);
                expect(container.stdin).to.equal(true);

                expect(container).to.include.keys('securityContext');
                expect(container.securityContext.runAsGroup).to.equal(20);
                expect(container.securityContext.runAsUser).to.equal(501);

                expect(container).to.include.keys('resources');
                expect(container.resources).to.include.keys('requests');
                expect(container.resources.requests).to.include.keys('cpu');
                expect(container.resources.requests.cpu).to.equal('1');
                expect(container.resources.requests).to.include.keys('memory');
                expect(container.resources.requests.memory).to.equal('100M');

                expect(container).to.include.keys('env');
                const envVars = container.env;
                const home = envVars.find((env) => env.name === 'HOME');
                expect(home.value).to.equal('/home/runai-home'); // home dir default for runAsUser, workingHome ignored
                if (!isRedirectedSubmission) {
                    const ldapUid = envVars.find((env) => env.name === 'LDAP_UID');
                    expect(ldapUid.value).to.equal('501');
                    const ldapGid = envVars.find((env) => env.name === 'LDAP_GID');
                    expect(ldapGid.value).to.equal('20');
                }

                expect(container).to.include.keys('volumeMounts');
                const mount = container.volumeMounts.find((env) => env.name === 'user-data');
                expect(mount.mountPath).to.equal('/home/runai-home');
            });

            test.describe('memory', function () {
                test('should have memory', async function () {
                    workload.memory = '1G';
                    submitFlags[SubmitFlagsMerged.cpuMemoryRequest] = workload.memory;
                    runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);

                    const pods = await waitUtils.waitForPods(jobName, namespace, 1);
                    expect(pods.length).to.equal(1);

                    expect(pods[0].spec.containers[0].resources.requests['memory']).to.equal(workload.memory);
                });

                test('should have memory-limit', async function () {
                    workload.memoryLimit = '1G';
                    submitFlags[SubmitFlagsMerged.cpuMemoryLimit] = workload.memoryLimit;
                    runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);

                    const pods = await waitUtils.waitForPods(jobName, namespace, 1);
                    expect(pods.length).to.equal(1);

                    expect(pods[0].spec.containers[0].resources.limits['memory']).to.equal(
                        workload.memoryLimit
                    );
                });
            });

            test.describe('volume (interactive job)', function () {
                // if this case fails, it might be because the nfs server is not available
                // check with devops if the nfs server is up (3.235.197.16/sgw-2023-03-08-f99b5)
                test('should be able to mount the volume content to file system', async function ({}, testInfo: TestInfo) {
                    await skipRedirectedGap(
                        testInfo,
                        'CLI exec cannot locate the redirected workspace workload'
                    );
                    await unsupportedPlatformsAsync([Platform.RKE2])(async function () {
                        submitFlags[SubmitFlagsMerged.nfs] = 'server=3.235.197.16,path=/sgw-2023-03-08-f99b5,mountpath=/volume1path';
                        submitFlags[SubmitFlagsMerged.command] = '-- sleep infinity';
                        submitFlags[SubmitFlagsMerged.imagePullPolicy] = 'IfNotPresent';
                        runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);
                        logger.debug(
                            `volume (interactive job) - pod for job name: ${jobName} namespace: ${namespace} was created`
                        );
                        const pods = await waitUtils.waitForPods(jobName, namespace, 1, 240, ['Running']);
                        logger.debug(
                            `volume (interactive job) - pod for job name: ${jobName} namespace: ${namespace} is Running`
                        );
                        expect(pods.length).to.equal(1);
                        const execFlag = {
                            [execFlags.project]: projectName,
                            [execFlags.command]: 'ls /volume1path'
                        } as Partial<typeof execFlags>;
                        const resultLs = runaiCliV2.exec(
                            execFlag,
                            jobName,
                            cliV2WorkloadType.workspace
                        );
                        logger.debug(`result of ls /volume1path: ${resultLs.stdout}`);
                        expect(resultLs.stdout.trim()).to.include('succeeded.txt');
                    });
                });
            });

            test.describe('cpu', function () {
                test('should have cpu', async function () {
                    submitFlags[SubmitFlagsMerged.cpuCoreRequest] = '1';
                    runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);

                    const pods = await waitUtils.waitForPods(jobName, namespace, 1);
                    expect(pods.length).to.equal(1);

                    expect(pods[0].spec.containers[0].resources.requests['cpu']).to.equal(workload.cpu);
                });

                test('should have cpu-limit', async function () {
                    workload.cpuLimit = '3';
                    submitFlags[SubmitFlagsMerged.cpuCoreLimit] = workload.cpuLimit;
                    runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);

                    const pods = await waitUtils.waitForPods(jobName, namespace, 1);
                    expect(pods.length).to.equal(1);

                    expect(pods[0].spec.containers[0].resources.limits['cpu']).to.equal(workload.cpuLimit);
                });
            });
            test.describe('git-sync (interactive job)', function () {
                test(`should be able to mount the git repository to file system ${tags.runOn.parallelCI}`, async function ({}, testInfo: TestInfo) {
                    await skipRedirectedGap(
                        testInfo,
                        'CLI exec cannot locate the redirected workspace workload'
                    );
                    submitFlags[SubmitFlagsMerged.command] = '-- sleep infinity';
                    submitFlags[SubmitFlagsMerged.cpuCoreRequest] = '0';
                    submitFlags[SubmitFlagsMerged.gitSync] = 'name=git-mount,repository=https://github.com/octocat/Hello-World,path=/gitmount';

                    runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);
                    logger.debug(`job name: ${jobName} namespace: ${namespace} was created`);
                    const pods = await waitUtils.waitForPods(jobName, namespace, 1, 150, ['Running']);
                    logger.debug(`pod for job name: ${jobName} namespace: ${namespace} was created`);
                    expect(pods.length).to.equal(1);

                    const execFlag = {
                        [execFlags.project]: projectName,
                        [execFlags.command]: 'ls /gitmount'
                    } as Partial<typeof execFlags>;
                    const resultLs = runaiCliV2.exec(execFlag, jobName, cliV2WorkloadType.workspace);
                    logger.debug(`result of ls /gitmount: ${resultLs.stdout}`);
                    expect(resultLs.stdout.trim()).to.include('Hello-World');
                });
            });

            test.describe('gpu', function () {
                test('should have gpu', async function () {
                    workload.gpu = '1';
                    const gpuSubmitFlags = {
                        [SubmitFlagsMerged.image]: workload.image,
                        [SubmitFlagsMerged.project]: projectName,
                        [SubmitFlagsMerged.gpuRequest]: workload.gpu,
                    } as Partial<typeof SubmitFlagsMerged>;
                    runaiCliV2.submit(gpuSubmitFlags, workload.name, cliV2WorkloadType.workspace);

                    const pods = await waitUtils.waitForPods(jobName, namespace, 1);
                    expect(pods.length).to.equal(1);

                    expect(pods[0].spec.containers[0].resources.limits['nvidia.com/gpu']).to.equal(
                        workload.gpu
                    );
                });
            });

            test.describe('test pvcs (interactive job)', function () {
                test('test pvcs (ephemeral + regular)', async function () {
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
                        },
                    };

                    submitFlags[SubmitFlagsMerged.gpuRequest] = '0';
                    submitFlags[SubmitFlagsMerged.command] = '-- sleep infinity';
                    submitFlags[SubmitFlagsMerged.newPvc] = [
                        `claimname=${regularClaimName},path=${pvc1Path},size=1Gi,ro`,
                        `claimname=${ephemeralClaimName},path=${pvc2Path},size=1Gi,ro,ephemeral`
                    ];

                    logger.debug('create workload with ephemeral + regular pvc');
                    runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);
                    await waitUtils.waitForPvcs(namespace, 1, 120, workload.name);
                    await waitUtils.waitForPvcs(namespace, 1, 120, regularClaimName);

                    const podName = (
                        await waitUtils.waitForPods(workload.name, namespace, 1, 120, ['Running'])
                    )[0].metadata.name;

                    const directories = bashCli
                        .exec(
                            `kubectl exec -it -n ${namespace} ${podName} -- /bin/sh -c "cd ../..; ls | grep ${pathPrefix}"`
                        )
                        .split('\n');
                    expect(directories.includes(pvc1Path.slice(1)), `directory ${pvc1Path} exist`).to.true;
                    expect(directories.includes(pvc2Path.slice(1)), `directory ${pvc2Path} exist`).to.true;

                    logger.debug('delete workload with ephemeral + regular pvc');
                    runaiCliV2.delete(
                        {
                            [DeleteFlagsMerged.project]: projectName
                        } as Partial<typeof DeleteFlagsMerged>,
                        workload.name,
                        cliV2WorkloadType.workspace
                    );
                    const numOfExpectedPvcs = versionUtils.isVersionSupported(
                        flowRootState.cluster.version,
                        '2.14'
                    )
                        ? 0
                        : 1;
                    await waitUtils.waitForPvcs(namespace, numOfExpectedPvcs, 120, regularClaimName);
                });

                test(`should have pvcs after create and automatically deleted after delete (ephemeral) ${tags.runOn.parallelCI}`, async function () {
                    const claimName = 'ephemeralclaimname';
                    workload.pvcs = {
                        pvc1: {
                            claimName: claimName,
                            path: '/path',
                            size: '2Gi',
                            readOnly: false,
                            ephemeral: true,
                            readOnlyMany: false,
                            readWriteMany: false,
                            readWriteOnce: true
                        }
                    };
                    submitFlags[SubmitFlagsMerged.command] = '-- sleep infinity';
                    submitFlags[SubmitFlagsMerged.newPvc] = `claimname=${claimName},path=/path,size=2Gi,accessmode-rwo,ephemeral`;

                    logger.debug('create workload with ephemeral pvc');
                    runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);
                    const pvcList = await waitUtils.waitForPvcs(namespace, 1, 120, workload.name);
                    expect(pvcList.items[0].spec.accessModes).to.be.deep.eq(['ReadWriteOnce']);
                    expect(pvcList.items[0].spec.resources.requests).to.be.deep.eq({ storage: '2Gi' });

                    logger.debug('delete workload with ephemeral pvc');
                    runaiCliV2.delete(
                        {[DeleteFlagsMerged.project]: projectName} as Partial<typeof DeleteFlagsMerged>,
                        workload.name,
                        cliV2WorkloadType.workspace
                    );
                    await waitUtils.waitForPvcs(namespace, 0, 120, workload.name);
                });

                test('should have pvcs after create and automatically deleted after delete (regular pvc)', async function () {
                    const claimName = 'regularclaimname';
                    workload.pvcs = {
                        pvc1: {
                            claimName: claimName,
                            path: '/path',
                            size: '1Gi',
                            readOnly: true,
                            readOnlyMany: false,
                            readWriteMany: true,
                            readWriteOnce: false
                        }
                    };
                    submitFlags[SubmitFlagsMerged.newPvc] = `claimname=${claimName},path=/path,size=1Gi,accessmode-rwm,ro`;
                    logger.debug('create workload with regular pvc');
                    runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);
                    let pvcList = await waitUtils.waitForPvcs(namespace, 1, 120, claimName);
                    expect(pvcList.items.length).to.equal(1);
                    expect(pvcList.items[0].spec.accessModes).to.be.deep.eq(['ReadWriteMany']);
                    expect(pvcList.items[0].spec.resources.requests).to.be.deep.eq({ storage: '1Gi' });

                    logger.debug('delete workload with regular pvc');
                    runaiCliV2.delete(
                        {[DeleteFlagsMerged.project]: projectName} as Partial<typeof DeleteFlagsMerged>,
                        workload.name,
                        cliV2WorkloadType.workspace
                    );
                    const numOfExpectedPvcs = versionUtils.isVersionSupported(
                        flowRootState.cluster.version,
                        '2.14'
                    )
                        ? 0
                        : 1;
                    pvcList = await waitUtils.waitForPvcs(namespace, numOfExpectedPvcs, 120, claimName);
                    expect(pvcList.items.length).to.equal(numOfExpectedPvcs);
                });
            });
        });

        test.describe('runai job parameters - large SHM', function () {
            let projectName: string, namespace: string;
            let project: Project;
            let workloadName: string;

            test.beforeAll(() => {
                cliLoginAsSystemAdmin();
            });

            test.beforeEach(async function () {
                projectName = objectGenerator.generateNameByPrefix('project');
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 2,
                    maxAllowedGpus: 2
                });

                workloadName = objectGenerator.generateShortNameByPrefix('workload-shm', 5);
                runaiCliV2.projectSet(project.name);
            });

            test.afterEach(async function () {
                try {
                    runaiCliV2.delete(undefined, workloadName, cliV2WorkloadType.workspace);
                    await waitUtils.waitForWorkloadPodNotExist(namespace, workloadName);
                    await operations.deleteProjectAndNamespace(project.id);
                } catch {
                    logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
                }
            });

            test('Should mount a shared memory volume and allow researchers to store data in it', async function ({}, testInfo: TestInfo) {
                await skipRedirectedGap(
                    testInfo,
                    'CLI logs cannot locate the redirected workspace workload'
                );
                const submitFlag = {
                    [SubmitFlagsMerged.image]: config.quickStartImage.ubuntu,
                    [SubmitFlagsMerged.largeShm]: '',
                    [SubmitFlagsMerged.command]: '-- sh -c "df /dev/shm"'
                } as Partial<typeof SubmitFlagsMerged>;
                runaiCliV2.submit(submitFlag, workloadName, cliV2WorkloadType.workspace);
                await waitUtils.waitForBackendWorkloadPhase(workloadName, project.id, Phase.Completed);

                const logs = runaiCliV2.workspaceLogs(workloadName, {});
                expect(logs.stdout).to.include('Filesystem');
                expect(logs.stdout).to.include('/dev/shm');
            });
        });

        test.describe('Submitted interactive job with hostIpc and hostNetwork', function () {
            let projectName: string, namespace: string;
            let project: Project;
            let workload: InteractiveWorkload;
            let jobName: string;

            test.beforeEach(async function () {
                projectName = objectGenerator.generateNameByPrefix('project');
                namespace = Utils.namespaceFromProjectName(projectName);
                project = await projectResourcesUtils.createProject(namespace, {
                    name: projectName,
                    deservedGpus: 1,
                    maxAllowedGpus: 1
                });

                jobName = objectGenerator.generateNameByPrefix('job-with-large-shm');

                workload = new InteractiveWorkload(config.quickStartImage.interactive, namespace, jobName);
                workload.hostIpc = true;
                workload.hostNetwork = true;
            });

            test.afterEach(async function () {
                try {
                    await operations.deleteWorkloads([workload]);
                    await operations.deleteProjectAndNamespace(project.id);
                } catch {
                    logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
                }
            });

            test('pod expected with true value for hostIpc and hostNetwork', async function () {
                const submitFlags = {
                    [SubmitFlagsMerged.image]: workload.image,
                    [SubmitFlagsMerged.hostIpc]: '',
                    [SubmitFlagsMerged.hostNetwork]: '',
                    [SubmitFlagsMerged.project]: projectName
                } as Partial<typeof SubmitFlagsMerged>;
                runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);

                const pods = await waitUtils.waitForPods(jobName, namespace, 1, 60, ['Running']);
                expect(pods.length).to.equal(1);
                expect(pods[0].spec.hostIPC).to.equal(true);
                expect(pods[0].spec.hostNetwork).to.equal(true);
            });
        });

        supportedFrom(SupportedVersions.V2_24)(function () {
            test.describe('preemptible', function () {
                let projectName: string, namespace: string;
                let project: Project;
                let workload: InteractiveWorkload;
                let jobName: string;

                test.beforeAll(() => {
                    cliLoginAsSystemAdmin();
                });

                test.beforeEach(async function () {
                    projectName = objectGenerator.generateNameByPrefix('project');
                    namespace = Utils.namespaceFromProjectName(projectName);
                    project = await projectResourcesUtils.createProject(namespace, {
                        name: projectName,
                        deservedGpus: 1,
                        maxAllowedGpus: 1
                    });

                    jobName = objectGenerator.generateNameByPrefix('preemptible');

                    workload = new InteractiveWorkload(
                        config.quickStartImage.interactive,
                        namespace,
                        jobName
                    );
                    workload.preemptibility = Preemptibility.Preemptible;
                });

                test.afterEach(async function () {
                    try {
                        await operations.deleteWorkloads([workload]);
                        await operations.deleteProjectAndNamespace(project.id);
                    } catch {
                        logger.error(`Failed to delete project ${projectName} and namespace ${namespace}`);
                    }
                });

                test('workload resource contains the correct label', async () => {
                    const submitFlags = {
                        [SubmitFlagsMerged.image]: workload.image,
                        [SubmitFlagsMerged.preemptible]: '',
                        [SubmitFlagsMerged.project]: projectName
                    } as Partial<typeof SubmitFlagsMerged>;
                    runaiCliV2.submit(submitFlags, workload.name, cliV2WorkloadType.workspace);

                    if (
                        await isAnyWorkloadSubmissionEnabled(AnyWorkloadSubmissionType.Workspace)
                    ) {
                        await waitUtils.waitForBatchJobToExist(jobName, namespace);
                        const job = await kubeCli.getBatchJob(jobName, namespace);
                        expect(job.metadata?.labels?.priorityClassName).to.equal('interactive-preemptible');
                    } else {
                        await waitUtils.waitForCrd(crds.runaiJobCrd, jobName, namespace);
                        const job = await kubeCli.getCrd(crds.runaiJobCrd, jobName, namespace);
                        expect(job['metadata']['labels']['priorityClassName']).to.equal(
                            'interactive-preemptible'
                        );
                    }
                });
            });
        });
    });
});
