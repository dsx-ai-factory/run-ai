CREATE TABLE IF NOT EXISTS reports (
                                       created_at TIMESTAMPTZ NOT NULL,
                                       created_by TEXT NOT NULL,
                                       id UUID NOT NULL PRIMARY KEY,
                                       name TEXT NOT NULL,
                                       tenant_id INT NOT NULL,
                                       status TEXT NOT NULL,
                                       status_updated_at TIMESTAMPTZ NOT NULL,
                                       status_message TEXT NOT NULL,
                                       path TEXT,
                                       labels JSONB NOT NULL DEFAULT '{}'::JSONB,
                                       start TIMESTAMPTZ NOT NULL,
                                       "end" TIMESTAMPTZ NOT NULL,
                                       additional_fields JSONB NOT NULL DEFAULT '{}'::JSONB
);

CREATE UNIQUE INDEX idx_reports_tenant_name ON reports (tenant_id, name);

CREATE TABLE IF NOT EXISTS report_metrics (
                                      report_id UUID,
                                      FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE,
                                      name TEXT NOT NULL,
                                      display_name TEXT NOT NULL,
                                      filters JSONB NOT NULL DEFAULT '{}'::JSONB
);

CREATE OR REPLACE FUNCTION update_report_status_update_time()
    RETURNS TRIGGER AS $$
BEGIN
    IF OLD is null OR NEW.status <> OLD.status THEN
        NEW.status_updated_at = NOW();
    ELSE
        NEW.status_updated_at = OLD.status_updated_at;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
BEGIN
    BEGIN
        CREATE TRIGGER set_report_status_update_time
            BEFORE INSERT OR UPDATE
            ON reports
            FOR EACH ROW
        EXECUTE FUNCTION update_report_status_update_time();
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;
END $$;
