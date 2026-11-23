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
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [EmailApi, HealthApi, NotificationChannelsApi, NotificationStateApi, NotificationTypesApi, SlackApi, SubscriptionsApi];


