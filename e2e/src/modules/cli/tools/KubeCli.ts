/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// https://github.com/kubernetes-client/javascript
import * as k8s from '@kubernetes/client-node';
import {
    V1Ingress,
    V1Pod,
    V1ReplicaSet,
    V1RoleBinding,
    V1ClusterRoleBinding,
    V1Secret,
    V1SecretList,
    V1Service,
    V1ServiceAccount,
    V1EnvVar,
    V1Deployment,
    V1Node,
    V1NodeList,
    V1NamespaceList,
    V1ConfigMap,
    V1PersistentVolumeClaimList,
    V1ValidatingWebhookConfigurationList,
    V1PersistentVolumeClaim,
    V1ObjectMeta,
    V1Job,
    KubernetesObject,
    V2HorizontalPodAutoscaler
} from '@kubernetes/client-node';
import * as yaml from 'js-yaml';
import { Writable, finished } from 'stream';
import {
    BaseWorkload,
    CliError,
    crds,
    CustomResourceDefinition,
    flowRootState,
    logger,
    NodeResources,
    settings,
    Utils,
    waitUtils,
    tenantUtils,
    TokenObject,
    config
} from '../../helpers';
import { KubeCliContextProvider } from './KubeCliContextProvider';
import { podLabelKeys } from '../../helpers/models/pods';
import { KnativeRevision, KnativeService } from '../../helpers/models/knativeService';
import { bashCli, kubeCli } from '../index';
import { LeaderWorkerSet } from '../../helpers/models/lws';
import child_process from 'node:child_process';
import { spawn } from 'child_process';
import { get } from 'lodash-es';

const MINIO_TEST_IMAGE = 'chainguard/minio:latest';
const S3_SEED_IMAGE = 'public.ecr.aws/aws-cli/aws-cli:2.27.60';

class KubeCli extends KubeCliContextProvider {
    CoreV1Api: k8s.CoreV1Api;

    RbacV1Api: k8s.RbacAuthorizationV1Api;

    CustomObjectsApi: k8s.CustomObjectsApi;

    AppsV1Api: k8s.AppsV1Api;

    KubernetesObjectApi: k8s.KubernetesObjectApi;

    AdmissionregistrationV1Api: k8s.AdmissionregistrationV1Api;

    StorageV1Api: k8s.StorageV1Api;

    networkingV1Api: k8s.NetworkingV1Api;

    BatchV1Api: k8s.BatchV1Api;

    AutoscalingV2Api: k8s.AutoscalingV2Api;

    ApiextensionsV1Api: k8s.ApiextensionsV1Api;

    private readonly versionApi: k8s.VersionApi;

    constructor(kubeconfigPath: string, clusterName?: string) {
        super(kubeconfigPath, clusterName);
        // Init API clients
        this.CoreV1Api = this.kubeClient.makeApiClient(k8s.CoreV1Api);
        this.AppsV1Api = this.kubeClient.makeApiClient(k8s.AppsV1Api);
        this.RbacV1Api = this.kubeClient.makeApiClient(k8s.RbacAuthorizationV1Api);
        this.CustomObjectsApi = this.kubeClient.makeApiClient(k8s.CustomObjectsApi);
        this.KubernetesObjectApi = this.kubeClient.makeApiClient(k8s.KubernetesObjectApi);
        this.AdmissionregistrationV1Api = this.kubeClient.makeApiClient(k8s.AdmissionregistrationV1Api);
        this.StorageV1Api = this.kubeClient.makeApiClient(k8s.StorageV1Api);
        this.networkingV1Api = this.kubeClient.makeApiClient(k8s.NetworkingV1Api);
        this.BatchV1Api = this.kubeClient.makeApiClient(k8s.BatchV1Api);
        this.AutoscalingV2Api = this.kubeClient.makeApiClient(k8s.AutoscalingV2Api);
        this.ApiextensionsV1Api = this.kubeClient.makeApiClient(k8s.ApiextensionsV1Api);
        this.versionApi = this.kubeClient.makeApiClient(k8s.VersionApi);
    }

    async getServerVersion(): Promise<string> {
        return (await this.versionApi.getCode()).gitVersion;
    }

    async listPodsByNamepsace(namespace: string): Promise<V1Pod[]> {
        return (await this.CoreV1Api.listNamespacedPod({ namespace })).items;
    }

    async listAllPods(): Promise<V1Pod[]> {
        return (await this.CoreV1Api.listPodForAllNamespaces()).items;
    }

    async listPodsByNamespaceAndPrefix(namespace: string, namePrefix: string): Promise<V1Pod[]> {
        const pods = (await this.CoreV1Api.listNamespacedPod({ namespace })).items;
        return pods.filter((pod) => {
            return pod.metadata.name.startsWith(namePrefix);
        });
    }

    async listNamespaces(): Promise<V1NamespaceList> {
        return await this.CoreV1Api.listNamespace();
    }

    async execInPod(
        podName: string,
        namespace: string,
        command: string[]
    ): Promise<{ stdout: string; stderr: string }> {
        const pod = await this.CoreV1Api.readNamespacedPod({ name: podName, namespace });
        const exec = new k8s.Exec(this.kubeClient);
        let output = '';
        let errorOutput = '';
        const stdout = new Writable({
            write: (chunk: Buffer, _encoding: string, callback: () => void): void => {
                output += chunk.toString();
                callback();
            }
        });
        const stderr = new Writable({
            write: (chunk: Buffer, _encoding: string, callback: () => void): void => {
                errorOutput += chunk.toString();
                callback();
            }
        });

        await exec.exec(
            namespace,
            podName,
            pod.spec?.containers?.[0]?.name,
            command,
            stdout,
            stderr,
            null,
            false
        );
        await Promise.all([
            new Promise((resolve) => finished(stdout, resolve)),
            new Promise((resolve) => finished(stderr, resolve))
        ]);
        return { stdout: output.trim(), stderr: errorOutput.trim() };
    }

    async getLogs(podNameSubString: string, namespace = 'runai', tailLines = 700): Promise<string> {
        const pods = (await this.CoreV1Api.listNamespacedPod({ namespace })).items;
        const filteredPods = pods.filter((pod) => {
            return pod.metadata.name.includes(podNameSubString);
        });
        let res = '';
        for (const pod of filteredPods) {
            res += `Showing ${pod.metadata.name} logs: \n`;
            const podLogs = await this.CoreV1Api.readNamespacedPodLog({
                name: pod.metadata.name,
                namespace,
                tailLines
            });
            res += podLogs;
        }

        return res;
    }

    async printLogs(podNameSubString: string, namespace = 'runai', tailLines = 700): Promise<void> {
        const logs = await this.getLogs(podNameSubString, namespace, tailLines);
        logger.info(logs);
    }

    async getKnativeService(name: string, namespace: string): Promise<KnativeService> {
        const kservice = await this.getCrd(crds.knativeService, name, namespace);
        return kservice as KnativeService;
    }

    async getKnativeRevision(name: string, namespace: string): Promise<KnativeRevision> {
        const revision = await this.getCrd(crds.knativeRevision, name, namespace);
        return revision as KnativeRevision;
    }

    async getLeaderWorkerSet(name: string, namespace: string): Promise<LeaderWorkerSet> {
        const lws = await this.getCrd(crds.lwsCrd, name, namespace);
        return lws as LeaderWorkerSet;
    }

    async getIngressHosts(namespace: string): Promise<string> {
        const hostUrl = await this.networkingV1Api.listNamespacedIngress({ namespace });
        return 'https://' + hostUrl.items[0].spec.rules[0].host;
    }

    async getRunaiJob(name: string, namespace: string): Promise<KubernetesObject> {
        const job = await this.getCrd(crds.runaiJobCrd, name, namespace);
        return job;
    }

    async getBatchJob(name: string, namespace: string): Promise<V1Job> {
        return await this.BatchV1Api.readNamespacedJob({ name, namespace });
    }

