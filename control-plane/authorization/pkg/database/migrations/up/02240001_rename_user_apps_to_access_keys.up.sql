-- Update permissions table: rename user-apps resource type to access-keys
UPDATE permissions 
SET resource_type = 'access-keys' 
WHERE resource_type = 'user-apps';

