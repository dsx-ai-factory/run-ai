/* eslint-disable */
import type { IHTTPOptions } from "@/models/http-response.model";
import { httpService } from "@/services/infra/https.service/http.service";

let _baseUrl = "";

export const clusterApiServiceOld = {
  get(endpoint: string, query?: any, options?: IHTTPOptions): Promise<any> {
    return httpService.request(_baseUrl, endpoint, "GET", { responseType: "json", ...options }, {}, query);
  },
  getFullData(endpoint: string, query?: any, options?: IHTTPOptions): Promise<any> {
    return httpService.requestFullData(_baseUrl, endpoint, "GET", { responseType: "json", ...options }, {}, query);
  },
  post(endpoint: string, data?: any, options?: IHTTPOptions): Promise<any> {
    return httpService.request(_baseUrl, endpoint, "POST", { responseType: "json", ...options }, data);
  },
  put(endpoint: string, data?: any, options?: IHTTPOptions): Promise<any> {
    return httpService.request(_baseUrl, endpoint, "PUT", { responseType: "json", ...options }, data);
  },
  delete(endpoint: string, data?: any, options?: IHTTPOptions): Promise<any> {
    return httpService.request(_baseUrl, endpoint, "DELETE", { responseType: "json", ...options }, data);
  },
  setBaseURL(baseUrl: string): void {
    _baseUrl = baseUrl && (baseUrl.endsWith("/") ? baseUrl : `${baseUrl}/`);
  },
};