    async createS3Bucket(podName: string, namespace: string): Promise<void> {
        // Complete the upload and verify it through S3 before the mount test starts.
        const podManifest: V1Pod = {
            apiVersion: 'v1',
            kind: 'Pod',
            metadata: { name: podName, namespace },
            spec: {
                restartPolicy: 'Never',
                activeDeadlineSeconds: 60,
                containers: [
                    {
                        name: 'seed-s3',
                        image: S3_SEED_IMAGE,
                        env: [
                            {
                                name: 'AWS_ACCESS_KEY_ID',
                                value: 'minioadmin'
                            },
                            {
                                name: 'AWS_SECRET_ACCESS_KEY',
                                value: 'minioadmin'
                            },
                            {
                                name: 'AWS_DEFAULT_REGION',
                                value: 'us-east-1'
                            },
                            {
                                name: 'AWS_MAX_ATTEMPTS',
                                value: '1'
                            }
                        ],
                        command: ['/bin/sh', '-ec'],
                        args: [
                            [
                                `endpoint=http://minio-service.${namespace}.svc:9000`,
                                'attempt=0',
                                'until aws --endpoint-url "$endpoint" s3api list-buckets --cli-connect-timeout 2 --cli-read-timeout 2 > /dev/null; do',
                                '  attempt=$((attempt + 1))',
                                '  if [ "$attempt" -ge 10 ]; then echo "MinIO did not become ready" >&2; exit 1; fi',
                                '  sleep 2',
                                'done',
                                'aws --endpoint-url "$endpoint" s3api head-bucket --bucket my-new-bucket || aws --endpoint-url "$endpoint" s3api create-bucket --bucket my-new-bucket',
                                'printf "s3 mount test\\n" > /tmp/e2e-test-file',
                                'aws --endpoint-url "$endpoint" s3api put-object --bucket my-new-bucket --key e2e-test-file --body /tmp/e2e-test-file',
                                'aws --endpoint-url "$endpoint" s3api head-object --bucket my-new-bucket --key e2e-test-file'
                            ].join('\n')
                        ]
                    }
                ]
            }
        };

        try {
            await this.CoreV1Api.createNamespacedPod({ namespace, body: podManifest });
            logger.info(`S3 seed pod ${podName} created in namespace ${namespace}`);
            await waitUtils.waitForCondition(
                async () => {
                    const pod = await this.CoreV1Api.readNamespacedPod({ name: podName, namespace });
                    return ['Succeeded', 'Failed'].includes(pod.status?.phase);
                },
                `S3 seed pod ${podName} to finish uploading and verifying e2e-test-file`,
                true,
                60
            );
            const pod = await this.CoreV1Api.readNamespacedPod({ name: podName, namespace });
            if (pod.status?.phase !== 'Succeeded') {
                throw new Error(`S3 bucket setup failed: ${JSON.stringify(pod.status)}`);
            }
        } catch (error) {
            await this.logS3Diagnostics(namespace);
            throw error;
        }
    }

    async deployMinio(namespace: string): Promise<void> {
        const minioDeployment = {
            apiVersion: 'apps/v1',
            kind: 'Deployment',
            metadata: {
                name: 'minio-server',
                namespace: namespace
            },
            spec: {
                replicas: 1,
                selector: {
                    matchLabels: {
                        app: 'minio-server'
                    }
                },
                template: {
                    metadata: {
                        labels: {
                            app: 'minio-server'
                        }
                    },
                    spec: {
                        securityContext: {
                            fsGroup: 65532
                        },
                        containers: [
                            {
                                name: 'minio-server',
                                image: MINIO_TEST_IMAGE,
                                readinessProbe: {
                                    httpGet: { path: '/minio/health/ready', port: 9000 },
                                    periodSeconds: 1,
                                    timeoutSeconds: 2
                                },
                                ports: [
                                    {
                                        containerPort: 9000
                                    }
                                ],
                                env: [
                                    {
                                        name: 'MINIO_ROOT_USER',
                                        value: 'minioadmin'
                                    },
                                    {
                                        name: 'MINIO_ROOT_PASSWORD',
                                        value: 'minioadmin'
                                    }
                                ],
                                args: ['server', '/data'],
                                volumeMounts: [
                                    {
                                        name: 'minio-data',
                                        mountPath: '/data'
                                    }
                                ]
                            }
                        ],
                        volumes: [
                            {
                                name: 'minio-data',
                                emptyDir: {}
                            }
                        ]
                    }
                }
            }
        };

        try {
            await this.AppsV1Api.createNamespacedDeployment({ namespace, body: minioDeployment });
            logger.info('Minio deployment created.');
            await waitUtils.waitForCondition(
                async () => {
                    const deployment = await this.AppsV1Api.readNamespacedDeployment({
                        name: 'minio-server',
                        namespace
                    });
                    return deployment.status?.availableReplicas === 1;
                },
                'MinIO deployment to be ready to serve S3 requests',
                true,
                60
            );
        } catch (error) {
            logger.error('Error deploying Minio:', error.message);
            await this.logS3Diagnostics(namespace);
            throw error;
        }
    }

    async logS3Diagnostics(namespace: string): Promise<void> {
        const isS3Pod = (name: string): boolean =>
            name?.startsWith('minio-server-') || name === 'my-new-bucket' || name === 'my-python-pod';
        try {
            const endpoints = await this.CoreV1Api.readNamespacedEndpoints({
                name: 'minio-service',
                namespace
            });
            logger.error(`MinIO service endpoints: ${JSON.stringify(endpoints.subsets)}`);
        } catch (error) {
            logger.warn(`Could not collect MinIO endpoints: ${error.message}`);
        }
        try {
            const events = await this.CoreV1Api.listNamespacedEvent({ namespace });
            logger.error(
                `S3 namespace events: ${JSON.stringify(
                    events.items
                        .filter((event) => isS3Pod(event.involvedObject?.name))
                        .map((event) => ({
                            object: event.involvedObject?.name,
                            reason: event.reason,
                            message: event.message
                        }))
                )}`
            );
            const pods = await this.CoreV1Api.listNamespacedPod({ namespace });
            for (const pod of pods.items.filter((item) => isS3Pod(item.metadata?.name))) {
                logger.error(`S3 pod ${pod.metadata.name} status: ${JSON.stringify(pod.status)}`);
                const containers = [
                    ...(pod.status?.initContainerStatuses ?? []),
                    ...(pod.status?.containerStatuses ?? [])
                ];
                for (const container of containers) {
                    for (const previous of container.restartCount > 0 ? [false, true] : [false]) {
                        try {
                            const logs = await this.CoreV1Api.readNamespacedPodLog({
                                name: pod.metadata.name,
                                namespace,
                                container: container.name,
                                tailLines: 100,
                                previous
                            });
                            logger.error(
                                `S3 logs ${pod.metadata.name}/${container.name} (previous=${previous}):\n${logs}`
                            );
                        } catch (error) {
                            logger.warn(
                                `Could not collect ${pod.metadata.name}/${container.name} logs: ${error.message}`
                            );
                        }
                    }
                }
            }
        } catch (error) {
            logger.warn(`Could not collect S3 pod diagnostics: ${error.message}`);
        }
    }

    async deleteMinio(namespace: string): Promise<void> {
        try {
            await this.AppsV1Api.deleteNamespacedDeployment({ name: 'minio-server', namespace });
            logger.info('Minio deployment deleted.');
        } catch (error) {
            logger.error('Error deleting Minio deployment:', error.message);
            throw error;
        }
    }

    async createMinioService(namespace: string): Promise<void> {
        // Tests reach minio through the service ClusterIP on port 9000 (see getMinioServiceIP),
        // so a ClusterIP service is all that's needed. Using NodePort would allocate a
        // cluster-wide port that collides across namespaces when parallel tests run concurrently.
        const service = {
            apiVersion: 'v1',
            kind: 'Service',
            metadata: {
                name: 'minio-service',
                namespace: namespace
            },
            spec: {
                type: 'ClusterIP',
                ports: [
                    {
                        port: 9000,
                        targetPort: 9000,
                        protocol: 'TCP',
                        name: 'http'
                    }
                ],
                selector: {
                    app: 'minio-server' // Ensure this label matches the label on the Minio pod
                }
            }
        };

        try {
            await this.CoreV1Api.createNamespacedService({ namespace, body: service });
            logger.info('Minio service created successfully.');
        } catch (error) {
            logger.error('Error creating Minio service:', error.message);
            throw error;
        }
    }

