/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { CustomResourceDefinition, DistributedJobType, logger, PodStatus } from '..';
import { kubeCli } from '../../cli';
import { KubernetesObject } from '@kubernetes/client-node/dist/types';
import { Preemptibility } from '../../api/apis/assetsApi/generated/model/preemptibility';

export enum WorkloadStatus {
    SUCCEEDED = 'Succeeded',
    START_ERROR = 'StartError',
    RUNNING = 'Running',
    INITIALIZING = 'Initializing',
    PENDING = 'Pending',
    TIMEOUT = 'Timeout',
    SUSPENDED = 'Suspended',
    SUSPENDING = 'Suspending',
    STOPPED = 'Stopped',
    STOPPING = 'Stopping',
    ERROR = 'Error',
    FAILED = 'Failed',
    RESUMING = 'Resuming',
    COMPLETED = 'Completed'
}

export class WorkloadPort {
    container: number;

    external: number;

    protocol?: string;

    serviceType?: string; // only available in cluster verison 2.17 and up
}

export class S3Bucket {
    bucket: string;

    key?: string;

    secret?: string;

    path: string;

    url?: string;
}

export class GitSync {
    repository: string;

    branch?: string;

    revision?: string;

    directory?: string;

    readOnly?: boolean;

    secretName?: string;

    username?: string;

    password?: string;
}

export class Pvc {
    claimName: string;

    storageClass?: string;

    size?: string;

    path: string;

    readOnly?: boolean;

    ephemeral?: boolean;

    existingPvc?: boolean;

    readWriteOnce?: boolean;

    readOnlyMany?: boolean;

    readWriteMany?: boolean;
}

export class Nfs {
    sourcePath: string;

    targetPath: string;

    readOnly?: boolean;

    nfsServer?: string;
}

export class Volume {
    sourcePath: string;

    targetPath: string;

    readOnly?: boolean;

    nfsServer?: string;
}

export class ConfigMap {
    name: string;

    mountPath: string;
}

export class EmptyDir {
    mountPath: string;

    medium?: string;

    sizeLimit?: string;
}

export class Toleration {
    operator: string;

    key: string;

    value?: string;

    effect?: string;

    seconds?: number;
}

export class ExposedUrl {
    containerPort: number;

    customUrl?: boolean;

    url?: string;
}

export enum WorkloadConditionType {
    Initialized = 'Initialized',
    Running = 'Running',
    Succeeded = 'Succeeded',
    Failed = 'Failed',
    Suspended = 'Suspended',
    Pending = 'Pending',
    Ready = 'Ready',
    LastReconcileSucceeded = 'LastReconcileSucceeded'
}

export class WorkloadCondition {
    type: WorkloadConditionType;

    status: boolean;

    reason: string;

    message: string;

    lastTransitionTime: string;

    lastProbeTime: string;

    static fromRawCondition(rawCondition: object): WorkloadCondition {
        const condition = new WorkloadCondition();
        if (!rawCondition) {
            return condition;
        }
        condition.type = rawCondition['type'];
        condition.status = rawCondition['status'];
        condition.reason = rawCondition['reason'];
        condition.message = rawCondition['message'];
        condition.lastTransitionTime = rawCondition['lastTransitionTime'];
        condition.lastProbeTime = rawCondition['lastProbeTime'];
        return condition;
    }
}

export class TestingSpecificFields {
    namespace: string;
}

export type CapabilitiesItem = {
    capability: string;
};
export type ExtendedResource = {
    name: string;
    quantity: string;
};

export abstract class BaseWorkload {
    testingSpecificFields: TestingSpecificFields;

    //   mandatory fields
    name: string;

    image: string;

    allowPrivilegeEscalation: boolean;

    arguments: string;

    backoffLimit: number;

    extendedResources: { [key: string]: ExtendedResource };

    capabilities: { [key: string]: CapabilitiesItem };

    command: string;

    completions: number;

    cpu: string; // should be a string to contain non-numeric cpu quanities

    cpuLimit: string;

    createHomeDir: boolean;

    environment: { [key: string]: string };

    exposedUrls: { [key: string]: ExposedUrl };

    gitSync: { [key: string]: GitSync };

    gpu: string; // should be a string to contain non-numeric gpu quantities

    gpuDevices: number;

    gpuLimit: string;

    gpuMemory: string;

    gpuMemoryLimit: string;

    hostIpc: boolean;

    active: boolean;

    hostNetwork: boolean;

