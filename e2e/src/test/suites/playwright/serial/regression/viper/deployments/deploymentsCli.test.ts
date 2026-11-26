/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { test } from '@playwright/test';
import {
    config,
    flowRootState,
    logger,
    objectGenerator,
    operations,
    orgUnitUtils,
    tags,
    Utils,
    waitUtils
} from '../../../../../../../modules/helpers';
import {
    waitAndValidateDeploymentCreatedSuccessfullyInBackend,
    validateDeploymentCreatedSuccessfullyInCluster,
    waitAndValidateDeploymentInBackendStatusAvailable,
    validateDeploymentsDeleted
} from '../../../../../../../modules/helpers/Validations';
import { kubeCli } from '../../../../../../../modules/cli';
import { V1Deployment } from '@kubernetes/client-node/';
import { Project } from '../../../../../../../modules/api/apis/orgUnitApi';

let namespaceName: string;
let projectName: string;
let deploymentName: string;
let replicas: number;

function getDeployment(
    name = deploymentName,
    namespace = namespaceName,
    project = projectName,
    replicas = 2
): V1Deployment {
    const deployment: V1Deployment = {
        apiVersion: 'apps/v1',
        kind: 'Deployment',
        metadata: {
            name,
            namespace,
            labels: {
                app: 'test-app',
                project
            }
        },
        spec: {
            replicas,
            selector: {
                matchLabels: {
                    app: 'test-app'
                }
            },
            strategy: {
                rollingUpdate: {
                    maxSurge: '25%',
                    maxUnavailable: '25%'
                },
                type: 'RollingUpdate'
            },
            template: {
                metadata: {
                    labels: {
                        app: 'test-app',
                        project
                    }
                },
                spec: {
                    containers: [
                        {
                            name: 'ubuntu',
                            image: 'ubuntu',
                            command: ['sleep'],
                            args: ['infinity'],
                            imagePullPolicy: 'Always',
                            resources: {
                                limits: {
                                    'nvidia.com/gpu': '1'
                                }
                            },
                            terminationMessagePath: '/dev/termination-log',
                            terminationMessagePolicy: 'File'
                        }
                    ],
                    dnsPolicy: 'ClusterFirst',
                    restartPolicy: 'Always',
                    schedulerName: 'runai-scheduler',
                    securityContext: {},
                    terminationGracePeriodSeconds: 5
                }
            }
        }
    };
    logger.debug(`Deployment structure: ${JSON.stringify(deployment)}`);
    return deployment;
}

test.describe(`deployment over kubectl and backend api, tags:  ${tags.resource.deployments} ${tags.tool.api} ${tags.teams.viper}`, () => {
    let clusterId: string;
    let project: Project;

    test.beforeAll(() => {
        replicas = 1;
        clusterId = flowRootState.cluster.uuid;
    });

    test.describe('Submit', () => {
        test.beforeEach(async () => {
            projectName = objectGenerator.generateNameByPrefix('project');
            namespaceName = Utils.namespaceFromProjectName(projectName);

            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 2, limit: 2 } }]
            });
            namespaceName = project.status.namespace;

            deploymentName = objectGenerator.generateNameByPrefix('deployment');
        });

        test.afterEach(async () => {
            try {
                await operations.deleteDeployment(deploymentName, namespaceName, clusterId);
                // sleep for just a little bit for the cluster-sync to delete the deployment from the backend
                await waitUtils.sleep(100);
            } catch (error) {
                if (error.code !== 404) {
                    logger.warn(`Failed to delete deployment ${deploymentName}, error: ${error}`);
                }
            }

            try {
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            } catch (error) {
                logger.error(`Failed to delete project ${projectName} and namespace ${namespaceName}`);
            }
        });

        test('creates deployment in cluster and control-plane', async () => {
            const result = await kubeCli.createDeployment(
                getDeployment(deploymentName, namespaceName, projectName, replicas)
            );
            let deployment = result;

            // Validate deployment creation in cluster
            validateDeploymentCreatedSuccessfullyInCluster(
                deployment,
                deploymentName,
                namespaceName,
                projectName,
                replicas,
                config.quickStartImage.ubuntu
            );

            // Validate creation in backend
            await waitAndValidateDeploymentCreatedSuccessfullyInBackend(clusterId, deployment, replicas);

            // Wait for deployment to become available
            await waitUtils.waitForDeploymentToBeAvailable(deploymentName, namespaceName);

            deployment = await operations.getDeployment(namespaceName, deploymentName);
            await waitAndValidateDeploymentInBackendStatusAvailable(clusterId, deployment);
        });
    });

    test.describe('Delete', () => {
        test.beforeEach(async () => {
            projectName = objectGenerator.generateNameByPrefix('project');
            namespaceName = Utils.namespaceFromProjectName(projectName);

            project = await orgUnitUtils.createProjectAndValidate({
                name: projectName,
                clusterId: flowRootState.cluster.uuid,
                resources: [{ gpu: { deserved: 2, limit: 2 } }]
            });
            namespaceName = project.status.namespace;

            deploymentName = objectGenerator.generateNameByPrefix('deployment');
        });

        test.afterEach(async () => {
            try {
                await operations.deleteDeployment(deploymentName, namespaceName, clusterId);
                // sleep for just a little bit for the cluster-sync to delete the deployment from the backend
                await waitUtils.sleep(100);
            } catch (error) {
                if (error.code !== 404) {
                    logger.warn(`Failed to delete deployment ${deploymentName}, error: ${error}`);
                }
            }

            try {
                await orgUnitUtils.deleteProjectAndNamespace(project.id, project.name);
            } catch (error) {
                logger.error(`Failed to delete project ${projectName} and namespace ${namespaceName}`);
            }
        });

        test('deletes the deployment', async () => {
            replicas = 1;
            const deployment = await kubeCli.createDeployment(
                getDeployment(deploymentName, namespaceName, projectName, replicas)
            );

            validateDeploymentCreatedSuccessfullyInCluster(
                deployment,
                deploymentName,
                namespaceName,
                projectName,
                replicas,
                config.quickStartImage.ubuntu
            );

            await operations.deleteDeployment(deploymentName, namespaceName, clusterId);

            await validateDeploymentsDeleted(deployment.metadata.uid, project.id.toString());
        });
    });
});