    async getMinioServiceIP(namespace: string): Promise<string> {
        try {
            const service = await this.getServiceByName(namespace, 'minio-service');
            const ip = service.spec.clusterIP;
            if (!ip) {
                throw new Error('Minio service IP not found');
            }
            return ip;
        } catch (error) {
            logger.error('Error getting Minio service IP:', error.message);
            throw error;
        }
    }

    async deleteMinioService(namespace: string): Promise<void> {
        try {
            await this.CoreV1Api.deleteNamespacedService({ name: 'minio-service', namespace });
            console.log('Minio service deleted successfully.');
        } catch (error) {
            console.error('Error deleting Minio service:', error);
        }
    }

    async deleteService(serviceName: string, namespace: string): Promise<void> {
        try {
            await this.CoreV1Api.deleteNamespacedService({ name: serviceName, namespace });
            logger.info(`Deleted service ${serviceName} in namespace ${namespace}`);
        } catch (error) {
            // A missing service is the expected no-op for pre-delete/idempotent cleanup.
            // Any other error (e.g. RBAC 403, API failure) is real — surface it, don't hide it.
            if ((error as any)?.code === 404) {
                logger.debug(`service ${serviceName} not found in namespace ${namespace}, nothing to delete`);
            } else {
                logger.error(`failed to delete service ${serviceName} in namespace ${namespace}: ${error}`);
            }
        }
    }

    async podLogs(podName: string, namespace: string, tailLines = 150): Promise<string> {
        const res = await this.CoreV1Api.readNamespacedPodLog({
            name: podName,
            namespace,
            tailLines
        });
        return res;
    }

    async getPodEnvVarByPrefix(
        podPrefix: string,
        envVarField: string,
        namespace = 'runai'
    ): Promise<V1EnvVar> {
        const a = await this.CoreV1Api.listNamespacedPod({
            namespace,
            labelSelector: `app=${podPrefix}`
        });
        return a.items[0].spec.containers[0].env.find((envVar) => envVar.name === envVarField);
    }

    async waitForPodToRun(namespace: string, prefix: string): Promise<void> {
        return waitUtils.waitForCondition(
            async () => {
                const pods = await this.CoreV1Api.listNamespacedPod({ namespace });
                return (
                    pods.items.filter(
                        (pod) => pod.metadata.name.startsWith(prefix) && pod.status.phase === 'Running'
                    ).length > 0
                );
            },
            `Waiting for pod with prefix ${prefix} to run`,
            true,
            60
        );
    }

    async listPodsOfJob(namespace: string, jobName: string, isPrefixName = false): Promise<V1Pod[]> {
        return (await this.CoreV1Api.listNamespacedPod({ namespace })).items.filter((pod) => {
            const owners = pod.metadata.ownerReferences ?? [];
            return (
                owners.some((reference) => reference.name === jobName) ||
                (isPrefixName && owners.some((reference) => reference.name.startsWith(jobName))) ||
                pod.metadata.labels?.[podLabelKeys.INFERENCE_WORKLOAD] === jobName
            );
        });
    }

    async waitForJobPodsRunning(namespace: string, jobName: string, timeoutSeconds = 180): Promise<void> {
        await waitUtils.waitForCondition(
            async () => {
                const pods = await this.listPodsOfJob(namespace, jobName, true);
                return pods.length > 0 && pods.every((pod) => pod.status?.phase === 'Running');
            },
            `wait for pods of ${jobName} in ${namespace} to reach Running`,
            true,
            timeoutSeconds,
            3
        );
    }

    async listPVCsByNamespace(namespace: string): Promise<V1PersistentVolumeClaimList> {
        const pvcs = await this.CoreV1Api.listNamespacedPersistentVolumeClaim({ namespace });
        return pvcs;
    }

    async listPodsByOwnerPrefix(namespace: string, ownerPrefix: string): Promise<V1Pod[]> {
        return (await this.CoreV1Api.listNamespacedPod({ namespace })).items.filter((pod) =>
            pod.metadata.ownerReferences.some((reference) => reference.name.startsWith(ownerPrefix))
        );
    }

    async listReplicaSetsByOwnerPrefix(namespace: string, ownerPrefix: string): Promise<V1ReplicaSet[]> {
        return (await this.AppsV1Api.listNamespacedReplicaSet({ namespace })).items.filter((replicaSet) =>
            replicaSet.metadata.ownerReferences.some((reference) => reference.name.startsWith(ownerPrefix))
        );
    }

    async listRoleBindings(namespace: string): Promise<V1RoleBinding[]> {
        return (await this.RbacV1Api.listNamespacedRoleBinding({ namespace })).items;
    }

    async getRoleBindingByName(namespace: string, roleBindingName: string): Promise<V1RoleBinding> {
        const allRoleBindings = await this.listRoleBindings(namespace);
        return allRoleBindings.find((allRoleBindings) => allRoleBindings.metadata.name === roleBindingName);
    }

    async listClusterRoleBindings(): Promise<V1ClusterRoleBinding[]> {
        return (await this.RbacV1Api.listClusterRoleBinding()).items;
    }

    async getClusterRoleBinding(name: string): Promise<V1ClusterRoleBinding> {
        const allClusterRoleBindings = await this.listClusterRoleBindings();
        return allClusterRoleBindings.find((clusterRoleBinding) => clusterRoleBinding.metadata.name === name);
    }

    async getRouteByName(namespace: string, ingressName: string): Promise<object> {
        const response = await this.CustomObjectsApi.listNamespacedCustomObject({
            group: 'route.openshift.io',
            version: 'v1',
            namespace,
            plural: 'routes',
            fieldSelector: `metadata.name=${ingressName}`
        });
        if (response['items'].length) {
            return response['items'][0];
        } else {
            return undefined;
        }
    }

    async getServiceByName(namespace: string, serviceName: string): Promise<V1Service> {
        const services = await this.CoreV1Api.listNamespacedService({
            namespace,
            fieldSelector: `metadata.name=${serviceName}`
        });
        if (services.items.length) {
            return services.items[0];
        } else {
            return undefined;
        }
    }

    async patchServiceStatus(
        namespace: string,
        serviceName: string,
        patchBody: Array<{ op: string; path: string; value: unknown }>
    ): Promise<void> {
        await this.CoreV1Api.patchNamespacedServiceStatus({
            name: serviceName,
            namespace,
            body: patchBody
        });
    }

    async getIngressByName(namespace: string, ingressName: string): Promise<V1Ingress> {
        const ingresses = await this.networkingV1Api.listNamespacedIngress({
            namespace,
            fieldSelector: `metadata.name=${ingressName}`
        });
        if (ingresses.items.length) {
            return ingresses.items[0];
        } else {
            return undefined;
        }
    }

    async getNodes(labelSelector?: string): Promise<V1NodeList> {
        const nodesList = await this.CoreV1Api.listNode({ labelSelector });
        return nodesList;
    }

    async getNodeLabels(nodeName: string): Promise<{ [key: string]: string }> {
        const res = await this.CoreV1Api.readNode({ name: nodeName });
        return res.metadata.labels;
    }

    async patchNode(nodeName: string, patchBody: object): Promise<void> {
        await this.CoreV1Api.patchNode({
            name: nodeName,
            body: patchBody
        });
    }

    async getNode(nodeName: string): Promise<V1Node> {
        return await this.CoreV1Api.readNode({ name: nodeName });
    }

    // Conditions live on the status subresource, which patchNode cannot reach.
    // The generated client always negotiates application/json-patch+json here,
    // so callers must pass JSON Patch operations rather than a merge document.
    async patchNodeStatus(
        nodeName: string,
        patchBody: Array<{ op: string; path: string; value?: unknown }>
    ): Promise<void> {
        await this.CoreV1Api.patchNodeStatus({
            name: nodeName,
            body: patchBody
        });
    }

    async labelNode(nodeName: string, labelKey: string, labelValue: string): Promise<void> {
        const patch = [
            {
                op: 'add',
                path: `/metadata/labels/${labelKey.replace(/\//g, '~1')}`,
                value: labelValue
            }
        ];
        await this.patchNode(nodeName, patch)
            .then(() => {
                logger.info(`node: ${nodeName} labeled: ${labelKey}/${labelValue}`);
            })
            .catch((err) => {
                logger.info(`Error labeling node: ${nodeName}, err: ${err}`);
            });
    }

