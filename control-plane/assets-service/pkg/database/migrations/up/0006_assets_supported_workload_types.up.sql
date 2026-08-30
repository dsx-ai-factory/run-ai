CREATE TABLE IF NOT EXISTS assets_supported_workload_types (
                                                               asset_id UUID NOT NULL PRIMARY KEY,
                                                               workspace BOOLEAN default true,
                                                               training BOOLEAN default true,
                                                               distributed BOOLEAN default true,
                                                               dist_framework TEXT
);