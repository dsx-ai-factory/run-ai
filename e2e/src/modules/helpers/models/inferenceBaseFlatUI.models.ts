/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Scope } from '../../api/apis/assetsApi';

export interface IAutoScalingOptions {
    minReplicas?: number;
    maxReplicas?: number;
    thresholdValue?: number;
    autoScalingPeriod?: string;
}

export interface IPVCOptions {
    name: string;
    scope?: Scope;
    scopeEntityName?: string;
    storageClass?: string;
    accessMode?: string;
    claimSize?: string;
    claimUnit?: string;
    containerPath?: string;
    volumeMode?: string;
}

export enum IServingPortAccessOptions {
    Public = 'public',
    Everyone = 'everyone',
    Group = 'group',
    Users = 'users'
}

export interface IServingPortOptions {
    protocol?: string;
    port?: number;
    access?: {
        option: IServingPortAccessOptions;
    };
}

export enum EImagePullPolicyOptions {
    ImageOnly = 'Pull the image only',
    ImageFromRegistry = 'Always pull the image from the registry',
    Never = 'Never pull (if the image is already present on the host, it will be used)'
}

export enum IInferenceType {
    Custom = 'custom',
    VLLM = 'vllm',
    TGI = 'tgi',
    Nim = 'nim'
}

export enum EPersistenceOption {
    Persistent = 'persistent',
    Ephemeral = 'ephemeral'
}

export enum EVolumeAccessMode {
    ReadWriteByOneNode = 'Read-write by one node',
    ReadOnlyByManyNodes = 'Read-only by many nodes',
    ReadWriteByManyNodes = 'Read-write by many nodes'
}

export interface IVolumeOptions {
    accessMode: EVolumeAccessMode;
    sizeUnit: string;
    path: string;
    persistenceOption?: EPersistenceOption;
}

export enum EEnvironmentVariableSource {
    Custom = 'Custom',
    Credentials = 'Credentials',
    ConfigMap = 'ConfigMap'
}

export enum EModelSourceForNim {
    HuggingFace = 'Hugging Face',
    Nim = 'NVIDIA NGC catalog '
}

export interface IEnvironmentVariableEntry {
    name: string;
    source?: EEnvironmentVariableSource; // defaults to 'Custom'
    value: string;
}
export interface IEnvironmentOptions {
    image?: string;
    imagePullPolicy?: EImagePullPolicyOptions;
    tool?: string;
    runtimeSettings?: {
        isEditCmd?: boolean;
        cmd?: string;
        args?: string;
        environmentVariablesValues?: Array<string>;
        environmentVariablesValuesByName?: Map<string, string>;
        fullEnvironmentVariables?: Array<IEnvironmentVariableEntry>;
    };
    servingPort?: IServingPortOptions;
    assetName?: string;
    newEnvironmentResource?: {
        assetName: string;
        projectName: string;
        image: string;
        servingPort: { // in environment this is required
            container: number;
        };
        runtimeSettings?: {
            cmd?: string;
            args?: string;
        };
    };
}

export interface IComputeOptions {
    newComputeResource?: {
        assetName: string;
        projectName: string;
        gpu?: {
            devicesInput?: number;
        };
        isProjectScope?: boolean;
    };
    gpu?: {
        devicesInput?: string;
        memoryRequest?: string;
    };
    cpu?: {
        compute?: {
            option?: 'Cores' | 'Millicores';
            request?: string;
            limit?: string;
        };
        memory?: {
            option?: 'MB' | 'GB';
            request?: string;
            limit?: string;
        };
    };
    extendedResource?: {
        increaseSharedMemory?: boolean;
        resources?: Array<{ resource: string; quantity: string }>;
    };
    autoScaling?: IAutoScalingOptions;
    assetName?: string;
    waitForUnifiedRequest?: boolean; // set to false only when you select an existing asset and the unified request is not needed
    nodes?: {
        nodeType?: string;
    };
}

export interface IDataSourceOptions {
    assetNames?: Array<string>;
    waitForUnifiedRequest?: boolean; // set to false only when you select an existing asset and the unified request is not needed
    nfs?: string;
    volume?: IVolumeOptions;
    pvc?: IPVCOptions;
}

export interface IGeneralSectionOptions {
    annotations?: Array<{ name: string; value: string }>;
    labels?: Array<{ name: string; value: string }>;
    initializationTimeout?: {
        minutes?: number;
        seconds?: number;
    };
    requestTimeout?: {
        minutes?: number;
        seconds?: number;
    };
    preemptibility?: string;
    startupPolicy?: string;
    priority?: string;
}

export enum ECredentialsSource {
    Shared = 'shared',
    Private = 'private'
}

export enum ECredentialsSourceType {
    NgcApiKey = 'ngc-api-key',
    GenericSecret = 'generic-secret',
}

export interface ICredentials {
    source: ECredentialsSource;
    sourceType?: ECredentialsSourceType;
    secretName: string;
    sourceDisabled?: boolean;
}

export interface IHuggingFaceModelOptions {
    modelName?: string;
    hfToken?: string;
    credentials?: ICredentials;
    servingPort?: IServingPortOptions;
    shouldOpenServingPortSection?: boolean;
}

export interface INimModelOptions {
    modelSource?: EModelSourceForNim;
    modelName?: string;
    hfToken?: string;
    apiKey?: string;
    fromStorageOnly?: boolean;
    profileOrHash?: string;
    servingPort?: IServingPortOptions;
    credentials?: ICredentials;
    nimPullImageCredential?: {
        dockerRegistry?: string;
        ngcApiKey?: string;
    }
}