    async annotateNode(nodeName: string, annotationKey: string, annotationValue: string): Promise<void> {
        const patch = [
            {
                op: 'add',
                path: `/metadata/annotations/${annotationKey.replace(/\//g, '~1')}`,
                value: annotationValue
            }
        ];
        await this.patchNode(nodeName, patch)
            .then(() => {
                logger.info(`node: ${nodeName} annotated: ${annotationKey}/${annotationValue}`);
            })
            .catch((err) => {
                logger.info(`Error annotating node: ${nodeName}, err: ${err}`);
            });
    }

    async unLabelNode(nodeName: string, labelKey: string): Promise<void> {
        const patch = [
            {
                op: 'remove',
                path: `/metadata/labels/${labelKey.replace(/\//g, '~1')}`
            }
        ];
        await this.patchNode(nodeName, patch)
            .then(() => {
                logger.info(`node: ${nodeName} removed label: ${labelKey}`);
            })
            .catch((err) => {
                logger.info(`Error un-labeling node: ${nodeName}, err: ${err}`);
            });
    }

    async patchNamespace(namespaceName: string, patchBody: object): Promise<void> {
        await this.CoreV1Api.patchNamespace({
            name: namespaceName,
            body: patchBody
        });
    }

    async labelNamespace(namespaceName: string, labelKey: string, labelValue: string): Promise<void> {
        const patch = [
            {
                op: 'add',
                path: `/metadata/labels/${labelKey.replace(/\//g, '~1')}`,
                value: labelValue
            }
        ];
        await this.patchNamespace(namespaceName, patch)
            .then(() => {
                logger.info(`namespace: ${namespaceName} labeled: ${labelKey}/${labelValue}`);
            })
            .catch((err) => {
                logger.info(`Error labeling namespace: ${namespaceName}, err: ${err}`);
            });
    }

    async getDeployment(deploymentName: string, namespace: string): Promise<V1Deployment> {
        const res = await this.AppsV1Api.readNamespacedDeployment({ name: deploymentName, namespace });
        return res;
    }

    async getHPA(name: string, namespace: string): Promise<V2HorizontalPodAutoscaler> {
        const res = await this.AutoscalingV2Api.readNamespacedHorizontalPodAutoscaler({
            name,
            namespace
        });
        return res;
    }

    async patchDeployment(
        deploymentName: string,
        namespace: string,
        patchBody: object | Array<any>
    ): Promise<void> {
        let patch: object | Array<any> = patchBody;

        // Convert strategic merge patch to JSON Patch format if needed
        if (!Array.isArray(patchBody)) {
            // For now, we'll convert common deployment patches to JSON Patch format
            // This handles the replicas update case from the test
            const jsonPatch = [];

            if ('spec' in patchBody && typeof patchBody['spec'] === 'object') {
                const spec = patchBody['spec'];
                if ('replicas' in spec) {
                    jsonPatch.push({
                        op: 'replace',
                        path: '/spec/replicas',
                        value: spec['replicas']
                    });
                }

                // Add more conversions as needed for other fields
                if ('template' in spec && typeof spec['template'] === 'object') {
                    const template = spec['template'];
                    const templateMetadata = template['metadata'];
                    if ('metadata' in template && 'annotations' in templateMetadata) {
                        const annotations = template['metadata']['annotations'];
                        for (const [key, value] of Object.entries(annotations)) {
                            jsonPatch.push({
                                op: 'add',
                                path: `/spec/template/metadata/annotations/${key.replace(/\//g, '~1')}`,
                                value: value
                            });
                        }
                    }
                }
            }

            patch = jsonPatch;
        }

        await this.AppsV1Api.patchNamespacedDeployment({
            name: deploymentName,
            namespace,
            body: patch as any
        });
    }

    async getNodesResources(): Promise<NodeResources[]> {
        const nodes = (await this.CoreV1Api.listNode()).items;
        const nodesResources = nodes.map((node) => {
            const resources = node.status.capacity;
            const nodeResources: NodeResources = {
                name: node.metadata.name,
                cpu: parseInt(resources.cpu),
                memoryMB: Math.floor(parseInt(resources.memory) / 1024),
                gpu: parseInt(node.metadata.labels['nvidia.com/gpu.count'] || '0')
            };
            return nodeResources;
        });
        return nodesResources;
    }

    async setFlowRootStateResources(adminTokens: TokenObject): Promise<void> {
        flowRootState.nodesResources = await this.getNodesResources();
        flowRootState.numOfGpus = flowRootState.nodesResources.reduce((sum, node) => sum + node.gpu, 0);
        flowRootState.currentTenant = await tenantUtils.getTenant(adminTokens.idToken);
    }

    async restartDeployment(deploymentName: string, namespace: string): Promise<void> {
        const now = new Date();
        const patch = {
            spec: {
                template: {
                    metadata: {
                        annotations: {
                            'kubectl.kubernetes.io/restartedAt': now.toISOString()
                        }
                    }
                }
            }
        };
        await this.patchDeployment(deploymentName, namespace, patch)
            .then(() => {
                logger.info(`deployment: ${deploymentName}/${namespace} restarted`);
            })
            .catch((err) => {
                logger.info(`Error restarting deployment ${deploymentName}/${namespace}, err: ${err}`);
            });
    }

    async deleteDeployment(deploymentName: string, namespace: string): Promise<void> {
        await this.AppsV1Api.deleteNamespacedDeployment({ name: deploymentName, namespace })
            .then(() => {
                logger.info(`deployment: ${deploymentName}/${namespace} deleted`);
            })
            .catch((err) => {
                logger.info(`Error deleting deployment ${deploymentName}/${namespace}, err: ${err}`);
            });
    }

    async deletePod(podName: string, namespace: string): Promise<void> {
        await this.CoreV1Api.deleteNamespacedPod({ name: podName, namespace });
        logger.info(`pod: ${podName}/${namespace} deleted`);
    }

    async killOrgUnitServicePod(): Promise<void> {
        const namespace = 'runai-backend';
        const podPrefix = `${namespace}-org-unit-service`;

        const pods = await this.CoreV1Api.listNamespacedPod({ namespace });
        const pod = pods.items.find((p) => p.metadata?.name?.startsWith(podPrefix));

        if (!pod?.metadata?.name) {
            throw new Error(`No pod found with prefix "${podPrefix}" in namespace "${namespace}"`);
        }

        await this.deletePod(pod.metadata.name, namespace);
    }

    async waitForOrgUnitPodRecreationAndReadiness(): Promise<void> {
        const namespace = 'runai-backend';
        const podPrefix = `${namespace}-org-unit-service`;
        const timeoutMs = 60000;
        const pollIntervalMs = 1000;

        const start = Date.now();

        while (Date.now() - start < timeoutMs) {
            const pods = await this.CoreV1Api.listNamespacedPod({ namespace });
            const pod = pods.items.find((p) => p.metadata?.name?.startsWith(podPrefix));

            if (pod?.status?.phase === 'Running') {
                const isReady = pod.status.conditions?.some((c) => c.type === 'Ready' && c.status === 'True');

                if (isReady) {
                    logger.info(`Org-unit pod ${pod.metadata.name} is Running and Ready`);
                    return;
                }
            }

            await new Promise((res) => setTimeout(res, pollIntervalMs));
        }

        throw new Error(
            `Org-unit service pod was not recreated and ready within ${timeoutMs / 1000} seconds`
        );
    }

