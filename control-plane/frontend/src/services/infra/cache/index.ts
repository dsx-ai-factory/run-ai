/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

export { cacheService, CacheService } from "./cache.service";
export { installCacheInterceptors, cacheUtils } from "./cache.interceptor";
export {
  entityTTL,
  pathToEntity,
  ECacheEntity,
  DEFAULT_TTL,
  SECONDS,
  MINUTES,
  HOURS,
  CACHE_DEBUG,
  excludeFromCache,
} from "./cache.config";
