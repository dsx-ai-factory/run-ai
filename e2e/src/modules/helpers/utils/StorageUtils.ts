/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { config, logger, waitUtils } from '..';
import { kubeCli } from '../../cli';
import {
    V1PersistentVolume,
    V1PersistentVolumeClaim,
    V1PersistentVolumeSpec,
    V1Pod,
    V1StorageClass
} from '@kubernetes/client-node';

const RUNAI_ASSET_LABEL = 'run.ai/asset';
const PVC_WAIT_FOR_FIRST_CONSUMER = 'WaitForFirstConsumer';

export class StorageUtils {
    async listStorageClasses(): Promise<Array<V1StorageClass>> {
        return (await kubeCli.StorageV1Api.listStorageClass()).items;
    }

    async createStorageClass(
        name: string,
        provisioner = 'kubernetes.io/aws-ebs',
        isDefault = false,
        allowVolumeExpansion = false
    ): Promise<V1StorageClass> {
        const storageClassManifest = {
            apiVersion: 'storage.k8s.io/v1',
            kind: 'StorageClass',
            metadata: {
                name: name,
                annotations: {
                    'storageclass.kubernetes.io/is-default-class': String(isDefault)
                }
            },
            provisioner: provisioner,
            reclaimPolicy: 'Retain', // Can be 'Delete', 'Retain', or 'Recycle'
            volumeBindingMode: 'Immediate', // Can be 'Immediate' or 'WaitForFirstConsumer'
            allowVolumeExpansion: allowVolumeExpansion,
            parameters: {
                type: 'gp2' // Example parameter for AWS EBS (customize for your provisioner)
            }
        };
        const res = await kubeCli.StorageV1Api.createStorageClass({ body: storageClassManifest });
        return res;
    }

    // async patchStorageClass(name: string, isDefault= false, allowVolumeExpansion = false): Promise<V1StorageClass> {
    //     const patch = {
    //         allowVolumeExpansion: allowVolumeExpansion
    //     };
    //     const res = await kubeCli.StorageV1Api.patchStorageClass(name, patch);
    //     return res.body;
    // }

    async deleteStorageClass(name: string): Promise<void> {
        await kubeCli.StorageV1Api.deleteStorageClass({ name });
    }

    async getFirstStorageClassNameWithProvisioner(provisioner: string): Promise<string> {
        const storageClasses = await this.listStorageClasses();
        const storageClass = storageClasses.find((sc) => sc.provisioner === provisioner);
        if (storageClass) {
            return storageClass.metadata.name;
        }
        return '';
    }

    async listPvs(): Promise<Array<V1PersistentVolume>> {
        return (await kubeCli.CoreV1Api.listPersistentVolume()).items;
    }

    async getFirstPvWithCsiDriver(csiDriver: string): Promise<V1PersistentVolume> {
        const pvs = await this.listPvs();
        return pvs.find((pv) => pv.spec.csi?.driver === csiDriver);
    }

    async getPv(name: string): Promise<V1PersistentVolume> {
        return await kubeCli.CoreV1Api.readPersistentVolume({ name });
    }

    async createPv(pv: V1PersistentVolume): Promise<V1PersistentVolume> {
        logger.debug(`Creating PV ${pv.metadata.name}`);
        return await kubeCli.CoreV1Api.createPersistentVolume({ body: pv });
    }

    async deletePv(name: string): Promise<void> {
        await kubeCli.CoreV1Api.deletePersistentVolume({ name });
    }

    private buildPvc(
        name: string,
        namespace: string,
        labels?: Record<string, string>,
        storageClass = 'standard'
    ): V1PersistentVolumeClaim {
        return {
            apiVersion: 'v1',
            kind: 'PersistentVolumeClaim',
            metadata: {
                name,
                namespace,
                ...(labels && { labels })
            },
            spec: {
                accessModes: ['ReadWriteOnce'],
                storageClassName: storageClass,
                resources: { requests: { storage: '1Mi' } }
            }
        };
    }

    buildUnlabeledPvc(
        name: string,
        namespace: string,
        storageClass = 'standard'
    ): V1PersistentVolumeClaim {
        return this.buildPvc(name, namespace, undefined, storageClass);
    }

    buildLabeledPvc(
        name: string,
        namespace: string,
        assetLabelKey = RUNAI_ASSET_LABEL,
        assetLabelValue = 'pvc',
        storageClass = 'standard'
    ): V1PersistentVolumeClaim {
        return this.buildPvc(name, namespace, { [assetLabelKey]: assetLabelValue }, storageClass);
    }