    async waitForOrgUnitServiceFullReadiness(timeoutSeconds = 180): Promise<void> {
        const namespace = 'runai-backend';
        const podPrefix = `${namespace}-org-unit-service`;

        await waitUtils.waitForCondition(
            async () => {
                try {
                    const pods = await this.CoreV1Api.listNamespacedPod({ namespace });
                    const pod = pods.items.find((p) => p.metadata?.name?.startsWith(podPrefix));

                    if (!pod) {
                        logger.warn('org-unit-service pod not found yet');
                        return false;
                    }

                    if (pod.status?.phase !== 'Running') {
                        logger.warn(
                            `org-unit-service pod phase is ${pod.status?.phase}, waiting for Running`
                        );
                        return false;
                    }

                    // Check all init containers have completed
                    for (const initContainer of pod.status?.initContainerStatuses || []) {
                        if (!initContainer.state?.terminated) {
                            logger.warn(`Init container ${initContainer.name} not yet terminated`);
                            return false;
                        }
                    }

                    // Check all containers are running (not waiting/PodInitializing)
                    for (const container of pod.status?.containerStatuses || []) {
                        if (container.state?.waiting) {
                            logger.warn(
                                `Container ${container.name} is waiting: ${container.state.waiting.reason}`
                            );
                            return false;
                        }
                        if (!container.state?.running || !container.ready) {
                            logger.warn(`Container ${container.name} is not running or not ready`);
                            return false;
                        }
                    }

                    const isReady = pod.status?.conditions?.some(
                        (c) => c.type === 'Ready' && c.status === 'True'
                    );
                    if (!isReady) {
                        logger.warn('org-unit-service pod Ready condition is not True');
                        return false;
                    }

                    logger.info(
                        `org-unit-service pod ${pod.metadata?.name} is fully running with all containers started`
                    );
                    return true;
                } catch (error) {
                    logger.warn(`Error checking org-unit-service pod status: ${error.message}`);
                    return false;
                }
            },
            'Wait for org-unit-service pod to be fully running with all containers started',
            true,
            timeoutSeconds,
            2
        );
    }

    async getConfigMap(configMapName: string, namespace: string): Promise<V1ConfigMap> {
        const res = await this.CoreV1Api.readNamespacedConfigMap({ name: configMapName, namespace });
        return res;
    }

    async updateConfigMap(configMapName: string, namespace: string, body: V1ConfigMap): Promise<void> {
        await this.CoreV1Api.replaceNamespacedConfigMap({ name: configMapName, namespace, body });
    }

    async createDeployment(deployment: V1Deployment): Promise<KubernetesObject> {
        return await this.AppsV1Api.createNamespacedDeployment({
            namespace: deployment.metadata.namespace,
            body: deployment
        });
    }

    async createFromYaml(yamlAsString: string): Promise<KubernetesObject> {
        const spec: k8s.KubernetesObject = yaml.load(yamlAsString);
        return await this.KubernetesObjectApi.create(spec);
    }

    async createAllFromYaml(yamlAsString: string): Promise<void> {
        const specs: k8s.KubernetesObject[] = yaml.loadAll(yamlAsString);
        for (const spec of specs) {
            if (spec) {
                try {
                    await this.KubernetesObjectApi.patch(spec);
                } catch {
                    await this.KubernetesObjectApi.create(spec);
                }
            }
        }
    }

    async createCrd(
        crd: CustomResourceDefinition,
        body: object,
        namespace?: string
    ): Promise<KubernetesObject> {
        try {
            if (namespace) {
                return await this.CustomObjectsApi.createNamespacedCustomObject({
                    group: crd.group,
                    version: crd.version,
                    namespace,
                    plural: crd.plural,
                    body
                });
            } else {
                return await this.CustomObjectsApi.createClusterCustomObject({
                    group: crd.group,
                    version: crd.version,
                    plural: crd.plural,
                    body
                });
            }
        } catch (err) {
            throw new CliError(
                new Error(`Could not create crd ${crd.plural}, reason: ${JSON.stringify(err)}`)
            );
        }
    }

    async patchCrd(
        crd: CustomResourceDefinition,
        patchBody: object,
        name?: string,
        namespace?: string
    ): Promise<any> {
        const maxAttempts = 6;
        let attempt = 0;
        let lastError: any;

        const isConflictError = (err: any): boolean => {
            try {
                if (err?.response?.status === 409) return true;
                if (err?.statusCode === 409) return true;
                if (typeof err?.body === 'object' && err?.body?.code === 409) return true;
                if (typeof err?.message === 'string') {
                    const msg = err.message.toLowerCase();
                    if (
                        msg.includes('409') ||
                        msg.includes('conflict') ||
                        msg.includes('object has been modified')
                    ) {
                        return true;
                    }
                }
            } catch {
                // ignore parsing issues
            }
            return false;
        };

        while (attempt < maxAttempts) {
            try {
                // Always fetch the latest object to get up-to-date resourceVersion
                const currentObject = namespace
                    ? await this.CustomObjectsApi.getNamespacedCustomObject({
                          group: crd.group,
                          version: crd.version,
                          namespace,
                          plural: crd.plural,
                          name
                      })
                    : await this.CustomObjectsApi.getClusterCustomObject({
                          group: crd.group,
                          version: crd.version,
                          plural: crd.plural,
                          name
                      });

                // Deep-merge the incoming patch into the latest object to avoid clobbering concurrent updates
                const mergedBody = this.deepMergeObjects(currentObject, patchBody);

                // Ensure resourceVersion matches the latest
                if (currentObject && currentObject['metadata']) {
                    mergedBody['metadata'] = {
                        ...mergedBody['metadata'],
                        resourceVersion: currentObject['metadata']['resourceVersion']
                    };
                }

                // Replace with the merged body
                if (namespace) {
                    return await this.CustomObjectsApi.replaceNamespacedCustomObject({
                        group: crd.group,
                        version: crd.version,
                        namespace,
                        plural: crd.plural,
                        name,
                        body: mergedBody
                    });
                } else {
                    return await this.CustomObjectsApi.replaceClusterCustomObject({
                        group: crd.group,
                        version: crd.version,
                        plural: crd.plural,
                        name,
                        body: mergedBody
                    });
                }
            } catch (err) {
                lastError = err;
                if (isConflictError(err) && attempt < maxAttempts - 1) {
                    const backoffMs = 200 * Math.pow(2, attempt);
                    logger.warn(
                        `Conflict patching ${crd.plural}/${name} (attempt ${attempt + 1}/${maxAttempts}), retrying in ${backoffMs}ms`
                    );
                    await new Promise((res) => setTimeout(res, backoffMs));
                    attempt++;
                    continue;
                }

                logger.error(
                    `Could not patch crd ${crd.plural} with name ${name}, reason: ${JSON.stringify(err)}`
                );
                throw new Error(
                    `Could not patch crd ${crd.plural} with name ${name}, reason: ${JSON.stringify(err)}`
                );
            }
        }

        throw new Error(
            `Could not patch crd ${crd.plural} with name ${name}, reason: ${JSON.stringify(lastError)}`
        );
    }

    async deleteCrd(
        crd: CustomResourceDefinition,
        name: string,
        namespace?: string
    ): Promise<KubernetesObject> {
        try {
            if (namespace) {
                return await this.CustomObjectsApi.deleteNamespacedCustomObject({
                    group: crd.group,
                    version: crd.version,
                    namespace,
                    plural: crd.plural,
                    name
                });
            } else {
                return await this.CustomObjectsApi.deleteClusterCustomObject({
                    group: crd.group,
                    version: crd.version,
                    plural: crd.plural,
                    name
                });
            }
        } catch (err) {
            throw new Error(
                `Could not delete crd ${crd.plural} with name ${name}, reason: ${JSON.stringify(err)}`
            );
        }
    }

    async getCrd<T extends KubernetesObject>(
        crd: CustomResourceDefinition,
        name: string,
        namespace?: string
    ): Promise<T> {
        try {
            return await Utils.retry(async () => {
                if (namespace) {
                    return await this.CustomObjectsApi.getNamespacedCustomObject({
                        group: crd.group,
                        version: crd.version,
                        namespace,
                        plural: crd.plural,
                        name
                    });
                } else {
                    return await this.CustomObjectsApi.getClusterCustomObject({
                        group: crd.group,
                        version: crd.version,
                        plural: crd.plural,
                        name
                    });
                }
            });
        } catch (err) {
            throw new Error(
                `Could not get crd ${crd.plural} with name ${name}, reason: ${JSON.stringify(err)}`
            );
        }
    }

