/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ResourceType, Action } from "@/swagger-models/authorization-client";

import type { RunaiRouteRecord } from "vue-router";
import { DATA_SOURCE_ROUTE_NAMES } from "@/router/data-source.routes/data-source.routes.names";
import { EMenuSubjectNames } from "@/components/infra/runai-layout/routes-model/routes-model";

export const dataSourceRoutes: RunaiRouteRecord = {
  path: "/data-sources",
  // route level code-splitting
  // this generates a separate chunk (department.[hash].js) for this route
  // which is lazy-loaded when the route is visited.
  component: () => import("@/views/data-source/data-source-app.vue"),
  children: [
    {
      path: "",
      name: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX,
      component: () => import("@/views/data-source/data-source-index.vue"),
      meta: {
        pageTitle: "Data sources",
        resourceType: [
          ResourceType.NfsAssets,
          ResourceType.S3Assets,
          ResourceType.GitAssets,
          ResourceType.PvcAssets,
          ResourceType.HostPathAssets,
          ResourceType.CmVolumeAssets,
          ResourceType.SecretVolumeAssets,
        ],
        minPermittedActions: [Action.Read],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
    },
    {
      path: "hostpath/new",
      name: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_HOST_PATH_NEW,
      component: () => import("@/views/data-source/host-path/host-path-new.vue"),
      meta: {
        pageTitle: "New data source",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.HostPathAssets,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
      },
    },
    {
      path: "s3/new",
      name: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_S3_NEW,
      component: () => import("@/views/data-source/s3/s3-new.vue"),
      meta: {
        pageTitle: "New data source",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.S3Assets,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
      },
    },
    {
      path: "git/new",
      name: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_GIT_NEW,
      component: () => import("@/views/data-source/git/git-new.vue"),
      meta: {
        pageTitle: "New data source",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.GitAssets,
        minPermittedActions: [Action.Create, Action.Update],
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        requiresAuth: true,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
      },
    },
    {
      path: "nfs/new",
      name: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_NFS_NEW,
      component: () => import("@/views/data-source/nfs/nfs-new.vue"),
      meta: {
        pageTitle: "New data source",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.NfsAssets,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
      },
    },
    {
      path: "pvc/new",
      name: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_PVC_NEW,
      component: () => import("@/views/data-source/pvc/pvc-new.vue"),
      meta: {
        pageTitle: "New data source",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.PvcAssets,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
      },
    },
    {
      path: "config-map/new",
      name: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_CONFIG_MAP_NEW,
      component: () => import("@/views/data-source/config-map/config-map-new.vue"),
      meta: {
        pageTitle: "New data source",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.CmVolumeAssets,
        minPermittedActions: [Action.Create, Action.Update],
        requiresAuth: true,
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
      },
    },
    {
      path: "secret-volume/new",
      name: DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_SECRET_VOLUME_NEW,
      component: () => import("@/views/data-source/secret-volume/secret-volume-new.vue"),
      meta: {
        pageTitle: "New data source",
        fullWidth: true,
        requestToLeave: true,
        resourceType: ResourceType.SecretVolumeAssets,
        minPermittedActions: [Action.Create, Action.Update],
        parentSubject: EMenuSubjectNames.WORKLOADS_MANAGER,
        requiresAuth: true,
      },
      beforeEnter: (to, from) => {
        if (from.meta.supportBack) {
          to.meta.backPageName = DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
          return;
        }
        to.meta.backPageName = (from.name as string) || DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
      },
    },
  ],
};
