import {
  Configuration,
  type ConfigurationParameters,
  SettingsApi,
  SystemNotificationApi,
} from "@/swagger-models/bff-service-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class BffServiceApi {
  configuration: Configuration;

  settingsApi: SettingsApi;

  systemNotificationApi: SystemNotificationApi;

  constructor() {
    this.configuration = new Configuration();
    this.systemNotificationApi = new SystemNotificationApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.settingsApi = new SettingsApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl(): string {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const bffServiceApi = new BffServiceApi();
