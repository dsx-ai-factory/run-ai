import { Configuration, K8sObjectApi, type ConfigurationParameters } from "@/swagger-models/cluster-api-client";
import { authenticatedAxios } from "@/services/infra/https.service/auth-axios.service";

class ClusterApi {
  configuration: Configuration;
  k8sObjectApi!: K8sObjectApi;
  private currentBaseUrl: string;

  constructor() {
    this.configuration = new Configuration();
    this.currentBaseUrl = window.location.origin;
    this.rebuildClients();
  }

  // Hit the cluster's own ingress: the SaaS CP doesn't proxy /cluster-api/* to remote clusters.
  setBaseUrl(url: string): void {
    const next = (url || window.location.origin).replace(/\/+$/, "");
    if (next === this.currentBaseUrl) return;
    this.currentBaseUrl = next;
    this.rebuildClients();
  }

  private rebuildClients(): void {
    this.k8sObjectApi = new K8sObjectApi(this.configuration, this.currentBaseUrl, authenticatedAxios);
  }

  public set setAccessToken(accessToken: ConfigurationParameters["accessToken"]) {
    this.configuration.accessToken = accessToken;
  }
}

export const clusterApi = new ClusterApi();
