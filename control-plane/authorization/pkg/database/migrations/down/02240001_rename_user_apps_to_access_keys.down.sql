-- Rollback: rename access-keys resource type back to user-apps
UPDATE permissions 
SET resource_type = 'user-apps' 
WHERE resource_type = 'access-keys';

