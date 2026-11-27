export * from './departmentsApi';
import { DepartmentsApi } from './departmentsApi';
export * from './departmentsV1Api';
import { DepartmentsV1Api } from './departmentsV1Api';
export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './nodeTypesApi';
import { NodeTypesApi } from './nodeTypesApi';
export * from './notificationsApi';
import { NotificationsApi } from './notificationsApi';
export * from './orgUnitApi';
import { OrgUnitApi } from './orgUnitApi';
export * from './projectsApi';
import { ProjectsApi } from './projectsApi';
export * from './projectsV1Api';
import { ProjectsV1Api } from './projectsV1Api';
export * from './reportsApi';
import { ReportsApi } from './reportsApi';
export * from './settingsApi';
import { SettingsApi } from './settingsApi';
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [DepartmentsApi, DepartmentsV1Api, HealthApi, NodeTypesApi, NotificationsApi, OrgUnitApi, ProjectsApi, ProjectsV1Api, ReportsApi, SettingsApi];


