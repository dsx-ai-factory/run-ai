CREATE TABLE IF NOT EXISTS helm_repos (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id TEXT,
    cluster_id UUID,
    department_id TEXT,
    project_id TEXT,
    scope_type TEXT NOT NULL,
    scope_id TEXT,
    name TEXT NOT NULL,
    url TEXT NOT NULL,
    description TEXT,
    version TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS helm_repos_scope_url_idx 
ON helm_repos (scope_type, tenant_id, scope_id, url)
WHERE deleted_at IS NULL;
