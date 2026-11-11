import { ClusterObjectsStatusApi } from './clusterObjectsStatusApi';
import { ConfigMapsApi } from './configMapsApi';
import { HealthApi } from './healthApi';
import { PVCsApi } from './pVCsApi';
import { SecretsApi } from './secretsApi';
import { StorageClassConfigurationApi } from './storageClassConfigurationApi';
import { StorageClassesApi } from './storageClassesApi';

export type { RequestFile } from '../model/models';

export const APIS = [ClusterObjectsStatusApi, ConfigMapsApi, HealthApi, PVCsApi, SecretsApi, StorageClassConfigurationApi, StorageClassesApi];

// Re-export everything at the end to avoid circular dependencies
export * from './clusterObjectsStatusApi';
export * from './configMapsApi';
export * from './healthApi';
export * from './pVCsApi';
export * from './secretsApi';
export * from './storageClassConfigurationApi';
export * from './storageClassesApi';