    /**
     * Whether the CRD itself (not any of its objects) is installed in the cluster.
     * Used to pick between two groups serving the same kind across a migration.
     */
    async isCrdInstalled(crd: CustomResourceDefinition): Promise<boolean> {
        const name = `${crd.plural}.${crd.group}`;
        try {
            await this.ApiextensionsV1Api.readCustomResourceDefinition({ name });
            return true;
        } catch (err) {
            logger.debug(`CRD ${name} is not installed, reason: ${JSON.stringify(err)}`);
            return false;
        }
    }

    async getCrds(crd: CustomResourceDefinition, namespace?: string): Promise<KubernetesObject[]> {
        try {
            return await Utils.retry(async () => {
                if (namespace) {
                    return (
                        await this.CustomObjectsApi.listNamespacedCustomObject({
                            group: crd.group,
                            version: crd.version,
                            namespace,
                            plural: crd.plural
                        })
                    )['items'];
                } else {
                    return (
                        await this.CustomObjectsApi.listClusterCustomObject({
                            group: crd.group,
                            version: crd.version,
                            plural: crd.plural
                        })
                    )['items'];
                }
            });
        } catch (err) {
            throw new Error(`Could not get crds ${crd.plural}, reason: ${JSON.stringify(err)}`);
        }
    }

    async getNumOfGpus(): Promise<number> {
        logger.debug(`using context ${this.getCurrentContext()}`);
        return (await this.CoreV1Api.listNode()).items.reduce((numOfGpus, node) => {
            logger.debug(
                `node ${node.metadata.name} has ${Number(
                    node.metadata.labels['nvidia.com/gpu.count'] || '0'
                )} gpus`
            );
            return numOfGpus + Number(node.metadata.labels['nvidia.com/gpu.count'] || '0');
        }, 0);
    }

    async getNumOfGpusOnNode(nodeName: string): Promise<number> {
        logger.debug(`using context ${this.getCurrentContext()}`);
        const nodeLabels = await this.getNodeLabels(nodeName);
        return Number(nodeLabels['nvidia.com/gpu.count'] || '0');
    }

    async getNumOfGpusOnEachNode(): Promise<{ [nodeName: string]: number }> {
        logger.debug(`using context ${this.getCurrentContext()}`);
        const nodes = await this.getNodes();

        const result = {};
        nodes.items.forEach((node) => {
            const nodeName = node.metadata.name;
            result[nodeName] = Number(node.metadata.labels['nvidia.com/gpu.count'] || '0');
        });
        return result;
    }

    // getNumOfGpusOnEachGpuNode - returning only node with GPU!
    // so all keys are > 0.
    // iof you want all the nodes - use "getNumOfGpusOnEachNode" func.
    async getNumOfGpusOnEachGpuNode(): Promise<{ [nodeName: string]: number }> {
        const numOfGpusOnEachGpuNode = {};
        const numOfGpusOnAllNodes = await kubeCli.getNumOfGpusOnEachNode();
        for (const [nodeName, numOfGpus] of Object.entries(numOfGpusOnAllNodes)) {
            if (numOfGpus > 0) {
                numOfGpusOnEachGpuNode[nodeName] = numOfGpus;
            }
        }
        return numOfGpusOnEachGpuNode;
    }

    private deepMergeObjects(target: any, source: any): any {
        if (!target || typeof target !== 'object') {
            return source;
        }
        if (!source || typeof source !== 'object') {
            return target;
        }
        const output: any = Array.isArray(target) ? [...target] : { ...target };
        Object.keys(source).forEach((key) => {
            const sourceValue = source[key];
            const targetValue = (target as any)[key];
            if (sourceValue && typeof sourceValue === 'object' && !Array.isArray(sourceValue)) {
                output[key] = this.deepMergeObjects(targetValue || {}, sourceValue);
            } else {
                output[key] = sourceValue;
            }
        });
        return output;
    }

    async getNumOfNodesWithGpus(): Promise<number> {
        logger.debug(`using context ${this.getCurrentContext()}`);
        return (await this.CoreV1Api.listNode()).items.reduce((numOfNodesWithGpus, node) => {
            const numOfGpus = Number(node.metadata.labels['nvidia.com/gpu.count'] || '0');
            if (numOfGpus > 0) {
                return numOfNodesWithGpus + 1;
            }
            return numOfNodesWithGpus;
        }, 0);
    }

    async updateWorkload(workload: BaseWorkload): Promise<KubernetesObject> {
        return await this.patchCrd(
            workload.getCrd(),
            workload.getBody(),
            workload.name,
            workload.testingSpecificFields.namespace
        );
    }

    async createWorkload(workload: BaseWorkload): Promise<KubernetesObject> {
        return await this.createCrd(
            workload.getCrd(),
            workload.getBody(),
            workload.testingSpecificFields.namespace
        );
    }

    async deleteWorkload(workload: BaseWorkload): Promise<KubernetesObject> {
        return await this.deleteCrd(
            workload.getCrd(),
            workload.name,
            workload.testingSpecificFields.namespace
        );
    }

    async deleteTrainingWorkloadOfJob(jobName: string, namespace: string): Promise<void> {
        const workload = await this.getTrainWorkloadOfRunaiJob(jobName, namespace);
        await this.deleteCrd(crds.runaiTrainingWorkloadCrd, workload.metadata.name, namespace);
    }

    async getWorkload(
        jobName: string,
        namespace: string,
        crd: CustomResourceDefinition
    ): Promise<KubernetesObject> {
        const workloadObjs = await this.getCrds(crd, namespace);
        return workloadObjs.find((workload) => workload.metadata.name.startsWith(jobName));
    }

    async getTrainWorkloadOfRunaiJob(jobName: string, namespace: string): Promise<KubernetesObject> {
        return this.getWorkload(jobName, namespace, crds.runaiTrainingWorkloadCrd);
    }

    async getInteractiveWorkload(jobName: string, namespace: string): Promise<KubernetesObject> {
        return this.getWorkload(jobName, namespace, crds.runaiInteractiveWorkloadCrd);
    }

    async getInferenceWorkload(jobName: string, namespace: string): Promise<KubernetesObject> {
        return this.getWorkload(jobName, namespace, crds.runaiInferenceWorkloadCrd);
    }

    async getDistributedInferenceWorkload(jobName: string, namespace: string): Promise<KubernetesObject> {
        return this.getWorkload(jobName, namespace, crds.runaiDistributedInferenceWorkloadCrd);
    }

    exportConfig(): string {
        return this.kubeClient.exportConfig();
    }

    async listValidatingWebhooks(): Promise<V1ValidatingWebhookConfigurationList> {
        const res = await this.AdmissionregistrationV1Api.listValidatingWebhookConfiguration();
        return res;
    }

    async patchPvc(pvcName: string, body: object, namespace: string): Promise<void> {
        await this.CoreV1Api.patchNamespacedPersistentVolumeClaim({
            name: pvcName,
            namespace,
            body
        });
        logger.info(`Patched pvc ${pvcName} in namespace ${namespace}`);
    }

    async patchSecret(secretName: string, body: object, namespace: string): Promise<void> {
        await this.CoreV1Api.patchNamespacedSecret({
            name: secretName,
            namespace,
            body
        });
        logger.info(`Patched secret ${secretName} in namespace ${namespace}`);
    }

    async deleteIngress(ingressName: string, namespace: string): Promise<void> {
        try {
            await this.networkingV1Api.deleteNamespacedIngress({ name: ingressName, namespace });
            logger.info(`Deleted ingress ${ingressName} in namespace ${namespace}`);
        } catch (err) {
            logger.error(`Failed to delete ingress ${ingressName} in namespace ${namespace}`);
            throw err;
        }
    }

    async getSecret(secretName: string, namespace: string): Promise<V1Secret> {
        const res = await this.CoreV1Api.readNamespacedSecret({ name: secretName, namespace });
        return res;
    }

    async createSecret(
        secretName: string,
        namespace: string,
        data: { [key: string]: string },
        type = 'Opaque',
        labels?: { [key: string]: string }
    ): Promise<V1Secret> {
        const secret: V1Secret = {
            metadata: {
                name: secretName,
                namespace: namespace,
                ...(labels && { labels })
            },
            type: type,
            data: data // data is already base64 encoded
        };

        const res = await this.CoreV1Api.createNamespacedSecret({ namespace, body: secret });
        return res;
    }

