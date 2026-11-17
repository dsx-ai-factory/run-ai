/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { mount, flushPromises, type VueWrapper } from "@vue/test-utils";
import { createMemoryHistory, createRouter, RouterView } from "vue-router";
import { createTestingPinia } from "@pinia/testing";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import LoginPage from "./login-page.vue";
import AuthCallback from "./auth-callback.vue";
import { useAuthStore } from "@/stores/auth.store";
import { STORAGE_KEYS } from "@/models/storage.model";
import { storageUtil } from "@/utils/storage.util";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import { LOGIN_ROUTE_NAMES } from "@/router/login.routes/login.routes.names";
import { ERROR_ROUTE_NAMES } from "@/router/error.routes/error.routes.names";
vi.mock("@/services/control-plane/auth.service/auth.service", () => ({
  authService: { isValidSsoEmailMapperToken: vi.fn().mockReturnValue(true) },
}));
let wrapper: VueWrapper | undefined;
async function render(path: string) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/", name: "home", component: { template: "<div>Home</div>" } },
      {
        path: "/node-pools",
        name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX,
        component: { template: "<div>Pool list</div>" },
      },
      {
        path: "/node-pools/edit/:id",
        name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT,
        component: { template: "<div>Saved pool</div>" },
      },
      { path: "/login", name: LOGIN_ROUTE_NAMES.LOGIN_PAGE, component: LoginPage },
      { path: "/callback", name: LOGIN_ROUTE_NAMES.LOGIN_CALLBACK, component: AuthCallback },
      {
        path: "/error/502",
        name: ERROR_ROUTE_NAMES.BAD_GATEWAY_ERROR,
        component: { template: "<div>Bad gateway</div>" },
      },
    ],
  });
  await router.push(path);
  const pinia = createTestingPinia({ createSpy: vi.fn });
  useAuthStore(pinia).accessToken = "unit-test-access-token";
  wrapper = mount(RouterView, { global: { plugins: [router, pinia] } });
  await router.isReady();
  return router;
}
beforeEach(() => {
  vi.clearAllMocks();
  storageUtil.remove(STORAGE_KEYS.LOCATION);
});
afterEach(() => {
  wrapper?.unmount();
  storageUtil.remove(STORAGE_KEYS.LOCATION);
});

describe("login return navigation", () => {
  it("returns to a cold Step 3 link with its saved pool ID, query and section", async () => {
    const path = "/node-pools/edit/100002?stepName=topology&clusterId=test-cluster#placement";
    const router = await render(path);
    await router.replace({ name: LOGIN_ROUTE_NAMES.LOGIN_PAGE });
    await flushPromises();
    await router.replace({ name: LOGIN_ROUTE_NAMES.LOGIN_CALLBACK });
    await flushPromises();
    expect(router.currentRoute.value.fullPath).toBe(path);
    expect(storageUtil.get(STORAGE_KEYS.LOCATION)).toBeNull();
  });
  it("keeps legacy saved route names working", async () => {
    const router = await render("/");
    storageUtil.save(STORAGE_KEYS.LOCATION, NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX);
    await router.replace({ name: LOGIN_ROUTE_NAMES.LOGIN_CALLBACK });
    await flushPromises();
    expect(router.currentRoute.value.name).toBe(NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX);
  });
  it("uses the landing route when no return destination was saved", async () => {
    const router = await render("/node-pools");
    await router.replace({ name: LOGIN_ROUTE_NAMES.LOGIN_CALLBACK });
    await flushPromises();
    expect(router.currentRoute.value.path).toBe("/");
  });
});