    imagePullPolicy: string;

    largeShm: boolean;

    memory: string;

    mountPropagation: boolean;

    memoryLimit: string;

    mpi: boolean;

    namePrefix: string;

    nodePool: string; //DEPRECATED IN 2.9; only available in 2.8

    nodePools: string;

    nodeType: string;

    preemptible: boolean; // DEPRECATED IN 2.24: Use preemptibility instead

    preemptibility: Preemptibility;

    parallelism: number;

    ports: { [key: string]: WorkloadPort };

    processes: number;

    pvcs: { [key: string]: Pvc };

    configMapVolumes: { [key: string]: ConfigMap };

    runAsGid: number;

    runAsUid: number;

    runAsUser: boolean;

    serviceType?: string;

    s3: { [key: string]: S3Bucket };

    stdin: boolean;

    supplementalGroups: string;

    tolerations: { [key: string]: Toleration };

    tty: boolean;

    username: string;

    userId: string;

    workingDir: string;

    nfs: { [key: string]: Nfs };

    volumes: { [key: string]: Volume };

    terminateAfterPreemption: boolean;

    readinessProbe: object;

    labels: object;

    annotations: object;

    protected constructor(name: string, namespace: string, image: string) {
        this.testingSpecificFields = {
            namespace: namespace
        };
        this.name = name;
        this.image = image;
        this.environment = {};
    }

    getMetadata(): KubernetesObject {
        const crd = this.getCrd();
        return {
            apiVersion: `${crd.group}/${crd.version}`,
            kind: crd.kind,
            metadata: {
                name: this.name,
                namespace: this.testingSpecificFields.namespace,
                annotations: {}
            }
        };
    }

    //
    //   this is a general function to compose k8s object for a given workload
    //   the returned object contains both metadata and spec.
    //   in spec, scalar items are coded as { value: value1 } while maps are
    //   coded as { items: { key1: { value: value1 }, { key2: { value: value2 } ... }
    //   or as { values: { key1: { value: value1 }, { key2: { value: value2 } ... },
    //   depends on the specific field.
    //
    getBody(): KubernetesObject {
        const result = this.getMetadata();
        result['spec'] = {};

        const flatFields = this.getFlatFields();
        const fieldsToBeMappedAsValues = this.getFieldsToBeMappedAsValues();

        for (const [key, value] of Object.entries(this)) {
            if (key === 'testingSpecificFields') {
                continue;
            }
            if (flatFields.includes(key)) {
                result['spec'][key] = value;
                continue;
            }
            if (typeof value === 'object') {
                const items: object = {};

                for (const [itemKey, itemValue] of Object.entries(value)) {
                    items[itemKey] = {
                        value: itemValue
                    };
                }

                // support both formats - "items:" and "values:"
                if (fieldsToBeMappedAsValues.includes(key)) {
                    result['spec'][key] = {
                        values: items
                    };
                } else {
                    result['spec'][key] = {
                        items: items
                    };
                }
            } else {
                // simple value
                if (value !== undefined) {
                    result['spec'][key] = {
                        value: value
                    };
                }
            }
        }
        if (this.namePrefix !== undefined && this.namePrefix !== '') {
            delete result['spec']['name'];
        }
        this.removeMpiMasterSpecGpu(result);

        return result;
    }

    // Since MPI doesnt allow to allocate gpu to the master pod, we need to remove the gpu field from the master spec
    removeMpiMasterSpecGpu(result: KubernetesObject): void {
        if (result['spec']['jobType'] === DistributedJobType.MpiJobType) {
            result['masterSpec'] = Object.assign({}, result['spec']);
            delete result['masterSpec']['gpu'];
        }
    }

    //  Some workloads, like DistributedJob, have "flat" fields. for those fields the getBody function
    //  should not attempt to add rules, value, etc. just keep them as is
    getFlatFields(): string[] {
        return [];
    }

    getFieldsToBeMappedAsValues(): string[] {
        return ['readinessProbe'];
    }

    abstract getCrd(): CustomResourceDefinition;

    async validateStatus(expectedStatus: PodStatus): Promise<boolean> {
        const pods = await kubeCli.listPodsOfJob(this.testingSpecificFields.namespace, this.name);
        if (pods.length === 0) {
            return false;
        }

        logger.debug(
            `workload ${this.name} pods statuses  ${JSON.stringify(pods.map((pod) => pod.status.phase))}`
        );
        return pods.every((pod) => pod.status.phase === expectedStatus);
    }
}
