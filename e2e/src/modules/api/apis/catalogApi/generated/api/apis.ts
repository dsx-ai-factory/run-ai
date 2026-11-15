export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './helmReposApi';
import { HelmReposApi } from './helmReposApi';
export * from './imagesApi';
import { ImagesApi } from './imagesApi';
export * from './inferencesApi';
import { InferencesApi } from './inferencesApi';
export * from './registriesApi';
import { RegistriesApi } from './registriesApi';
export * from './settingsApi';
import { SettingsApi } from './settingsApi';
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [HealthApi, HelmReposApi, ImagesApi, InferencesApi, RegistriesApi, SettingsApi];


