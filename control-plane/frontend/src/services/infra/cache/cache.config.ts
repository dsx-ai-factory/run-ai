/**
 * Cache configuration for API requests.
 *
 * This configuration uses explicit path-to-entity mapping for precise control
 * over which endpoints are cached and how they're grouped for invalidation.
 *
 * The cache is automatically invalidated when:
 * - A POST/PUT/DELETE request is made to any path mapped to the same entity
 * - The TTL expires
 */

// TTL helper constants (in milliseconds)
export const SECONDS = 1000;
export const MINUTES = 60 * SECONDS;
export const HOURS = 60 * MINUTES;

/**
 * Default TTL for entities not explicitly configured.
 * Set to 0 to disable caching for unconfigured entities.
 */
export const DEFAULT_TTL = 0;

export enum ECacheEntity {
  Setting = "setting",
}

/**
 * Entity TTL configuration.
 *
 * Key: Entity name (used for grouping and invalidation)
 * Value: TTL in milliseconds. Set to 0 to disable caching for that entity.
 */
export const entityTTL: Record<ECacheEntity, number> = {
  [ECacheEntity.Setting]: 5 * MINUTES,
};

/**
 * Path to Entity mapping.
 *
 * Maps URL path patterns to entity names. When a request matches a pattern,
 * it uses the corresponding entity's TTL and invalidation group.
 *
 * Patterns:
 * - Use exact path segments (e.g., "/v1/k8s/setting")
 * - Paths are matched as prefixes, so "/v1/k8s/setting" matches "/v1/k8s/setting/123"
 * - Order matters: first match wins, so put more specific paths first
 *
 * Multiple paths can map to the same entity for grouped invalidation.
 * Example: POST to /v1/k8s/setting will invalidate cache for all "setting" entity paths.
 */
export const pathToEntity: Record<string, ECacheEntity> = {
  "/v1/k8s/setting": ECacheEntity.Setting,
};

/**
 * URL patterns to always exclude from caching.
 * Useful for endpoints that should never be cached regardless of method.
 */
export const excludeFromCache: RegExp[] = [
  /\/health/,
  /\/ready/,
  /\/token/,
  /\/login/,
  /\/logout/,
  /\/auth\//,
  /\/metrics\//,
  /\/audit\//,
];

/**
 * Whether to enable debug logging for cache operations.
 * Set to true during development to see cache hits/misses in the console.
 */
export const CACHE_DEBUG = false;