    async deleteSecret(secretName: string, namespace: string): Promise<V1Secret> {
        const res = await this.CoreV1Api.deleteNamespacedSecret({ name: secretName, namespace });
        return res;
    }

    async deletePvc(name: string, namespace: string): Promise<V1PersistentVolumeClaim> {
        const pvc = await this.CoreV1Api.deleteNamespacedPersistentVolumeClaim({ name, namespace });
        return pvc;
    }

    async deleteConfigMap(meta: V1ObjectMeta): Promise<V1ConfigMap> {
        const configMap = await this.CoreV1Api.deleteNamespacedConfigMap({
            name: meta.name,
            namespace: meta.namespace
        });
        return configMap;
    }

    async listSecretsByNamepsace(namespace: string): Promise<V1SecretList> {
        return await this.CoreV1Api.listNamespacedSecret({ namespace });
    }

    async getServiceAccount(serviceAccountName: string, namespace: string): Promise<V1ServiceAccount> {
        const res = await this.CoreV1Api.readNamespacedServiceAccount({
            name: serviceAccountName,
            namespace
        });
        return res;
    }

    create_backend_tlsSecret(): k8s.V1Secret {
        try {
            logger.debug('Creating secret...');
            const encodedCertificate = settings.tls_cert;
            const encodedPrivateKey = settings.tls_key;
            const secret: V1Secret = {
                metadata: {
                    name: 'runai-backend-tls',
                    namespace: 'runai-backend'
                },
                type: 'kubernetes.io/tls',
                data: {
                    'tls.crt': encodedCertificate,
                    'tls.key': encodedPrivateKey
                }
            };
            return secret;
        } catch (err) {
            throw new Error(`Could not create secret, reason: ${JSON.stringify(err)}`);
        }
    }

    async statefulSetDelete(name: string, namespace: string): Promise<void> {
        await this.AppsV1Api.deleteNamespacedStatefulSet({ name, namespace });
    }

    async deleteTfJob(name: string, namespace: string): Promise<void> {
        logger.info(`Deleting TFJob ${name} in namespace ${namespace}`);
        await this.deleteCrd(crds.tfJobCrd, name, namespace);
    }

    async deleteXGBoostJob(name: string, namespace: string): Promise<void> {
        logger.info(`Deleting XGBoostJob ${name} in namespace ${namespace}`);
        await this.deleteCrd(crds.xgboostJobCrd, name, namespace);
    }

    async deletePyTorchJob(name: string, namespace: string): Promise<void> {
        logger.info(`Deleting PyTorchJob ${name} in namespace ${namespace}`);
        await this.deleteCrd(crds.pyTorchJobCrd, name, namespace);
    }

    async deleteRayCluster(name: string, namespace: string): Promise<void> {
        logger.info(`Deleting RayCluster ${name} in namespace ${namespace}`);
        await this.deleteCrd(crds.rayClusterCrd, name, namespace);
    }

    async createOAuthClient(
        runaiEnvUrl: string,
        clientId: string,
        idpSecret: string
    ): Promise<{ clientId: string; idpSecret: string }> {
        const kc = new k8s.KubeConfig();
        kc.loadFromDefault();
        const k8sApi = kc.makeApiClient(k8s.CustomObjectsApi);

        const oauthClient = {
            kind: 'OAuthClient',
            metadata: {
                name: clientId
            },
            grantMethod: 'auto',
            redirectURIs: [`${runaiEnvUrl}`],
            secret: idpSecret
        };

        try {
            await k8sApi.createClusterCustomObject({
                group: 'oauth.openshift.io',
                version: 'v1',
                plural: 'oauthclients',
                body: oauthClient
            });
            logger.debug('OAuthClient is created');
        } catch (error) {
            logger.error(
                `Failed to verify OAuthClient, error: ${error.response ? JSON.stringify(error.response.body) : error.message}`
            );

            throw new Error(`Failed to create OAuthClient, error: ${error}`);
        }

        try {
            const result = await k8sApi.getClusterCustomObject({
                group: 'oauth.openshift.io',
                version: 'v1',
                plural: 'oauthclients',
                name: clientId
            });
            if (result) {
                logger.debug('OAuthClient creation verified');
            } else {
                throw new Error('OAuthClient verification failed');
            }
        } catch (error) {
            throw new Error(`Failed to verify OAuthClient, error: ${error}`);
        }

        return { clientId: clientId, idpSecret: idpSecret };
    }

    async checkPodRunningTime(
        podName: string,
        namespace = 'runai',
        runningTimeSec = 60,
        phase = 'Running'
    ): Promise<boolean> {
        const kc = new k8s.KubeConfig();
        kc.loadFromDefault();

        const k8sApi = kc.makeApiClient(k8s.CoreV1Api);

        try {
            const res = await k8sApi.readNamespacedPod({ name: podName, namespace });
            const pod = res;

            if (pod.status && pod.status.startTime && pod.status.phase === phase) {
                const startTime = new Date(pod.status.startTime.toString());
                const currentTime = new Date();

                const diffInMilliseconds = currentTime.getTime() - startTime.getTime();
                const diffInSec = diffInMilliseconds / 1000;

                if (diffInSec > runningTimeSec) {
                    return true;
                } else {
                    return false;
                }
            } else {
                console.log(`Could not determine start time for pod ${podName}.`);
                return false;
            }
        } catch (err) {
            console.error(`Failed to get pod ${podName}: ${err}`);
            return false;
        }
    }

    async getPodYaml(podName: string, namespace: string): Promise<object> {
        try {
            // Get the pod object
            const response = await this.CoreV1Api.readNamespacedPod({ name: podName, namespace });

            // Clean up the pod object by removing unnecessary fields
            const pod = response;
            const cleanPod = {
                apiVersion: 'v1',
                kind: 'Pod',
                metadata: {
                    name: pod.metadata.name,
                    namespace: pod.metadata.namespace,
                    labels: pod.metadata.labels,
                    annotations: pod.metadata.annotations
                },
                spec: pod.spec
            };

            // Convert to YAML
            return cleanPod;
        } catch (err) {
            throw new Error(
                `Could not get YAML for pod ${podName} in namespace ${namespace}, reason: ${JSON.stringify(err)}`
            );
        }
    }

    async getCrdYaml(crd: CustomResourceDefinition, name: string, namespace: string): Promise<object> {
        try {
            const response = await this.CustomObjectsApi.getNamespacedCustomObject({
                group: crd.group,
                version: crd.version,
                namespace,
                plural: crd.plural,
                name
            });
            const resource = response;
            const cleanResource: any = {
                apiVersion: `${crd.group}/${crd.version}`,
                kind: crd.kind,
                metadata: {}
            };
            if (resource['metadata']) {
                cleanResource.metadata.name = resource['metadata']['name'];
                cleanResource.metadata.namespace = resource['metadata']['namespace'];

                if (resource['metadata']['labels']) {
                    cleanResource.metadata.labels = resource['metadata']['labels'];
                }

                if (resource['metadata']['annotations']) {
                    cleanResource.metadata.annotations = resource['metadata']['annotations'];
                }
            }
            if (resource['spec']) {
                cleanResource.spec = resource['spec'];
            }
            if (resource['status']) {
                cleanResource.status = resource['status'];
            }

            return cleanResource;
        } catch (err) {
            throw new Error(
                `Could not get YAML for ${crd.kind} ${name} in namespace ${namespace}, reason: ${JSON.stringify(err)}`
            );
        }
    }

    async getCrdWithIssues(crd: CustomResourceDefinition, name: string, namespace: string): Promise<object> {
        const resource = await this.getCrdYaml(crd, name, namespace);
        if (!resource['status']?.['issues']) {
            try {
                const kubectlCmd = `kubectl get dw -n ${namespace} ${name} -o json`;
                const kubectlResult = bashCli.exec(kubectlCmd);
                const kubectlResource = JSON.parse(kubectlResult);

                if (kubectlResource?.status?.issues) {
                    if (!resource['status']) {
                        resource['status'] = {};
                    }
                    resource['status']['issues'] = kubectlResource.status.issues;
                }
            } catch (err) {
                console.warn(`Failed to get issues using kubectl: ${err.message}`);
            }
        }
        return resource;
    }

