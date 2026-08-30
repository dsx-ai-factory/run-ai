DO $$
BEGIN
    UPDATE assets.assets_metadata
    SET
        created_by = CASE
            WHEN created_by LIKE '%runai%' THEN 'Run:ai'
            ELSE created_by
        END,
        updated_by = CASE
            WHEN updated_by LIKE '%runai%' THEN 'Run:ai'
            ELSE updated_by
        END
    WHERE
        created_by LIKE '%runai%'
        OR updated_by LIKE '%runai%';
END $$;