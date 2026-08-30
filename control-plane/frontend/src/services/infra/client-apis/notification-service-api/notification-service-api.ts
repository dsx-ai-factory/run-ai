import {
  NotificationChannelsApi,
  NotificationStateApi,
  NotificationTypesApi,
  SubscriptionsApi,
  SlackApi,
  EmailApi,
  Configuration,
  type ConfigurationParameters,
} from "@/swagger-models/notifications-proxy-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class NotificationServiceApi {
  configuration: Configuration;

  notificationChannelsApi: NotificationChannelsApi;
  notificationStateApi: NotificationStateApi;
  notificationTypesApi: NotificationTypesApi;
  subscriptionApi: SubscriptionsApi;
  slackApi: SlackApi;
  emailApi: EmailApi;

  constructor() {
    this.configuration = new Configuration();

    this.notificationChannelsApi = new NotificationChannelsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.notificationStateApi = new NotificationStateApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.notificationTypesApi = new NotificationTypesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.subscriptionApi = new SubscriptionsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.emailApi = new EmailApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.slackApi = new SlackApi(this.configuration, this.baseUrl, authenticatedAxios);
  }
  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const notificationServiceApi = new NotificationServiceApi();
