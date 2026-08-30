ALTER TABLE department
    ADD CONSTRAINT fk_parent FOREIGN KEY (parent_id) REFERENCES department(id);
