/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import Axios, { type AxiosInstance, type AxiosRequestConfig } from "axios";
import { useAuthStore } from "@/stores/auth.store";
import { installCacheInterceptors } from "@/services/infra/cache";

const TOKEN_REFRESH_LOCK = "runai-token-refresh";

let isStartupPhase = true;

interface RetryableAxiosRequestConfig extends AxiosRequestConfig {
  _retryAfter401?: boolean;
}

let forceTokenRefresh = false;

function isAuthEndpoint(url: string | undefined): boolean {
  if (!url) return false;
  const urlPath = url.split("?")[0];
  return urlPath.endsWith("/token") || url.includes("/auth/") || url.includes("/login");
}

function shouldSkip401Handling(url: string | undefined): boolean {
  if (!url) return false;
  return url.endsWith("/me/password") || url.includes("/api/v1/logout");
}

const createAuthenticatedAxios = (): AxiosInstance => {
  const axios = Axios.create({
    withCredentials: true,
  });

  axios.interceptors.request.use(
    async (config) => {
      const authStore = useAuthStore();

      if (isAuthEndpoint(config.url) || !authStore.accessToken) {
        return config;
      }

      if (isStartupPhase && authStore.accessToken) {
        isStartupPhase = false;
      }

      try {
        await performCoordinatedTokenRefresh();

        if (config.headers && authStore.accessToken) {
          config.headers["Authorization"] = `Bearer ${authStore.accessToken}`;
        }
      } catch (error: unknown) {
        console.error("Failed to refresh token on request:", error);
        throw new Error("Authentication failed - unable to refresh token");
      }

      return config;
    },
    (error) => {
      return Promise.reject(error);
    },
  );

  axios.interceptors.response.use(
    (response) => {
      return response;
    },
    async (error) => {
      const authStore = useAuthStore();

      if (Axios.isAxiosError(error) && error.response?.status === 400 && error.config?.url?.includes("/token")) {
        isStartupPhase = false;

        authStore.clearTokens();
        await authStore.logout();
        return Promise.reject(error);
      }

      if (Axios.isAxiosError(error) && error.response?.status === 401 && error.config) {
        const config = error.config as RetryableAxiosRequestConfig;

        // Skip 401 handling for endpoints where 401 means "invalid input" (e.g., wrong current password)
        if (shouldSkip401Handling(config.url)) {
          return Promise.reject(error);
        }

        if (isAuthEndpoint(config.url)) {
          if (isStartupPhase) {
            isStartupPhase = false;
            return Promise.reject(error);
          }

          authStore.clearTokens();
          return Promise.reject(error);
        }

        if (config._retryAfter401) {
          if (isStartupPhase) {
            isStartupPhase = false;
            return Promise.reject(error);
          }

          console.error("Still getting 401 after retry - logging out");
          await authStore.logout();
          return Promise.reject(error);
        }

        forceTokenRefresh = true;
        config._retryAfter401 = true;

        try {
          await performCoordinatedTokenRefresh();
          if (config.headers && authStore.accessToken)
            config.headers["Authorization"] = `Bearer ${authStore.accessToken}`;
          return await axios(config);
        } catch (refreshError) {
          return Promise.reject(refreshError);
        }
      }

      return Promise.reject(error);
    },
  );

  // Install caching interceptors for API response caching
  installCacheInterceptors(axios);

  return axios;
};

function isTokenValid(token: string): boolean {
  if (forceTokenRefresh) {
    return false;
  }

  const authStore = useAuthStore();
  return !authStore.isTokenAboutToExpire(token);
}

function isAuthenticationError(error: unknown): boolean {
  if (Axios.isAxiosError(error) && error.response && error.response.status) {
    return error.response.status >= 400;
  }
  return false;
}

async function doTokenRefresh(): Promise<void> {
  const authStore = useAuthStore();

  try {
    await authStore.updateAccessToken();
    forceTokenRefresh = false;
  } catch (error: unknown) {
    const isNetworkError = Axios.isAxiosError(error) && !error.response;

    if (!isNetworkError) {
      console.error("Token refresh failed:", error);
    }

    if (isAuthenticationError(error)) {
      if (isStartupPhase) {
        isStartupPhase = false;
        throw error;
      }

      await authStore.logout();
    } else {
      throw error;
    }
  }
}

async function performCoordinatedTokenRefresh(): Promise<void> {
  await navigator.locks.request(TOKEN_REFRESH_LOCK, async () => {
    const authStore = useAuthStore();
    const tokenIsValid = isTokenValid(authStore.accessToken);

    if (!tokenIsValid) {
      await doTokenRefresh();
    }
  });
}

export const authenticatedAxios = createAuthenticatedAxios();
