-- 000002_add_new_features.down.sql
-- 新機能ロールバック

DROP TABLE IF EXISTS approval_steps;
DROP TABLE IF EXISTS approval_flows;
DROP TABLE IF EXISTS holidays;
DROP TABLE IF EXISTS time_entries;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS attendance_corrections;
DROP TABLE IF EXISTS leave_balances;
DROP TABLE IF EXISTS overtime_requests;

ALTER TABLE attendances DROP COLUMN IF EXISTS clock_in_latitude;
ALTER TABLE attendances DROP COLUMN IF EXISTS clock_in_longitude;
ALTER TABLE attendances DROP COLUMN IF EXISTS clock_out_latitude;
ALTER TABLE attendances DROP COLUMN IF EXISTS clock_out_longitude;
