export * from './emailApi';
import { EmailApi } from './emailApi';
export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './notificationChannelsApi';
import { NotificationChannelsApi } from './notificationChannelsApi';
export * from './notificationStateApi';
import { NotificationStateApi } from './notificationStateApi';
export * from './notificationTypesApi';
import { NotificationTypesApi } from './notificationTypesApi';
export * from './slackApi';
import { SlackApi } from './slackApi';
export * from './subscriptionsApi';
import { SubscriptionsApi } from './subscriptionsApi';
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

export const APIS = [EmailApi, HealthApi, NotificationChannelsApi, NotificationStateApi, NotificationTypesApi, SlackApi, SubscriptionsApi];


