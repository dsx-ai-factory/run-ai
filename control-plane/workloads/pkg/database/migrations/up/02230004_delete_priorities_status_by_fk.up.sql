ALTER TABLE priority_statuses DROP CONSTRAINT fk_priority;
ALTER TABLE priority_statuses ADD CONSTRAINT fk_priority FOREIGN KEY (priority_id) REFERENCES priorities(id) ON DELETE CASCADE;
