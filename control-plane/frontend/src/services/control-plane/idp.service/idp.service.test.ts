/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import { AxiosError } from "axios";
import { idpService } from "./idp.service";
import { identityManagerApi } from "@/services/infra/client-apis/identity-manager-api/identity-manager-api";

vi.mock("@/services/infra/client-apis/identity-manager-api/identity-manager-api", () => ({
  identityManagerApi: { idpsApi: { updateIdp: vi.fn(), createIdp: vi.fn() } },
}));

describe("idpService.validateSamlCertificate", () => {
  const samlData = { metadataXmlFile: "xml", fileName: "idp.xml" };
  const updateIdp = vi.mocked(identityManagerApi.idpsApi.updateIdp);
  const createIdp = vi.mocked(identityManagerApi.idpsApi.createIdp);

  const badRequest = (data: unknown) =>
    new AxiosError("Bad Request", "ERR_BAD_REQUEST", undefined, undefined, {
      status: 400,
      data,
    } as never);

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("calls the update endpoint as a dry run with the given alias", async () => {
    updateIdp.mockResolvedValue({
      data: { alias: "my-idp", samlData: { certificateExpirationDate: "2030-01-01" } },
    } as never);

    const result = await idpService.validateSamlCertificate(samlData, "my-idp");

    expect(updateIdp).toHaveBeenCalledWith("my-idp", expect.objectContaining({ dryRun: true, samlData }));
    expect(createIdp).not.toHaveBeenCalled();
    expect(result.samlData?.certificateExpirationDate).toBe("2030-01-01");
  });

  it("calls the create endpoint as a dry run for a new IdP", async () => {
    createIdp.mockResolvedValue({ data: { samlData: { certificateExpirationDate: "2031-01-01" } } } as never);

    const result = await idpService.validateSamlCertificate(samlData);

    expect(createIdp).toHaveBeenCalledWith(expect.objectContaining({ dryRun: true, samlData }));
    expect(updateIdp).not.toHaveBeenCalled();
    expect(result.samlData?.certificateExpirationDate).toBe("2031-01-01");
  });

  it.each([
    ["new", undefined, createIdp],
    ["existing", "my-idp", updateIdp],
  ])("returns the parsed config of a rejected (400) %s IdP instead of throwing", async (_name, alias, api) => {
    api.mockRejectedValue(badRequest({ alias: "saml", samlData: { certificateExpirationDate: "2020-01-01" } }));

    const result = await idpService.validateSamlCertificate(samlData, alias);

    expect(result.samlData?.certificateExpirationDate).toBe("2020-01-01");
    expect(result.errorMessage).toBe("Invalid SAML metadata");
  });

  it("prefers the backend message when the 400 has one", async () => {
    createIdp.mockRejectedValue(badRequest({ code: 400, message: "saml idp requires metadata" }));

    expect((await idpService.validateSamlCertificate(samlData)).errorMessage).toBe("saml idp requires metadata");
  });
});
