/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig, AxiosError } from "axios";
import { cacheService } from "./cache.service";
import type { ECacheEntity } from "./cache.config";

// Extend InternalAxiosRequestConfig to include our custom properties
interface CacheableRequestConfig extends InternalAxiosRequestConfig {
  __cacheKey?: string;
  __skipCache?: boolean;
  __fromCache?: boolean;
}

/**
 * HTTP methods that should be cached.
 */
const CACHEABLE_METHODS = ["get", "GET"];

/**
 * HTTP methods that should invalidate the cache.
 */
const INVALIDATING_METHODS = ["post", "POST", "put", "PUT", "delete", "DELETE", "patch", "PATCH"];

/**
 * Installs caching interceptors on an Axios instance.
 *
 * Request Interceptor:
 * - For GET requests: Checks if there's a valid cached response
 * - If cached: Returns the cached response immediately (via rejected promise with special marker)
 * - If not cached: Proceeds with the request normally
 *
 * Response Interceptor:
 * - For successful GET responses: Caches the response
 * - For POST/PUT/DELETE/PATCH responses: Invalidates cache for the affected entity
 *
 * @param axios The Axios instance to install interceptors on
 */
export function installCacheInterceptors(axios: AxiosInstance): void {
  // Request interceptor - check cache before making request
  axios.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      const cacheConfig = config as CacheableRequestConfig;

      // Only cache GET requests
      if (!CACHEABLE_METHODS.includes(config.method || "")) {
        return config;
      }

      const url = config.url || "";

      // Check if this URL should be excluded from caching
      if (cacheService.shouldExclude(url)) {
        cacheConfig.__skipCache = true;
        return config;
      }

      // Get entity and TTL
      const entity = cacheService.getEntityFromUrl(url);
      const ttl = cacheService.getTTLForEntity(entity);

      // If TTL is 0, skip caching
      if (ttl <= 0) {
        cacheConfig.__skipCache = true;
        return config;
      }

      // Generate cache key
      const cacheKey = cacheService.getCacheKey(url, config.params as Record<string, unknown>);
      cacheConfig.__cacheKey = cacheKey;

      // Check for cached response
      const cachedResponse = cacheService.get<AxiosResponse>(cacheKey);
      if (cachedResponse) {
        // Return cached response by throwing a special "error" that we'll catch in the response interceptor
        // This is a pattern used because axios interceptors don't have a direct way to short-circuit
        const cachedResult: CachedResponseMarker = {
          __isCachedResponse: true,
          response: cachedResponse,
        };
        return Promise.reject(cachedResult);
      }

      // Check for pending identical request (deduplication)
      const pendingRequest = cacheService.getPendingRequest(cacheKey);
      if (pendingRequest) {
        const pendingResult: PendingRequestMarker = {
          __isPendingRequest: true,
          promise: pendingRequest,
        };
        return Promise.reject(pendingResult);
      }

      return config;
    },
    (error) => Promise.reject(error),
  );

  // Response interceptor - cache successful responses and handle cache markers
  axios.interceptors.response.use(
    (response: AxiosResponse) => {
      const config = response.config as CacheableRequestConfig;

      // Handle mutating requests - invalidate cache
      if (INVALIDATING_METHODS.includes(config.method || "")) {
        const url = config.url || "";
        const entity = cacheService.getEntityFromUrl(url);
        if (entity) {
          cacheService.invalidateEntity(entity);
        }
        return response;
      }

      // Handle GET requests - cache the response
      if (CACHEABLE_METHODS.includes(config.method || "") && !config.__skipCache && config.__cacheKey) {
        const entity = cacheService.getEntityFromUrl(config.url || "");
        const ttl = cacheService.getTTLForEntity(entity);

        if (ttl > 0 && entity) {
          // Clone the response to avoid issues with response object mutation
          const responseToCache: AxiosResponse = {
            data: response.data,
            status: response.status,
            statusText: response.statusText,
            headers: response.headers,
            config: response.config,
          };
          cacheService.set(config.__cacheKey, responseToCache, ttl, entity);
        }

        // Clear pending request
        cacheService.clearPendingRequest(config.__cacheKey);
      }

      return response;
    },
    (error: AxiosError | CachedResponseMarker | PendingRequestMarker) => {
      // Handle cached response marker
      if (isCachedResponseMarker(error)) {
        return Promise.resolve(error.response);
      }

      // Handle pending request marker (deduplication)
      if (isPendingRequestMarker(error)) {
        return error.promise as Promise<AxiosResponse>;
      }

      // Handle real errors
      if (isAxiosError(error)) {
        const config = error.config as CacheableRequestConfig | undefined;

        // Clear pending request on error
        if (config?.__cacheKey) {
          cacheService.clearPendingRequest(config.__cacheKey);
        }

        // Invalidate cache on mutating request errors too (to be safe)
        if (config && INVALIDATING_METHODS.includes(config.method || "")) {
          const entity = cacheService.getEntityFromUrl(config.url || "");
          if (entity) {
            cacheService.invalidateEntity(entity);
          }
        }
      }

      return Promise.reject(error);
    },
  );
}

// Type guards and marker interfaces

interface CachedResponseMarker {
  __isCachedResponse: true;
  response: AxiosResponse;
}

interface PendingRequestMarker {
  __isPendingRequest: true;
  promise: Promise<unknown>;
}

function isCachedResponseMarker(obj: unknown): obj is CachedResponseMarker {
  return (
    typeof obj === "object" &&
    obj !== null &&
    "__isCachedResponse" in obj &&
    (obj as CachedResponseMarker).__isCachedResponse === true
  );
}

function isPendingRequestMarker(obj: unknown): obj is PendingRequestMarker {
  return (
    typeof obj === "object" &&
    obj !== null &&
    "__isPendingRequest" in obj &&
    (obj as PendingRequestMarker).__isPendingRequest === true
  );
}

function isAxiosError(error: unknown): error is AxiosError {
  return (
    typeof error === "object" && error !== null && "isAxiosError" in error && (error as AxiosError).isAxiosError === true
  );
}

/**
 * Utility functions for manual cache management.
 * These can be used in components or services when needed.
 */
export const cacheUtils = {
  /**
   * Manually invalidate cache for a specific entity.
   * Useful when you know data has changed but didn't go through the API.
   */
  invalidateEntity(entity: ECacheEntity): void {
    cacheService.invalidateEntity(entity);
  },

  /**
   * Clear the entire cache.
   * Useful on logout or when user switches context.
   */
  clearAll(): void {
    cacheService.clear();
  },

  /**
   * Get cache statistics for debugging.
   */
  getStats() {
    return cacheService.getStats();
  },

  /**
   * Manually remove a specific cache entry by URL.
   */
  invalidateUrl(url: string, params?: Record<string, unknown>): void {
    const key = cacheService.getCacheKey(url, params);
    cacheService.delete(key);
  },
};
