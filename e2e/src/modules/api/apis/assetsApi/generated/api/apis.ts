export * from './assetsApi';
import { AssetsApi } from './assetsApi';
export * from './computeApi';
import { ComputeApi } from './computeApi';
export * from './configMapApi';
import { ConfigMapApi } from './configMapApi';
export * from './credentialsApi';
import { CredentialsApi } from './credentialsApi';
export * from './datasourcesApi';
import { DatasourcesApi } from './datasourcesApi';
export * from './environmentApi';
import { EnvironmentApi } from './environmentApi';
export * from './eventsApi';
import { EventsApi } from './eventsApi';
export * from './gitApi';
import { GitApi } from './gitApi';
export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './hostPathApi';
import { HostPathApi } from './hostPathApi';
export * from './inferencesApi';
import { InferencesApi } from './inferencesApi';
export * from './modelApi';
import { ModelApi } from './modelApi';
export * from './modelInferenceApi';
import { ModelInferenceApi } from './modelInferenceApi';
export * from './nFSApi';
import { NFSApi } from './nFSApi';
export * from './pVCApi';
import { PVCApi } from './pVCApi';
export * from './registryApi';
import { RegistryApi } from './registryApi';
export * from './s3Api';
import { S3Api } from './s3Api';
export * from './secretApi';
import { SecretApi } from './secretApi';
export * from './synchronizationApi';
import { SynchronizationApi } from './synchronizationApi';
export * from './templateApi';
import { TemplateApi } from './templateApi';
export * from './trainingsApi';
import { TrainingsApi } from './trainingsApi';
export * from './workloadApi';
import { WorkloadApi } from './workloadApi';
export * from './workspacesApi';
import { WorkspacesApi } from './workspacesApi';
import * as http from 'http';

export class HttpError extends Error {
    constructor (public response: http.IncomingMessage, public body: any, public statusCode?: number, requestOptions?: any) {
        /////Start: Modified by @Run.AI /////
        super(`HTTP request failed with the following statusCode ${statusCode},\n Request: ${JSON.stringify(requestOptions)},\n Response: ${JSON.stringify(body)}`);
        /////End: Modified by @Run.AI /////
        this.name = 'HttpError';
    }
}

export type { RequestFile } from '../model/models';

export const APIS = [AssetsApi, ComputeApi, ConfigMapApi, CredentialsApi, DatasourcesApi, EnvironmentApi, EventsApi, GitApi, HealthApi, HostPathApi, InferencesApi, ModelApi, ModelInferenceApi, NFSApi, PVCApi, RegistryApi, S3Api, SecretApi, SynchronizationApi, TemplateApi, TrainingsApi, WorkloadApi, WorkspacesApi];


