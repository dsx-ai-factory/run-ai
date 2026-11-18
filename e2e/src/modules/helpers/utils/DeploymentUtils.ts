/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { kubeCli } from '../../cli';
import { logger, nodePoolUtils } from '..';
import { waitUtils } from '..';
import { V1Deployment } from '@kubernetes/client-node';

export class InferenceClientCreator {
    async createClient(
        clientName: string,
        namespace: string,
        apiVersion: string,
        clientImage: string,
        clientCommand: string[]
    ): Promise<void> {
        const client = {
            apiVersion: apiVersion,
            kind: 'Pod',
            metadata: {
                name: clientName,
                namespace: namespace
            },
            spec: {
                containers: [
                    {
                        name: 'client',
                        image: clientImage,
                        command: clientCommand
                    }
                ]
            }
        };
        await kubeCli.CoreV1Api.createNamespacedPod({ namespace, body: client });
        waitUtils.waitForPods(clientName, namespace, 1, 20, ['Running'], true);
        logger.info(`Client is ${clientName} created and Running`);
    }

    async getKnativeServiceRoute(deploymentName: string, namespace: string): Promise<string> {
        const kservice = await kubeCli.getKnativeService(deploymentName, namespace);
        logger.info(`route is ${kservice.status.url.replace('http://', '')}`);
        return kservice.status.url.replace('http://', '');
    }

    getDeployment(
        namespaceName: string,
        projectName: string,
        deploymentName: string,
        replicas: number,
        image: string,
        nodePoolName?: string
    ): V1Deployment {
        const deployment = {
            apiVersion: 'apps/v1',
            kind: 'Deployment',
            metadata: {
                name: deploymentName,
                namespace: namespaceName,
                labels: {
                    app: 'test-app',
                    project: projectName
                }
            },
            spec: {
                replicas: replicas,
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
                            project: projectName
                        }
                    },
                    spec: {
                        containers: [
                            {
                                command: ['sleep'],
                                args: ['infinity'],
                                image: image,
                                imagePullPolicy: 'Always',
                                name: 'ubuntu',
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
        if (nodePoolName !== 'default' && nodePoolName !== null) {
            deployment['spec']['template']['metadata']['labels']['runai/node-pool'] = nodePoolName;
        }

        logger.debug(`Deployment structure: ${JSON.stringify(deployment)}`);
        return deployment;
    }

    getDeploymentWithMultiNodePools(
        namespaceName: string,
        projectName: string,
        deploymentName: string,
        replicas: number,
        image: string,
        nodePools: { name: string; labelKey: string; labelValue: string }[]
    ): V1Deployment {
        const deployment = this.getDeployment(namespaceName, projectName, deploymentName, replicas, image);
        const affinity = nodePoolUtils.buildNodeAffinityForNodePools(nodePools);

        logger.debug(`Deployment's affinity structure: ${JSON.stringify(affinity)}`);
        deployment['spec']['template']['spec']['affinity'] = affinity;
        return deployment;
    }

    async setDeploymentReplica(
        deploymentName: string,
        value: number,
        namespace = 'runai-backend'
    ): Promise<void> {
        const patch = {
            spec: {
                replicas: value
            }
        };
        await kubeCli.patchDeployment(deploymentName, namespace, patch);
    }
}
