/* eslint-disable */
import { httpResponseService } from "@/services/infra/http-response.service/http-response.service";
import type { HttpErrorResponse, IHTTPOptions, RequestHeaders } from "@/models/http-response.model";
import { ForbiddenError, UnauthorizedError } from "@/models/http-response.model";
import { useAuthStore } from "@/stores/auth.store";
import { authenticatedAxios } from "./auth-axios.service";
import qs from "qs";

type Method = "POST" | "GET" | "DELETE" | "PUT" | "PATCH";

export const httpService = {
  request,
  requestFullData,
  handleHttpError,
};

const axios = authenticatedAxios;

async function requestFullData<T = any>(
  baseUrl: string,
  endpoint: string,
  method: Method = "GET",
  options: IHTTPOptions,
  data: any = {},
  query: any = {},
): Promise<any> {
  const authStore = useAuthStore();

  // Configure arrayFormat to serialize arrays as comma-separated strings
  const paramsSerializer = (params: any) => {
    return qs.stringify(params, {
      arrayFormat: "comma",
      // Other qs options can be added here based on your needs
    });
  };

  try {
    let headers: RequestHeaders = { "content-type": "application/json" };
    if (options.headers) {
      headers = { ...headers, ...options.headers };
    }

    headers["Authorization"] = `Bearer ${authStore.accessToken}`;

    if (options.useIdToken) {
      headers["Authorization"] = `Bearer ${authStore.idToken}`;
    }

    const res = await axios({
      url: `${baseUrl}${endpoint}`,
      method,
      data,
      headers,
      responseType: options.responseType,
      params: query,
      paramsSerializer: paramsSerializer,
    });
    return res;
  } catch (err: any) {
    throw handleHttpError(err);
  }
}

async function request<T = any>(
  baseUrl: string,
  endpoint: string,
  method: Method = "GET",
  options: IHTTPOptions,
  data: any = {},
  query: any = {},
): Promise<any> {
  const res = await requestFullData(baseUrl, endpoint, method, options, data, query);
  return res.data;
}

function _handleWorkLoadError(data: any, generalMSG?: string): string {
  let toastErrorMsg: string = generalMSG || "Something went wrong... please try again later";
  if (!data.details) {
    try {
      toastErrorMsg = data.message || JSON.stringify(data);
    } catch (e: unknown) {
      console.error(e);
    } finally {
      return toastErrorMsg;
    }
  }
  const errDetails: string = data.details;
  if (errDetails.includes("Internal error occurred")) {
    toastErrorMsg = "Internal error occurred";
  } else if (errDetails.includes("denied the request:")) {
    toastErrorMsg = errDetails.split("denied the request:")[1];
  } else {
    toastErrorMsg = errDetails;
  }
  return toastErrorMsg;
}

function handleHttpError(err: any) {
  const authStore = useAuthStore();

  if (err.response && err.response.status === UnauthorizedError.statusCode) {
    authStore.clearTokens();
    return;
  }

  let errorObj: HttpErrorResponse;
  if (err.response && err.response.status === ForbiddenError.statusCode) {
    errorObj = httpResponseService.getError(ForbiddenError.statusCode);
    throw errorObj;
  }

  if (typeof err === "object" && !err.response) {
    errorObj = httpResponseService.getError(-1);
    errorObj.message = err.message;

    throw errorObj;
  }

  // Workloads submission can return a different error format which describes the issue in a more detailed way
  if (err.response.data.complianceIssues?.length) {
    errorObj = httpResponseService.getSubmissionError(err.response.data);
    throw errorObj;
  }

  errorObj = httpResponseService.getError(err.response.data.code || err.response.status);
  if (err.response.data.details?.includes("workload-controller")) {
    errorObj.message = _handleWorkLoadError(err.response.data);
  } else if (err.response.data.message) {
    errorObj.message = err.response.data.message;
  } else if (err.response.data.description) {
    errorObj.message = err.response.data.description;
  }

  throw errorObj;
}
