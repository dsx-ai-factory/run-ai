import type { ConfigurationParameters } from "@/swagger-models/catalog-service-client";
import {
  Configuration,
  RegistriesApi,
  ImagesApi,
  InferencesApi,
  HelmReposApi,
  SettingsApi,
} from "@/swagger-models/catalog-service-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class CatalogServiceApi {
  configuration: Configuration;

  registriesApi: RegistriesApi;
  imagesApi: ImagesApi;
  inferencesApi: InferencesApi;
  helmReposApi: HelmReposApi;
  settingsApi: SettingsApi;

  constructor() {
    this.configuration = new Configuration();

    this.registriesApi = new RegistriesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.imagesApi = new ImagesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.inferencesApi = new InferencesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.helmReposApi = new HelmReposApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.settingsApi = new SettingsApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const catalogServiceApi = new CatalogServiceApi();
