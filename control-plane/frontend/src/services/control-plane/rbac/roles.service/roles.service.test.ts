/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//@ts-nocheck
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";
import { authorizationApi } from "@/services/infra/client-apis/authorization-api/authorization-api";
import {
  permissionMockNarrow,
  permissionsMockWide,
  rolesMock,
} from "@/services/control-plane/rbac/roles.service/roles.mock";

vi.mock("@/services/control-plane/control-plane.service/control-plane.service");
describe("rolesService", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    setActivePinia(createPinia());
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("list", () => {
    it("returns the roles of a single page", async () => {
      const getRoles = vi
        .spyOn(authorizationApi.rolesApi, "getRoles")
        .mockResolvedValueOnce({ data: { roles: [rolesMock[0], rolesMock[1]] } });

      const roles = await rolesService.list();

      expect(roles).toEqual([rolesMock[0], rolesMock[1]]);
      expect(getRoles).toHaveBeenCalledTimes(1);
      expect(getRoles).toHaveBeenCalledWith(500, 0);
    });

    it("follows next until all pages are fetched", async () => {
      const getRoles = vi
        .spyOn(authorizationApi.rolesApi, "getRoles")
        .mockResolvedValueOnce({ data: { roles: [rolesMock[0]], next: 500 } })
        .mockResolvedValueOnce({ data: { roles: [rolesMock[1]] } });

      const roles = await rolesService.list();

      expect(roles).toEqual([rolesMock[0], rolesMock[1]]);
      expect(getRoles).toHaveBeenCalledTimes(2);
      expect(getRoles).toHaveBeenNthCalledWith(1, 500, 0);
      expect(getRoles).toHaveBeenNthCalledWith(2, 500, 500);
    });

    it.each([
      ["next is 0", 0],
      ["next equals the current offset", 500],
    ])("stops when next does not move forward (%s)", async (_, staleNext) => {
      const getRoles = vi
        .spyOn(authorizationApi.rolesApi, "getRoles")
        .mockResolvedValueOnce({ data: { roles: [rolesMock[0]], next: 500 } })
        .mockResolvedValue({ data: { roles: [rolesMock[1]], next: staleNext } });

      const roles = await rolesService.list();

      expect(roles).toEqual([rolesMock[0], rolesMock[1]]);
      expect(getRoles).toHaveBeenCalledTimes(2);
    });
  });

  it("isPermittedRole with wide set of permissions", async () => {
    expect(rolesService.isPermittedRole(rolesMock[0], permissionsMockWide)).toBeTruthy();
    //  add one action in role which is not exist in permission events-history
    //  { resourceType: "events-history", actions: ["read", "create"] }
    expect(rolesService.isPermittedRole(rolesMock[1], permissionsMockWide)).toBeFalsy();
    expect(rolesService.isPermittedRole(rolesMock[2], permissionsMockWide)).toBeTruthy();
    expect(rolesService.isPermittedRole(rolesMock[3], permissionsMockWide)).toBeTruthy();
    //  add one permission which is not exist in role
    //  { resourceType: "another-permission", actions: ["read"] },
    expect(rolesService.isPermittedRole(rolesMock[4], permissionsMockWide)).toBeFalsy();
  });

  it("isPermittedRole with narrow set of permissions", async () => {
    expect(rolesService.isPermittedRole(rolesMock[0], permissionMockNarrow)).toBeFalsy();
    expect(rolesService.isPermittedRole(rolesMock[1], permissionMockNarrow)).toBeFalsy();
    expect(rolesService.isPermittedRole(rolesMock[2], permissionMockNarrow)).toBeFalsy();
    expect(rolesService.isPermittedRole(rolesMock[3], permissionMockNarrow)).toBeFalsy();
    expect(rolesService.isPermittedRole(rolesMock[4], permissionMockNarrow)).toBeFalsy();
    expect(rolesService.isPermittedRole(rolesMock[5], permissionMockNarrow)).toBeTruthy();
  });
});
