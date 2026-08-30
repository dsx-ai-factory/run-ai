import { Configuration, type ConfigurationParameters, TenantApi } from "@/swagger-models/temp-backend-settings-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class BackendApi {
  configuration: Configuration;

  tenantApi: TenantApi;

  constructor() {
    this.configuration = new Configuration();

    this.tenantApi = new TenantApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const backendApi = new BackendApi();
