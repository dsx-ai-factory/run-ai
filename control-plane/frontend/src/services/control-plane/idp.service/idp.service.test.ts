/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import { AxiosError } from "axios";
import { idpService } from "./idp.service";
import { IdpCreationRequestTypeEnum } from "@/swagger-models/identity-manager-client";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";

vi.mock("@/services/infra/client-apis/identity-manager-api/identity-manager-api", () => ({
  identityManagerApi: { idpsApi: { updateIdp: vi.fn(), createIdp: vi.fn() } },
}));

describe("idpService dry run", () => {
  const request = { type: IdpCreationRequestTypeEnum.Saml, samlData: { metadataXmlFile: "xml", fileName: "idp.xml" } };
  const updateIdp = vi.mocked(identityManagerApi.idpsApi.updateIdp);
  const createIdp = vi.mocked(identityManagerApi.idpsApi.createIdp);

  const badRequest = (data: unknown) =>
    new AxiosError("Bad Request", "ERR_BAD_REQUEST", undefined, undefined, { status: 400, data } as never);

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("updateIdp sends dryRun to the update endpoint and returns the parsed config", async () => {
    updateIdp.mockResolvedValue({
      data: { alias: "my-idp", samlData: { certificateExpirationDate: "2030-01-01" } },
    } as never);

    const result = await idpService.updateIdp("my-idp", request, true);

    expect(updateIdp).toHaveBeenCalledWith("my-idp", { ...request, dryRun: true });
    expect(createIdp).not.toHaveBeenCalled();
    expect(result?.samlData?.certificateExpirationDate).toBe("2030-01-01");
  });

  it("addIdp sends dryRun to the create endpoint", async () => {
    createIdp.mockResolvedValue({
      data: { alias: "saml", samlData: { certificateExpirationDate: "2031-01-01" } },
    } as never);

    const result = await idpService.addIdp(request, true);

    expect(createIdp).toHaveBeenCalledWith({ ...request, dryRun: true });
    expect(updateIdp).not.toHaveBeenCalled();
    expect(result.samlData?.certificateExpirationDate).toBe("2031-01-01");
  });

  it("does not add dryRun to a regular request", async () => {
    createIdp.mockResolvedValue({ data: { alias: "saml" } } as never);

    await idpService.addIdp(request);

    expect(createIdp).toHaveBeenCalledWith(request);
  });

  it.each([
    ["addIdp", () => idpService.addIdp(request, true), createIdp],
    ["updateIdp", () => idpService.updateIdp("my-idp", request, true), updateIdp],
  ])("%s rethrows the raw 400 on a dry run so the body stays readable", async (_name, call, api) => {
    const error = badRequest({ alias: "saml", samlData: { certificateExpirationDate: "2020-01-01" } });
    api.mockRejectedValue(error);

    await expect(call()).rejects.toBe(error);
  });
});
