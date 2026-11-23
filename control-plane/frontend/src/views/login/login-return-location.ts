/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { RouteLocationNormalized, RouteLocationRaw, Router } from "vue-router";
import { STORAGE_KEYS } from "@/models/storage.model";
import { LOGIN_ROUTE_NAMES } from "@/router/login.routes/login.routes.names";
import { OIDC_ROUTE_NAMES } from "@/router/oidc.routes/oidc.routes.names";
import { storageUtil } from "@/utils/storage.util";

function isReturnRoute(name: RouteLocationNormalized["name"] | null): boolean {
  return Boolean(
    name &&
      name !== LOGIN_ROUTE_NAMES.LOGIN_PAGE &&
      name !== LOGIN_ROUTE_NAMES.LOGIN_CALLBACK &&
      name !== OIDC_ROUTE_NAMES.OAUTH_CODE,
  );
}

export function rememberLoginLocation(from: RouteLocationNormalized): void {
  if (isReturnRoute(from.name)) {
    sessionStorage.setItem(STORAGE_KEYS.LOCATION, from.fullPath);
  } else {
    sessionStorage.removeItem(STORAGE_KEYS.LOCATION);
  }
}

export function takeLoginLocation(router: Router): RouteLocationRaw {
  const path = sessionStorage.getItem(STORAGE_KEYS.LOCATION);
  sessionStorage.removeItem(STORAGE_KEYS.LOCATION);
  try {
    if (path !== null) {
      if (
        !path.startsWith("/") ||
        path.startsWith("//") ||
        new URL(path, window.location.origin).origin !== window.location.origin
      ) {
        return "/";
      }
      const route = router.resolve(path);
      return isReturnRoute(route.name) ? { path: route.path, query: route.query, hash: route.hash } : "/";
    }

    // Sign-ins started on an older frontend stored only a route name.
    const name = storageUtil.get<string>(STORAGE_KEYS.LOCATION);
    storageUtil.remove(STORAGE_KEYS.LOCATION);
    if (name && isReturnRoute(name) && router.hasRoute(name)) {
      return router.resolve({ name }).fullPath;
    }
  } catch {
    // A stale target, including a legacy name missing required params, falls back to the landing page.
  }
  return "/";
}
