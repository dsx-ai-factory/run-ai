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