    async createPvc(pvc: V1PersistentVolumeClaim): Promise<V1PersistentVolumeClaim> {
        logger.debug(`Creating PVC ${pvc.metadata.namespace}/${pvc.metadata.name}`);
        return await kubeCli.CoreV1Api.createNamespacedPersistentVolumeClaim({
            namespace: pvc.metadata.namespace,
            body: pvc
        });
    }

    async deletePvc(name: string, namespace: string): Promise<void> {
        await kubeCli.CoreV1Api.deleteNamespacedPersistentVolumeClaim({ name, namespace });
    }

    async getPvc(name: string, namespace: string): Promise<V1PersistentVolumeClaim> {
        return await kubeCli.CoreV1Api.readNamespacedPersistentVolumeClaim({ name, namespace });
    }

    async listPvcsInNamespace(namespace: string): Promise<Array<V1PersistentVolumeClaim>> {
        return (await kubeCli.CoreV1Api.listNamespacedPersistentVolumeClaim({ namespace })).items;
    }

    buildNewPvObjFromPv(
        newPvName: string,
        storageClassName: string,
        pvToCopy: V1PersistentVolume
    ): V1PersistentVolume {
        const pvSpec = Object.assign(
            {},
            JSON.parse(JSON.stringify(pvToCopy.spec)) as V1PersistentVolumeSpec
        ) as V1PersistentVolumeSpec;
        pvSpec.claimRef = undefined;
        pvSpec.accessModes = ['ReadOnlyMany'];
        pvSpec.persistentVolumeReclaimPolicy = 'Retain';
        pvSpec.storageClassName = storageClassName;
        return {
            metadata: {
                name: newPvName
            },
            spec: pvSpec
        };
    }

    buildPvcObjFromPv(
        pv: V1PersistentVolume,
        pvcName: string,
        pvcNamespace: string
    ): V1PersistentVolumeClaim {
        const accessModes = pv.spec.accessModes;
        const storage = pv.spec.capacity.storage;
        const volumeMode = pv.spec.volumeMode;
        const storageClassName = pv.spec.storageClassName;

        return {
            metadata: {
                name: pvcName,
                namespace: pvcNamespace
            },
            spec: {
                accessModes: accessModes,
                volumeMode: volumeMode,
                volumeName: pv.metadata.name,
                storageClassName: storageClassName,
                resources: {
                    requests: {
                        storage: storage
                    }
                }
            }
        } as V1PersistentVolumeClaim;
    }

    async getPvcBoundVolumeName(pvcName: string, pvcNamespace: string): Promise<string> {
        const pvc = await this.getPvc(pvcName, pvcNamespace);
        if (pvc.status.phase === 'Bound') {
            return pvc.spec.volumeName;
        }
        return '';
    }

    async waitForPvcToBeBoundToPv(
        pvcName: string,
        pvcNamespace: string,
        expectedBoundPvName: string
    ): Promise<void> {
        const boundPvName = await this.waitForPvcToBeBound(pvcName, pvcNamespace);
        if (boundPvName !== expectedBoundPvName) {
            throw new Error(
                `PVC ${pvcNamespace}/${pvcName} is bound to a different PV than expected; bound to: ${boundPvName}, expected: ${expectedBoundPvName}`
            );
        }
    }

    async ensurePvcBound(
        pvcName: string,
        pvcNamespace: string,
        binderImage: string = config.quickStartImage.ubuntu
    ): Promise<string> {
        const boundPvName = await this.getPvcBoundVolumeName(pvcName, pvcNamespace);
        if (boundPvName !== '') {
            return boundPvName;
        }

        if (!(await this.pvcNeedsConsumerToBind(pvcName, pvcNamespace))) {
            return await this.waitForPvcToBeBound(pvcName, pvcNamespace);
        }

        const binderPodName = `pvc-binder-${pvcName}`.slice(0, 63);
        logger.info(
            `PVC ${pvcNamespace}/${pvcName} uses ${PVC_WAIT_FOR_FIRST_CONSUMER}; creating pod ${binderPodName} to trigger binding`
        );
        await kubeCli.CoreV1Api.createNamespacedPod({
            namespace: pvcNamespace,
            body: {
                apiVersion: 'v1',
                kind: 'Pod',
                metadata: { name: binderPodName, namespace: pvcNamespace },
                spec: {
                    restartPolicy: 'Never',
                    containers: [
                        {
                            name: 'binder',
                            image: binderImage,
                            command: ['sh', '-c', 'sleep 10'],
                            volumeMounts: [{ name: 'pvc', mountPath: '/mnt/pvc' }]
                        }
                    ],
                    volumes: [
                        { name: 'pvc', persistentVolumeClaim: { claimName: pvcName } }
                    ]
                }
            } as V1Pod
        });

        try {
            return await this.waitForPvcToBeBound(pvcName, pvcNamespace);
        } finally {
            try {
                await kubeCli.CoreV1Api.deleteNamespacedPod({
                    name: binderPodName,
                    namespace: pvcNamespace
                });
            } catch (error) {
                logger.warn(`failed to delete binder pod ${binderPodName}: ${error}`);
            }
        }
    }

