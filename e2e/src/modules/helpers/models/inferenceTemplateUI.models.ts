/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    IComputeOptions,
    IDataSourceOptions,
    IEnvironmentOptions,
    IGeneralSectionOptions,
    IInferenceType,
    IHuggingFaceModelOptions,
    INimModelOptions
} from './inferenceBaseFlatUI.models';

export interface IInferenceTemplateMeta {
    clusterName?: string;
    projectName?: string;
    inferenceType: IInferenceType;
    templateName: string;
}

export interface IInferenceTemplate {
    templateMeta: IInferenceTemplateMeta;
    environmentSection?: IEnvironmentOptions;
    computeSection?: IComputeOptions;
    dataSourceOptions?: IDataSourceOptions;
    generalSectionOptions?: IGeneralSectionOptions;
}

export interface INimTemplate {
    templateMeta: IInferenceTemplateMeta;
    modelSection: INimModelOptions;
    computeSection?: IComputeOptions;
    dataSourceOptions?: IDataSourceOptions;
    advancedSetup?: boolean;
}

export interface IHuggingFaceTemplate {
    templateMeta: IInferenceTemplateMeta;
    modelSection: IHuggingFaceModelOptions;
    computeSection?: IComputeOptions;
    dataSourceOptions?: IDataSourceOptions;
    advancedSetup?: boolean;
}
