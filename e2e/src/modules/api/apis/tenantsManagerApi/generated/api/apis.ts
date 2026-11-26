export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './internalApi';
import { InternalApi } from './internalApi';
export * from './logoApi';
import { LogoApi } from './logoApi';
export * from './tenantApi';
import { TenantApi } from './tenantApi';
export * from './tenantsApi';
import { TenantsApi } from './tenantsApi';
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [HealthApi, InternalApi, LogoApi, TenantApi, TenantsApi];


