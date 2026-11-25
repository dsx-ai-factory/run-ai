/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { flushPromises, mount } from "@vue/test-utils";
import { createMemoryHistory, createRouter, RouterView } from "vue-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { STORAGE_KEYS } from "@/models/storage.model";
import { ERROR_ROUTE_NAMES } from "@/router/error.routes/error.routes.names";
import { LOGIN_ROUTE_NAMES } from "@/router/login.routes/login.routes.names";
import { NODE_POOL_ROUTE_NAMES } from "@/router/node-pool.routes/node-pool.routes.names";
import { storageUtil } from "@/utils/storage.util";
import AuthCallback from "./auth-callback.vue";
import LoginPage from "./login-page.vue";

const auth = vi.hoisted(() => ({ login: vi.fn(), postLogin: vi.fn(), accessToken: "valid" }));
const validEmailMapper = vi.hoisted(() => vi.fn());
vi.mock("@/stores/auth.store", () => ({ useAuthStore: () => auth }));
vi.mock("@/services/control-plane/auth.service/auth.service", () => ({
  authService: { isValidSsoEmailMapperToken: validEmailMapper },
}));

const empty = { template: "<div/>" };
function routerForPageLoad() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: "/", component: empty },
      { path: "/node-pools", name: NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX, component: empty },
      { path: "/node-pools/edit/:id", name: NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT, component: empty },
      { path: "/login", name: LOGIN_ROUTE_NAMES.LOGIN_PAGE, component: LoginPage },
      { path: "/login/callback", name: LOGIN_ROUTE_NAMES.LOGIN_CALLBACK, component: AuthCallback },
      { path: "/denied", name: ERROR_ROUTE_NAMES.NO_ROLES, component: empty },
      { path: "/error/502", name: ERROR_ROUTE_NAMES.BAD_GATEWAY_ERROR, component: empty },
      { path: "/missing-mapper", name: ERROR_ROUTE_NAMES.NO_SSO_EMAIL_MAPPER, component: empty },
      { path: "/:catchAll(.*)", redirect: "/" },
    ],
  });
}

async function completeCallback(router = routerForPageLoad()) {
  await router.push("/login/callback");
  const view = mount(RouterView, { global: { plugins: [router] } });
  try {
    await flushPromises();
    return router.currentRoute.value;
  } finally {
    view.unmount();
  }
}

beforeEach(() => {
  sessionStorage.clear();
  localStorage.clear();
  vi.clearAllMocks();
  validEmailMapper.mockReturnValue(true);
  auth.accessToken = "valid";
});

describe("returning to the app after sign-in", () => {
  it.each(["scheduling", "topology"])("restores the %s edit URL on a new page load", async (step) => {
    const destination = `/node-pools/edit/100042?stepName=${step}&filter=a%26b&filter=c&flag#resource%20table`;
    const initial = routerForPageLoad();
    await initial.push(destination);
    const view = mount(RouterView, { global: { plugins: [initial] } });
    try {
      await initial.replace({ name: LOGIN_ROUTE_NAMES.LOGIN_PAGE });
      await flushPromises();
      expect(auth.login).toHaveBeenCalledOnce();
      expect(sessionStorage.getItem(STORAGE_KEYS.LOCATION)).toBe(destination);
      expect(localStorage.getItem(STORAGE_KEYS.LOCATION)).toBeNull();
    } finally {
      view.unmount();
    }

    const restored = await completeCallback();
    expect(restored.fullPath).toBe(destination);
    expect(restored.params.id).toBe("100042");
    expect(restored.query.filter).toEqual(["a&b", "c"]);
    expect(restored.hash).toBe("#resource table");
    expect(sessionStorage.getItem(STORAGE_KEYS.LOCATION)).toBeNull();
  });

  it.each([
    "https://example.invalid/outside",
    "//example.invalid/outside",
    "/unregistered",
    "/login/callback",
    "/login",
  ])("uses the landing page for an invalid or authentication return location %s", async (path) => {
    sessionStorage.setItem(STORAGE_KEYS.LOCATION, path);
    expect((await completeCallback()).path).toBe("/");
  });

  it.each([
    [NODE_POOL_ROUTE_NAMES.NODE_POOL_INDEX, "/node-pools"],
    [NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT, "/"],
    ["retired-route-name", "/"],
  ])("handles a legacy saved name %s", async (name, expected) => {
    storageUtil.save(STORAGE_KEYS.LOCATION, name);
    expect((await completeCallback()).path).toBe(expected);
    expect(localStorage.getItem(STORAGE_KEYS.LOCATION)).toBeNull();
  });

  it("runs the destination's navigation guards after authentication", async () => {
    const router = routerForPageLoad();
    router.beforeEach((to) =>
      to.name === NODE_POOL_ROUTE_NAMES.NODE_POOL_EDIT ? { name: ERROR_ROUTE_NAMES.NO_ROLES } : true,
    );
    sessionStorage.setItem(STORAGE_KEYS.LOCATION, "/node-pools/edit/7?stepName=topology");
    expect((await completeCallback(router)).name).toBe(ERROR_ROUTE_NAMES.NO_ROLES);
  });

  it("does not restore a route before token and mapper validation", async () => {
    sessionStorage.setItem(STORAGE_KEYS.LOCATION, "/node-pools/edit/7");
    auth.accessToken = "";
    expect((await completeCallback()).name).toBe(LOGIN_ROUTE_NAMES.LOGIN_CALLBACK);
    expect(auth.login).toHaveBeenCalledOnce();
    auth.accessToken = "valid";
    validEmailMapper.mockReturnValue(false);
    expect((await completeCallback()).name).toBe(ERROR_ROUTE_NAMES.NO_SSO_EMAIL_MAPPER);
    expect(sessionStorage.getItem(STORAGE_KEYS.LOCATION)).toBe("/node-pools/edit/7");
  });
});
