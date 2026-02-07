-- 000001_init_schema.down.sql
-- 初期スキーマの削除

DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS shifts;
DROP TABLE IF EXISTS leave_requests;
DROP TABLE IF EXISTS attendances;
ALTER TABLE IF EXISTS departments DROP CONSTRAINT IF EXISTS fk_departments_manager;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS departments;