    async describeCrd(crd: CustomResourceDefinition): Promise<string> {
        try {
            let describeResults = `=== Describe All ${crd.kind}s ===\n\n`;

            // Get all CRDs across all namespaces using kubeClient
            const crdList = await this.getCrds(crd);

            if (!crdList || crdList.length === 0) {
                describeResults += `No ${crd.kind}s found in any namespace.\n`;
                return describeResults;
            }

            for (const item of crdList) {
                const name = item.metadata?.name;
                const namespace = item.metadata?.namespace;

                if (!name) {
                    continue;
                }

                describeResults += `\n${'='.repeat(80)}\n`;
                describeResults += `Describing ${crd.kind}: ${name} (namespace: ${namespace || 'default'})\n`;
                describeResults += `${'='.repeat(80)}\n\n`;

                try {
                    // Use kubectl describe command for detailed output
                    const describeCmd = `kubectl describe ${crd.plural} ${name} -n ${namespace || 'default'}`;
                    const describeOutput = bashCli.exec(describeCmd);
                    describeResults += describeOutput;
                } catch (describeErr) {
                    describeResults += `Error describing ${name}: ${describeErr}\n`;
                }

                describeResults += '\n\n';
            }

            return describeResults;
        } catch (err) {
            throw new Error(`Could not describe CRDs of type ${crd.plural}, reason: ${JSON.stringify(err)}`);
        }
    }

    async annotatePod(podName: string, namespace: string, annotation: string): Promise<void> {
        try {
            // Parse the annotation string to get key and value
            const [key, value] = annotation.split('=');

            // Use JSON Patch format (array of operations)
            const patchBody = [
                {
                    op: 'add',
                    path: `/metadata/annotations/${key.replace(/\//g, '~1')}`, // escape forward slashes per JSON Patch RFC
                    value: value
                }
            ];

            await this.CoreV1Api.patchNamespacedPod({
                name: podName,
                namespace,
                body: patchBody
            });

            logger.info(`Successfully annotated pod ${podName} in namespace ${namespace} with ${annotation}`);
        } catch (err) {
            logger.error(`Error annotating pod ${podName} in namespace ${namespace}: ${err}`);
            throw err;
        }
    }

    listEventsByNamespace(namespace: string): string {
        try {
            const eventsCmd = `kubectl get events -n ${namespace}`;
            return bashCli.exec(eventsCmd);
        } catch (err) {
            logger.debug(`Error getting events in namespace ${namespace}: ${err}`);
        }
    }

    async waitForCrdToBeDeleted(
        crd: CustomResourceDefinition,
        name: string,
        namespace?: string,
        timeout = 40
    ): Promise<void> {
        const doesCrdExist = async (): Promise<boolean> => {
            try {
                await this.getCrd(crd, name, namespace);
                return true;
            } catch {
                return false;
            }
        };

        await waitUtils.waitForCondition(
            doesCrdExist,
            `waitForCrdToBeDeleted ${crd.plural} name: ${name}`,
            false,
            timeout
        );
    }

    portForwardPostgres(
        namespace: string = 'runai-backend',
        localPort: number = 5432,
        remotePort: number = 5432
    ): child_process.ChildProcess {
        const podName = 'runai-backend-postgresql-0';
        const command = `kubectl port-forward -n ${namespace} pod/${podName} ${localPort}:${remotePort}`;
        logger.info(`Starting port-forward: ${command}`);

        // Use spawn for long-running processes like port-forward
        const child = spawn(
            'kubectl',
            ['port-forward', '-n', namespace, `pod/${podName}`, `${localPort}:${remotePort}`],
            {
                stdio: 'inherit',
                env: {
                    ...process.env,
                    KUBECONFIG: config.kubeconfigPath
                }
            }
        );

        child.on('error', (error) => {
            logger.error(`Port-forward error: ${error}`);
        });

        child.on('exit', (code) => {
            logger.info(`Port-forward exited with code: ${code}`);
        });

        return child;
    }

    async waitForCrdPhaseWithWatch(
        crd: CustomResourceDefinition,
        name: string,
        namespace: string,
        expectedPhase: string,
        timeoutSec = 60,
        invalidPhase?: string
    ): Promise<any> {
        return new Promise((resolve, reject) => {
            const kc = new k8s.KubeConfig();
            kc.loadFromDefault();
            const watch = new k8s.Watch(kc);
            const path = `/apis/${crd.group}/${crd.version}/namespaces/${namespace}/${crd.plural}`;

            // Determine the phase path: use custom getPhase() if available, otherwise default to 'status.phase'
            const phasePath = crd.getPhasePath ? crd.getPhasePath() : 'status.phase';
            const phaseFieldName = phasePath.split('.').pop(); // Extract 'phase' or 'state' for logging

            logger.info(
                `[waitForCrdPhaseWithWatch] Start API watch: path=${path}, expectedPhase=${expectedPhase}, phaseField=${phasePath}`
            );
            let req: any;
            const timeout: NodeJS.Timeout = setTimeout(() => {
                if (req) req.abort();
                logger.error(
                    `[waitForCrdPhaseWithWatch] Timeout waiting for CRD ${crd.kind} ${name} ${phaseFieldName} to be ${expectedPhase}`
                );
                reject(
                    new Error(
                        `Timeout waiting for CRD ${crd.kind} ${name} ${phaseFieldName} to be ${expectedPhase}`
                    )
                );
            }, timeoutSec * 1000);
            watch
                .watch(
                    path,
                    {},
                    (type, obj) => {
                        const objName = obj.metadata?.name;
                        const phase = get(obj, phasePath);

                        logger.debug(
                            `[waitForCrdPhaseWithWatch] event: type=${type}, name=${objName}, ${phaseFieldName}=${phase}`
                        );
                        if (objName === name) {
                            if (phase === expectedPhase) {
                                clearTimeout(timeout);
                                req.abort();
                                logger.info(
                                    `[waitForCrdPhaseWithWatch] Matched expected ${phaseFieldName}: ${phase}`
                                );
                                resolve(obj);
                            } else if (invalidPhase && phase === invalidPhase) {
                                clearTimeout(timeout);
                                req.abort();
                                logger.error(
                                    `[waitForCrdPhaseWithWatch] Invalid ${phaseFieldName} detected: ${phase}`
                                );
                                reject(new Error(`Invalid ${phaseFieldName} detected: ${phase}`));
                            } else {
                                logger.info(`[waitForCrdPhaseWithWatch] Current ${phaseFieldName}: ${phase}`);
                            }
                        }
                    },
                    (err) => {
                        clearTimeout(timeout);
                        if (err && err.message && err.message.includes('aborted')) {
                            logger.debug(
                                '[waitForCrdPhaseWithWatch] Watch aborted (expected after phase match)'
                            );
                        } else if (err) {
                            logger.error('[waitForCrdPhaseWithWatch] Watch error: ' + err);
                            reject(err);
                        }
                    }
                )
                .then((r) => {
                    req = r;
                })
                .catch((err) => {
                    clearTimeout(timeout);
                    logger.error(`[waitForCrdPhaseWithWatch] Watch setup error: ${err}`);
                    reject(err);
                });
        });
    }

    setWorkloadActive(workloadName: string, namespace: string, crdKind: string, active: boolean): void {
        // prettier-ignore
        const cmd = `kubectl patch ${crdKind} ${workloadName} -n ${namespace} --type json -p '[{"op": "replace", "path": "/spec/active/value", "value": ${active}}]'`;

        try {
            bashCli.exec(cmd);
        } catch (err) {
            logger.error(`Failed to set workload ${workloadName} active to ${active}: ${err}`);
            throw err;
        }
    }

    async patchRunaiConfig(patch: Array<any>): Promise<boolean> {
        const request: k8s.CustomObjectsApiPatchNamespacedCustomObjectRequest = {
            group: 'run.ai',
            version: 'v1',
            namespace: 'runai',
            plural: 'runaiconfigs',
            name: 'runai',
            body: patch
        };
        try {
            await this.CustomObjectsApi.patchNamespacedCustomObject(request);
            logger.info('runaiconfig updated successfully');
            return true;
        } catch (error) {
            logger.error('Error updating runaiconfig:', error);
            return false;
        }
    }
}

export { KubeCli };
