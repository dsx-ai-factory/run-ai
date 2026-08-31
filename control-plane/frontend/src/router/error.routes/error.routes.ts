/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import NoSsoEmailMapper from "@/views/errors/no-sso-email-mapper.vue";
import ErrorApp from "@/views/errors/error-app.vue";
import { ERROR_ROUTE_NAMES } from "./error.routes.names";
import NotFound from "@/views/errors/not-found.vue";
import ServerError from "@/views/errors/server-error.vue";
import BadGateway from "@/views/errors/bad-gateway.vue";
import MissingClaim from "@/views/errors/missing-claim.vue";
import InternalServerError from "@/views/errors/internal-server-error.vue";
import UnknownTimeout from "@/views/errors/unknown-timeout.vue";
import type { RunaiRouteRecord } from "vue-router";

export const errorRoutes: RunaiRouteRecord[] = [
  {
    path: "/error",
    component: ErrorApp,
    meta: {
      fullPage: true,
      requiresAuth: false,
      minPermittedActions: null,
      resourceType: null,
    },
    children: [
      {
        path: "/error/404",
        component: NotFound,
        name: ERROR_ROUTE_NAMES.NOT_FOUND,
        meta: {
          pageTitle: "404",
          fullPage: true,
          requiresAuth: false,
          minPermittedActions: null,
          resourceType: null,
        },
      },
      {
        path: "/error/502",
        component: BadGateway,
        name: ERROR_ROUTE_NAMES.BAD_GATEWAY_ERROR,
        meta: {
          pageTitle: "502",
          requiresAuth: false,
          fullPage: true,
          minPermittedActions: null,
          resourceType: null,
        },
      },
      {
        path: "/error/500",
        component: InternalServerError,
        name: ERROR_ROUTE_NAMES.INTERNAL_SERVER_ERROR,
        meta: {
          pageTitle: "500",
          requiresAuth: false,
          fullPage: true,
          minPermittedActions: null,
          resourceType: null,
        },
      },
      {
        path: "/error/503",
        component: ServerError,
        name: ERROR_ROUTE_NAMES.SERVICE_UNAVAILABLE_ERROR,
        meta: {
          pageTitle: "503",
          requiresAuth: false,
          fullPage: true,
          minPermittedActions: null,
          resourceType: null,
        },
      },
      {
        path: "/no-sso-email-mapper",
        component: NoSsoEmailMapper,
        name: ERROR_ROUTE_NAMES.NO_SSO_EMAIL_MAPPER,
        meta: {
          pageTitle: "401",
          requiresAuth: false,
          fullPage: true,
          minPermittedActions: null,
          resourceType: null,
        },
      },
      {
        path: "/error/missing-claim",
        component: MissingClaim,
        name: ERROR_ROUTE_NAMES.MISSING_CLAIM,
        meta: {
          pageTitle: "401",
          requiresAuth: false,
          fullPage: true,
          minPermittedActions: null,
          resourceType: null,
        },
      },
      {
        path: "/error/unknown-timeout",
        component: UnknownTimeout,
        name: ERROR_ROUTE_NAMES.UNKNOWN_TIMEOUT,
        meta: {
          pageTitle: "",
          requiresAuth: false,
          fullPage: true,
          minPermittedActions: null,
          resourceType: null,
        },
      },
    ],
  },
];
