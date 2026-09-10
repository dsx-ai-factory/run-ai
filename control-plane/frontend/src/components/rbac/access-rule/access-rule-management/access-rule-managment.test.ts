/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { Quasar } from "quasar";
import { flushPromises, shallowMount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import AccessRuleManagement from "./access-rule-managment.vue";
import NewManagementAccessRuleForm from "@/components/rbac/access-rule/access-rule-management-modal/new-management-access-rule-form/new-management-access-rule-form.vue";
import { EAccessRuleModalPage } from "@/models/access-rule.model";
import type { AccessRule, AccessRulesBatchCreateResponse } from "@/swagger-models/authorization-client";

vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: {
    getAccessRulesCount: vi.fn().mockResolvedValue(0),
    getAccessRulesV2: vi.fn().mockResolvedValue([]),
    deleteAccessRule: vi.fn(),
  },
}));

function mountComponent(): VueWrapper<InstanceType<typeof AccessRuleManagement>> {
  const wrapper = shallowMount(AccessRuleManagement, {
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      directives: { permission: vi.fn(), clickOutside: vi.fn() },
      // QCardSection wraps the form; shallowMount's default auto-stub would otherwise
      // stub it out entirely and skip rendering NewManagementAccessRuleForm as a child.
      // QInfiniteScroll needs reset()/resume() so the post-save refresh (setTimeout) doesn't throw.
      stubs: {
        QCardSection: false,
        QInfiniteScroll: { template: "<div />", methods: { reset: vi.fn(), resume: vi.fn() } },
      },
    },
    props: {
      formOptions: { page: EAccessRuleModalPage.Project, scopeId: "1001", scopeType: "tenant" },
      hideSearch: true,
      autoOpenForm: true,
    },
  });
  wrapper.vm.$q.notify = vi.fn();
  return wrapper;
}

async function emitSave(
  wrapper: VueWrapper<InstanceType<typeof AccessRuleManagement>>,
  succeeded: Array<Partial<AccessRule>>,
  failedCount: number,
) {
  const response = {
    succeeded: succeeded as AccessRule[],
    failed: Array.from({ length: failedCount }, () => ({ code: 403, message: "Insufficient permissions" })),
  } as AccessRulesBatchCreateResponse;
  const form = wrapper.findComponent(NewManagementAccessRuleForm);
  await form.vm.$emit("save", response);
  await flushPromises();
  // resetAndTriggerInfiniteScroll defers its reset()/resume() call via setTimeout(0).
  await new Promise((resolve) => setTimeout(resolve, 0));
}

describe("AccessRuleManagement - notifyCreatedAccessRules", () => {
  beforeEach(() => vi.clearAllMocks());

  it("shows a success toast when every rule is created", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    await emitSave(wrapper, [{ id: 1, subjectId: "alice@run.ai" }], 0);

    expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(
      expect.objectContaining({ type: "positive", message: "Access rule created" }),
    );
  });

  it("shows a pluralized success toast for multiple created rules", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    await emitSave(
      wrapper,
      [
        { id: 1, subjectId: "alice@run.ai" },
        { id: 2, subjectId: "bob@run.ai" },
      ],
      0,
    );

    expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(
      expect.objectContaining({ type: "positive", message: "2 access rule(s) created" }),
    );
  });

  it("shows an error toast instead of a fractional success count when the only rule fails", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    await emitSave(wrapper, [], 1);

    expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(
      expect.objectContaining({ type: "negative", message: "Access rule couldn't be created" }),
    );
  });

  it("shows a pluralized error toast reporting the failed count on partial failure", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    await emitSave(wrapper, [{ id: 1, subjectId: "alice@run.ai" }], 2);

    expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(
      expect.objectContaining({ type: "negative", message: "2 access rule(s) couldn't be created" }),
    );
  });
});
