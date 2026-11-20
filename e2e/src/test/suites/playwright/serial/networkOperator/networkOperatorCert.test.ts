/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { expect } from 'chai';
import {
    config,
    logger,
    objectGenerator,
    operations,
    projectResourcesUtils,
    tags,
    TrainingWorkload,
    Utils,
    waitUtils
} from '../../../../../modules/helpers';
import { kubeCli } from '../../../../../modules/cli';
import { Project } from '../../../../../modules/api/apis/projectsDepartmentsApi';
import { test } from '@playwright/test';
import { shouldRunNativeWorkloadCrdTests } from '../../../../../modules/helpers/utils/SupporetedEnvsUtils';

test.describe(`Network Operator Certification ${tags.teams.pegasus} ${tags.runOn.networkCert}`, () => {
    const NETWORK_OPERATOR_NAMESPACE = 'nvidia-network-operator';
    const OPENSHIFT_NFD_NAMESPACE = 'openshift-nfd';
    const OPENSHIFT_MULTUS_NAMESPACE = 'openshift-multus';
    const isOpenShift = process.env.CLUSTER_PLATFORM === 'openshift';
    const MACVLAN_NETWORK_NAME = 'macvlan-net';
    const MACVLAN_MASTER_FALLBACK = 'ens5';
    const SECONDARY_NETWORK_ANNOTATION = 'k8s.v1.cni.cncf.io/networks';
    const NETWORK_STATUS_ANNOTATION = 'k8s.v1.cni.cncf.io/network-status';
    const nicClusterPolicy = {
        apiVersion: 'mellanox.com/v1alpha1',
        kind: 'NicClusterPolicy',
        metadata: { name: 'nic-cluster-policy' },
        spec: {
            rdmaSharedDevicePlugin: {
                image: 'k8s-rdma-shared-dev-plugin',
                repository: 'ghcr.io/mellanox',
                version: 'latest',
                config: JSON.stringify({
                    configList: [
                        {
                            resourceName: 'rdma_shared_device_a',
                            rdmaHcaMax: 63,
                            selectors: { ifNames: ['eth0'] }
                        }
                    ]
                })
            },
            ...(isOpenShift
                ? {}
                : {
                      secondaryNetwork: {
                          multus: {
                              image: 'multus-cni',
                              repository: 'ghcr.io/k8snetworkplumbingwg',
                              version: 'v4.1.0'
                          },
                          cniPlugins: {
                              image: 'plugins',
                              repository: 'ghcr.io/k8snetworkplumbingwg',
                              version: 'v1.5.0'
                          }
                      }
                  })
        }
    };
    let projectName: string;
    let namespace: string;
    let project: Project;
    let jobName: string;
    let workload: TrainingWorkload;

    const detectMasterInterface = async (): Promise<string> => {
        try {
            const nodeIp = (await kubeCli.CoreV1Api.listNode()).items[0]?.status?.addresses?.find(
                (address) => address.type === 'InternalIP'
            )?.address;
            if (!nodeIp) {
                return MACVLAN_MASTER_FALLBACK;
            }
            const hostPods = (await kubeCli.CoreV1Api.listPodForAllNamespaces()).items.filter(
                (pod) => pod.spec?.hostNetwork && pod.status?.phase === 'Running'
            );
            const candidates = [
                ...hostPods.filter((pod) => /flannel|multus|cni|kube-proxy|ovn/i.test(pod.metadata.name)),
                ...hostPods
            ];
            for (const pod of candidates) {
                try {
                    const { stdout } = await kubeCli.execInPod(pod.metadata.name, pod.metadata.namespace, [
                        'ip',
                        '-o',
                        '-4',
                        'addr',
                        'show'
                    ]);
                    const iface = stdout
                        .split('\n')
                        .find((line) => line.includes(`${nodeIp}/`))
                        ?.trim()
                        .split(/\s+/)[1];
                    if (iface) {
                        return iface;
                    }
                } catch (error) {
                    logger.debug(`interface probe failed on ${pod.metadata.name}: ${error.message}`);
                }
            }
            return MACVLAN_MASTER_FALLBACK;
        } catch (error) {
            logger.debug(
                `master interface auto-detect failed, using ${MACVLAN_MASTER_FALLBACK}: ${error.message}`
            );
            return MACVLAN_MASTER_FALLBACK;
        }
    };

    test.beforeAll(async () => {
        projectName = objectGenerator.generateNameByPrefix('netcert-project');
        namespace = Utils.namespaceFromProjectName(projectName);
        project = await projectResourcesUtils.createProject(namespace, {
            name: projectName,
            deservedGpus: 0,
            maxAllowedGpus: 0
        });
        try {
            await kubeCli.createFromYaml(JSON.stringify(nicClusterPolicy));
        } catch (error) {
            logger.debug(`NicClusterPolicy may already exist on the cluster: ${error.message}`);
        }
        const masterInterface = await detectMasterInterface();
        logger.info(`macvlan master interface resolved to: ${masterInterface}`);
        const macvlanNetwork = {
            apiVersion: 'mellanox.com/v1alpha1',
            kind: 'MacvlanNetwork',
            metadata: { name: MACVLAN_NETWORK_NAME },
            spec: {
                networkNamespace: namespace,
                master: masterInterface,
                mode: 'bridge',
                ipam: JSON.stringify({
                    type: 'host-local',
                    subnet: '192.168.2.0/24',
                    rangeStart: '192.168.2.10',
                    rangeEnd: '192.168.2.100'
                })
            }
        };
        try {
            await kubeCli.createFromYaml(JSON.stringify(macvlanNetwork));
        } catch (error) {
            logger.debug(`MacvlanNetwork may already exist on the cluster: ${error.message}`);
        }
    });

    test('network operator controller is available', async () => {
        const pods = (await kubeCli.CoreV1Api.listNamespacedPod({ namespace: NETWORK_OPERATOR_NAMESPACE }))
            .items;
        const controller = pods.find((pod) => pod.metadata.name.startsWith('network-operator'));
        expect(controller, 'network-operator controller pod should exist').to.not.be.undefined;
        expect(controller.status.phase, 'network-operator controller should be Running').to.equal('Running');
    });

    test('network operator node-feature-discovery is running', async () => {
        const pods = (await kubeCli.CoreV1Api.listPodForAllNamespaces()).items.filter((pod) =>
            [NETWORK_OPERATOR_NAMESPACE, OPENSHIFT_NFD_NAMESPACE].includes(pod.metadata.namespace)
        );
        const nfd = pods.filter((pod) => /node-feature-discovery|^nfd-/.test(pod.metadata.name));
        expect(
            nfd.length,
            'node-feature-discovery should run, deployed by the network operator or by the OpenShift NFD Operator'
        ).to.be.greaterThan(0);
        expect(
            nfd.every((pod) => pod.status.phase === 'Running'),
            'all node-feature-discovery pods should be Running'
        ).to.equal(true);
    });

    test('network operator accepts the NicClusterPolicy', async () => {
        const policy = await kubeCli.CustomObjectsApi.getClusterCustomObject({
            group: 'mellanox.com',
            version: 'v1alpha1',
            plural: 'nicclusterpolicies',
            name: 'nic-cluster-policy'
        });
        expect(policy, 'NicClusterPolicy should be applied on the cluster').to.not.be.undefined;
    });

    test('network operator reconciles the NicClusterPolicy into its device-plugin daemonset', async () => {
        await waitUtils.waitForCondition(
            async () => {
                const daemonsets = (
                    await kubeCli.AppsV1Api.listNamespacedDaemonSet({
                        namespace: NETWORK_OPERATOR_NAMESPACE
                    })
                ).items;
                return daemonsets.some((ds) => ds.metadata.name.includes('rdma-shared-dp'));
            },
            'network operator to reconcile the NicClusterPolicy into the rdma device-plugin daemonset',
            true,
            180
        );
    });

    test('multus is available for secondary networks', async () => {
        const multusNamespace = isOpenShift ? OPENSHIFT_MULTUS_NAMESPACE : NETWORK_OPERATOR_NAMESPACE;
        await waitUtils.waitForCondition(
            async () => {
                const daemonsets = (
                    await kubeCli.AppsV1Api.listNamespacedDaemonSet({ namespace: multusNamespace })
                ).items;
                return daemonsets.some((ds) => ds.metadata.name.includes('multus'));
            },
            isOpenShift
                ? 'the OpenShift multus daemonset to serve the network operator secondary networks'
                : 'network operator to deploy the multus daemonset for secondary networks',
            true,
            180
        );
    });

    test('network operator generates a NetworkAttachmentDefinition from the MacvlanNetwork', async () => {
        await waitUtils.waitForCondition(
            async () => {
                try {
                    const nad = await kubeCli.CustomObjectsApi.getNamespacedCustomObject({
                        group: 'k8s.cni.cncf.io',
                        version: 'v1',
                        namespace,
                        plural: 'network-attachment-definitions',
                        name: MACVLAN_NETWORK_NAME
                    });
                    return nad !== undefined;
                } catch (error) {
                    logger.debug(`NetworkAttachmentDefinition not ready yet: ${error.message}`);
                    return false;
                }
            },
            'network operator to generate a NetworkAttachmentDefinition from the MacvlanNetwork',
            true,
            180
        );
    });

    shouldRunNativeWorkloadCrdTests(() =>
    test('runai workload attaches to the network operator secondary network', async () => {
        jobName = objectGenerator.generateNameByPrefix('netcert-macvlan');
        workload = new TrainingWorkload(config.quickStartImage.emptyImage, namespace, jobName);
        workload.annotations = { [SECONDARY_NETWORK_ANNOTATION]: MACVLAN_NETWORK_NAME };
        await kubeCli.createWorkload(workload);

        await waitUtils.waitForPods(jobName, namespace, 1, 120, ['Running']);

        let networkStatus: string;
        await waitUtils.waitForCondition(
            async () => {
                const pods = await kubeCli.listPodsOfJob(namespace, jobName);
                networkStatus = pods
                    .map((pod) => pod.metadata.annotations?.[NETWORK_STATUS_ANNOTATION])
                    .find((status) => status?.includes('net1'));
                return networkStatus !== undefined;
            },
            'multus to attach the secondary network interface (net1) to the runai workload pod',
            true,
            120
        );
        expect(
            networkStatus,
            `pod should report the ${MACVLAN_NETWORK_NAME} secondary network`
        ).to.contain(MACVLAN_NETWORK_NAME);
        })
    );

    shouldRunNativeWorkloadCrdTests(() =>
    test('runai workload exchanges traffic over the network operator secondary network', async () => {
        const clientJob = objectGenerator.generateNameByPrefix('netcert-traffic');
        const client = new TrainingWorkload('busybox', namespace, clientJob);
        client.command = 'sleep 3600';
        client.annotations = { [SECONDARY_NETWORK_ANNOTATION]: MACVLAN_NETWORK_NAME };
        await kubeCli.createWorkload(client);
        const clientPod = (await waitUtils.waitForPods(clientJob, namespace, 1, 120, ['Running'])).pop();

        const peerName = `${clientJob}-peer`;
        const peerPod = {
            apiVersion: 'v1',
            kind: 'Pod',
            metadata: {
                name: peerName,
                namespace,
                annotations: { [SECONDARY_NETWORK_ANNOTATION]: MACVLAN_NETWORK_NAME }
            },
            spec: {
                nodeName: clientPod.spec.nodeName,
                restartPolicy: 'Never',
                containers: [
                    {
                        name: 'peer',
                        image: 'busybox',
                        command: [
                            'sh',
                            '-c',
                            'mkdir -p /tmp/www && echo netcert-ok > /tmp/www/index.html && httpd -f -p 8080 -h /tmp/www'
                        ]
                    }
                ]
            }
        };
        await kubeCli.createFromYaml(JSON.stringify(peerPod));
        await waitUtils.waitForCondition(
            async () =>
                (await kubeCli.CoreV1Api.readNamespacedPod({ name: peerName, namespace })).status
                    ?.phase === 'Running',
            'the peer pod on the secondary network to be Running',
            true,
            120
        );

        const peerAddr = await kubeCli.execInPod(peerName, namespace, [
            'ip',
            '-o',
            '-4',
            'addr',
            'show',
            'net1'
        ]);
        const peerIp = peerAddr.stdout.trim().split(/\s+/)[3]?.split('/')[0];
            expect(peerIp, 'peer pod should get a net1 address from the operator macvlan network').to.match(
                /^192\.168\.2\./
            );

        const probe = await kubeCli.execInPod(clientPod.metadata.name, namespace, [
            'wget',
            '-qO-',
            '--timeout=5',
            `http://${peerIp}:8080/`
        ]);
        expect(
            probe.stdout,
            `runai workload should reach the peer over the operator secondary network (stderr: ${probe.stderr})`
        ).to.contain('netcert-ok');
        })
    );

    test.afterAll(async () => {
        try {
            await kubeCli.CustomObjectsApi.deleteClusterCustomObject({
                group: 'mellanox.com',
                version: 'v1alpha1',
                plural: 'macvlannetworks',
                name: MACVLAN_NETWORK_NAME
            });
        } catch (error) {
            logger.debug(`MacvlanNetwork cleanup skipped: ${error.message}`);
        }
        try {
            await operations.deleteProjectNamespaceAndWorkloads(project, 120);
        } catch (error) {
            logger.error(`network-cert cleanup failed: ${error.message}`);
        }
    });
});
