import {
  DatavolumesApi,
  Configuration,
  type ConfigurationParameters,
  SettingsApi,
} from "@/swagger-models/datavolumes-service-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class DataVolumesServiceApi {
  configuration: Configuration;
  dataVolumesApi: DatavolumesApi;
  settingsApi: SettingsApi;

  constructor() {
    this.configuration = new Configuration();

    this.dataVolumesApi = new DatavolumesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.settingsApi = new SettingsApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const dataVolumesServiceApi = new DataVolumesServiceApi();
