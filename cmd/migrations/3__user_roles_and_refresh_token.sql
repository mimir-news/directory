-- +migrate Up
ALTER TABLE app_user 
ADD COLUMN role VARCHAR(50) DEFAULT 'USER';

ALTER TABLE session 
ADD COLUMN refresh_token VARCHAR(200);

ALTER TABLE session 
ADD COLUMN is_active BOOLEAN;

ALTER TABLE session 
ADD COLUMN deleted_at TIMESTAMP;

-- +migrate Down
ALTER TABLE session DROP COLUMN deleted_at;
ALTER TABLE session DROP COLUMN is_active;
ALTER TABLE session DROP COLUMN refresh_token;
ALTER TABLE app_user DROP COLUMN role;