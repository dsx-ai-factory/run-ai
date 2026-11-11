export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './settingsApi';
import { SettingsApi } from './settingsApi';
export * from './systemNotificationApi';
import { SystemNotificationApi } from './systemNotificationApi';
export * from './versionApi';
import { VersionApi } from './versionApi';
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [HealthApi, SettingsApi, SystemNotificationApi, VersionApi];


