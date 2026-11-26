/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { crds, CustomResourceDefinition } from './customResourceDefinition';
import { BaseWorkload } from './baseWorkload';
import { KubernetesObject } from '@kubernetes/client-node';

interface JobResources {
    limits?: {
        cpu?: string;
        memory?: string;
        'nvidia.com/gpu'?: string;
    };
    requests?: {
        cpu?: string;
        memory?: string;
        'nvidia.com/gpu'?: string;
    };
}

interface TFJobSpec {
    tfReplicaSpecs: {
        Worker: {
            replicas: number;
            restartPolicy: string;
            template: {
                spec: {
                    schedulerName: string;
                    containers: {
                        name: string;
                        image: string;
                        command: string[];
                        resources?: JobResources;
                    }[];
                };
            };
        };
    };
}

export class TFJob extends BaseWorkload {
    spec: TFJobSpec;

    resources: JobResources;

    constructor(
        name: string,
        namespace: string,
        image: string,
        command = 'sleep infinity',
        workerReplicas = 2,
        restartPolicy = 'OnFailure',
        resources?: JobResources
    ) {
        super(name, namespace, image);
        this.command = command;
        this.resources = resources;
        this.initSpec(workerReplicas, restartPolicy);
    }

    getCrd(): CustomResourceDefinition {
        return crds.tfJobCrd;
    }

    getBody(): KubernetesObject {
        const result = this.getMetadata();
        result['spec'] = this.spec;
        return result;
    }

    initSpec(workerReplicas, restartPolicy): void {
        this.spec = {
            tfReplicaSpecs: {
                Worker: {
                    replicas: workerReplicas,
                    restartPolicy: restartPolicy,
                    template: {
                        spec: {
                            schedulerName: 'runai-scheduler',
                            containers: [
                                {
                                    name: 'tensorflow',
                                    image: this.image,
                                    command: this.command.split(' '),
                                    resources: this.resources
                                }
                            ]
                        }
                    }
                }
            }
        };
    }
}

interface XGBoostJobSpec {
    xgbReplicaSpecs: {
        Master: {
            replicas: number;
            restartPolicy: string;
            template: {
                spec: {
                    schedulerName: string;
                    containers: {
                        name: string;
                        image: string;
                        command: string[];
                        resources?: JobResources;
                    }[];
                };
            };
        };
        Worker: {
            replicas: number;
            restartPolicy: string;
            template: {
                spec: {
                    schedulerName: string;
                    containers: {
                        name: string;
                        image: string;
                        command: string[];
                        resources?: JobResources;
                    }[];
                };
            };
        };
    };
}

export class XGBoostJob extends BaseWorkload {
    spec: XGBoostJobSpec;

    resourcesMaster: JobResources;

    resourcesWorker: JobResources;

    constructor(
        name: string,
        namespace: string,
        image: string,
        command = 'sleep infinity',
        masterReplicas = 1,
        workerReplicas = 2,
        restartPolicy = 'OnFailure',
        resourcesMaster?: JobResources,
        resourcesWorker?: JobResources
    ) {
        super(name, namespace, image);
        this.command = command;
        this.resourcesMaster = resourcesMaster;
        this.resourcesWorker = resourcesWorker;
        this.initSpec(masterReplicas, workerReplicas, restartPolicy);
    }

    getCrd(): CustomResourceDefinition {
        return crds.xgboostJobCrd;
    }

    getBody(): KubernetesObject {
        const result = this.getMetadata();
        result['spec'] = this.spec;
        return result;
    }

    initSpec(masterReplicas, workerReplicas, restartPolicy): void {
        this.spec = {
            xgbReplicaSpecs: {
                Master: {
                    replicas: masterReplicas,
                    restartPolicy: restartPolicy,
                    template: {
                        spec: {
                            schedulerName: 'runai-scheduler',
                            containers: [
                                {
                                    name: 'xgboost',
                                    image: this.image,
                                    command: this.command.split(' '),
                                    resources: this.resourcesMaster
                                }
                            ]
                        }
                    }
                },
                Worker: {
                    replicas: workerReplicas,
                    restartPolicy: restartPolicy,
                    template: {
                        spec: {
                            schedulerName: 'runai-scheduler',
                            containers: [
                                {
                                    name: 'xgboost',
                                    image: this.image,
                                    command: this.command.split(' '),
                                    resources: this.resourcesWorker
                                }
                            ]
                        }
                    }
                }
            }
        };
    }
}

