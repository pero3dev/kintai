-- seed.sql
-- 開発環境用初期データ
-- パスワードは全て「password123」

-- ===== 部署データ =====
INSERT INTO departments (id, name, created_at, updated_at) VALUES
    ('d0000001-0000-0000-0000-000000000001'::uuid, '経営企画部', NOW(), NOW()),
    ('d0000002-0000-0000-0000-000000000002'::uuid, '開発部', NOW(), NOW()),
    ('d0000003-0000-0000-0000-000000000003'::uuid, '営業部', NOW(), NOW()),
    ('d0000004-0000-0000-0000-000000000004'::uuid, '人事部', NOW(), NOW()),
    ('d0000005-0000-0000-0000-000000000005'::uuid, '総務部', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- ===== ユーザーデータ =====
-- パスワード: password123 (bcryptハッシュ)
-- $2a$10$kOA.AHA4Q2p.hD8w8fP5DeIdVBaN0f0QGFbn6aIhNZg9acHNN3la.

-- 管理者
INSERT INTO users (id, email, password_hash, first_name, last_name, role, department_id, is_active, created_at, updated_at) VALUES
    ('a0000001-0000-0000-0000-000000000001'::uuid, 'admin@example.com', '$2a$10$kOA.AHA4Q2p.hD8w8fP5DeIdVBaN0f0QGFbn6aIhNZg9acHNN3la.', '管理者', 'ユーザー', 'admin', 'd0000001-0000-0000-0000-000000000001'::uuid, true, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- マネージャー
INSERT INTO users (id, email, password_hash, first_name, last_name, role, department_id, is_active, created_at, updated_at) VALUES
    ('b0000001-0000-0000-0000-000000000002'::uuid, 'manager@example.com', '$2a$10$kOA.AHA4Q2p.hD8w8fP5DeIdVBaN0f0QGFbn6aIhNZg9acHNN3la.', '部門', 'マネージャー', 'manager', 'd0000002-0000-0000-0000-000000000002'::uuid, true, NOW(), NOW()),
    ('b0000002-0000-0000-0000-000000000003'::uuid, 'manager2@example.com', '$2a$10$kOA.AHA4Q2p.hD8w8fP5DeIdVBaN0f0QGFbn6aIhNZg9acHNN3la.', '営業', 'マネージャー', 'manager', 'd0000003-0000-0000-0000-000000000003'::uuid, true, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- 一般従業員
INSERT INTO users (id, email, password_hash, first_name, last_name, role, department_id, is_active, created_at, updated_at) VALUES
    ('c0000001-0000-0000-0000-000000000004'::uuid, 'employee1@example.com', '$2a$10$kOA.AHA4Q2p.hD8w8fP5DeIdVBaN0f0QGFbn6aIhNZg9acHNN3la.', '山田', '太郎', 'employee', 'd0000002-0000-0000-0000-000000000002'::uuid, true, NOW(), NOW()),
    ('c0000002-0000-0000-0000-000000000005'::uuid, 'employee2@example.com', '$2a$10$kOA.AHA4Q2p.hD8w8fP5DeIdVBaN0f0QGFbn6aIhNZg9acHNN3la.', '佐藤', '花子', 'employee', 'd0000002-0000-0000-0000-000000000002'::uuid, true, NOW(), NOW()),
    ('c0000003-0000-0000-0000-000000000006'::uuid, 'employee3@example.com', '$2a$10$kOA.AHA4Q2p.hD8w8fP5DeIdVBaN0f0QGFbn6aIhNZg9acHNN3la.', '鈴木', '一郎', 'employee', 'd0000003-0000-0000-0000-000000000003'::uuid, true, NOW(), NOW()),
    ('c0000004-0000-0000-0000-000000000007'::uuid, 'employee4@example.com', '$2a$10$kOA.AHA4Q2p.hD8w8fP5DeIdVBaN0f0QGFbn6aIhNZg9acHNN3la.', '田中', '美咲', 'employee', 'd0000003-0000-0000-0000-000000000003'::uuid, true, NOW(), NOW()),
    ('c0000005-0000-0000-0000-000000000008'::uuid, 'employee5@example.com', '$2a$10$kOA.AHA4Q2p.hD8w8fP5DeIdVBaN0f0QGFbn6aIhNZg9acHNN3la.', '高橋', '健太', 'employee', 'd0000004-0000-0000-0000-000000000004'::uuid, true, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

-- 部署のマネージャーを設定
UPDATE departments SET manager_id = 'a0000001-0000-0000-0000-000000000001'::uuid WHERE id = 'd0000001-0000-0000-0000-000000000001'::uuid;
UPDATE departments SET manager_id = 'b0000001-0000-0000-0000-000000000002'::uuid WHERE id = 'd0000002-0000-0000-0000-000000000002'::uuid;
UPDATE departments SET manager_id = 'b0000002-0000-0000-0000-000000000003'::uuid WHERE id = 'd0000003-0000-0000-0000-000000000003'::uuid;

-- ===== サンプル勤怠データ（今日の分）=====
INSERT INTO attendances (id, user_id, date, clock_in, status, created_at, updated_at) VALUES
    (gen_random_uuid(), 'c0000001-0000-0000-0000-000000000004'::uuid, CURRENT_DATE, NOW() - interval '2 hours', 'present', NOW(), NOW()),
    (gen_random_uuid(), 'c0000002-0000-0000-0000-000000000005'::uuid, CURRENT_DATE, NOW() - interval '3 hours', 'present', NOW(), NOW())
ON CONFLICT (user_id, date) DO NOTHING;

-- ===== サンプル休暇申請 =====
INSERT INTO leave_requests (id, user_id, leave_type, start_date, end_date, reason, status, created_at, updated_at) VALUES
    (gen_random_uuid(), 'c0000003-0000-0000-0000-000000000006'::uuid, 'paid', CURRENT_DATE + interval '7 days', CURRENT_DATE + interval '9 days', '家族旅行のため', 'pending', NOW(), NOW()),
    (gen_random_uuid(), 'c0000004-0000-0000-0000-000000000007'::uuid, 'sick', CURRENT_DATE + interval '3 days', CURRENT_DATE + interval '3 days', '通院のため', 'pending', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ===== サンプルシフト（今週分）=====
INSERT INTO shifts (id, user_id, date, shift_type, created_at, updated_at) VALUES
    (gen_random_uuid(), 'c0000001-0000-0000-0000-000000000004'::uuid, CURRENT_DATE, 'morning', NOW(), NOW()),
    (gen_random_uuid(), 'c0000002-0000-0000-0000-000000000005'::uuid, CURRENT_DATE, 'day', NOW(), NOW()),
    (gen_random_uuid(), 'c0000003-0000-0000-0000-000000000006'::uuid, CURRENT_DATE, 'evening', NOW(), NOW()),
    (gen_random_uuid(), 'c0000001-0000-0000-0000-000000000004'::uuid, CURRENT_DATE + interval '1 day', 'day', NOW(), NOW()),
    (gen_random_uuid(), 'c0000002-0000-0000-0000-000000000005'::uuid, CURRENT_DATE + interval '1 day', 'morning', NOW(), NOW())
ON CONFLICT (user_id, date) DO NOTHING;
