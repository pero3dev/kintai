-- seed_performance.sql
-- パフォーマンステスト用データ
-- 実行前に seed.sql を実行しておくこと

-- ===== 追加ユーザー（30人）=====
-- パスワード: password123
INSERT INTO users (id, email, password_hash, first_name, last_name, role, department_id, is_active, created_at, updated_at)
SELECT
    gen_random_uuid(),
    'user' || n || '@example.com',
    '$2a$10$kOA.AHA4Q2p.hD8w8fP5DeIdVBaN0f0QGFbn6aIhNZg9acHNN3la.',
    (ARRAY['田中', '佐藤', '鈴木', '高橋', '伊藤', '渡辺', '山本', '中村', '小林', '加藤'])[1 + (n % 10)],
    (ARRAY['太郎', '花子', '一郎', '美咲', '健太', '陽子', '翔太', '愛', '大輔', '真由'])[1 + (n % 10)],
    'employee',
    (ARRAY['d0000002-0000-0000-0000-000000000002', 'd0000003-0000-0000-0000-000000000003', 'd0000004-0000-0000-0000-000000000004', 'd0000005-0000-0000-0000-000000000005']::uuid[])[1 + (n % 4)],
    true,
    NOW(),
    NOW()
FROM generate_series(10, 39) n
ON CONFLICT (email) DO NOTHING;

-- 全ユーザーIDを取得するための一時テーブル
CREATE TEMP TABLE temp_users AS
SELECT id, row_number() OVER () as rn FROM users WHERE is_active = true;