interface PyTorchJobSpec {
    pytorchReplicaSpecs: {
        Master: {
            replicas: number;
            restartPolicy: string;
            template: {
                spec: {
                    schedulerName: string;
                    containers: {
                        name: string;
                        image: string;
                        imagePullPolicy: string;
                        command: string[];
                        resources?: JobResources;
                    }[];
                };
            };
        };
        Worker: {
            replicas: number;
            restartPolicy: string;
            template: {
                spec: {
                    schedulerName: string;
                    containers: {
                        name: string;
                        image: string;
                        imagePullPolicy: string;
                        command: string[];
                        resources?: JobResources;
                    }[];
                };
            };
        };
    };
}

export class PyTorchJob extends BaseWorkload {
    spec: PyTorchJobSpec;

    resourcesMaster: JobResources;

    resourcesWorker: JobResources;

    constructor(
        name: string,
        namespace: string,
        image: string,
        command = 'sleep infinity',
        masterReplicas = 1,
        workerReplicas = 1,
        restartPolicy = 'OnFailure',
        resourcesMaster?: JobResources,
        resourcesWorker?: JobResources
    ) {
        super(name, namespace, image);
        this.command = command;
        this.resourcesMaster = resourcesMaster;
        this.resourcesWorker = resourcesWorker;
        this.initSpec(masterReplicas, workerReplicas, restartPolicy);
    }

    getCrd(): CustomResourceDefinition {
        return crds.pyTorchJobCrd;
    }

    getBody(): KubernetesObject {
        const result = this.getMetadata();
        result['spec'] = this.spec;
        return result;
    }

    initSpec(masterReplicas, workerReplicas, restartPolicy): void {
        this.spec = {
            pytorchReplicaSpecs: {
                Master: {
                    replicas: masterReplicas,
                    restartPolicy: restartPolicy,
                    template: {
                        spec: {
                            schedulerName: 'runai-scheduler',
                            containers: [
                                {
                                    name: 'pytorch',
                                    image: this.image,
                                    imagePullPolicy: 'Always',
                                    command: this.command.split(' '),
                                    resources: this.resourcesMaster
                                }
                            ]
                        }
                    }
                },
                Worker: {
                    replicas: workerReplicas,
                    restartPolicy: restartPolicy,
                    template: {
                        spec: {
                            schedulerName: 'runai-scheduler',
                            containers: [
                                {
                                    name: 'pytorch',
                                    image: this.image,
                                    imagePullPolicy: 'Always',
                                    command: this.command.split(' '),
                                    resources: this.resourcesWorker
                                }
                            ]
                        }
                    }
                }
            }
        };
    }
}

export interface RayClusterSpec {
    rayVersion: string;
    headGroupSpec: any;
    workerGroupSpecs: any[];
    [key: string]: any;
}

export class RayClusterWorkload extends BaseWorkload {
    spec: RayClusterSpec;

    resources: JobResources;

    memoryRequest: string;

    gpuRequest: string;

    cpuRequest: string;

    rayVersion: string;

    maxReplicas: number;

    minReplicas: number;

    replicas: number;

    groupName: string;

    nodePoolSelectors?: Array<{ key: string; value: string }>;

    /**
     * @param name
     * @param namespace
     * @param nodePoolSelectors Optional. Array of key-value pairs for node affinity.
     * @param image
     * @param rayVersion
     * @param cpuLimit
     * @param cpuRequest
     * @param memoryLimit
     * @param memoryRequest
     * @param gpuLimit
     * @param gpuRequest
     * @param maxReplicas
     * @param minReplicas
     * @param replicas
     * @param groupName
     */
    constructor(
        name: string,
        namespace: string,
        nodePoolSelectors?: Array<{ key: string; value: string }>,
        image = 'rayproject/ray:2.4.0',
        rayVersion = '2.4.0',
        cpuLimit = '500m',
        cpuRequest = '500m',
        memoryLimit = '2G',
        memoryRequest = '2G',
        gpuLimit = '2',
        gpuRequest = '2',
        maxReplicas = 2,
        minReplicas = 2,
        replicas = 2,
        groupName = 'small-group'
    ) {
        super(name, namespace, image);
        this.image = image;
        this.rayVersion = rayVersion;
        this.cpuLimit = cpuLimit;
        this.cpuRequest = cpuRequest;
        this.memoryLimit = memoryLimit;
        this.memoryRequest = memoryRequest;
        this.gpuLimit = gpuLimit;
        this.gpuRequest = gpuRequest;
        this.maxReplicas = maxReplicas;
        this.minReplicas = minReplicas;
        this.replicas = replicas;
        this.groupName = groupName;
        this.nodePoolSelectors = nodePoolSelectors;
        this.resources = {
            limits: {
                cpu: this.cpuLimit,
                memory: this.memoryLimit,
                'nvidia.com/gpu': this.gpuLimit
            },
            requests: {
                cpu: this.cpuRequest,
                memory: this.memoryRequest,
                'nvidia.com/gpu': this.gpuRequest
            }
        };
        this.initSpec();
    }

