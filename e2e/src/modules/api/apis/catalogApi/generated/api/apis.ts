export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './imagesApi';
import { ImagesApi } from './imagesApi';
export * from './inferencesApi';
import { InferencesApi } from './inferencesApi';
export * from './registriesApi';
import { RegistriesApi } from './registriesApi';
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

export const APIS = [HealthApi, ImagesApi, InferencesApi, RegistriesApi];