    /** True when the PVC's StorageClass only binds once a pod consumes the claim. */
    private async pvcNeedsConsumerToBind(pvcName: string, pvcNamespace: string): Promise<boolean> {
        const pvc = await this.getPvc(pvcName, pvcNamespace);
        const storageClassName = pvc.spec?.storageClassName;
        if (!storageClassName) {
            return false;
        }
        try {
            const storageClass = await kubeCli.StorageV1Api.readStorageClass({
                name: storageClassName
            });
            return storageClass.volumeBindingMode === PVC_WAIT_FOR_FIRST_CONSUMER;
        } catch (error) {
            logger.warn(`failed to read storage class ${storageClassName}: ${error}`);
            return false;
        }
    }

    async waitForPvcToBeBound(pvcName: string, pvcNamespace: string): Promise<string> {
        let boundPvName: string;
        const isPvcBoundToPv = async (): Promise<boolean> => {
            boundPvName = await this.getPvcBoundVolumeName(pvcName, pvcNamespace);
            if (boundPvName === '') {
                logger.debug(`PVC ${pvcNamespace}/${pvcName} is not bound`);
                return false;
            }
            logger.debug(`PVC ${pvcNamespace}/${pvcName} is bound to PV ${boundPvName}`);
            return true;
        };

        await waitUtils.waitForCondition(
            isPvcBoundToPv,
            `waiting for PVC ${pvcNamespace}/${pvcName} to be bound`,
            true,
            120,
            5
        );

        return boundPvName;
    }

    async waitForPvcAndPvOfDatavolumeForNamespace(
        dvOwnerName: string,
        namespace: string,
        desiredResult = true
    ): Promise<V1PersistentVolumeClaim> {
        const pvc = await this.waitForPvcOfDatavolumeInNamespace(dvOwnerName, namespace, desiredResult);
        await this.waitForPvOfDatavolumeForNamespace(dvOwnerName, namespace, desiredResult);
        return pvc;
    }

    async waitForPvcOfDatavolumeInNamespace(
        dvOwnerName: string,
        namespace: string,
        desiredResult = true
    ): Promise<V1PersistentVolumeClaim> {
        let foundPvc;
        const doesPvcExist = async (): Promise<boolean> => {
            const pvcs = await this.listPvcsInNamespace(namespace);
            foundPvc = pvcs.find((pvc) => pvc.metadata?.labels?.['run.ai/datavolume-name'] === dvOwnerName);
            return foundPvc !== undefined;
        };

        await waitUtils.waitForCondition(
            doesPvcExist,
            `expected pvc owned by datavolume ${dvOwnerName} in namespace ${namespace} to exist, desiredResult: ${desiredResult}`,
            desiredResult,
            30
        );
        return foundPvc;
    }

    async waitForPvOfDatavolumeForNamespace(
        dvOwnerName: string,
        namespace: string,
        desiredResult = true
    ): Promise<V1PersistentVolume> {
        let foundPv;
        const doesPvExist = async (): Promise<boolean> => {
            const pvs = await this.listPvs();
            foundPv = pvs.find((pv) => pv.metadata?.labels?.['run.ai/datavolume-name'] === dvOwnerName);
            return foundPv !== undefined;
        };

        await waitUtils.waitForCondition(
            doesPvExist,
            `expected pv owned by datavolume ${dvOwnerName} for namespace ${namespace} to exist, desiredResult: ${desiredResult}`,
            desiredResult,
            30
        );
        return foundPv;
    }

    //TODO: use in clsuter-wide-secrets-tests
    async waitForPvcCreation(
        name: string,
        namespace: string,
        desiredResult = true
    ): Promise<V1PersistentVolumeClaim> {
        let foundPvc;
        const doesPvcExist = async (): Promise<boolean> => {
            const pvcs = await this.listPvcsInNamespace(namespace);
            foundPvc = pvcs.find((pvc) => pvc.metadata.name === name);
            return foundPvc !== undefined;
        };

        await waitUtils.waitForCondition(
            doesPvcExist,
            `expected pvc ${namespace}/${name} to exist, desiredResult: ${desiredResult}`,
            desiredResult,
            60
        );
        return foundPvc;
    }
}
