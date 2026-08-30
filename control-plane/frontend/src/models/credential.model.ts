import { CREDENTIAL_ROUTE_NAMES } from "@/router/credential.routes/credential.routes.names";
import type { IScopeModel, IUIAssetCreationMeta } from "./global.model";
import type {
  Scope,
  AccessKeyUpdateRequest,
  PasswordUpdateRequest,
  DockerRegistryUpdateRequest,
  GenericSecretUpdateRequest,
  NgcApiKeyUpdateRequest,
  AccessKeyCreationSpec,
  DockerRegistryCreationSpec,
  PasswordCreationSpec,
  GenericSecretCreationSpec,
  NgcApiKeyCreationSpec,
} from "@/swagger-models/assets-service-client";
import { AssetKind } from "@/swagger-models/assets-service-client";
import { SecretType } from "@/swagger-models/k8s-objects-tracker-client";

enum CREDENTIAL_UI_LABELS {
  DOCKER_REGISTRY = "Docker registry",
  GENERIC_SECRET = "Generic secret",
  NGC_API_KEY = "NGC API key",
}

export const credentialTextForUI: Record<TCredentialKinds, string> = {
  accessKey: CREDENTIAL_UI_LABELS.GENERIC_SECRET,
  password: CREDENTIAL_UI_LABELS.GENERIC_SECRET,
  dockerRegistry: CREDENTIAL_UI_LABELS.DOCKER_REGISTRY,
  genericSecret: CREDENTIAL_UI_LABELS.GENERIC_SECRET,
  ngcApiKey: CREDENTIAL_UI_LABELS.NGC_API_KEY,
};

export type TCredentialKinds =
  | (typeof AssetKind)["AccessKey"]
  | (typeof AssetKind)["Password"]
  | (typeof AssetKind)["DockerRegistry"]
  | (typeof AssetKind)["GenericSecret"]
  | (typeof AssetKind)["NgcApiKey"];

export type TCredentialUpdateRequests =
  | AccessKeyUpdateRequest
  | PasswordUpdateRequest
  | DockerRegistryUpdateRequest
  | GenericSecretUpdateRequest
  | NgcApiKeyUpdateRequest;

export type TCredentialCreationRoutes =
  | typeof CREDENTIAL_ROUTE_NAMES.CREDENTIAL_ACCESS_KEY_NEW
  | typeof CREDENTIAL_ROUTE_NAMES.CREDENTIAL_USERPASS_NEW
  | typeof CREDENTIAL_ROUTE_NAMES.CREDENTIAL_DOCKER_REGISTRY_NEW
  | typeof CREDENTIAL_ROUTE_NAMES.CREDENTIAL_GENERIC_SECRET_NEW
  | typeof CREDENTIAL_ROUTE_NAMES.CREDENTIAL_NGC_API_KEY_NEW;

export const CREDENTIAL_CREATION_ROUTES_TO_CREDENTIAL_KINDS: Record<TCredentialCreationRoutes, TCredentialKinds> = {
  [CREDENTIAL_ROUTE_NAMES.CREDENTIAL_ACCESS_KEY_NEW]: AssetKind.AccessKey,
  [CREDENTIAL_ROUTE_NAMES.CREDENTIAL_USERPASS_NEW]: AssetKind.Password,
  [CREDENTIAL_ROUTE_NAMES.CREDENTIAL_DOCKER_REGISTRY_NEW]: AssetKind.DockerRegistry,
  [CREDENTIAL_ROUTE_NAMES.CREDENTIAL_GENERIC_SECRET_NEW]: AssetKind.GenericSecret,
  [CREDENTIAL_ROUTE_NAMES.CREDENTIAL_NGC_API_KEY_NEW]: AssetKind.NgcApiKey,
};

export const SECRET_TYPE_TO_ASSET_KIND: Record<SecretType, TCredentialKinds> = {
  [SecretType.AccessKey]: AssetKind.AccessKey,
  [SecretType.Password]: AssetKind.Password,
  [SecretType.Generic]: AssetKind.GenericSecret,
  [SecretType.DockerRegistry]: AssetKind.DockerRegistry,
};

export const TYPE_SSH_KEY = "generic-ssh-key";

export type TGenericFormats =
  | typeof AssetKind.GenericSecret
  | typeof AssetKind.AccessKey
  | typeof AssetKind.Password
  | typeof TYPE_SSH_KEY;

export interface IScopeCredentialsList extends Omit<IScopeModel, "scope" | "clusterId"> {
  scope?: Scope;
  clusterUuid?: string;
}
export interface IUIAccessKeyCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: AccessKeyCreationSpec;
}
export interface IUIDockerRegistryCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: DockerRegistryCreationSpec;
}
export interface IUINgcApiKeyCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: NgcApiKeyCreationSpec;
}
export interface IUIPasswordCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: PasswordCreationSpec;
}

export interface IUIGeneralCredentialCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: GenericSecretCreationSpec;
}

export interface INewGeneralSecret {
  key: string;
  value: string;
}

export interface IUIGenericSecretCreationRequest {
  meta: IUIAssetCreationMeta;
  spec: IUIGenericSecretSpecDetails;
}
export interface IUIGenericSecretSpecDetails
  extends GenericSecretCreationSpec,
    PasswordCreationSpec,
    AccessKeyCreationSpec {}
