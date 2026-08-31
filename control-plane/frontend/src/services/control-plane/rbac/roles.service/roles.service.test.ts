/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

//@ts-nocheck
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";
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
