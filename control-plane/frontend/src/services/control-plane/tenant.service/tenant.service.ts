import { tenantManagerApi } from "@/services/infra/client-apis/tenants-manager-api/tenants-manager-api";
import { httpService } from "@/services/infra/https.service/http.service";

import type { GetTenantLogo200Response, Tenant, LogoCreationRequest } from "@/swagger-models/tenants-manager-client";

export const tenantService = {
  get,
  getTenantLogo,
  uploadTenantLogo,
};

async function get(): Promise<Tenant[]> {
  try {
    const response = await tenantManagerApi.tenantApi.getTenants();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getTenantLogo(): Promise<GetTenantLogo200Response> {
  try {
    const response = await tenantManagerApi.logoApi.getTenantLogo();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function uploadTenantLogo(logoCreationRequest?: LogoCreationRequest): Promise<void> {
  try {
    const response = await tenantManagerApi.logoApi.uploadTenantLogo("", logoCreationRequest);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
