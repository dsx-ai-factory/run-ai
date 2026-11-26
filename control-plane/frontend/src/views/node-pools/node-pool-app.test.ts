/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import { createRouter, createMemoryHistory, useRoute } from "vue-router";
import { describe, expect, it } from "vitest";
import NodePoolApp from "./node-pool-app.vue";

describe("node pool view transitions", () => {
  it("initializes the shared view for each new/edit entity without resetting it for step queries", async () => {
    let mounts = 0;
    const SharedView = defineComponent({
      setup() {
        const route = useRoute();
        const initialIdentity = `${String(route.name)}:${route.params.id ?? ""}`;
        mounts++;
        return () => h("p", initialIdentity);
      },
    });
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: "/new", name: "new", component: SharedView },
        { path: "/edit/:id", name: "edit", component: SharedView },
      ],
    });
    await router.push("/new");
    await router.isReady();
    const wrapper = mount(NodePoolApp, { global: { plugins: [router] } });
    expect(wrapper.text()).toBe("new:");
    await router.push("/edit/created?stepName=scheduling");
    await flushPromises();
    expect(wrapper.text()).toBe("edit:created");
    await router.push("/edit/created?stepName=topology");
    await flushPromises();
    expect(mounts).toBe(2);
    await router.push("/edit/other");
    await flushPromises();
    expect(wrapper.text()).toBe("edit:other");
    await router.push("/new");
    await flushPromises();
    expect(wrapper.text()).toBe("new:");
    expect(mounts).toBe(4);
    wrapper.unmount();
  });
});
