// types
import {
  AssetKind,
  type CredentialsListResponseEntry,
  type PasswordCreationRequest,
  type AccessKeyCreationRequest,
  type AccessKey,
  type Password,
  type PasswordUpdateRequest,
  type AccessKeyUpdateRequest,
  type DockerRegistryCreationRequest,
  type DockerRegistry,
  type DockerRegistryUpdateRequest,
  type HttpResponse,
  type GenericSecretCreationRequest,
  type GenericSecret,
  type GenericSecretUpdateRequest,
  type KeyValuePair,
  type NgcApiKeyCreationRequest,
  type AssetMeta,
  type NgcApiKey,
  type NgcApiKeyUpdateRequest,
} from "@/swagger-models/assets-service-client";
import type { SecretListInfo, SecretType } from "@/swagger-models/k8s-objects-tracker-client";
import {
  type TCredentialKinds,
  type IUIGenericSecretCreationRequest,
  type TGenericFormats,
  TYPE_SSH_KEY,
} from "@/models/credential.model";
import { assetsServiceApi } from "@/services/infra/client-apis/assets-service-api/assets-service-api";
import { httpService } from "@/services/infra/https.service/http.service";

// constants
import type { IAssetsFilter } from "@/models/filter.model";
import { assetsUtil } from "@/utils/assets.util/assets.util";

// services
import { pick } from "@/utils/common.util";
import { k8sObjectTrackerApi } from "@/services/infra/client-apis/k8s-object-tracker-api/k8s-object-tracker-api";

export const credentialService = {
  listCredentials,
  remove,
  getEmptyDockerRegistryModel,
  getEmptyUserpassModel,
  getEmptyAccessKeyModel,
  getEmptyGenericSecretCredentialModel,
  getEmptyNgcApiKeyModel,
  createDockerRegistry,
  createUserpass,
  createAccessKey,
  createGeneralCredential,
  createNgcApiKey,
  updateUserpass,
  updateAccessKey,
  updateDockerRegistry,
  updateGenericSecret,
  updateNgcApiKey,
  listExistingSecrets,
  getCredentialById,
  prepareGenericSecretForServerRequest,
  getEmptyGenericFormatSecretModel,
  createFormatBasedGenericSecret,
  listGenericSecret,
  listNgcApiKey,
  listDockerRegistries,
};

