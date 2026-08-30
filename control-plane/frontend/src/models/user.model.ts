export enum EUserType {
  REGULAR_USER = "regular-user",
  SSO_USER = "sso-user",
}

export interface IMe {
  username: string;
  identityProvider: string;
  id: string;
  tenantId: number;
  firstName?: string;
  lastName?: string;
}
// this is how we get the values from the jwt
export interface IAuthProviderConfig {
  email: string;
  identity_provider: string;
  tenant_id: number;
  subject_uuid: string;
}

export interface IUser {
  email: string; // todo: remove - use username instead
  name: string; // todo: remove
  roles: Array<string>;
  entityType: EUserType; // todo: remove
  needToChangePassword: boolean;
  permitAllClusters: boolean;
  permittedClusters: [];
  username: string;
  tenantId?: number;
  userId?: string; // todo: remove - use id instead
  lastUpdated?: string;
  password?: string; // todo: remove - use tempPassword instead
  id: string;
  createdAt?: string;
  lastLogin?: string;
  createdBy?: string;
  updatedAt?: string;
  tempPassword?: string;
  isLocal?: boolean;
  identityProvider?: string;
  groups?: Array<string>;
  firstName?: string;
  lastName?: string;
}
