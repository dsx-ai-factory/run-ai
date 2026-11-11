export * from './adminCLIApi';
import { AdminCLIApi } from './adminCLIApi';
export * from './auditApi';
import { AuditApi } from './auditApi';
export * from './nodeTypesApi';
import { NodeTypesApi } from './nodeTypesApi';
import * as http from 'http';

export class HttpError extends Error {
    constructor (public response: http.IncomingMessage, public body: any, public statusCode?: number) {
        super(`HTTP request failed with the following error: statusCode ${statusCode}, body: ${JSON.stringify(body)} fullResponse: ${JSON.stringify(response)}`);
        this.name = 'HttpError';
    }
}

export type { RequestFile } from '../model/models';

export const APIS = [AdminCLIApi, AuditApi, NodeTypesApi];
