export * from './clusterObjectsStatusApi';
import { ClusterObjectsStatusApi } from './clusterObjectsStatusApi';
export * from './configMapsApi';
import { ConfigMapsApi } from './configMapsApi';
export * from './deviceClassesApi';
import { DeviceClassesApi } from './deviceClassesApi';
export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './pVCsApi';
import { PVCsApi } from './pVCsApi';
export * from './secretsApi';
import { SecretsApi } from './secretsApi';
export * from './storageClassConfigurationApi';
import { StorageClassConfigurationApi } from './storageClassConfigurationApi';
export * from './storageClassesApi';
import { StorageClassesApi } from './storageClassesApi';
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [ClusterObjectsStatusApi, ConfigMapsApi, DeviceClassesApi, HealthApi, PVCsApi, SecretsApi, StorageClassConfigurationApi, StorageClassesApi];


