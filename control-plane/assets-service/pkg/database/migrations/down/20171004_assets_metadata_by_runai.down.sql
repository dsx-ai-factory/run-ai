DO $$
BEGIN
    UPDATE assets.assets_metadata
    SET
        created_by = CASE
            WHEN created_by LIKE '%Run:ai%' THEN 'runai'
            ELSE created_by
        END,
        updated_by = CASE
            WHEN updated_by LIKE '%Run:ai%' THEN 'runai'
            ELSE updated_by
        END
    WHERE
        created_by LIKE '%Run:ai%'
        OR updated_by LIKE '%Run:ai%';
END $$;