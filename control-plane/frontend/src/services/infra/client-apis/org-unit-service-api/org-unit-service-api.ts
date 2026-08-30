import { Configuration, type ConfigurationParameters } from "@/swagger-models/workloads-client";
import {
  OrgUnitApi,
  DepartmentsApi,
  ProjectsApi,
  ReportsApi,
  NodeTypesApi,
  SettingsApi,
} from "@/swagger-models/org-unit-service-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class OrgUnitServiceApi {
  configuration: Configuration;

  orgUnitApi: OrgUnitApi;
  nodeTypesApi: NodeTypesApi;
  projectsApi: ProjectsApi;
  departmentsApi: DepartmentsApi;
  reportsApi: ReportsApi;
  settingsApi: SettingsApi;

  constructor() {
    this.configuration = new Configuration();

    this.orgUnitApi = new OrgUnitApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.projectsApi = new ProjectsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.departmentsApi = new DepartmentsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.reportsApi = new ReportsApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.nodeTypesApi = new NodeTypesApi(this.configuration, this.baseUrl, authenticatedAxios);
    this.settingsApi = new SettingsApi(this.configuration, this.baseUrl, authenticatedAxios);
  }

  get baseUrl() {
    return `${window.location.origin}`;
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const orgUnitServiceApi = new OrgUnitServiceApi();
