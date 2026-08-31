/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { entityTTL, DEFAULT_TTL, excludeFromCache, CACHE_DEBUG, pathToEntity, type ECacheEntity } from "./cache.config";

interface CacheEntry<T = unknown> {
  data: T;
  timestamp: number;
  ttl: number;
  entity: ECacheEntity;
}

interface PendingRequest {
  promise: Promise<unknown>;
  timestamp: number;
}

/**
 * API Response Cache Service
 *
 * This service provides a caching layer for API requests with the following features:
 * - TTL-based cache expiration per entity type
 * - Automatic cache invalidation on mutating requests (POST/PUT/DELETE/PATCH)
 * - Request deduplication (prevents duplicate in-flight requests)
 * - Manual cache invalidation API
 * - Explicit path-to-entity mapping for precise control
 */
class CacheService {
  private cache: Map<string, CacheEntry> = new Map();
  private pendingRequests: Map<string, PendingRequest> = new Map();

  /**
   * Generates a unique cache key from the request URL and params.
   */
  getCacheKey(url: string, params?: Record<string, unknown>): string {
    const baseKey = this.getPathFromUrl(url);
    if (params && Object.keys(params).length > 0) {
      const sortedParams = this.sortObjectKeys(params);
      return `${baseKey}?${JSON.stringify(sortedParams)}`;
    }
    return baseKey;
  }

  /**
   * Extracts the path from a URL (removes origin if present).
   */
  private getPathFromUrl(url: string): string {
    try {
      // Handle full URLs
      if (url.startsWith("http://") || url.startsWith("https://")) {
        const urlObj = new URL(url);
        return urlObj.pathname;
      }
      // Already a path
      return url.split("?")[0];
    } catch {
      return url.split("?")[0];
    }
  }

  /**
   * Sorts object keys recursively for consistent cache key generation.
   */
  private sortObjectKeys(obj: unknown): unknown {
    if (typeof obj !== "object" || obj === null) return obj;
    if (Array.isArray(obj)) return obj.map((item) => this.sortObjectKeys(item));

    const record = obj as Record<string, unknown>;
    return Object.keys(record)
      .sort()
      .reduce(
        (result, key) => {
          result[key] = this.sortObjectKeys(record[key]);
          return result;
        },
        {} as Record<string, unknown>,
      );
  }

  /**
   * Gets the entity name for a URL using the explicit path mapping.
   * Returns null if no mapping is found.
   */
  getEntityFromUrl(url: string): ECacheEntity | null {
    const path = this.getPathFromUrl(url);

    // Find matching path pattern (prefix match)
    for (const [pattern, entity] of Object.entries(pathToEntity)) {
      if (path === pattern || path.startsWith(pattern + "/")) {
        return entity;
      }
    }

    return null;
  }

  /**
   * Gets all paths that map to a given entity.
   */
  getPathsForEntity(entity: string): string[] {
    return Object.entries(pathToEntity)
      .filter(([_, e]) => e === entity)
      .map(([path, _]) => path);
  }

  /**
   * Gets the TTL for a given entity type.
   */
  getTTLForEntity(entity: ECacheEntity | null): number {
    if (!entity) return DEFAULT_TTL;
    return entityTTL[entity] ?? DEFAULT_TTL;
  }

  /**
   * Checks if a URL should be excluded from caching.
   */
  shouldExclude(url: string): boolean {
    return excludeFromCache.some((pattern) => pattern.test(url));
  }

  /**
   * Gets a cached response if it exists and is still valid.
   */
  get<T>(key: string): T | null {
    const entry = this.cache.get(key);

    if (!entry) {
      this.log("MISS", key);
      return null;
    }

    const age = Date.now() - entry.timestamp;
    if (age > entry.ttl) {
      this.log("EXPIRED", key, `age: ${age}ms, ttl: ${entry.ttl}ms`);
      this.cache.delete(key);
      return null;
    }

    this.log("HIT", key, `age: ${age}ms, ttl: ${entry.ttl}ms`);
    return entry.data as T;
  }

  /**
   * Stores a response in the cache.
   */
  set<T>(key: string, data: T, ttl: number, entity: ECacheEntity): void {
    if (ttl <= 0) return;

    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
      entity,
    });
    this.log("SET", key, `entity: ${entity}, ttl: ${ttl}ms`);
  }

  /**
   * Removes a specific entry from the cache.
   */
  delete(key: string): void {
    this.cache.delete(key);
    this.log("DELETE", key);
  }

  /**
   * Invalidates all cache entries for a specific entity.
   * Called automatically when a mutating request (POST/PUT/DELETE/PATCH) is made.
   */
  invalidateEntity(entity: ECacheEntity): void {
    const keysToDelete: string[] = [];

    this.cache.forEach((entry, key) => {
      if (entry.entity === entity) {
        keysToDelete.push(key);
      }
    });

    keysToDelete.forEach((key) => {
      this.cache.delete(key);
    });

    if (keysToDelete.length > 0) {
      this.log("INVALIDATE", entity, `cleared ${keysToDelete.length} entries`);
    }
  }

  /**
   * Clears the entire cache.
   */
  clear(): void {
    this.cache.clear();
    this.pendingRequests.clear();
    this.log("CLEAR", "all");
  }

  /**
   * Returns the current cache size.
   */
  get size(): number {
    return this.cache.size;
  }

  /**
   * Request deduplication: checks if there's already an identical request in flight.
   */
  getPendingRequest(key: string): Promise<unknown> | null {
    const pending = this.pendingRequests.get(key);
    if (pending) {
      // Clean up stale pending requests (older than 30 seconds)
      if (Date.now() - pending.timestamp > 30000) {
        this.pendingRequests.delete(key);
        return null;
      }
      this.log("DEDUPE", key);
      return pending.promise;
    }
    return null;
  }

  /**
   * Removes a pending request after it completes.
   */
  clearPendingRequest(key: string): void {
    this.pendingRequests.delete(key);
  }

  /**
   * Debug logging helper.
   */
  private log(action: string, key: string, details?: string): void {
    if (CACHE_DEBUG) {
      const entity = this.getEntityFromUrl(key) || "unknown";
      const message = `[Cache ${action}] ${entity}: ${key}`;
      console.log(details ? `${message} (${details})` : message);
    }
  }

  /**
   * Returns cache statistics for debugging.
   */
  getStats(): { size: number; entries: { key: string; entity: ECacheEntity; age: number; ttl: number }[] } {
    const entries: { key: string; entity: ECacheEntity; age: number; ttl: number }[] = [];

    this.cache.forEach((entry, key) => {
      entries.push({
        key,
        entity: entry.entity,
        age: Date.now() - entry.timestamp,
        ttl: entry.ttl,
      });
    });

    return {
      size: this.cache.size,
      entries,
    };
  }
}

// Export singleton instance
export const cacheService = new CacheService();

// Also export class for testing purposes
export { CacheService };
