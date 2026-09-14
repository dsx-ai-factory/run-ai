/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from "vitest";
import { Quasar } from "quasar";
import { flushPromises, shallowMount, type VueWrapper } from "@vue/test-utils";
import { createTestingPinia } from "@pinia/testing";
import AccessRuleIndex from "./access-rule-index.vue";
import NewAccessRuleModal from "@/components/rbac/access-rule/new-access-rule-modal/new-access-rule-modal.vue";
import type { AccessRule, AccessRulesBatchCreateResponse } from "@/swagger-models/authorization-client";

vi.mock("vue-router", () => ({
  useRouter: vi.fn().mockReturnValue({ push: vi.fn(), resolve: vi.fn().mockReturnValue({ href: "/" }) }),
  useRoute: vi.fn().mockReturnValue({ name: "", query: {}, matched: [], params: {}, path: "", fullPath: "", hash: "" }),
}));

vi.mock("@/services/control-plane/rbac/access-rule.service/access-rule.service", () => ({
  accessRuleService: {
    list: vi.fn().mockResolvedValue([]),
    getAccessRulesCount: vi.fn().mockResolvedValue(0),
    getAccessRule: vi.fn(),
    getAccessRulesV2: vi.fn().mockResolvedValue([]),
    accessRulesBatch: vi.fn(),
  },
}));

function mountComponent(): VueWrapper<InstanceType<typeof AccessRuleIndex>> {
  const wrapper = shallowMount(AccessRuleIndex, {
    global: {
      plugins: [Quasar, createTestingPinia({ createSpy: vi.fn })],
      directives: { permission: vi.fn(), clickOutside: vi.fn() },
    },
  });
  wrapper.vm.$q.notify = vi.fn();
  return wrapper;
}

async function emitClose(
  wrapper: VueWrapper<InstanceType<typeof AccessRuleIndex>>,
  succeeded: Array<Partial<AccessRule>>,
  failedCount: number,
) {
  const response = {
    succeeded: succeeded as AccessRule[],
    failed: Array.from({ length: failedCount }, () => ({ code: 403, message: "Insufficient permissions" })),
  } as AccessRulesBatchCreateResponse;

  wrapper.vm.isNewAccessRuleModalOpen = true;
  await wrapper.vm.$nextTick();

  const modal = wrapper.findComponent(NewAccessRuleModal);
  await modal.vm.$emit("close", response);
  await flushPromises();
}

describe("access-rule-index - notifyCreatedAccessRules", () => {
  beforeEach(() => vi.clearAllMocks());

  it("shows a success toast naming the subject when a single rule is created", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    await emitClose(wrapper, [{ id: 1, subjectId: "alice@run.ai" }], 0);

    expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(
      expect.objectContaining({ type: "positive", message: "Access rule for alice@run.ai created" }),
    );
  });

  it("shows a pluralized success toast for multiple created rules", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    await emitClose(
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

  it("shows an error toast instead of a misleading '0/1 created' success when the only rule fails", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    await emitClose(wrapper, [], 1);

    expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(
      expect.objectContaining({ type: "negative", message: "Access rule couldn't be created" }),
    );
  });

  it("shows a pluralized error toast reporting the failed count on partial failure", async () => {
    const wrapper = mountComponent();
    await flushPromises();

    await emitClose(wrapper, [{ id: 1, subjectId: "alice@run.ai" }], 2);

    expect(wrapper.vm.$q.notify).toHaveBeenCalledWith(
      expect.objectContaining({ type: "negative", message: "2 access rule(s) couldn't be created" }),
    );
  });
});