    private getAffinity(): any | undefined {
        const selectors =
            this.nodePoolSelectors && this.nodePoolSelectors.length > 0 ? this.nodePoolSelectors : [];
        if (selectors.length > 0) {
            return {
                nodeAffinity: {
                    requiredDuringSchedulingIgnoredDuringExecution: {
                        nodeSelectorTerms: [
                            {
                                matchExpressions: selectors.map((sel) => ({
                                    key: sel.key,
                                    operator: 'In',
                                    values: [sel.value]
                                }))
                            }
                        ]
                    }
                }
            };
        }
        return undefined;
    }

    initSpec(): void {
        const affinity = this.getAffinity();
        this.spec = {
            rayVersion: this.rayVersion,
            headGroupSpec: {
                rayStartParams: {
                    'dashboard-host': '0.0.0.0'
                },
                template: {
                    spec: {
                        schedulerName: 'runai-scheduler',
                        containers: [
                            {
                                name: 'ray-head',
                                image: this.image,
                                resources: this.resources,
                                ports: [
                                    { containerPort: 6379, name: 'gcs' },
                                    { containerPort: 8265, name: 'dashboard' },
                                    { containerPort: 10001, name: 'client' }
                                ],
                                lifecycle: {
                                    preStop: {
                                        exec: {
                                            command: ['/bin/sh', '-c', 'ray stop']
                                        }
                                    }
                                }
                            }
                        ],
                        ...(affinity ? { affinity } : {})
                    }
                }
            },
            workerGroupSpecs: [
                {
                    groupName: this.groupName,
                    replicas: this.replicas,
                    minReplicas: this.minReplicas,
                    maxReplicas: this.maxReplicas,
                    rayStartParams: {},
                    template: {
                        spec: {
                            schedulerName: 'runai-scheduler',
                            containers: [
                                {
                                    name: 'ray-worker',
                                    image: this.image,
                                    resources: {
                                        limits: {
                                            cpu: this.cpuLimit,
                                            memory: this.memoryLimit,
                                            'nvidia.com/gpu': this.gpuLimit
                                        },
                                        requests: {
                                            cpu: this.cpuRequest,
                                            memory: this.memoryRequest,
                                            'nvidia.com/gpu': this.gpuRequest
                                        }
                                    },
                                    lifecycle: {
                                        preStop: {
                                            exec: {
                                                command: ['/bin/sh', '-c', 'ray stop']
                                            }
                                        }
                                    }
                                }
                            ],
                            ...(affinity ? { affinity } : {})
                        }
                    }
                }
            ]
        };
    }

    getCrd(): CustomResourceDefinition {
        return crds.rayClusterCrd;
    }

    getBody(): KubernetesObject {
        const result = this.getMetadata();
        result['spec'] = this.spec;
        return result;
    }
}

export const DynamoDefaultImage = 'python:3.11-slim';

// The Dynamo operator injects health probes into every component: the frontend is probed on
// port 8000 and other components on the "system" port 9090. This command answers those probes
// so the DGD reaches "successful" without running a real inference backend.
export const dynamoHealthCommand = (port: number): string[] => [
    'python3',
    '-c',
    'from http.server import HTTPServer, BaseHTTPRequestHandler\n' +
        'class H(BaseHTTPRequestHandler):\n' +
        '    def do_GET(self):\n' +
        '        self.send_response(200)\n' +
        '        self.end_headers()\n' +
        '        self.wfile.write(b"ok")\n' +
        '    def log_message(self, *a): pass\n' +
        `HTTPServer(("0.0.0.0", ${port}), H).serve_forever()`
];

