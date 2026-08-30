import {
  Configuration,
  SettingsApi,
  WorkloadsV2Api,
  type ConfigurationParameters,
} from "@/swagger-models/workloads-manager-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class WorkloadsManagerApi {
  configuration: Configuration;

  workloadsV2Api: WorkloadsV2Api;
  settingsApi: SettingsApi;

  constructor() {
    this.configuration = new Configuration();

    this.workloadsV2Api = new WorkloadsV2Api(this.configuration, this.baseUrl, authenticatedAxios);
    this.settingsApi = new SettingsApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const workloadsManagerApi = new WorkloadsManagerApi();
