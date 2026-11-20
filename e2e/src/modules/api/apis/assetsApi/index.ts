/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
    ComputeApi,
    DatasourcesApi,
    EnvironmentApi,
    GitApi,
    HostPathApi,
    NFSApi,
    PVCApi,
    S3Api,
    CredentialsApi,
    WorkspacesApi,
    TemplateApi,
    TrainingsApi,
    InferencesApi,
    ModelApi,
    ModelInferenceApi,
    ConfigMapApi,
    SecretApi
} from './generated/api';
import { API_DEFAULT_HEADERS } from '../apiDefaults';

export * from './generated/model/models';

export class AssetsApi {
    computeApi: ComputeApi;

    datasourcesApi: DatasourcesApi;

    environmentApi: EnvironmentApi;

    gitApi: GitApi;

    hostPathApi: HostPathApi;

    nFSApi: NFSApi;

    pVCApi: PVCApi;

    s3Api: S3Api;

    configMapApi: ConfigMapApi;

    secretApi: SecretApi;

    credentialsApi: CredentialsApi;

    workspaceApi: WorkspacesApi;

    templateApi: TemplateApi;

    trainingApi: TrainingsApi;

    inferenceApi: InferencesApi;

    modelApi: ModelApi;

    modelInferenceApi: ModelInferenceApi;

    constructor(url: string) {
        this.computeApi = new ComputeApi(url);
        this.computeApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.datasourcesApi = new DatasourcesApi(url);
        this.datasourcesApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.environmentApi = new EnvironmentApi(url);
        this.environmentApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.gitApi = new GitApi(url);
        this.gitApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.hostPathApi = new HostPathApi(url);
        this.hostPathApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.nFSApi = new NFSApi(url);
        this.nFSApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.pVCApi = new PVCApi(url);
        this.pVCApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.s3Api = new S3Api(url);
        this.s3Api.defaultHeaders = API_DEFAULT_HEADERS;

        this.configMapApi = new ConfigMapApi(url);
        this.configMapApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.secretApi = new SecretApi(url);
        this.secretApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.credentialsApi = new CredentialsApi(url);
        this.credentialsApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.workspaceApi = new WorkspacesApi(url);
        this.workspaceApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.templateApi = new TemplateApi(url);
        this.templateApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.trainingApi = new TrainingsApi(url);
        this.trainingApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.inferenceApi = new InferencesApi(url);
        this.inferenceApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.modelApi = new ModelApi(url);
        this.modelApi.defaultHeaders = API_DEFAULT_HEADERS;

        this.modelInferenceApi = new ModelInferenceApi(url);
        this.modelInferenceApi.defaultHeaders = API_DEFAULT_HEADERS;
    }
}