export const DynamoDefaultCommand = dynamoHealthCommand(8000);

export const DynamoSystemPort = 9090;

export interface NodepoolConfig {
    name: string;
    labelKey?: string;
    labelValue?: string;
    setGpuNode?: boolean;
}

export interface ServiceConfig {
    componentType: string;
    runtimeVersionOverride?: string;
    subComponentType?: string;
    replicas: number;
    envFromSecret?: string;
    modelName?: string;
    resources?: {
        limits?: {
            gpu?: string;
            cpu?: string;
            memory?: string;
        };
        requests?: {
            gpu?: string;
            cpu?: string;
            memory?: string;
        };
    };
    image?: string;
    command?: string[];
    args?: string[];
    workingDir?: string;
}

interface DynamoServiceSpec {
    componentType: string;
    runtimeVersionOverride?: string;
    subComponentType?: string;
    replicas: number;
    envFromSecret?: string;
    modelRef?: {
        name: string;
    };
    dynamoNamespace?: string;
    extraPodSpec: {
        mainContainer: {
            image: string;
            command?: string[];
            args?: string[];
            workingDir?: string;
            resources?: {
                limits?: { [key: string]: string };
                requests?: { [key: string]: string };
            };
        };
        affinity?: {
            nodeAffinity: {
                requiredDuringSchedulingIgnoredDuringExecution: {
                    nodeSelectorTerms: Array<{
                        matchExpressions: Array<{
                            key: string;
                            operator: string;
                            values?: string[];
                        }>;
                    }>;
                };
            };
        };
    };
}

export const LabelPriorityClassName = 'priorityClassName';
export const LabelPreemptibilityName = 'kai.scheduler/preemptibility';

// DynamoGraphDeployment CRD from Kubernetes
export interface DynamoGraphDeploymentCrd extends KubernetesObject {
    spec: {
        services: {
            [key: string]: any;
        };
    };
    status?: {
        state?: string;
        conditions?: any[];
        services?: {
            [key: string]: any;
        };
    };
}

// ExternalWorkload CRD from Kubernetes
export interface ExternalWorkloadSpec {
    /** If we have more fields we want to pass to the server, we can use this field */
    additionalFields?: string;
    /** The CreationTimestamp of the associated workload object */
    creationTimestamp?: string;
    departmentId?: string;
    departmentName?: string;
    group?: string;
    imageName?: string;
    originalWorkloadName?: string;
    /** UID is a type that holds unique ID values, including UUIDs */
    originalWorkloadUID?: string;
    preemptibility?: string;
    priorityClassName?: string;
    projectId?: string;
    projectName?: string;
    /** The following fields will be populated if the workload object contains the info */
    requestedNodePools?: string[];
    requestedPods?: { [key: string]: any };
    /** Request and limit of each resource type we need */
    requestedResources?: { [key: string]: any };
    topology?: { [key: string]: any };
    /** The type of the workload - usually the k8s object's kind */
    type?: string;
    /** URLs of the workload, if exist */
    urls?: string[];
}

export interface ExternalWorkloadCrd extends KubernetesObject {
    spec?: ExternalWorkloadSpec;
    status?: {
        observedStatus?: {
            phase?: string;
            [key: string]: any;
        };
        [key: string]: any;
    };
}

export function createServiceConfig(
    componentType: string,
    subComponentType?: string,
    modelName?: string
): ServiceConfig {
    return {
        componentType,
        subComponentType,
        modelName,
        replicas: 1,
        resources: {
            limits: {
                gpu: '1'
            }
        },
        command: ['/bin/sh', '-c', 'sleep infinity'],
        args: ['--model', 'TinyLlama/TinyLlama-1.1B-Chat-v1.0', '--gpu-memory-utilization', '0.9']
    };
}

export class DynamoGraphDeployment extends BaseWorkload {
    spec: any;

    runtimeVersionOverride?: string;

    replicas: number;

    host: string;

    project: string;

    resources: JobResources;

    nodepools?: NodepoolConfig[];

    services?: { [serviceName: string]: ServiceConfig[] };

    customAnnotations?: { [key: string]: string };

    customLabels?: { [key: string]: string };

