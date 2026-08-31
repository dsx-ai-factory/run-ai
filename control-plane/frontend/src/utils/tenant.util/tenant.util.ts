/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ITenantClientConfigInfo, ITenantServerConfigInfo } from "@/models/tenant.model";

export const tenantUtil = {
  getClientConfigInfo,
  getServerConfigInfo,
};

const DEFAULT_SERVER_CLIENT_ID = "runai";

function getClientConfigInfo(issuerUrl: string, authRealm: string, isSso: boolean): ITenantClientConfigInfo {
  return isSso
    ? {
        airgapped: true,
        authFlow: "remote-browser",
        realm: authRealm,
        clientId: "runai-cli",
        idpIssuerUrl: issuerUrl,
        redirectUri: `${window.location.origin}/oauth-code`,
      }
    : {
        airgapped: true,
        authFlow: "cli",
        realm: authRealm,
        clientId: "runai-cli",
        idpIssuerUrl: issuerUrl,
        redirectUri: undefined,
      };
}

function getServerConfigInfo(issuerUrl: string, isSso: boolean): ITenantServerConfigInfo {
  return isSso
    ? {
        clientId: DEFAULT_SERVER_CLIENT_ID,
        issuerUrl: issuerUrl,
        groupsClaim: "groups",
      }
    : {
        clientId: DEFAULT_SERVER_CLIENT_ID,
        issuerUrl: issuerUrl,
        groupsClaim: undefined,
      };
}
