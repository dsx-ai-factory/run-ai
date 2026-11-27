export * from './applicationApi';
import { ApplicationApi } from './applicationApi';
export * from './authenticationApi';
import { AuthenticationApi } from './authenticationApi';
export * from './authorizationApi';
import { AuthorizationApi } from './authorizationApi';
export * from './clustersApi';
import { ClustersApi } from './clustersApi';
export * from './departmentsApi';
import { DepartmentsApi } from './departmentsApi';
export * from './internalApi';
import { InternalApi } from './internalApi';
export * from './jobsApi';
import { JobsApi } from './jobsApi';
export * from './organizationalApi';
import { OrganizationalApi } from './organizationalApi';
export * from './projectsApi';
import { ProjectsApi } from './projectsApi';
export * from './ssoApi';
import { SsoApi } from './ssoApi';
export * from './tenantApi';
import { TenantApi } from './tenantApi';
export * from './wandbApi';
import { WandbApi } from './wandbApi';
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

export const APIS = [ApplicationApi, AuthenticationApi, AuthorizationApi, ClustersApi, DepartmentsApi, InternalApi, JobsApi, OrganizationalApi, ProjectsApi, SsoApi, TenantApi, WandbApi];


