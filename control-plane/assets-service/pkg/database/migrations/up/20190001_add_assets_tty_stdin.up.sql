ALTER TABLE assets_environment ADD COLUMN if not exists stdin boolean;
ALTER TABLE assets_environment ADD COLUMN if not exists tty boolean;