    constructor(
        name: string,
        namespace: string,
        project: string,
        replicas: number = 1,
        host: string = 'frontend',
        image = DynamoDefaultImage,
        cpuLimit?: string,
        cpuRequest?: string,
        memoryLimit?: string,
        memoryRequest?: string,
        gpuLimit?: string,
        gpuRequest?: string,
        nodepools?: NodepoolConfig[],
        services?: { [serviceName: string]: ServiceConfig[] },
        annotations?: { [key: string]: string },
        labels?: { [key: string]: string },
        runtimeVersionOverride?: string
    ) {
        super(name, namespace, image);
        this.replicas = replicas;
        this.host = host;
        this.project = project;
        this.nodepools = nodepools;
        this.services = services;
        this.customAnnotations = annotations;
        this.customLabels = labels;
        this.runtimeVersionOverride = runtimeVersionOverride;

        // Only set resources if at least one value is provided
        if (cpuLimit || cpuRequest || memoryLimit || memoryRequest || gpuLimit || gpuRequest) {
            this.resources = {};
            if (cpuLimit || memoryLimit || gpuLimit) {
                this.resources.limits = {};
                if (cpuLimit) this.resources.limits.cpu = cpuLimit;
                if (memoryLimit) this.resources.limits.memory = memoryLimit;
                if (gpuLimit) this.resources.limits['nvidia.com/gpu'] = gpuLimit;
            }
            if (cpuRequest || memoryRequest || gpuRequest) {
                this.resources.requests = {};
                if (cpuRequest) this.resources.requests.cpu = cpuRequest;
                if (memoryRequest) this.resources.requests.memory = memoryRequest;
                if (gpuRequest) this.resources.requests['nvidia.com/gpu'] = gpuRequest;
            }
        }

        this.initSpec();
    }

    private buildNodeAffinity(): DynamoServiceSpec['extraPodSpec']['affinity'] | undefined {
        if (!this.nodepools || this.nodepools.length === 0) {
            return undefined;
        }

        const nodeSelectorTerms = [];

        for (const nodepool of this.nodepools) {
            if (nodepool.name === 'default') {
                nodeSelectorTerms.push({
                    matchExpressions: [
                        {
                            key: 'runai/node-pool',
                            operator: 'DoesNotExist'
                        }
                    ]
                });
            } else {
                if (nodepool.labelKey && nodepool.labelValue) {
                    nodeSelectorTerms.push({
                        matchExpressions: [
                            {
                                key: nodepool.labelKey,
                                operator: 'In',
                                values: [nodepool.labelValue]
                            }
                        ]
                    });
                }
            }
        }

        return {
            nodeAffinity: {
                requiredDuringSchedulingIgnoredDuringExecution: {
                    nodeSelectorTerms
                }
            }
        };
    }

