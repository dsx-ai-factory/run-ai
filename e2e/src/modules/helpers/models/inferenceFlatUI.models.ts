/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    IComputeOptions,
    IDataSourceOptions,
    IEnvironmentOptions,
    IGeneralSectionOptions,
    IHuggingFaceModelOptions,
    IInferenceType,
    INimModelOptions
} from './inferenceBaseFlatUI.models';

export interface IInferenceMeta {
    clusterName: string;
    projectName: string;
    workloadName: string;
    templateName?: string;
    distributed?: boolean;
    inferenceType?: IInferenceType;
}

export interface IFlatWorkload {
    clusterName: string;
    projectName: string;
    workloadName: string;
    templateName?: string;
    inferenceType?: IInferenceType;
    environmentSection?: IEnvironmentOptions;
    computeSection?: IComputeOptions;
    dataSourceOptions?: IDataSourceOptions;
    generalSectionOptions?: IGeneralSectionOptions;
}

export interface IHuggingFaceFlatWorkload {
    clusterName: string;
    projectName: string;
    workloadName: string;
    inferenceType: IInferenceType;
    modelSection?: IHuggingFaceModelOptions;
    computeSection?: IComputeOptions;
    dataSourceOptions?: IDataSourceOptions;
    advancedSetup?: boolean;
}

export interface INimFlatWorkload {
    clusterName: string;
    projectName: string;
    workloadName: string;
    modelSection?: INimModelOptions;
    computeSection?: IComputeOptions;
    dataSourceOptions?: IDataSourceOptions;
    advancedSetup?: boolean;
}

export interface IUpdateWorkload {
    environmentSection?: IEnvironmentOptions;
    computeSection?: IComputeOptions;
    generalSectionOptions?: IGeneralSectionOptions;
}

export interface IDistributedInferenceWorkload {
    clusterName: string;
    projectName: string;
    workloadName: string;
    leader: {
        environmentSection?: IEnvironmentOptions;
        computeSection?: IComputeOptions;
        dataSourceOptions?: IDataSourceOptions;
        generalSectionOptions?: IGeneralSectionOptions;
    };
    worker?: {
        environmentSection?: IEnvironmentOptions;
        computeSection?: IComputeOptions;
        dataSourceOptions?: IDataSourceOptions;
        generalSectionOptions?: IGeneralSectionOptions;
    };
    workloadSetup?: {
        computeSection?: IComputeOptions;
        generalSectionOptions?: IGeneralSectionOptions;
    };
}
