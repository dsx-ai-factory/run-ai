/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { shallowMount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import RuleSelectorParams from "./rule-selector-params.vue";
import EnforcementValueInput from "./enforcement-value-input.vue";
import { errorMessages } from "@/common/error-message.constant";

describe("rule-selector-params.vue — API mapping-type validation", () => {
  it("applies ipaddress validation to selector argsInfo parameters", () => {
    const wrapper = shallowMount(RuleSelectorParams, {
      props: {
        params: [{ name: "hostAliasIp", type: "ipaddress" }],
        values: {},
      },
    });
    const rules = wrapper.findComponent(EnforcementValueInput).props("rules") as Array<
      (value: string) => boolean | string
    >;

    expect(rules).toHaveLength(1);
    expect(rules[0]("not-an-ip")).toBe(errorMessages.ENTER_VALID_IP_ADDRESS);
    expect(rules[0]("10.0.0.1")).toBe(true);
  });

  it("does not add format validation to an unvalidated mapping type", () => {
    const wrapper = shallowMount(RuleSelectorParams, {
      props: {
        params: [{ name: "containerName", type: "string" }],
        values: {},
      },
    });

    expect(wrapper.findComponent(EnforcementValueInput).props("rules")).toEqual([]);
  });
});
