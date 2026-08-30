import {
  Configuration,
  AccessRulesApi,
  PermissionsApi,
  RolesApi,
  type ConfigurationParameters,
  MeApi,
} from "@/swagger-models/authorization-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class AuthorizationApi {
  configuration: Configuration;

  accessRulesApi: AccessRulesApi;
  permissionsApi: PermissionsApi;
  rolesApi: RolesApi;
  meApi: MeApi;

  constructor() {
    this.configuration = new Configuration();

    this.accessRulesApi = new AccessRulesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.permissionsApi = new PermissionsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.rolesApi = new RolesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.meApi = new MeApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const authorizationApi = new AuthorizationApi();
