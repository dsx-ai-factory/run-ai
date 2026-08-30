-- department.parent_id is unused (multi-hierarchy plan was dropped) and
-- is always NULL. The self-referencing FK forces Postgres to scan the referencing
-- column on every DELETE FROM department, which is prohibitively slow on the
-- soft-delete-heavy table. Dropping the FK avoids the scan and removes the need
-- for an index on parent_id.

ALTER TABLE department DROP CONSTRAINT IF EXISTS fk_parent;
