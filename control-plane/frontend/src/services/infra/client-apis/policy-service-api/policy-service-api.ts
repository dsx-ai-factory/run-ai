import {
  Configuration,
  type ConfigurationParameters,
  PolicyApi,
  SpecMappingApi,
  SynchronizationApi,
} from "@/swagger-models/policy-service-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class PolicyServiceApi {
  configuration: Configuration;

  policyApi: PolicyApi;
  synchronizationApi: SynchronizationApi;
  specMappingApi: SpecMappingApi;

  constructor() {
    this.configuration = new Configuration();

    this.policyApi = new PolicyApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.synchronizationApi = new SynchronizationApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.specMappingApi = new SpecMappingApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const policyServiceApi = new PolicyServiceApi();