-- ===== 過去90日分の勤怠データ（全ユーザー）=====
INSERT INTO attendances (id, user_id, date, clock_in, clock_out, clock_in_latitude, clock_in_longitude, clock_out_latitude, clock_out_longitude, work_minutes, overtime_minutes, status, note, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.id,
    d::date,
    (d::date + time '09:00:00' + (random() * interval '30 minutes'))::timestamp,
    (d::date + time '18:00:00' + (random() * interval '120 minutes'))::timestamp,
    35.6762 + (random() - 0.5) * 0.01,  -- 東京周辺
    139.6503 + (random() - 0.5) * 0.01,
    35.6762 + (random() - 0.5) * 0.01,
    139.6503 + (random() - 0.5) * 0.01,
    480 + floor(random() * 120)::int,  -- 8-10時間
    floor(random() * 120)::int,         -- 0-2時間残業
    'present',
    CASE WHEN random() < 0.1 THEN '在宅勤務' ELSE NULL END,
    NOW(),
    NOW()
FROM temp_users u
CROSS JOIN generate_series(CURRENT_DATE - interval '90 days', CURRENT_DATE - interval '1 day', interval '1 day') d
WHERE EXTRACT(DOW FROM d) NOT IN (0, 6)  -- 土日除外
  AND random() > 0.05  -- 5%は欠勤
ON CONFLICT (user_id, date) DO NOTHING;

-- ===== 残業申請データ（過去3ヶ月、各ユーザー月5件程度）=====
INSERT INTO overtime_requests (id, user_id, date, planned_minutes, actual_minutes, reason, status, approved_by, approved_at, rejected_reason, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.id,
    (CURRENT_DATE - (floor(random() * 90)::int || ' days')::interval)::date,
    30 + floor(random() * 5)::int * 30,  -- 30, 60, 90, 120, 150分
    30 + floor(random() * 5)::int * 30,
    (ARRAY['納期対応のため', 'クライアント対応', 'システム障害対応', 'リリース準備', '月末処理'])[1 + floor(random() * 5)::int],
    (ARRAY['pending', 'approved', 'approved', 'approved', 'rejected'])[1 + floor(random() * 5)::int],
    CASE WHEN random() > 0.3 THEN 'a0000001-0000-0000-0000-000000000001'::uuid ELSE NULL END,
    CASE WHEN random() > 0.3 THEN NOW() - interval '1 day' ELSE NULL END,
    CASE WHEN random() < 0.1 THEN '改善が必要です' ELSE NULL END,
    NOW() - (floor(random() * 90) || ' days')::interval,
    NOW()
FROM temp_users u
CROSS JOIN generate_series(1, 15) n
ON CONFLICT DO NOTHING;

-- ===== 有給残日数（全ユーザー）=====
INSERT INTO leave_balances (id, user_id, fiscal_year, leave_type, total_days, used_days, carried_over, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.id,
    2025,
    lt,
    CASE lt
        WHEN 'paid' THEN 10 + floor(random() * 10)::int
        WHEN 'sick' THEN 5
        WHEN 'special' THEN 3
    END,
    floor(random() * 5)::int,
    CASE WHEN lt = 'paid' THEN floor(random() * 5)::int ELSE 0 END,
    NOW(),
    NOW()
FROM temp_users u
CROSS JOIN (VALUES ('paid'), ('sick'), ('special')) AS lt_table(lt)
ON CONFLICT DO NOTHING;

-- ===== 追加の休暇申請（過去3ヶ月）=====
INSERT INTO leave_requests (id, user_id, leave_type, start_date, end_date, reason, status, approved_by, approved_at, rejected_reason, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.id,
    (ARRAY['paid', 'paid', 'paid', 'sick', 'special'])[1 + floor(random() * 5)::int],
    start_d,
    start_d + (floor(random() * 3) || ' days')::interval,
    (ARRAY['私用のため', '家族の用事', '通院', '旅行', 'リフレッシュ'])[1 + floor(random() * 5)::int],
    (ARRAY['pending', 'approved', 'approved', 'approved', 'rejected'])[1 + floor(random() * 5)::int],
    CASE WHEN random() > 0.2 THEN 'a0000001-0000-0000-0000-000000000001'::uuid ELSE NULL END,
    CASE WHEN random() > 0.2 THEN NOW() - interval '2 days' ELSE NULL END,
    CASE WHEN random() < 0.1 THEN '人員調整が必要' ELSE NULL END,
    NOW() - (floor(random() * 60) || ' days')::interval,
    NOW()
FROM temp_users u
CROSS JOIN generate_series(1, 5) n
CROSS JOIN LATERAL (SELECT (CURRENT_DATE - (floor(random() * 60)::int || ' days')::interval)::date AS start_d) sd
ON CONFLICT DO NOTHING;

-- ===== 勤怠修正申請 =====
INSERT INTO attendance_corrections (id, user_id, attendance_id, date, original_clock_in, original_clock_out, corrected_clock_in, corrected_clock_out, reason, status, approved_by, approved_at, created_at, updated_at)
SELECT
    gen_random_uuid(),
    a.user_id,
    a.id,
    a.date,
    a.clock_in,
    a.clock_out,
    (a.date + time '09:00:00')::timestamp,
    (a.date + time '18:30:00')::timestamp,
    (ARRAY['打刻忘れ', 'システム障害で打刻できず', '時間修正依頼', 'リモートワークで打刻漏れ'])[1 + floor(random() * 4)::int],
    (ARRAY['pending', 'approved', 'approved', 'rejected'])[1 + floor(random() * 4)::int],
    CASE WHEN random() > 0.3 THEN 'a0000001-0000-0000-0000-000000000001'::uuid ELSE NULL END,
    CASE WHEN random() > 0.3 THEN NOW() ELSE NULL END,
    NOW(),
    NOW()
FROM attendances a
WHERE random() < 0.03  -- 3%の勤怠に修正申請
LIMIT 100;

-- ===== 通知（全ユーザー、各10-20件）=====
INSERT INTO notifications (id, user_id, title, message, type, is_read, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.id,
    (ARRAY['休暇申請が承認されました', '残業申請が承認されました', '勤怠修正が完了しました', 'シフトが割り当てられました', '承認待ちの申請があります'])[1 + floor(random() * 5)::int],
    (ARRAY['申請内容をご確認ください', '詳細は勤怠画面からご確認ください', '新しい通知があります', 'ご確認お願いします', 'アクションが必要です'])[1 + floor(random() * 5)::int],
    (ARRAY['leave_approved', 'overtime_approved', 'correction_result', 'shift_assigned', 'general'])[1 + floor(random() * 5)::int],
    random() < 0.7,  -- 70%は既読
    NOW() - (floor(random() * 30) || ' days')::interval,
    NOW()
FROM temp_users u
CROSS JOIN generate_series(1, 15) n
ON CONFLICT DO NOTHING;

-- ===== プロジェクト =====
INSERT INTO projects (id, name, code, description, manager_id, budget_hours, status, created_at, updated_at) VALUES
    ('10000001-0000-0000-0000-000000000001'::uuid, '基幹システム刷新', 'PRJ-001', '既存基幹システムのモダナイゼーション', 'b0000001-0000-0000-0000-000000000002'::uuid, 5000, 'active', NOW(), NOW()),
    ('20000002-0000-0000-0000-000000000002'::uuid, 'モバイルアプリ開発', 'PRJ-002', '新規モバイルアプリの開発', 'b0000001-0000-0000-0000-000000000002'::uuid, 3000, 'active', NOW(), NOW()),
    ('30000003-0000-0000-0000-000000000003'::uuid, 'AI導入プロジェクト', 'PRJ-003', 'AI活用による業務効率化', 'a0000001-0000-0000-0000-000000000001'::uuid, 2000, 'active', NOW(), NOW()),
    ('40000004-0000-0000-0000-000000000004'::uuid, 'クラウド移行', 'PRJ-004', 'オンプレからAWSへの移行', 'b0000001-0000-0000-0000-000000000002'::uuid, 4000, 'active', NOW(), NOW()),
    ('50000005-0000-0000-0000-000000000005'::uuid, 'セキュリティ強化', 'PRJ-005', 'セキュリティ対策の強化', 'a0000001-0000-0000-0000-000000000001'::uuid, 1500, 'active', NOW(), NOW()),
    ('60000006-0000-0000-0000-000000000006'::uuid, '既存システム保守', 'PRJ-006', '既存システムの保守運用', 'b0000002-0000-0000-0000-000000000003'::uuid, 2400, 'active', NOW(), NOW()),
    ('70000007-0000-0000-0000-000000000007'::uuid, '完了済みプロジェクト', 'PRJ-007', '昨年度完了したプロジェクト', 'b0000001-0000-0000-0000-000000000002'::uuid, 3000, 'completed', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ===== 工数データ（過去3ヶ月、各ユーザー1日2-3件）=====
INSERT INTO time_entries (id, user_id, project_id, date, minutes, description, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.id,
    proj.id,
    d::date,
    60 + floor(random() * 6)::int * 30,  -- 60-240分
    (ARRAY['設計作業', 'コーディング', 'テスト', 'ドキュメント作成', 'ミーティング', 'コードレビュー', '調査・検討', 'バグ修正'])[1 + floor(random() * 8)::int],
    NOW(),
    NOW()
FROM temp_users u
CROSS JOIN generate_series(CURRENT_DATE - interval '90 days', CURRENT_DATE - interval '1 day', interval '1 day') d
CROSS JOIN generate_series(1, 2) entry_num
CROSS JOIN LATERAL (
    SELECT id FROM projects WHERE status = 'active' ORDER BY random() LIMIT 1
) proj
WHERE EXTRACT(DOW FROM d) NOT IN (0, 6)
  AND random() > 0.1
ON CONFLICT DO NOTHING;

-- ===== 祝日データ（2025-2026年度）=====
INSERT INTO holidays (id, date, name, holiday_type, is_recurring, created_at, updated_at) VALUES
    -- 2025年
    (gen_random_uuid(), '2025-01-01', '元日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-01-13', '成人の日', 'national', false, NOW(), NOW()),
    (gen_random_uuid(), '2025-02-11', '建国記念の日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-02-23', '天皇誕生日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-03-20', '春分の日', 'national', false, NOW(), NOW()),
    (gen_random_uuid(), '2025-04-29', '昭和の日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-05-03', '憲法記念日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-05-04', 'みどりの日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-05-05', 'こどもの日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-07-21', '海の日', 'national', false, NOW(), NOW()),
    (gen_random_uuid(), '2025-08-11', '山の日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-09-15', '敬老の日', 'national', false, NOW(), NOW()),
    (gen_random_uuid(), '2025-09-23', '秋分の日', 'national', false, NOW(), NOW()),
    (gen_random_uuid(), '2025-10-13', 'スポーツの日', 'national', false, NOW(), NOW()),
    (gen_random_uuid(), '2025-11-03', '文化の日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-11-23', '勤労感謝の日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-12-29', '年末休暇', 'company', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-12-30', '年末休暇', 'company', true, NOW(), NOW()),
    (gen_random_uuid(), '2025-12-31', '年末休暇', 'company', true, NOW(), NOW()),
    -- 2026年
    (gen_random_uuid(), '2026-01-01', '元日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2026-01-02', '年始休暇', 'company', true, NOW(), NOW()),
    (gen_random_uuid(), '2026-01-03', '年始休暇', 'company', true, NOW(), NOW()),
    (gen_random_uuid(), '2026-01-12', '成人の日', 'national', false, NOW(), NOW()),
    (gen_random_uuid(), '2026-02-11', '建国記念の日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2026-02-23', '天皇誕生日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2026-03-20', '春分の日', 'national', false, NOW(), NOW()),
    (gen_random_uuid(), '2026-04-29', '昭和の日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2026-05-03', '憲法記念日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2026-05-04', 'みどりの日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2026-05-05', 'こどもの日', 'national', true, NOW(), NOW()),
    (gen_random_uuid(), '2026-05-06', '振替休日', 'national', false, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ===== 承認フロー =====
INSERT INTO approval_flows (id, name, flow_type, is_active, created_at, updated_at) VALUES
    ('f0000001-0000-0000-0000-000000000001'::uuid, '休暇申請フロー（標準）', 'leave', true, NOW(), NOW()),
    ('f0000002-0000-0000-0000-000000000002'::uuid, '残業申請フロー（標準）', 'overtime', true, NOW(), NOW()),
    ('f0000003-0000-0000-0000-000000000003'::uuid, '勤怠修正フロー', 'correction', true, NOW(), NOW()),
    ('f0000004-0000-0000-0000-000000000004'::uuid, '休暇申請フロー（長期）', 'leave', false, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ===== 承認ステップ =====
INSERT INTO approval_steps (id, flow_id, step_order, step_type, approver_role, approver_id, created_at, updated_at) VALUES
    -- 休暇申請フロー（標準）- 1段階
    (gen_random_uuid(), 'f0000001-0000-0000-0000-000000000001'::uuid, 1, 'role', 'manager', NULL, NOW(), NOW()),
    -- 残業申請フロー - 1段階
    (gen_random_uuid(), 'f0000002-0000-0000-0000-000000000002'::uuid, 1, 'role', 'manager', NULL, NOW(), NOW()),
    -- 勤怠修正フロー - 2段階
    (gen_random_uuid(), 'f0000003-0000-0000-0000-000000000003'::uuid, 1, 'role', 'manager', NULL, NOW(), NOW()),
    (gen_random_uuid(), 'f0000003-0000-0000-0000-000000000003'::uuid, 2, 'role', 'admin', NULL, NOW(), NOW()),
    -- 休暇申請フロー（長期）- 2段階
    (gen_random_uuid(), 'f0000004-0000-0000-0000-000000000004'::uuid, 1, 'role', 'manager', NULL, NOW(), NOW()),
    (gen_random_uuid(), 'f0000004-0000-0000-0000-000000000004'::uuid, 2, 'specific_user', NULL, 'a0000001-0000-0000-0000-000000000001'::uuid, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ===== シフトデータ（今週〜来月）=====
INSERT INTO shifts (id, user_id, date, shift_type, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.id,
    d::date,
    (ARRAY['morning', 'day', 'day', 'evening', 'night', 'off'])[1 + floor(random() * 6)::int],
    NOW(),
    NOW()
FROM temp_users u
CROSS JOIN generate_series(CURRENT_DATE - interval '7 days', CURRENT_DATE + interval '30 days', interval '1 day') d
ON CONFLICT (user_id, date) DO NOTHING;

-- 一時テーブル削除
DROP TABLE temp_users;

-- 統計情報更新
ANALYZE users;
ANALYZE attendances;
ANALYZE overtime_requests;
ANALYZE leave_balances;
ANALYZE leave_requests;
ANALYZE attendance_corrections;
ANALYZE notifications;
ANALYZE projects;
ANALYZE time_entries;
ANALYZE holidays;
ANALYZE approval_flows;
ANALYZE approval_steps;
ANALYZE shifts;

-- データ件数確認用クエリ
SELECT 'users' as table_name, COUNT(*) as count FROM users
UNION ALL SELECT 'attendances', COUNT(*) FROM attendances
UNION ALL SELECT 'overtime_requests', COUNT(*) FROM overtime_requests
UNION ALL SELECT 'leave_balances', COUNT(*) FROM leave_balances
UNION ALL SELECT 'leave_requests', COUNT(*) FROM leave_requests
UNION ALL SELECT 'attendance_corrections', COUNT(*) FROM attendance_corrections
UNION ALL SELECT 'notifications', COUNT(*) FROM notifications
UNION ALL SELECT 'projects', COUNT(*) FROM projects
UNION ALL SELECT 'time_entries', COUNT(*) FROM time_entries
UNION ALL SELECT 'holidays', COUNT(*) FROM holidays
UNION ALL SELECT 'approval_flows', COUNT(*) FROM approval_flows
UNION ALL SELECT 'shifts', COUNT(*) FROM shifts
ORDER BY table_name;