    initSpec(): void {
        const servicesSpec: { [serviceName: string]: DynamoServiceSpec } = {};

        if (this.services && Object.keys(this.services).length > 0) {
            // Build services from the provided services map
            for (const [serviceName, serviceConfigs] of Object.entries(this.services)) {
                // Process each ServiceConfig in the array (typically one per service name)
                for (let i = 0; i < serviceConfigs.length; i++) {
                    const service = serviceConfigs[i];
                    const serviceSpec: Partial<DynamoServiceSpec> = {
                        componentType: service.componentType,
                        replicas: service.replicas
                    };
                    const runtimeVersionOverride =
                        service.runtimeVersionOverride ?? this.runtimeVersionOverride;
                    if (runtimeVersionOverride !== undefined) {
                        serviceSpec.runtimeVersionOverride = runtimeVersionOverride;
                    }

                    // Add subComponentType if specified
                    if (service.subComponentType) {
                        serviceSpec.subComponentType = service.subComponentType;
                    }

                    // Add envFromSecret if specified
                    if (service.envFromSecret) {
                        serviceSpec.envFromSecret = service.envFromSecret;
                    }

                    // Add modelRef if modelName is specified
                    if (service.modelName) {
                        serviceSpec.modelRef = {
                            name: service.modelName
                        };
                    }

                    // Build extraPodSpec
                    const extraPodSpec: Partial<DynamoServiceSpec['extraPodSpec']> = {};
                    const mainContainer: Partial<DynamoServiceSpec['extraPodSpec']['mainContainer']> = {
                        image: service.image || this.image,
                        command: service.command || ['/bin/sh', '-c', 'sleep infinity']
                    };

                    if (service.args) {
                        mainContainer.args = service.args;
                    }

                    if (service.workingDir) {
                        mainContainer.workingDir = service.workingDir;
                    }

                    // Add resources if specified
                    if (service.resources) {
                        mainContainer.resources = {};
                        if (service.resources.limits) {
                            mainContainer.resources.limits = {};
                            if (service.resources.limits.gpu) {
                                mainContainer.resources.limits['nvidia.com/gpu'] =
                                    service.resources.limits.gpu;
                            }
                            if (service.resources.limits.cpu) {
                                mainContainer.resources.limits.cpu = service.resources.limits.cpu;
                            }
                            if (service.resources.limits.memory) {
                                mainContainer.resources.limits.memory = service.resources.limits.memory;
                            }
                        }
                        if (service.resources.requests) {
                            mainContainer.resources.requests = {};
                            if (service.resources.requests.gpu) {
                                mainContainer.resources.requests['nvidia.com/gpu'] =
                                    service.resources.requests.gpu;
                            }
                            if (service.resources.requests.cpu) {
                                mainContainer.resources.requests.cpu = service.resources.requests.cpu;
                            }
                            if (service.resources.requests.memory) {
                                mainContainer.resources.requests.memory = service.resources.requests.memory;
                            }
                        }
                    }

                    extraPodSpec.mainContainer =
                        mainContainer as DynamoServiceSpec['extraPodSpec']['mainContainer'];

                    // Remove name property if it exists (not needed for DGD spec)
                    delete (extraPodSpec.mainContainer as any).name;

                    // Add node affinity if nodepools are specified
                    const affinity = this.buildNodeAffinity();
                    if (affinity) {
                        extraPodSpec.affinity = affinity;
                    }

                    serviceSpec.extraPodSpec = extraPodSpec as DynamoServiceSpec['extraPodSpec'];

                    // Use the map key as the service name (append index if multiple configs)
                    const finalServiceName = serviceConfigs.length > 1 ? `${serviceName}-${i}` : serviceName;

                    servicesSpec[finalServiceName] = serviceSpec as DynamoServiceSpec;
                }
            }
        } else {
            // Backward compatibility: create default Frontend service
            const mainContainer: Partial<DynamoServiceSpec['extraPodSpec']['mainContainer']> = {
                image: this.image,
                command: DynamoDefaultCommand
            };

            if (this.resources && Object.keys(this.resources).length > 0) {
                mainContainer.resources = this.resources as any;
            }

            const extraPodSpec: Partial<DynamoServiceSpec['extraPodSpec']> = {
                mainContainer: mainContainer as DynamoServiceSpec['extraPodSpec']['mainContainer']
            };

            // Remove name property if it exists (not needed for DGD spec)
            delete (extraPodSpec.mainContainer as any).name;

            // Add node affinity if nodepools are specified
            const affinity = this.buildNodeAffinity();
            if (affinity) {
                extraPodSpec.affinity = affinity;
            }

            servicesSpec.Frontend = {
                dynamoNamespace: 'hello-world',
                componentType: 'frontend',
                replicas: this.replicas,
                ...(this.runtimeVersionOverride !== undefined
                    ? { runtimeVersionOverride: this.runtimeVersionOverride }
                    : {}),
                extraPodSpec: extraPodSpec as DynamoServiceSpec['extraPodSpec']
            };
        }

        this.spec = {
            services: servicesSpec
        };
    }

    getCrd(): CustomResourceDefinition {
        return crds.dynamoGraphDeploymentCrd;
    }

    getBody(): KubernetesObject {
        const result = this.getMetadata();
        result.metadata.annotations = {
            ...result.metadata.annotations,
            ...(this.customAnnotations || {})
        };
        result.metadata.labels = {
            ...result.metadata.labels,
            ...(this.customLabels || {})
        };
        result['spec'] = this.spec;
        return result;
    }
}

// A component of a v1beta1 DynamoGraphDeployment. v1beta1 replaced the v1alpha1
// `spec.services` map with a `spec.components` list, and `extraPodSpec.mainContainer`
// with a native pod template whose primary container is named "main".
export interface DynamoComponentConfig {
    name: string;
    type: string;
    runtimeVersionOverride?: string;
    replicas?: number;
    image?: string;
    command?: string[];
    resources?: JobResources;
    // The topology level this component must be packed into. The Dynamo webhook requires it
    // to be narrower than or equal to the deployment-level packDomain.
    packDomain?: string;
}

