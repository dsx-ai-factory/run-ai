export * from './clustersApi';
import { ClustersApi } from './clustersApi';
export * from './deploymentsApi';
import { DeploymentsApi } from './deploymentsApi';
export * from './eventsApi';
import { EventsApi } from './eventsApi';
export * from './groupsApi';
import { GroupsApi } from './groupsApi';
export * from './jobsApi';
import { JobsApi } from './jobsApi';
export * from './nodePoolsApi';
import { NodePoolsApi } from './nodePoolsApi';
export * from './podGroupsApi';
import { PodGroupsApi } from './podGroupsApi';
export * from './podsApi';
import { PodsApi } from './podsApi';
export * from './servicesApi';
import { ServicesApi } from './servicesApi';
export * from './storageClassesApi';
import { StorageClassesApi } from './storageClassesApi';
export * from './trainingsApi';
import { TrainingsApi } from './trainingsApi';
export * from './usersApi';
import { UsersApi } from './usersApi';
export * from './workloadHistoryApi';
import { WorkloadHistoryApi } from './workloadHistoryApi';
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

export const APIS = [ClustersApi, DeploymentsApi, EventsApi, GroupsApi, JobsApi, NodePoolsApi, PodGroupsApi, PodsApi, ServicesApi, StorageClassesApi, TrainingsApi, UsersApi, WorkloadHistoryApi, WorkspacesApi];


