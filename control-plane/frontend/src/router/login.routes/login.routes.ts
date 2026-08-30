import type { RunaiRouteRecord } from "vue-router";

import { LOGIN_ROUTE_NAMES } from "./login.routes.names";

export const loginRoutes: Array<RunaiRouteRecord> = [
  {
    // We use this route for the authentication flow (in App.vue)
    // Since vue does not support defining routes without component
    // or redirect values and we don't have any logic in this
    // component - I've created an empty component.
    path: "/login/callback",
    name: LOGIN_ROUTE_NAMES.LOGIN_CALLBACK,
    component: () => import("@/views/login/auth-callback.vue"),
  },
  {
    path: "/login",
    name: LOGIN_ROUTE_NAMES.LOGIN_PAGE,
    component: () => import("@/views/login/login-page.vue"),
  },
];