export interface DynamoTopologyConstraint {
    clusterTopologyName: string;
    packDomain?: string;
}

export class DynamoGraphDeploymentV1Beta1 extends BaseWorkload {
    spec: any;

    components: DynamoComponentConfig[];

    topologyConstraint?: DynamoTopologyConstraint;

    constructor(
        name: string,
        namespace: string,
        components: DynamoComponentConfig[],
        topologyConstraint?: DynamoTopologyConstraint,
        image = DynamoDefaultImage
    ) {
        super(name, namespace, image);
        this.components = components;
        this.topologyConstraint = topologyConstraint;
        this.initSpec();
    }

    initSpec(): void {
        this.spec = {
            components: this.components.map((component) => {
                const componentSpec: any = {
                    name: component.name,
                    type: component.type,
                    replicas: component.replicas ?? 1,
                    ...(component.runtimeVersionOverride !== undefined
                        ? { runtimeVersionOverride: component.runtimeVersionOverride }
                        : {}),
                    podTemplate: {
                        spec: {
                            containers: [
                                {
                                    name: 'main',
                                    image: component.image || this.image,
                                    command: component.command || DynamoDefaultCommand,
                                    ...(component.resources ? { resources: component.resources } : {})
                                }
                            ]
                        }
                    }
                };

                if (component.packDomain) {
                    componentSpec.topologyConstraint = { packDomain: component.packDomain };
                }

                return componentSpec;
            })
        };

        if (this.topologyConstraint) {
            this.spec.topologyConstraint = this.topologyConstraint;
        }
    }

    getCrd(): CustomResourceDefinition {
        return crds.dynamoGraphDeploymentV1Beta1Crd;
    }

    getBody(): KubernetesObject {
        const result = this.getMetadata();
        result['spec'] = this.spec;
        return result;
    }
}

export interface NIMServiceResources {
    requests?: { cpu?: string; memory?: string; 'nvidia.com/gpu'?: string };
    limits?: { cpu?: string; memory?: string; 'nvidia.com/gpu'?: string };
}

export class NIMServiceWorkload extends BaseWorkload {
    spec: any;

    replicas: number;

    authSecret: string;

    imageRepository: string;

    imageTag: string;

    servingPort: number;

    resources: NIMServiceResources;

    pvcSize: string;

    constructor(
        name: string,
        namespace: string,
        image: string,
        authSecret = 'ngc-secret',
        replicas = 1,
        servingPort = 8000,
        resources: NIMServiceResources = { requests: { cpu: '100m', memory: '10M' } },
        pvcSize = '10M'
    ) {
        super(name, namespace, image);
        const [repository, tag] = NIMServiceWorkload.splitImage(image);
        this.imageRepository = repository;
        this.imageTag = tag;
        this.authSecret = authSecret;
        this.replicas = replicas;
        this.servingPort = servingPort;
        this.resources = resources;
        this.pvcSize = pvcSize;
        this.initSpec();
    }

    // Split a "repo:tag" image reference while preserving a registry port (host:port/path[:tag]).
    private static splitImage(image: string): [string, string] {
        const lastSlash = image.lastIndexOf('/');
        const lastColon = image.lastIndexOf(':');
        if (lastColon > lastSlash) {
            return [image.substring(0, lastColon), image.substring(lastColon + 1)];
        }
        return [image, 'latest'];
    }

    initSpec(): void {
        this.spec = {
            schedulerName: 'runai-scheduler',
            image: {
                repository: this.imageRepository,
                tag: this.imageTag,
                pullPolicy: 'IfNotPresent'
            },
            authSecret: this.authSecret,
            storage: {
                pvc: {
                    create: true,
                    size: this.pvcSize,
                    volumeAccessMode: 'ReadWriteOnce'
                }
            },
            resources: this.resources,
            expose: {
                service: {
                    type: 'ClusterIP',
                    port: this.servingPort
                }
            },
            replicas: this.replicas
        };
    }

    getCrd(): CustomResourceDefinition {
        return crds.nimServiceCrd;
    }

    getBody(): KubernetesObject {
        const result = this.getMetadata();
        result['spec'] = this.spec;
        return result;
    }
}
