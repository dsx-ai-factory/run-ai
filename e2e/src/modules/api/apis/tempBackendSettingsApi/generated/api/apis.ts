export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './internalApi';
import { InternalApi } from './internalApi';
export * from './tenantApi';
import { TenantApi } from './tenantApi';
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [HealthApi, InternalApi, TenantApi];


