/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable */
import { useAuthStore } from "@/stores/auth.store";
import type { IWebSocketSubscription } from "@/models/global.model";

const webSocketsMap: Record<string, WebSocket> = {};

export function getUrlByServer(baseUrl: string, endpoint: string, params?: any, isQueryParams: boolean = true): string {
  const fixedUrl = baseUrl.endsWith("/") ? baseUrl : `${baseUrl}/`;
  const wsUrl: string = fixedUrl.replace("http", "ws");
  let url = `${wsUrl}${endpoint}`;

  if (isQueryParams && params && Object.keys(params).length > 0) {
    const queryParams = new URLSearchParams();
    Object.keys(params).forEach((key) => {
      if (params[key] !== undefined && params[key] !== null) {
        queryParams.append(key, params[key].toString());
      }
    });
    if (queryParams.toString()) {
      url += `?${queryParams.toString()}`;
    }
  }

  return url;
}

function unsubscribe(socketId: string): void {
  const ws: WebSocket = webSocketsMap[socketId];
  if (!ws) {
    console.error(`WebSocket ${socketId} not found`);
    return;
  }
  ws.close();
  console.info(`WebSocket ${socketId} closed`);
  delete webSocketsMap[socketId];
}

export const webSocketsService = {
  // INFRA
  subscribe(
    baseUrl: string,
    endpoint: string,
    params: any,
    version: string | null | undefined,
    isUseClusterApi: boolean = true,
  ): IWebSocketSubscription {
    // if cluster and backend apis use different tokens this will fail authentication. currently we use/support ws for cluster
    const accessToken = useAuthStore().accessToken;
    const protocols = isUseClusterApi ? [`auth.bearer.${accessToken}`, "Bearer"] : ["bearer", accessToken];

    const ws: WebSocket = new WebSocket(getUrlByServer(baseUrl, endpoint, params, isUseClusterApi), protocols);
    const socketId: string = `${endpoint}-${Date.now()}`;
    webSocketsMap[socketId] = ws;

    const delegations: IWebSocketSubscription = {
      onMessage: (messageEvent: MessageEvent) => {},
      onError: (errorEvent: Event) => {},
      onOpen: (openEvent: Event) => {},
      onClose: (closeEvent: CloseEvent) => {},
      unsubscribe: () => unsubscribe.bind(this)(socketId),
      websocket: ws,
    };

    ws.onopen = (openEvent: Event): void => {
      if (!isUseClusterApi && params) {
        ws.send(JSON.stringify(params));
      }
      delegations.onOpen && delegations.onOpen(openEvent);
    };

    ws.onmessage = (message: MessageEvent): void => {
      delegations.onMessage && delegations.onMessage(message);
    };

    ws.onerror = (error): void => {
      delegations.onError && delegations.onError(error);
    };

    ws.onclose = (closeEvent: CloseEvent): void => {
      delegations.onClose && delegations.onClose(closeEvent);
      console.info(`WebSocket closed`);
      return;
    };

    return delegations;
  },
};
