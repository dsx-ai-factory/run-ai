export * from './datavolumesApi';
import { DatavolumesApi } from './datavolumesApi';
export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './settingsApi';
import { SettingsApi } from './settingsApi';
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [DatavolumesApi, HealthApi, SettingsApi];


