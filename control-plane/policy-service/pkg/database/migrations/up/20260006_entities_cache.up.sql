CREATE TABLE IF NOT EXISTS entities_cache (
    id            UUID     NOT NULL PRIMARY KEY,

    tenant_id     integer  NOT NULL,
    entity_type   TEXT     NOT NULL CHECK (entity_type IN ('breakdown')),
    entity_id     TEXT     NOT NULL,

    cached_info   jsonb    NOT NULL,

    UNIQUE (tenant_id, entity_type, entity_id),

    -- required so entities_dependency can use a composite FK (id, tenant_id).
    -- a plain PK on id alone does not satisfy a multi-column REFERENCES clause.
    UNIQUE (id, tenant_id)
);

CREATE TABLE IF NOT EXISTS entities_dependency (
    -- composite FK: a dependency row can only reference a cache row that
    -- belongs to the *same* tenant. a single-column REFERENCES entities_cache(id)
    -- would allow cross-tenant links because it ignores tenant_id.
    -- no ON DELETE CASCADE: all deletion logic is handled in code.
    cache_id       UUID     NOT NULL,
    tenant_id      integer  NOT NULL,

    component_type TEXT     NOT NULL CHECK (component_type IN (
                             'mapping', 'field_group', 'formatter')),
    component_id   TEXT     NOT NULL,

    PRIMARY KEY (component_type, component_id, cache_id),
    FOREIGN KEY (cache_id, tenant_id) REFERENCES entities_cache (id, tenant_id)
);

-- secondary index on cache_id - used by both deletion paths in code:
--   entity deleted     -> delete deps WHERE cache_id = ?
--   component modified -> collect cache_ids WHERE component_type=? AND component_id=?,
--                         then delete deps and cache rows
CREATE INDEX IF NOT EXISTS entities_dependency_by_cache
    ON entities_dependency (cache_id);
