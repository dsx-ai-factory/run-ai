/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { usePermissionStore } from "./permissions.store";
import { Action, ResourceType, ResourceTypeGroupId, type Permission } from "@/swagger-models/authorization-client";

function permission(resourceType: ResourceType, actions: Action[]): Permission {
  return {
    resourceType,
    displayName: resourceType,
    groupId: ResourceTypeGroupId.Organization,
    actions,
  };
}

describe("permissions.store", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  describe("setPermissions", () => {
    it("stores the permissions list", () => {
      const store = usePermissionStore();
      const perms = [permission(ResourceType.Settings, [Action.Read])];

      store.setPermissions(perms);

      expect(store.permissions).toEqual(perms);
      expect(store.permissionsList).toEqual(perms);
    });

    it("rebuilds resourceTypeMap keyed by resourceType", () => {
      const store = usePermissionStore();
      store.setPermissions([
        permission(ResourceType.Settings, [Action.Read]),
        permission(ResourceType.SecuritySettings, [Action.Read, Action.Update]),
      ]);

      expect(store.resourceTypeMap[ResourceType.Settings].actions).toEqual([Action.Read]);
      expect(store.resourceTypeMap[ResourceType.SecuritySettings].actions).toEqual([Action.Read, Action.Update]);
    });

    it("replaces previous permissions on a second call", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read])]);
      store.setPermissions([permission(ResourceType.BrandingSettings, [Action.Create])]);

      expect(store.permissions).toHaveLength(1);
      expect(store.resourceTypeMap[ResourceType.Settings]).toBeUndefined();
      expect(store.resourceTypeMap[ResourceType.BrandingSettings].actions).toEqual([Action.Create]);
    });
  });

  describe("isReadOnly", () => {
    it("returns true when the only action is Read", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read])]);

      expect(store.isReadOnly(ResourceType.Settings)).toBe(true);
    });

    it("returns false when Read is combined with Update", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read, Action.Update])]);

      expect(store.isReadOnly(ResourceType.Settings)).toBe(false);
    });

    it("returns false when Read is combined with Create", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read, Action.Create])]);

      expect(store.isReadOnly(ResourceType.Settings)).toBe(false);
    });

    it("returns false when actions are empty", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [])]);

      expect(store.isReadOnly(ResourceType.Settings)).toBe(false);
    });

    it("returns false when the resourceType is not present at all", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read])]);

      expect(store.isReadOnly(ResourceType.SecuritySettings)).toBe(false);
    });

    it("returns false when the only action is Update (not Read)", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Update])]);

      expect(store.isReadOnly(ResourceType.Settings)).toBe(false);
    });
  });

  describe("hasPermission", () => {
    it("returns true when the resourceType has the requested action", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read, Action.Update])]);

      expect(store.hasPermission(ResourceType.Settings, Action.Update)).toBe(true);
    });

    it("returns false when the resourceType lacks the requested action", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read])]);

      expect(store.hasPermission(ResourceType.Settings, Action.Update)).toBe(false);
    });

    it("returns false when the resourceType is missing", () => {
      const store = usePermissionStore();
      store.setPermissions([]);

      expect(store.hasPermission(ResourceType.Settings, Action.Read)).toBe(false);
    });

    it("returns true when any resourceType in the array has the action", () => {
      const store = usePermissionStore();
      store.setPermissions([
        permission(ResourceType.Settings, [Action.Read]),
        permission(ResourceType.SecuritySettings, [Action.Update]),
      ]);

      expect(store.hasPermission([ResourceType.Settings, ResourceType.SecuritySettings], Action.Update)).toBe(true);
    });

    it("returns false when none of the resourceTypes in the array have the action", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read])]);

      expect(store.hasPermission([ResourceType.SecuritySettings, ResourceType.BrandingSettings], Action.Read)).toBe(
        false,
      );
    });
  });

  describe("hasSomePermittedActions", () => {
    it("returns true when at least one action overlaps", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read])]);

      expect(store.hasSomePermittedActions(ResourceType.Settings, [Action.Create, Action.Read])).toBe(true);
    });

    it("returns false when no actions overlap", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read])]);

      expect(store.hasSomePermittedActions(ResourceType.Settings, [Action.Create, Action.Update])).toBe(false);
    });

    it("returns false when the resourceType is missing", () => {
      const store = usePermissionStore();
      store.setPermissions([]);

      expect(store.hasSomePermittedActions(ResourceType.Settings, [Action.Read])).toBe(false);
    });

    it("returns true if any of multiple resourceTypes grants any of the actions", () => {
      const store = usePermissionStore();
      store.setPermissions([
        permission(ResourceType.Settings, []),
        permission(ResourceType.BrandingSettings, [Action.Update]),
      ]);

      expect(
        store.hasSomePermittedActions(
          [ResourceType.Settings, ResourceType.BrandingSettings],
          [Action.Create, Action.Update],
        ),
      ).toBe(true);
    });

    it("returns false across all resourceTypes when none grant any action", () => {
      const store = usePermissionStore();
      store.setPermissions([
        permission(ResourceType.Settings, [Action.Read]),
        permission(ResourceType.BrandingSettings, [Action.Read]),
      ]);

      expect(
        store.hasSomePermittedActions(
          [ResourceType.Settings, ResourceType.BrandingSettings],
          [Action.Create, Action.Update, Action.Delete],
        ),
      ).toBe(false);
    });
  });

  describe("getPermittedActions", () => {
    it("returns the actions array for a known resourceType", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read, Action.Update])]);

      expect(store.getPermittedActions(ResourceType.Settings)).toEqual([Action.Read, Action.Update]);
    });

    it("returns an empty array for a missing resourceType", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read])]);

      expect(store.getPermittedActions(ResourceType.SecuritySettings)).toEqual([]);
    });

    it("returns an empty array when no permissions are loaded", () => {
      const store = usePermissionStore();

      expect(store.getPermittedActions(ResourceType.Settings)).toEqual([]);
    });
  });

  describe("hasAllPermissions", () => {
    it("returns true when every check passes", () => {
      const store = usePermissionStore();
      store.setPermissions([
        permission(ResourceType.Settings, [Action.Read]),
        permission(ResourceType.BrandingSettings, [Action.Update]),
      ]);

      expect(
        store.hasAllPermissions([
          { resourceType: ResourceType.Settings, action: Action.Read },
          { resourceType: ResourceType.BrandingSettings, action: Action.Update },
        ]),
      ).toBe(true);
    });

    it("returns false when any check fails", () => {
      const store = usePermissionStore();
      store.setPermissions([permission(ResourceType.Settings, [Action.Read])]);

      expect(
        store.hasAllPermissions([
          { resourceType: ResourceType.Settings, action: Action.Read },
          { resourceType: ResourceType.BrandingSettings, action: Action.Update },
        ]),
      ).toBe(false);
    });

    it("returns true for an empty list of checks", () => {
      const store = usePermissionStore();

      expect(store.hasAllPermissions([])).toBe(true);
    });
  });
});
