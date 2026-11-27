export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './nVIDIANIMApi';
import { NVIDIANIMApi } from './nVIDIANIMApi';
export * from './settingsApi';
import { SettingsApi } from './settingsApi';
export * from './workloadsV2Api';
import { WorkloadsV2Api } from './workloadsV2Api';
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [HealthApi, NVIDIANIMApi, SettingsApi, WorkloadsV2Api];