async function listCredentials(filterBy: IAssetsFilter): Promise<CredentialsListResponseEntry[]> {
  const filters: IAssetsFilter = pick(
    filterBy,
    "projectId",
    "scope",
    "usageInfo",
    "clusterUuid",
    "statusInfo",
    "privateOnly",
    "includePrivate",
  );
  try {
    return assetsServiceApi.credentialApi
      .listCredentialsAssets(
        undefined,
        filters.scope,
        Number(filters.projectId) || undefined,
        filters.departmentId || undefined,
        filters.clusterUuid?.toString(),
        filters.usageInfo,
        filters.statusInfo,
        filters.privateOnly,
        filters.includePrivate,
      )
      .then((res) => res.data.entries);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listExistingSecrets(
  scopeType: string,
  scopeId: string,
  type?: SecretType,
  origin?: "all" | "assets" | "existing",
  includeAncestors = false,
): Promise<SecretListInfo[]> {
  try {
    const response = await k8sObjectTrackerApi.secretsApi.listSecrets(
      scopeType,
      scopeId,
      undefined,
      type,
      undefined,
      origin,
      includeAncestors,
    );
    return response.data.secrets;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listGenericSecret(filterBy: IAssetsFilter): Promise<GenericSecret[]> {
  const filters: IAssetsFilter = pick(
    filterBy,
    "projectId",
    "scope",
    "usageInfo",
    "clusterUuid",
    "statusInfo",
    "privateOnly",
  );
  try {
    return assetsServiceApi.credentialApi
      .listGenericSecret(
        undefined,
        filters.scope,
        Number(filters.projectId) || undefined,
        filters.departmentId || undefined,
        filters.clusterUuid?.toString(),
        filters.usageInfo,
        filters.statusInfo,
        filters.privateOnly,
      )
      .then((res) => res.data.entries || []);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listNgcApiKey(filterBy: IAssetsFilter): Promise<NgcApiKey[]> {
  const filters: IAssetsFilter = pick(
    filterBy,
    "projectId",
    "scope",
    "usageInfo",
    "clusterUuid",
    "departmentId",
    "statusInfo",
    "privateOnly",
    "includePrivate",
  );
  try {
    return assetsServiceApi.credentialApi
      .listNgcApiKey(
        undefined,
        filters.scope,
        Number(filters.projectId) || undefined,
        filters.departmentId || undefined,
        filters.clusterUuid?.toString(),
        filters.usageInfo,
        filters.statusInfo,
        filters.privateOnly,
        filters.includePrivate,
      )
      .then((res) => res.data.entries || []);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listDockerRegistries(filterBy: IAssetsFilter): Promise<DockerRegistry[]> {
  const filters: IAssetsFilter = pick(
    filterBy,
    "projectId",
    "scope",
    "usageInfo",
    "clusterUuid",
    "statusInfo",
    "privateOnly",
  );
  try {
    return assetsServiceApi.credentialApi
      .listDockerRegistries(
        undefined,
        filters.scope,
        Number(filters.projectId) || undefined,
        filters.departmentId || undefined,
        filters.clusterUuid?.toString(),
        filters.usageInfo,
        filters.statusInfo,
        filters.privateOnly,
      )
      .then((res) => res.data.entries || []);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function remove(credentialId: string, kind: TCredentialKinds): Promise<HttpResponse | void> {
  switch (kind) {
    case AssetKind.AccessKey:
      return removeAccessKey(credentialId);
    case AssetKind.DockerRegistry:
      return removeDockerRegistry(credentialId);
    case AssetKind.Password:
      return removeUserpass(credentialId);
    case AssetKind.GenericSecret:
      return removeGenericSecret(credentialId);
    case AssetKind.NgcApiKey:
      return removeNgcApiKey(credentialId);
    default:
      throw `Wrong credential type: ${kind}`;
  }
}
interface ICredentialGetByIdParams {
  usageInfo?: boolean;
  statusInfo?: boolean;
}

async function getCredentialById(
  credentialId: string,
  kind: TCredentialKinds,
  options: ICredentialGetByIdParams = {},
): Promise<DockerRegistry | AccessKey | Password | GenericSecret> {
  switch (kind) {
    case AssetKind.AccessKey:
      return getAccesskeyById(credentialId, options);
    case AssetKind.DockerRegistry:
      return getDockerRegistryById(credentialId, options);
    case AssetKind.Password:
      return getUserpassById(credentialId, options);
    case AssetKind.GenericSecret:
      return getGenericSecretById(credentialId, options);
    default:
      throw `Wrong credential type: ${kind}`;
  }
}

async function getDockerRegistryById(
  dockerRegistryId: string,
  options: ICredentialGetByIdParams = {},
): Promise<DockerRegistry> {
  try {
    const params = pick(options, "usageInfo", "statusInfo");

    const response = await assetsServiceApi.credentialApi.getDockerRegistryById(
      dockerRegistryId,
      params.usageInfo,
      params.statusInfo,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function getAccesskeyById(accesskeyId: string, options: ICredentialGetByIdParams = {}): Promise<AccessKey> {
  try {
    const params = pick(options, "usageInfo", "statusInfo");

    const response = await assetsServiceApi.credentialApi.getAccessKeyById(
      accesskeyId,
      params.usageInfo,
      params.statusInfo,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
async function getUserpassById(userpassId: string, options: ICredentialGetByIdParams = {}): Promise<Password> {
  try {
    const params = pick(options, "usageInfo", "statusInfo");

    const response = await assetsServiceApi.credentialApi.getPasswordById(
      userpassId,
      params.usageInfo,
      params.statusInfo,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getGenericSecretById(
  genericSecretId: string,
  options: ICredentialGetByIdParams = {},
): Promise<GenericSecret> {
  try {
    const params = pick(options, "usageInfo", "statusInfo");

    const response = await assetsServiceApi.credentialApi.getGenericSecretById(
      genericSecretId,
      params.usageInfo,
      params.statusInfo,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function removeDockerRegistry(dockerRegistryId: string): Promise<HttpResponse> {
  try {
    const response = await assetsServiceApi.credentialApi.deleteDockerRegistry(dockerRegistryId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function removeAccessKey(accesskeyId: string): Promise<HttpResponse> {
  try {
    const response = await assetsServiceApi.credentialApi.deleteAccessKey(accesskeyId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function removeUserpass(userpassId: string): Promise<HttpResponse> {
  try {
    const response = await assetsServiceApi.credentialApi.deletePassword(userpassId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function removeGenericSecret(genericSecretId: string): Promise<void> {
  try {
    const response = await assetsServiceApi.credentialApi.deleteGenericSecret(genericSecretId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createDockerRegistry(dockerRegistry: DockerRegistryCreationRequest): Promise<DockerRegistry> {
  try {
    const response = await assetsServiceApi.credentialApi.createDockerRegistry(dockerRegistry);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createUserpass(userpass: PasswordCreationRequest): Promise<Password> {
  try {
    const response = await assetsServiceApi.credentialApi.createPassword(userpass);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createAccessKey(accessKey: AccessKeyCreationRequest): Promise<AccessKey> {
  try {
    const response = await assetsServiceApi.credentialApi.createAccessKey(accessKey);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createGeneralCredential(generalCredential: GenericSecretCreationRequest): Promise<GenericSecret> {
  try {
    const response = await assetsServiceApi.credentialApi.createGenericSecret(generalCredential);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateUserpass(userpassId: string, userpass: PasswordUpdateRequest): Promise<Password> {
  try {
    const response = await assetsServiceApi.credentialApi.updatePassword(userpassId, userpass);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateAccessKey(accessKeyId: string, accessKey: AccessKeyUpdateRequest): Promise<AccessKey> {
  try {
    const response = await assetsServiceApi.credentialApi.updateAccessKey(accessKeyId, accessKey);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateDockerRegistry(
  dockerRegistryId: string,
  dockerRegistry: DockerRegistryUpdateRequest,
): Promise<DockerRegistry> {
  try {
    const response = await assetsServiceApi.credentialApi.updateDockerRegistry(dockerRegistryId, dockerRegistry);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateGenericSecret(
  genericSecretId: string,
  genericSecret: GenericSecretUpdateRequest,
): Promise<GenericSecret> {
  try {
    const response = await assetsServiceApi.credentialApi.updateGenericSecret(genericSecretId, genericSecret);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateNgcApiKey(ngcApiKeyId: string, ngcApiKey: NgcApiKeyUpdateRequest): Promise<NgcApiKey> {
  try {
    const response = await assetsServiceApi.credentialApi.updateNgcApiKey(ngcApiKeyId, ngcApiKey);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createNgcApiKey(ngcApiKey: NgcApiKeyCreationRequest): Promise<NgcApiKey> {
  try {
    const response = await assetsServiceApi.credentialApi.createNgcApiKey(ngcApiKey);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function removeNgcApiKey(ngcApiKeyId: string): Promise<void> {
  try {
    const response = await assetsServiceApi.credentialApi.deleteNgcApiKey(ngcApiKeyId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
function getEmptyDockerRegistryModel(): DockerRegistryCreationRequest {
  return {
    meta: assetsUtil.getEmptyAssetMeta() as AssetMeta,
    spec: {
      existingSecretName: null,
      password: "",
      user: "",
      url: "",
    },
  };
}

function getEmptyUserpassModel(): PasswordCreationRequest {
  return {
    meta: assetsUtil.getEmptyAssetMeta() as AssetMeta,
    spec: {
      existingSecretName: null,
      password: "",
      user: "",
    },
  };
}

function getEmptyAccessKeyModel(): AccessKeyCreationRequest {
  return {
    meta: assetsUtil.getEmptyAssetMeta() as AssetMeta,
    spec: {
      existingSecretName: null,
      accessKeyId: "",
      secretAccessKey: "",
    },
  };
}

function getEmptyGenericSecretCredentialModel(): GenericSecretCreationRequest {
  return {
    meta: assetsUtil.getEmptyAssetMeta() as AssetMeta,
    spec: {
      existingSecretName: "",
    },
  };
}

function getEmptyGenericFormatSecretModel(): IUIGenericSecretCreationRequest {
  return {
    meta: assetsUtil.getEmptyAssetMeta() as AssetMeta,
    spec: {},
  };
}

function getEmptyNgcApiKeyModel(): NgcApiKeyCreationRequest {
  return {
    meta: assetsUtil.getEmptyAssetMeta() as AssetMeta,
    spec: {
      ngc_api_key: "",
    },
  };
}

function prepareGenericSecretForServerRequest(request: GenericSecretCreationRequest): GenericSecretCreationRequest {
  const requestCopy = JSON.parse(JSON.stringify(request));

  if (requestCopy.spec?.keyValuePairs) {
    requestCopy.spec.keyValuePairs.forEach((pair: KeyValuePair) => {
      if (pair.value && pair.value.includes("\n") && !pair.value.endsWith("\n")) {
        pair.value += "\n";
      }
    });
  }

  return requestCopy;
}

function prepareFormatBasedGenericSecretForServerRequest(
  request: IUIGenericSecretCreationRequest,
): GenericSecretCreationRequest {
  const genSec: GenericSecretCreationRequest = { meta: request.meta as AssetMeta, spec: {} };
  if (request.spec.existingSecretName) {
    genSec.spec.existingSecretName = request.spec.existingSecretName;
  } else {
    genSec.spec.keyValuePairs = request.spec.keyValuePairs;
  }

  return prepareGenericSecretForServerRequest(genSec);
}

function prepareAccessKeyForServerRequest(request: IUIGenericSecretCreationRequest): AccessKeyCreationRequest {
  const accessKey: AccessKeyCreationRequest = { meta: request.meta as AssetMeta, spec: {} };
  if (request.spec.existingSecretName) {
    accessKey.spec.existingSecretName = request.spec.existingSecretName;
  } else {
    accessKey.spec.accessKeyId = request.spec.accessKeyId;
    accessKey.spec.secretAccessKey = request.spec.secretAccessKey;
  }

  return accessKey;
}

function prepareUserpassForServerRequest(request: IUIGenericSecretCreationRequest): PasswordCreationRequest {
  const userpass: PasswordCreationRequest = { meta: request.meta as AssetMeta, spec: {} };
  if (request.spec.existingSecretName) {
    userpass.spec.existingSecretName = request.spec.existingSecretName;
  } else {
    userpass.spec.password = request.spec.password;
    userpass.spec.user = request.spec.user;
  }
  return userpass;
}

async function createFormatBasedGenericSecret(model: IUIGenericSecretCreationRequest, type: TGenericFormats) {
  switch (type) {
    case AssetKind.AccessKey:
      return createAccessKey(prepareAccessKeyForServerRequest(model));
    case AssetKind.Password:
      return createUserpass(prepareUserpassForServerRequest(model));
    case AssetKind.GenericSecret:
    case TYPE_SSH_KEY:
      return createGeneralCredential(prepareFormatBasedGenericSecretForServerRequest(model));
  }
}
