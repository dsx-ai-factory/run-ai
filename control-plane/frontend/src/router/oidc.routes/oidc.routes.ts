import OidcIndex from "@/views/oidc/oidc-index.vue";
import type { RunaiRouteRecord } from "vue-router";
import { OIDC_ROUTE_NAMES } from "@/router/oidc.routes/oidc.routes.names";

export const oidcRoutes: RunaiRouteRecord = {
  path: "/oauth-code",
  component: OidcIndex,
  name: OIDC_ROUTE_NAMES.OAUTH_CODE,
  meta: {
    requiresAuth: false,
    fullPage: true,
    minPermittedActions: null,
    resourceType: null,
  },
};
