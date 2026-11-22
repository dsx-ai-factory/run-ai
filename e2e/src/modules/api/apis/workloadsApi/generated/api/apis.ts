export * from './aIApplicationsApi';
import { AIApplicationsApi } from './aIApplicationsApi';
export * from './distributedApi';
import { DistributedApi } from './distributedApi';
export * from './distributedInferencesApi';
import { DistributedInferencesApi } from './distributedInferencesApi';
export * from './eventsApi';
import { EventsApi } from './eventsApi';
export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './inferencesApi';
import { InferencesApi } from './inferencesApi';
export * from './metricsApi';
import { MetricsApi } from './metricsApi';
export * from './podGroupsApi';
import { PodGroupsApi } from './podGroupsApi';
export * from './podsApi';
import { PodsApi } from './podsApi';
export * from './revisionsApi';
import { RevisionsApi } from './revisionsApi';
export * from './settingsApi';
import { SettingsApi } from './settingsApi';
export * from './syncApi';
import { SyncApi } from './syncApi';
export * from './trainingsApi';
import { TrainingsApi } from './trainingsApi';
export * from './unifiedApi';
import { UnifiedApi } from './unifiedApi';
export * from './workloadPropertiesApi';
import { WorkloadPropertiesApi } from './workloadPropertiesApi';
export * from './workloadTemplatesApi';
import { WorkloadTemplatesApi } from './workloadTemplatesApi';
export * from './workloadsApi';
import { WorkloadsApi } from './workloadsApi';
export * from './workloadsBatchApi';
import { WorkloadsBatchApi } from './workloadsBatchApi';
export * from './workspacesApi';
import { WorkspacesApi } from './workspacesApi';
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [AIApplicationsApi, DistributedApi, DistributedInferencesApi, EventsApi, HealthApi, InferencesApi, MetricsApi, PodGroupsApi, PodsApi, RevisionsApi, SettingsApi, SyncApi, TrainingsApi, UnifiedApi, WorkloadPropertiesApi, WorkloadTemplatesApi, WorkloadsApi, WorkloadsBatchApi, WorkspacesApi];


