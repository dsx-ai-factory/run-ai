Delete from tenants where status= 'Failed' and name = 'runai';

ALTER TABLE tenants ADD COLUMN IF NOT EXISTS "smg" BOOLEAN NOT NULL DEFAULT(TRUE);
