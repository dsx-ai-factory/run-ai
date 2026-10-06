/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useRoleStore } from "./role.store";
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";
import type { Role } from "@/swagger-models/authorization-client";

vi.mock("@/services/control-plane/rbac/roles.service/roles.service", () => ({
  rolesService: {
    list: vi.fn(),
  },
}));

describe("role store", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    setActivePinia(createPinia());
  });

  describe("loadRoles", () => {
    it("loads the roles and resets loading", async () => {
      const roles = [{ id: 1, name: "System administrator" }] as Role[];
      vi.mocked(rolesService.list).mockResolvedValue(roles);
      const store = useRoleStore();

      await store.loadRoles();

      expect(store.roles.map((role) => role.id)).toEqual([1]);
      expect(store.loading).toBe(false);
    });

    it("resets loading when loading the roles fails", async () => {
      vi.mocked(rolesService.list).mockRejectedValue(new Error("request failed"));
      const store = useRoleStore();

      await expect(store.loadRoles()).rejects.toThrow("request failed");

      expect(store.loading).toBe(false);
    });
  });
});
