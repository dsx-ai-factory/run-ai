export * from './accessKeysApi';
import { AccessKeysApi } from './accessKeysApi';
export * from './applicationsApi';
import { ApplicationsApi } from './applicationsApi';
export * from './clientsApi';
import { ClientsApi } from './clientsApi';
export * from './configApi';
import { ConfigApi } from './configApi';
export * from './healthApi';
import { HealthApi } from './healthApi';
export * from './idpsApi';
import { IdpsApi } from './idpsApi';
export * from './loginsApi';
import { LoginsApi } from './loginsApi';
export * from './logoutApi';
import { LogoutApi } from './logoutApi';
export * from './openIDConnectApi';
import { OpenIDConnectApi } from './openIDConnectApi';
export * from './realmApi';
import { RealmApi } from './realmApi';
export * from './serviceAccountsApi';
import { ServiceAccountsApi } from './serviceAccountsApi';
export * from './settingsApi';
import { SettingsApi } from './settingsApi';
export * from './tokensApi';
import { TokensApi } from './tokensApi';
export * from './userApplicationsApi';
import { UserApplicationsApi } from './userApplicationsApi';
export * from './usersApi';
import { UsersApi } from './usersApi';
import { HttpError } from './httpError';
export { HttpError };

export type { RequestFile } from '../model/models';

export const APIS = [AccessKeysApi, ApplicationsApi, ClientsApi, ConfigApi, HealthApi, IdpsApi, LoginsApi, LogoutApi, OpenIDConnectApi, RealmApi, ServiceAccountsApi, SettingsApi, TokensApi, UserApplicationsApi, UsersApi];


