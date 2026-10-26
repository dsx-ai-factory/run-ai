/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { VueWrapper } from "@vue/test-utils";
import { mount } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { setActivePinia } from "pinia";
import { createTestingPinia } from "@pinia/testing";
import { Quasar } from "quasar";
import RoleReview from "./role-review.vue";
import { useRoleStore } from "@/stores/role.store";
import { Actions } from "@/models/roles.model";
import { roleReviewMock } from "@/views/rbac/roles/role-review.mock";
import { excludedResourcesTypes } from "@/models/access-rule.model";
import { ResourceType, type Action } from "@/swagger-models/authorization-client";
import { stringUtil } from "@/utils/string.util";

vi.mock("@/services/control-plane/roles.service/roles.service.ts");

const pinia = createTestingPinia({
  createSpy: vi.fn,
  initialState: {
    Permission: {
      resourceTypeMap: Object.fromEntries(roleReviewMock.permissionsWithGroup.map((p) => [p.resourceType, p])),
    },
  },
});

setActivePinia(pinia);
const rolesStore = useRoleStore(pinia);
rolesStore.$patch({
  roles: roleReviewMock.roles,
});
rolesStore.getRoleById = vi.mocked(rolesStore.getRoleById).mockReturnValue(roleReviewMock.roles[0]);
describe("role-review.vue", () => {
  let roleReviewCmp: VueWrapper<InstanceType<typeof RoleReview>>;

  beforeEach(() => {
    roleReviewCmp = mount(RoleReview, {
      global: {
        plugins: [Quasar, pinia],
        mocks: {
          $route: {
            meta: {
              prevRoute: "",
            },
            query: "",
            params: {
              id: "2",
            },
          },
          $router: {
            back: vi.fn(),
          },
        },
      },
    });
    roleReviewCmp.vm.$q.notify = vi.fn(); // disabling notify to prevent it from failing tests
  });
  afterEach(() => {
    vi.clearAllMocks();
  });

  test("Component is loading properly", () => {
    const html = roleReviewCmp.html();
    expect(html).toContain("Role name");
    expect(html).toContain("Permissions");
  });

  test("Actions should be checked / unchecked according to role permission", () => {
    const getActionSelector = (displayName: string, action: string) => {
      return `[aid=action-${stringUtil.slug(displayName)}-${action}]`;
    };
    const s3Assets = roleReviewMock.roles[0].permissions.find(
      (permission) => permission.resourceType === ResourceType.S3Assets,
    );

    s3Assets?.actions.forEach((action) => {
      expect(
        roleReviewCmp.find(getActionSelector(s3Assets.displayName as string, action)).attributes("aria-checked"),
      ).toBe("true");
    });

    roleReviewMock.roles[0].permissions.forEach((permission) => {
      if (excludedResourcesTypes.has(permission.resourceType) || permission.displayName === s3Assets?.displayName) {
        return;
      }

      expect(roleReviewCmp.html()).toContain(permission.displayName);

      const allActions = Actions.map((action) => action.id);
      const permissionActions: Action[] = permission.actions;
      const validateActions = (action: string, expectedValue: string) => {
        expect(
          roleReviewCmp.find(getActionSelector(permission.displayName as string, action)).attributes("aria-checked"),
        ).toBe(expectedValue);
      };
      permissionActions.forEach((action) => validateActions(action, "true"));

      allActions
        .filter((action) => !permissionActions.includes(action))
        .forEach((action) => validateActions(action, "false"));
    });
  });

  test("close button should call $router.back", () => {
    const closeButton = roleReviewCmp.find("[aid=close-button]");
    closeButton.trigger("click");
    expect(roleReviewCmp.vm.$router.back).toBeCalled();
  });
});
