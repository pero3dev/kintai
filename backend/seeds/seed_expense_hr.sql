-- seed_expense_hr.sql
-- 経費精算・人事管理モジュール用大量データ
-- 実行前に seed.sql を実行しておくこと
-- GORM AutoMigrate でテーブルが作成済みであること

-- ============================================================
-- ■ HR部門データ（階層構造）
-- ============================================================
INSERT INTO hr_departments (id, name, code, description, parent_id, manager_id, budget, created_at, updated_at) VALUES
    -- 本社（トップレベル）
    ('e1000001-0000-0000-0000-000000000001'::uuid, '本社', 'HQ', '本社組織', NULL, NULL, 500000000, NOW(), NOW()),
    -- 事業部
    ('e1000002-0000-0000-0000-000000000002'::uuid, '経営企画本部', 'MGMT', '経営企画・戦略立案', 'e1000001-0000-0000-0000-000000000001'::uuid, NULL, 80000000, NOW(), NOW()),
    ('e1000003-0000-0000-0000-000000000003'::uuid, '開発本部', 'DEV', 'プロダクト開発・技術部門', 'e1000001-0000-0000-0000-000000000001'::uuid, NULL, 150000000, NOW(), NOW()),
    ('e1000004-0000-0000-0000-000000000004'::uuid, '営業本部', 'SALES', '営業・顧客対応部門', 'e1000001-0000-0000-0000-000000000001'::uuid, NULL, 100000000, NOW(), NOW()),
    ('e1000005-0000-0000-0000-000000000005'::uuid, '人事総務本部', 'HR', '人事・総務・労務管理', 'e1000001-0000-0000-0000-000000000001'::uuid, NULL, 60000000, NOW(), NOW()),
    ('e1000006-0000-0000-0000-000000000006'::uuid, '財務経理部', 'FIN', '財務・経理・予算管理', 'e1000001-0000-0000-0000-000000000001'::uuid, NULL, 40000000, NOW(), NOW()),
    -- 開発本部の子部門
    ('e1000007-0000-0000-0000-000000000007'::uuid, 'フロントエンド課', 'DEV-FE', 'フロントエンド開発', 'e1000003-0000-0000-0000-000000000003'::uuid, NULL, 50000000, NOW(), NOW()),
    ('e1000008-0000-0000-0000-000000000008'::uuid, 'バックエンド課', 'DEV-BE', 'バックエンド・API開発', 'e1000003-0000-0000-0000-000000000003'::uuid, NULL, 50000000, NOW(), NOW()),
    ('e1000009-0000-0000-0000-000000000009'::uuid, 'インフラ課', 'DEV-INFRA', 'インフラ・DevOps', 'e1000003-0000-0000-0000-000000000003'::uuid, NULL, 40000000, NOW(), NOW()),
    ('e1000010-0000-0000-0000-000000000010'::uuid, 'QA課', 'DEV-QA', '品質保証・テスト', 'e1000003-0000-0000-0000-000000000003'::uuid, NULL, 30000000, NOW(), NOW()),
    -- 営業本部の子部門
    ('e1000011-0000-0000-0000-000000000011'::uuid, '営業第一課', 'SALES-1', '法人営業（大手顧客担当）', 'e1000004-0000-0000-0000-000000000004'::uuid, NULL, 40000000, NOW(), NOW()),
    ('e1000012-0000-0000-0000-000000000012'::uuid, '営業第二課', 'SALES-2', '法人営業（中小顧客担当）', 'e1000004-0000-0000-0000-000000000004'::uuid, NULL, 35000000, NOW(), NOW()),
    ('e1000013-0000-0000-0000-000000000013'::uuid, 'カスタマーサポート課', 'SALES-CS', '顧客サポート・問い合わせ対応', 'e1000004-0000-0000-0000-000000000004'::uuid, NULL, 25000000, NOW(), NOW()),
    -- 人事総務本部の子部門
    ('e1000014-0000-0000-0000-000000000014'::uuid, '採用課', 'HR-REC', '新卒・中途採用', 'e1000005-0000-0000-0000-000000000005'::uuid, NULL, 20000000, NOW(), NOW()),
    ('e1000015-0000-0000-0000-000000000015'::uuid, '労務課', 'HR-LAB', '労務管理・給与計算', 'e1000005-0000-0000-0000-000000000005'::uuid, NULL, 15000000, NOW(), NOW()),
    ('e1000016-0000-0000-0000-000000000016'::uuid, '総務課', 'HR-GA', '総務・施設管理', 'e1000005-0000-0000-0000-000000000005'::uuid, NULL, 20000000, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- ============================================================
-- ■ HR社員データ（50名）
-- ============================================================

-- まず管理職クラスを固定UUIDで作成（後で部下や評価で参照するため）
INSERT INTO hr_employees (id, user_id, employee_code, first_name, last_name, email, phone, position, grade, department_id, manager_id, employment_type, status, hire_date, birth_date, address, base_salary, created_at, updated_at) VALUES
    -- 経営層
    ('e2000001-0000-0000-0000-000000000001'::uuid, 'a0000001-0000-0000-0000-000000000001'::uuid, 'EMP-001', '管理者', 'ユーザー', 'admin@example.com', '03-1234-5678', '代表取締役', 'E1', 'e1000002-0000-0000-0000-000000000002'::uuid, NULL, 'full_time', 'active', '2015-04-01', '1975-06-15', '東京都千代田区丸の内1-1-1', 1200000, NOW(), NOW()),
    -- 開発本部長
    ('e2000002-0000-0000-0000-000000000002'::uuid, 'b0000001-0000-0000-0000-000000000002'::uuid, 'EMP-002', '部門', 'マネージャー', 'manager@example.com', '03-2345-6789', '開発本部長', 'M1', 'e1000003-0000-0000-0000-000000000003'::uuid, 'e2000001-0000-0000-0000-000000000001'::uuid, 'full_time', 'active', '2016-04-01', '1980-03-22', '東京都渋谷区渋谷2-2-2', 900000, NOW(), NOW()),
    -- 営業本部長
    ('e2000003-0000-0000-0000-000000000003'::uuid, 'b0000002-0000-0000-0000-000000000003'::uuid, 'EMP-003', '営業', 'マネージャー', 'manager2@example.com', '03-3456-7890', '営業本部長', 'M1', 'e1000004-0000-0000-0000-000000000004'::uuid, 'e2000001-0000-0000-0000-000000000001'::uuid, 'full_time', 'active', '2016-10-01', '1982-11-08', '東京都港区赤坂3-3-3', 850000, NOW(), NOW()),
    -- 既存社員（seed.sqlのユーザーに対応）
    ('e2000004-0000-0000-0000-000000000004'::uuid, 'c0000001-0000-0000-0000-000000000004'::uuid, 'EMP-004', '山田', '太郎', 'employee1@example.com', '090-1111-1111', 'シニアエンジニア', 'S2', 'e1000007-0000-0000-0000-000000000007'::uuid, 'e2000002-0000-0000-0000-000000000002'::uuid, 'full_time', 'active', '2018-04-01', '1990-07-20', '東京都新宿区新宿4-4-4', 550000, NOW(), NOW()),
    ('e2000005-0000-0000-0000-000000000005'::uuid, 'c0000002-0000-0000-0000-000000000005'::uuid, 'EMP-005', '佐藤', '花子', 'employee2@example.com', '090-2222-2222', 'エンジニア', 'S1', 'e1000008-0000-0000-0000-000000000008'::uuid, 'e2000002-0000-0000-0000-000000000002'::uuid, 'full_time', 'active', '2019-04-01', '1993-01-15', '東京都目黒区目黒5-5-5', 480000, NOW(), NOW()),
    ('e2000006-0000-0000-0000-000000000006'::uuid, 'c0000003-0000-0000-0000-000000000006'::uuid, 'EMP-006', '鈴木', '一郎', 'employee3@example.com', '090-3333-3333', '営業担当', 'J2', 'e1000011-0000-0000-0000-000000000011'::uuid, 'e2000003-0000-0000-0000-000000000003'::uuid, 'full_time', 'active', '2020-04-01', '1995-09-10', '東京都品川区品川6-6-6', 420000, NOW(), NOW()),
    ('e2000007-0000-0000-0000-000000000007'::uuid, 'c0000004-0000-0000-0000-000000000007'::uuid, 'EMP-007', '田中', '美咲', 'employee4@example.com', '090-4444-4444', '営業担当', 'J1', 'e1000012-0000-0000-0000-000000000012'::uuid, 'e2000003-0000-0000-0000-000000000003'::uuid, 'full_time', 'active', '2021-04-01', '1997-04-05', '東京都世田谷区世田谷7-7-7', 380000, NOW(), NOW()),
    ('e2000008-0000-0000-0000-000000000008'::uuid, 'c0000005-0000-0000-0000-000000000008'::uuid, 'EMP-008', '高橋', '健太', 'employee5@example.com', '090-5555-5555', '人事担当', 'J2', 'e1000014-0000-0000-0000-000000000014'::uuid, NULL, 'full_time', 'active', '2020-10-01', '1994-12-25', '東京都杉並区杉並8-8-8', 430000, NOW(), NOW()),
    -- 追加管理職
    ('e2000009-0000-0000-0000-000000000009'::uuid, NULL, 'EMP-009', '伊藤', '裕子', 'ito.yuko@example.com', '03-4567-8901', '人事総務本部長', 'M1', 'e1000005-0000-0000-0000-000000000005'::uuid, 'e2000001-0000-0000-0000-000000000001'::uuid, 'full_time', 'active', '2016-04-01', '1979-08-18', '東京都中央区銀座9-9-9', 880000, NOW(), NOW()),
    ('e2000010-0000-0000-0000-000000000010'::uuid, NULL, 'EMP-010', '渡辺', '誠', 'watanabe.makoto@example.com', '03-5678-9012', '財務経理部長', 'M2', 'e1000006-0000-0000-0000-000000000006'::uuid, 'e2000001-0000-0000-0000-000000000001'::uuid, 'full_time', 'active', '2017-04-01', '1981-05-03', '東京都文京区本郷10-10', 750000, NOW(), NOW()),
    ('e2000011-0000-0000-0000-000000000011'::uuid, NULL, 'EMP-011', '中村', '浩二', 'nakamura.koji@example.com', '03-6789-0123', 'フロントエンド課長', 'M3', 'e1000007-0000-0000-0000-000000000007'::uuid, 'e2000002-0000-0000-0000-000000000002'::uuid, 'full_time', 'active', '2017-10-01', '1985-02-14', '東京都台東区浅草11-11', 650000, NOW(), NOW()),
    ('e2000012-0000-0000-0000-000000000012'::uuid, NULL, 'EMP-012', '小林', '恵美', 'kobayashi.emi@example.com', '03-7890-1234', 'バックエンド課長', 'M3', 'e1000008-0000-0000-0000-000000000008'::uuid, 'e2000002-0000-0000-0000-000000000002'::uuid, 'full_time', 'active', '2018-04-01', '1986-07-30', '東京都豊島区池袋12-12', 640000, NOW(), NOW()),
    ('e2000013-0000-0000-0000-000000000013'::uuid, NULL, 'EMP-013', '加藤', '大輔', 'kato.daisuke@example.com', '03-8901-2345', 'インフラ課長', 'M3', 'e1000009-0000-0000-0000-000000000009'::uuid, 'e2000002-0000-0000-0000-000000000002'::uuid, 'full_time', 'active', '2018-04-01', '1984-10-12', '東京都北区赤羽13-13', 640000, NOW(), NOW()),
    ('e2000014-0000-0000-0000-000000000014'::uuid, NULL, 'EMP-014', '吉田', '千春', 'yoshida.chiharu@example.com', '03-9012-3456', 'QA課長', 'M3', 'e1000010-0000-0000-0000-000000000010'::uuid, 'e2000002-0000-0000-0000-000000000002'::uuid, 'full_time', 'active', '2018-10-01', '1987-01-28', '東京都練馬区光が丘14-14', 620000, NOW(), NOW())
ON CONFLICT (employee_code) DO NOTHING;

-- 部門のマネージャーを設定
UPDATE hr_departments SET manager_id = 'e2000001-0000-0000-0000-000000000001'::uuid WHERE id = 'e1000001-0000-0000-0000-000000000001'::uuid;
UPDATE hr_departments SET manager_id = 'e2000001-0000-0000-0000-000000000001'::uuid WHERE id = 'e1000002-0000-0000-0000-000000000002'::uuid;
UPDATE hr_departments SET manager_id = 'e2000002-0000-0000-0000-000000000002'::uuid WHERE id = 'e1000003-0000-0000-0000-000000000003'::uuid;
UPDATE hr_departments SET manager_id = 'e2000003-0000-0000-0000-000000000003'::uuid WHERE id = 'e1000004-0000-0000-0000-000000000004'::uuid;
UPDATE hr_departments SET manager_id = 'e2000009-0000-0000-0000-000000000009'::uuid WHERE id = 'e1000005-0000-0000-0000-000000000005'::uuid;
UPDATE hr_departments SET manager_id = 'e2000010-0000-0000-0000-000000000010'::uuid WHERE id = 'e1000006-0000-0000-0000-000000000006'::uuid;
UPDATE hr_departments SET manager_id = 'e2000011-0000-0000-0000-000000000011'::uuid WHERE id = 'e1000007-0000-0000-0000-000000000007'::uuid;
UPDATE hr_departments SET manager_id = 'e2000012-0000-0000-0000-000000000012'::uuid WHERE id = 'e1000008-0000-0000-0000-000000000008'::uuid;
UPDATE hr_departments SET manager_id = 'e2000013-0000-0000-0000-000000000013'::uuid WHERE id = 'e1000009-0000-0000-0000-000000000009'::uuid;
UPDATE hr_departments SET manager_id = 'e2000014-0000-0000-0000-000000000014'::uuid WHERE id = 'e1000010-0000-0000-0000-000000000010'::uuid;

-- 一般社員を大量追加（36名追加で合計50名）
INSERT INTO hr_employees (id, employee_code, first_name, last_name, email, phone, position, grade, department_id, manager_id, employment_type, status, hire_date, birth_date, address, base_salary, created_at, updated_at)
SELECT
    gen_random_uuid(),
    'EMP-' || LPAD(n::text, 3, '0'),
    (ARRAY['松本', '井上', '木村', '林', '清水', '斎藤', '山口', '森', '池田', '橋本',
           '阿部', '石川', '前田', '藤田', '小川', '岡田', '長谷川', '村上', '近藤', '坂本',
           '遠藤', '青木', '藤井', '西村', '福田', '太田', '三浦', '藤原', '松田', '岩崎',
           '中島', '原田', '小野', '竹内', '金子', '和田'])[n - 14],
    (ARRAY['翔太', '結衣', '大翔', '陽菜', '悠真', '凜', '蓮', '咲良', '陽斗', '莉子',
           '悠人', '美桜', '湊', '葵', '朝陽', '芽依', '樹', '紬', '律', '詩',
           '颯', '凛', '暖', '杏', '蒼', '花', '新', '心結', '晴', '楓',
           '大和', '彩花', '瑛太', '七海', '隼人', '真央'])[n - 14],
    'emp' || n || '@example.com',
    '090-' || LPAD((1000 + n)::text, 4, '0') || '-' || LPAD((5000 + n * 7)::text, 4, '0'),
    (ARRAY['エンジニア', 'シニアエンジニア', 'デザイナー', 'プロジェクトマネージャー',
           '営業担当', 'カスタマーサポート', '人事担当', '経理担当', 'QAエンジニア', 'テクニカルライター'])[1 + (n % 10)],
    (ARRAY['J1', 'J2', 'S1', 'S2', 'J1', 'J2', 'S1', 'J1', 'J2', 'S1'])[1 + (n % 10)],
    (ARRAY[
        'e1000007-0000-0000-0000-000000000007', 'e1000008-0000-0000-0000-000000000008',
        'e1000009-0000-0000-0000-000000000009', 'e1000010-0000-0000-0000-000000000010',
        'e1000011-0000-0000-0000-000000000011', 'e1000012-0000-0000-0000-000000000012',
        'e1000013-0000-0000-0000-000000000013', 'e1000014-0000-0000-0000-000000000014',
        'e1000015-0000-0000-0000-000000000015', 'e1000016-0000-0000-0000-000000000016'
    ]::uuid[])[1 + (n % 10)],
    (ARRAY[
        'e2000011-0000-0000-0000-000000000011', 'e2000012-0000-0000-0000-000000000012',
        'e2000013-0000-0000-0000-000000000013', 'e2000014-0000-0000-0000-000000000014',
        'e2000003-0000-0000-0000-000000000003', 'e2000003-0000-0000-0000-000000000003',
        'e2000013-0000-0000-0000-000000000013', 'e2000009-0000-0000-0000-000000000009',
        'e2000009-0000-0000-0000-000000000009', 'e2000009-0000-0000-0000-000000000009'
    ]::uuid[])[1 + (n % 10)],
    CASE WHEN n < 45 THEN 'full_time'
         WHEN n < 48 THEN 'contract'
         ELSE 'part_time' END,
    CASE WHEN n = 49 THEN 'on_leave' ELSE 'active' END,
    (DATE '2017-04-01' + (n * 73 % 2000 || ' days')::interval)::date,
    (DATE '1985-01-01' + (n * 137 % 5000 || ' days')::interval)::date,
    '東京都' || (ARRAY['千代田区', '中央区', '港区', '新宿区', '文京区', '台東区', '墨田区', '江東区', '品川区', '目黒区'])[1 + (n % 10)] || '住所' || n,
    300000 + (n * 17 % 30) * 10000,
    NOW(),
    NOW()
FROM generate_series(15, 50) n
ON CONFLICT (employee_code) DO NOTHING;

-- HR社員の一時テーブル
CREATE TEMP TABLE temp_hr_employees AS
SELECT id, employee_code, department_id, manager_id, row_number() OVER () as rn
FROM hr_employees WHERE status = 'active';

-- ============================================================
-- ■ 評価サイクル・評価データ
-- ============================================================
INSERT INTO evaluation_cycles (id, name, start_date, end_date, is_active, created_at, updated_at) VALUES
    ('ec000001-0000-0000-0000-000000000001'::uuid, '2024年度上期評価', '2024-04-01', '2024-09-30', false, NOW(), NOW()),
    ('ec000002-0000-0000-0000-000000000002'::uuid, '2024年度下期評価', '2024-10-01', '2025-03-31', false, NOW(), NOW()),
    ('ec000003-0000-0000-0000-000000000003'::uuid, '2025年度上期評価', '2025-04-01', '2025-09-30', false, NOW(), NOW()),
    ('ec000004-0000-0000-0000-000000000004'::uuid, '2025年度下期評価', '2025-10-01', '2026-03-31', true, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 全社員の評価データ（直近2サイクル分）
INSERT INTO evaluations (id, employee_id, cycle_id, reviewer_id, status, self_score, manager_score, final_score, self_comment, manager_comment, goals, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.id,
    c.id,
    e.manager_id,
    CASE
        WHEN c.id = 'ec000003-0000-0000-0000-000000000003'::uuid THEN 'finalized'
        WHEN c.id = 'ec000004-0000-0000-0000-000000000004'::uuid THEN
            (ARRAY['draft', 'submitted', 'reviewed', 'draft'])[1 + (e.rn % 4)::int]
    END,
    CASE WHEN c.id = 'ec000003-0000-0000-0000-000000000003'::uuid THEN 2.5 + (random() * 2.0) ELSE NULL END,
    CASE WHEN c.id = 'ec000003-0000-0000-0000-000000000003'::uuid THEN 2.5 + (random() * 2.0) ELSE NULL END,
    CASE WHEN c.id = 'ec000003-0000-0000-0000-000000000003'::uuid THEN 2.5 + (random() * 2.0) ELSE NULL END,
    CASE WHEN c.id = 'ec000003-0000-0000-0000-000000000003'::uuid THEN
        (ARRAY['今期はプロジェクトの立ち上げに注力し、チーム内のコミュニケーション改善に努めました。',
               '新機能の設計・実装をリードし、品質向上に貢献できたと考えています。',
               '顧客対応力を向上させ、クレーム件数の削減に成功しました。',
               'チームメンバーの育成に注力し、新人のオンボーディングを円滑に進めました。'])[1 + (e.rn % 4)::int]
    ELSE NULL END,
    CASE WHEN c.id = 'ec000003-0000-0000-0000-000000000003'::uuid THEN
        (ARRAY['期待以上の成果を出しており、今後のリーダーシップ発揮を期待します。',
               '安定した業務遂行力があります。来期はより積極的な提案を期待します。',
               '着実に成長しています。次のステップとして後輩指導にも取り組んでください。',
               '目標達成に向けて努力が見られます。引き続き頑張ってください。'])[1 + (e.rn % 4)::int]
    ELSE NULL END,
    '次期の目標設定中',
    NOW(),
    NOW()
FROM temp_hr_employees e
CROSS JOIN (
    SELECT id FROM evaluation_cycles WHERE id IN (
        'ec000003-0000-0000-0000-000000000003'::uuid,
        'ec000004-0000-0000-0000-000000000004'::uuid
    )
) c
WHERE e.manager_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 目標データ（各社員2-4件）
-- ============================================================
INSERT INTO hr_goals (id, employee_id, title, description, category, status, progress, start_date, due_date, weight, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.id,
    (ARRAY[
        'プロジェクト完遂率の向上',
        'コードレビュー件数を月20件以上達成',
        '資格取得（AWS Solutions Architect）',
        '顧客満足度スコアの改善',
        'チーム内ナレッジ共有の促進',
        '新規提案件数を四半期5件以上',
        'バグ修正対応時間の短縮',
        '業務マニュアルの整備',
        'コスト削減提案の実施',
        '後輩育成プランの策定と実行',
        '英語力の向上（TOEIC 800点）',
        'マネジメントスキルの向上'
    ])[1 + (e.rn * n % 12)::int],
    '具体的なアクションプランを設定し、四半期ごとに進捗を確認する。',
    (ARRAY['performance', 'development', 'behavior', 'performance'])[1 + (n % 4)],
    (ARRAY['in_progress', 'completed', 'not_started', 'in_progress'])[1 + ((e.rn + n) % 4)],
    CASE
        WHEN (ARRAY['in_progress', 'completed', 'not_started', 'in_progress'])[1 + ((e.rn + n) % 4)] = 'completed' THEN 100
        WHEN (ARRAY['in_progress', 'completed', 'not_started', 'in_progress'])[1 + ((e.rn + n) % 4)] = 'not_started' THEN 0
        ELSE 10 + (e.rn * n % 80)::int
    END,
    DATE '2025-10-01',
    DATE '2026-03-31',
    CASE WHEN n = 1 THEN 3 WHEN n = 2 THEN 2 ELSE 1 END,
    NOW(),
    NOW()
FROM temp_hr_employees e
CROSS JOIN generate_series(1, 3) n
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 研修プログラム
-- ============================================================
INSERT INTO training_programs (id, title, description, category, instructor_name, status, start_date, end_date, max_participants, location, is_online, created_at, updated_at) VALUES
    ('a1000001-0000-0000-0000-000000000001'::uuid, 'ビジネスマナー研修', '新入社員向け基本的なビジネスマナーの習得', 'ビジネス基礎', '鈴木講師', 'completed', '2025-04-07', '2025-04-09', 30, '本社 大会議室A', false, NOW(), NOW()),
    ('a1000002-0000-0000-0000-000000000002'::uuid, 'AWS基礎研修', 'AWSの基本サービスの理解と実践', '技術', '山田講師', 'completed', '2025-05-15', '2025-05-17', 20, NULL, true, NOW(), NOW()),
    ('a1000003-0000-0000-0000-000000000003'::uuid, 'リーダーシップ研修', 'チームリーダーに必要なマネジメントスキル', 'マネジメント', '外部講師 田中氏', 'completed', '2025-06-10', '2025-06-11', 15, '本社 研修室B', false, NOW(), NOW()),
    ('a1000004-0000-0000-0000-000000000004'::uuid, 'セキュリティ研修', '情報セキュリティの基礎と実践的対策', '技術', '加藤講師', 'completed', '2025-09-01', '2025-09-02', 50, NULL, true, NOW(), NOW()),
    ('a1000005-0000-0000-0000-000000000005'::uuid, 'React/TypeScript実践研修', 'フロントエンド開発の最新技術習得', '技術', '外部講師 佐藤氏', 'completed', '2025-10-20', '2025-10-24', 15, NULL, true, NOW(), NOW()),
    ('a1000006-0000-0000-0000-000000000006'::uuid, 'プロジェクトマネジメント研修', 'PMP準拠のプロジェクト管理手法', 'マネジメント', '外部講師 伊藤氏', 'in_progress', '2026-01-15', '2026-01-17', 20, '本社 大会議室B', false, NOW(), NOW()),
    ('a1000007-0000-0000-0000-000000000007'::uuid, 'Kubernetes入門', 'コンテナオーケストレーション基礎', '技術', '中村講師', 'scheduled', '2026-03-10', '2026-03-12', 15, NULL, true, NOW(), NOW()),
    ('a1000008-0000-0000-0000-000000000008'::uuid, 'コミュニケーション研修', '効果的なコミュニケーション手法', 'ビジネス基礎', '外部講師 渡辺氏', 'scheduled', '2026-04-15', '2026-04-16', 25, '本社 研修室A', false, NOW(), NOW()),
    ('a1000009-0000-0000-0000-000000000009'::uuid, 'データ分析入門', 'ビジネスデータ分析の基礎', '技術', '森講師', 'scheduled', '2026-05-20', '2026-05-22', 20, NULL, true, NOW(), NOW()),
    ('a1000010-0000-0000-0000-000000000010'::uuid, 'ハラスメント防止研修', '全社員対象のコンプライアンス研修', 'コンプライアンス', '人事部', 'scheduled', '2026-02-25', '2026-02-25', 100, NULL, true, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 研修受講データ
INSERT INTO training_enrollments (id, program_id, employee_id, status, completed_at, score, feedback, created_at, updated_at)
SELECT
    gen_random_uuid(),
    tp.id,
    e.id,
    CASE
        WHEN tp.status = 'completed' THEN 'completed'
        WHEN tp.status = 'in_progress' THEN 'enrolled'
        ELSE 'enrolled'
    END,
    CASE WHEN tp.status = 'completed' THEN tp.end_date::timestamp ELSE NULL END,
    CASE WHEN tp.status = 'completed' THEN 60 + (random() * 40)::int ELSE NULL END,
    CASE WHEN tp.status = 'completed' THEN
        (ARRAY['非常に参考になりました。実務に活かしたいです。',
               '分かりやすい内容でした。もう少し深い内容も知りたかったです。',
               '実践的な演習が多く、理解が深まりました。',
               '基礎から学べて良かったです。次のステップアップ研修も受講したいです。',
               '業務に直結する内容で大変有益でした。'])[1 + (e.rn % 5)::int]
    ELSE NULL END,
    NOW(),
    NOW()
FROM temp_hr_employees e
CROSS JOIN (SELECT id, status, end_date FROM training_programs) tp
WHERE random() < 0.25  -- 各研修に約25%が参加
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 採用ポジション・応募者データ
-- ============================================================
INSERT INTO recruitment_positions (id, title, department_id, description, requirements, status, openings, location, salary_min, salary_max, created_at, updated_at) VALUES
    ('a2000001-0000-0000-0000-000000000001'::uuid, 'シニアフロントエンドエンジニア', 'e1000007-0000-0000-0000-000000000007'::uuid,
     'React/TypeScriptを用いたWebアプリケーション開発をリードするエンジニアを募集します。',
     'React 3年以上、TypeScript経験必須、チームリード経験歓迎', 'open', 2, '東京（リモート可）', 600000, 900000, NOW(), NOW()),
    ('a2000002-0000-0000-0000-000000000002'::uuid, 'バックエンドエンジニア（Go）', 'e1000008-0000-0000-0000-000000000008'::uuid,
     'Go言語を用いたマイクロサービス開発に携わるエンジニアを募集します。',
     'Go 2年以上、REST API設計経験、Docker/Kubernetes経験歓迎', 'open', 3, '東京（リモート可）', 500000, 800000, NOW(), NOW()),
    ('a2000003-0000-0000-0000-000000000003'::uuid, 'SREエンジニア', 'e1000009-0000-0000-0000-000000000009'::uuid,
     'サービスの信頼性向上とインフラ自動化を担当するSREエンジニアを募集します。',
     'Linux運用経験3年以上、AWS/GCP経験、IaC経験（Terraform等）', 'open', 1, '東京（リモート可）', 600000, 1000000, NOW(), NOW()),
    ('a2000004-0000-0000-0000-000000000004'::uuid, '法人営業（IT業界経験者）', 'e1000011-0000-0000-0000-000000000011'::uuid,
     'SaaS/ITサービスの法人営業を担当するメンバーを募集します。',
     '法人営業経験3年以上、IT業界での営業経験歓迎', 'open', 2, '東京', 400000, 700000, NOW(), NOW()),
    ('a2000005-0000-0000-0000-000000000005'::uuid, 'プロダクトマネージャー', 'e1000003-0000-0000-0000-000000000003'::uuid,
     'プロダクト戦略の立案から実行までをリードするPMを募集します。',
     'PM経験3年以上、アジャイル開発経験、データ分析スキル', 'open', 1, '東京', 700000, 1100000, NOW(), NOW()),
    ('a2000006-0000-0000-0000-000000000006'::uuid, '人事企画担当', 'e1000014-0000-0000-0000-000000000014'::uuid,
     '人事制度の企画・運用を担当するメンバーを募集します。',
     '人事経験3年以上、制度設計経験歓迎', 'closed', 1, '東京', 450000, 650000, NOW(), NOW()),
    ('a2000007-0000-0000-0000-000000000007'::uuid, 'QAエンジニア', 'e1000010-0000-0000-0000-000000000010'::uuid,
     'テスト自動化とQAプロセス改善を担当するエンジニアを募集します。',
     'テスト自動化経験2年以上、CI/CD経験', 'on_hold', 1, '東京（リモート可）', 450000, 700000, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- 応募者データ（各ポジションに5-15名）
INSERT INTO applicants (id, position_id, name, email, phone, resume_url, stage, notes, rating, applied_at, created_at, updated_at)
SELECT
    gen_random_uuid(),
    rp.id,
    (ARRAY['山本太郎', '田村花子', '松井一郎', '石田美咲', '木下健太',
           '上田裕子', '川崎誠', '吉川恵美', '中野大輔', '福島千春',
           '高田翔太', '武田結衣', '宮崎大翔', '島田陽菜', '平野悠真'])[n],
    'applicant' || n || '_' || SUBSTRING(rp.id::text, 1, 4) || '@gmail.com',
    '080-' || LPAD((2000 + n * 100)::text, 4, '0') || '-' || LPAD((3000 + n * 50)::text, 4, '0'),
    '/resumes/resume_' || n || '_' || SUBSTRING(rp.id::text, 1, 8) || '.pdf',
    (ARRAY['new', 'screening', 'interview', 'offer', 'hired', 'rejected',
           'new', 'screening', 'interview', 'new', 'screening',
           'new', 'rejected', 'screening', 'interview'])[n],
    CASE WHEN n <= 5 THEN '書類選考通過。技術面接を予定。'
         WHEN n <= 10 THEN '応募書類確認中。'
         ELSE NULL END,
    CASE WHEN random() < 0.6 THEN 1 + floor(random() * 5)::int ELSE NULL END,
    NOW() - (floor(random() * 60) || ' days')::interval,
    NOW(),
    NOW()
FROM (SELECT id FROM recruitment_positions WHERE status = 'open') rp
CROSS JOIN generate_series(1, 15) n
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ HR書類データ
-- ============================================================
INSERT INTO hr_documents (id, employee_id, title, type, file_name, file_path, file_size, mime_type, uploaded_by, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.id,
    doc.title,
    doc.doc_type,
    doc.file_name,
    '/documents/' || e.employee_code || '/' || doc.file_name,
    50000 + floor(random() * 500000)::int,
    doc.mime_type,
    'e2000009-0000-0000-0000-000000000009'::uuid,
    NOW() - (floor(random() * 365) || ' days')::interval,
    NOW()
FROM temp_hr_employees e
CROSS JOIN (VALUES
    ('雇用契約書', 'contract', 'employment_contract.pdf', 'application/pdf'),
    ('身分証明書コピー', 'identification', 'id_copy.pdf', 'application/pdf'),
    ('給与振込先届', 'banking', 'bank_account.pdf', 'application/pdf'),
    ('通勤経路届', 'commute', 'commute_route.pdf', 'application/pdf')
) AS doc(title, doc_type, file_name, mime_type)
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ お知らせデータ
-- ============================================================
INSERT INTO hr_announcements (id, title, content, priority, author_id, is_published, published_at, expires_at, created_at, updated_at) VALUES
    (gen_random_uuid(), '年末年始休暇のお知らせ', '2025年12月29日（月）〜2026年1月3日（土）は年末年始休暇となります。各自、業務の引き継ぎを確実に行ってください。', 'high', 'e2000009-0000-0000-0000-000000000009'::uuid, true, '2025-12-01 09:00:00', '2026-01-05', NOW(), NOW()),
    (gen_random_uuid(), '健康診断実施のお知らせ', '2026年2月16日〜20日に定期健康診断を実施します。対象者は全社員です。予約は社内ポータルから行ってください。', 'high', 'e2000009-0000-0000-0000-000000000009'::uuid, true, '2026-01-15 09:00:00', '2026-02-28', NOW(), NOW()),
    (gen_random_uuid(), '社内研修のご案内', 'ハラスメント防止研修を2026年2月25日にオンラインで実施します。全社員必須参加です。', 'urgent', 'e2000009-0000-0000-0000-000000000009'::uuid, true, '2026-02-01 09:00:00', '2026-02-26', NOW(), NOW()),
    (gen_random_uuid(), 'オフィス移転のお知らせ', '2026年4月より本社オフィスを渋谷に移転します。詳細は追ってご連絡いたします。', 'normal', 'e2000001-0000-0000-0000-000000000001'::uuid, true, '2026-01-20 09:00:00', '2026-04-30', NOW(), NOW()),
    (gen_random_uuid(), '新人事制度導入のお知らせ', '2026年度より新しい人事評価制度を導入します。説明会を3月に開催予定です。', 'high', 'e2000009-0000-0000-0000-000000000009'::uuid, true, '2026-02-05 09:00:00', '2026-04-01', NOW(), NOW()),
    (gen_random_uuid(), '社内サークル活動費補助について', '社内サークルの活動費補助制度を開始します。月額上限5,000円まで補助します。申請方法は総務課まで。', 'low', 'e2000009-0000-0000-0000-000000000009'::uuid, true, '2025-11-01 09:00:00', NULL, NOW(), NOW()),
    (gen_random_uuid(), 'リモートワーク制度の改定', '2026年1月よりリモートワーク制度を改定し、週3日までのリモート勤務が可能になります。', 'normal', 'e2000001-0000-0000-0000-000000000001'::uuid, true, '2025-12-15 09:00:00', NULL, NOW(), NOW()),
    (gen_random_uuid(), '四半期全体会議のご案内', '2026年度Q1全体会議を3月25日14:00から開催します。各部門の報告と来期計画の共有を行います。', 'normal', 'e2000001-0000-0000-0000-000000000001'::uuid, false, NULL, NULL, NOW(), NOW()),
    (gen_random_uuid(), '福利厚生制度の拡充について', '社員の健康増進のため、スポーツジム利用補助制度を新設します。月額3,000円まで補助対象。', 'normal', 'e2000009-0000-0000-0000-000000000009'::uuid, true, '2026-01-10 09:00:00', NULL, NOW(), NOW()),
    (gen_random_uuid(), 'インフルエンザ予防接種補助', '2025年度のインフルエンザ予防接種費用を全額補助します。領収書を総務課まで提出してください。', 'normal', 'e2000009-0000-0000-0000-000000000009'::uuid, true, '2025-10-01 09:00:00', '2026-01-31', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 1on1ミーティングデータ
-- ============================================================
INSERT INTO one_on_one_meetings (id, manager_id, employee_id, scheduled_date, status, frequency, agenda, notes, mood, action_items, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.manager_id,
    e.id,
    (CURRENT_DATE - (n * 14 || ' days')::interval)::timestamp + time '14:00:00',
    CASE WHEN n > 1 THEN 'completed' ELSE 'scheduled' END,
    'biweekly',
    CASE WHEN n <= 2 THEN
        (ARRAY['・前回のアクションアイテムの確認\n・現在の業務進捗\n・困っていること\n・キャリアについて',
               '・プロジェクトの進捗確認\n・チーム内の課題\n・今後のスキルアップ計画',
               '・目標の進捗確認\n・業務負荷の状況\n・来月の計画策定'])[1 + (e.rn % 3)::int]
    ELSE NULL END,
    CASE WHEN n > 1 THEN
        (ARRAY['プロジェクトは順調に進行中。チーム内のコミュニケーションを更に改善する方向で合意。',
               '目標に対して良い進捗。来月の資格試験に向けて勉強時間の確保を支援する。',
               '業務負荷が高めなため、タスクの優先順位を見直し。一部業務の委譲を検討。',
               'キャリアパスについて話し合い。次年度のリーダー候補として育成計画を策定。',
               '良い成果が出ている。チーム横断の取り組みにも参加してもらうことで合意。',
               '体調面の不安あり。在宅勤務の活用を推奨。産業医面談も提案。'])[1 + ((e.rn + n) % 6)::int]
    ELSE NULL END,
    CASE WHEN n > 1 THEN
        (ARRAY['positive', 'positive', 'neutral', 'concerned', 'positive', 'neutral'])[1 + ((e.rn + n) % 6)::int]
    ELSE NULL END,
    CASE WHEN n > 1 THEN
        ('["' || (ARRAY['ドキュメント整備を来週までに完了', '技術勉強会の資料準備', 'チームMTGのアジェンダ改善案を作成',
                         '業務改善提案書を次回までに提出', '資格試験の申し込み'])[1 + ((e.rn + n) % 5)::int] || '"]')::jsonb
    ELSE NULL END,
    NOW(),
    NOW()
FROM temp_hr_employees e
CROSS JOIN generate_series(0, 5) n
WHERE e.manager_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ スキルマップデータ
-- ============================================================
INSERT INTO employee_skills (id, employee_id, skill_name, category, level, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.id,
    skill.name,
    skill.cat,
    1 + floor(random() * 5)::int,
    NOW(),
    NOW()
FROM temp_hr_employees e
CROSS JOIN (VALUES
    ('JavaScript', 'technical'), ('TypeScript', 'technical'), ('React', 'technical'),
    ('Go', 'technical'), ('Python', 'technical'), ('SQL', 'technical'),
    ('AWS', 'technical'), ('Docker', 'technical'), ('Git', 'technical'),
    ('コミュニケーション', 'soft_skill'), ('リーダーシップ', 'soft_skill'),
    ('プレゼンテーション', 'soft_skill'), ('問題解決力', 'soft_skill'),
    ('プロジェクト管理', 'management'), ('アジャイル', 'management')
) AS skill(name, cat)
WHERE random() < 0.3  -- 各社員にランダムに30%のスキルを割り当て
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 給与記録データ
-- ============================================================
INSERT INTO salary_records (id, employee_id, base_salary, allowances, deductions, net_salary, effective_date, reason, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.id,
    e_full.base_salary + (n - 1) * 20000,
    30000 + floor(random() * 20000)::int,
    50000 + floor(random() * 30000)::int,
    e_full.base_salary + (n - 1) * 20000 + 30000 - 50000,
    (DATE '2023-04-01' + ((n - 1) * 365 || ' days')::interval)::date,
    CASE n WHEN 1 THEN '入社時' WHEN 2 THEN '昇給（定期）' WHEN 3 THEN '昇給（評価反映）' END,
    NOW(),
    NOW()
FROM temp_hr_employees e
JOIN hr_employees e_full ON e.id = e_full.id
CROSS JOIN generate_series(1, 3) n
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ オンボーディングテンプレート・データ
-- ============================================================
INSERT INTO onboarding_templates (id, name, description, tasks, created_at, updated_at) VALUES
    ('a3000001-0000-0000-0000-000000000001'::uuid, 'エンジニア向けオンボーディング', 'エンジニア新入社員用の標準オンボーディングテンプレート',
     '[{"name": "PC・アカウント設定", "description": "社用PC・メール・Slackアカウントの設定", "day": 1},
       {"name": "開発環境構築", "description": "IDE、Git、Docker等の開発環境をセットアップ", "day": 1},
       {"name": "セキュリティ研修", "description": "情報セキュリティポリシーの理解", "day": 2},
       {"name": "コードベース説明", "description": "プロジェクトのアーキテクチャ・コードベースの説明", "day": 3},
       {"name": "チームミーティング参加", "description": "各種定例MTGへの参加開始", "day": 5},
       {"name": "最初のタスク着手", "description": "メンターのサポートの下、最初のタスクに着手", "day": 5},
       {"name": "1週間振り返り", "description": "メンターとの1on1で1週間の振り返り", "day": 5},
       {"name": "2週間チェックイン", "description": "上長との面談、業務状況の確認", "day": 10}]'::jsonb, NOW(), NOW()),
    ('a3000002-0000-0000-0000-000000000002'::uuid, '営業向けオンボーディング', '営業職新入社員用の標準オンボーディングテンプレート',
     '[{"name": "PC・アカウント設定", "description": "社用PC・メール・CRMアカウントの設定", "day": 1},
       {"name": "製品知識研修", "description": "自社製品・サービスの理解", "day": 2},
       {"name": "営業プロセス研修", "description": "営業フロー・ツールの使い方", "day": 3},
       {"name": "先輩同行", "description": "先輩社員の顧客訪問に同行", "day": 5},
       {"name": "ロールプレイング", "description": "商談シミュレーション", "day": 7},
       {"name": "顧客引き継ぎ", "description": "担当顧客の引き継ぎ開始", "day": 10}]'::jsonb, NOW(), NOW()),
    ('a3000003-0000-0000-0000-000000000003'::uuid, '全職種共通オンボーディング', '全職種共通の基本オンボーディングテンプレート',
     '[{"name": "入社手続き", "description": "各種書類の提出・手続き", "day": 1},
       {"name": "会社紹介", "description": "会社の沿革・ミッション・バリューの理解", "day": 1},
       {"name": "ツール設定", "description": "メール・チャット・勤怠システムの設定", "day": 1},
       {"name": "社内ルール説明", "description": "就業規則・各種制度の説明", "day": 2},
       {"name": "部門紹介", "description": "配属部門のメンバー紹介・業務概要", "day": 2}]'::jsonb, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- オンボーディングデータ（最近入社した社員）
INSERT INTO onboardings (id, employee_id, template_id, mentor_id, status, start_date, tasks, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.id,
    CASE WHEN e.rn % 3 = 0 THEN 'a3000001-0000-0000-0000-000000000001'::uuid
         WHEN e.rn % 3 = 1 THEN 'a3000002-0000-0000-0000-000000000002'::uuid
         ELSE 'a3000003-0000-0000-0000-000000000003'::uuid END,
    (SELECT id FROM temp_hr_employees WHERE rn != e.rn ORDER BY random() LIMIT 1),
    CASE WHEN e.rn <= 5 THEN 'completed'
         WHEN e.rn <= 8 THEN 'in_progress'
         ELSE 'pending' END,
    (CURRENT_DATE - (e.rn * 30 || ' days')::interval)::date,
    '[{"name": "PC設定完了", "status": "completed"}, {"name": "研修受講", "status": "in_progress"}]'::jsonb,
    NOW(),
    NOW()
FROM temp_hr_employees e
WHERE e.rn <= 12  -- 最近の12名分
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ オフボーディングデータ
-- ============================================================
INSERT INTO offboardings (id, employee_id, reason, status, last_working_date, notes, exit_interview, checklist, created_at, updated_at) VALUES
    (gen_random_uuid(), (SELECT id FROM temp_hr_employees WHERE rn = 40), 'resignation', 'completed', '2025-12-31',
     '一身上の都合により退職。円満退社。',
     '在籍期間中の経験は非常に有意義でした。チームメンバーには感謝しています。今後はスタートアップに挑戦したいと思います。',
     '[{"item": "社員証返却", "completed": true}, {"item": "PC・備品返却", "completed": true}, {"item": "アカウント削除", "completed": true}, {"item": "引き継ぎ完了", "completed": true}]'::jsonb,
     NOW(), NOW()),
    (gen_random_uuid(), (SELECT id FROM temp_hr_employees WHERE rn = 41), 'resignation', 'in_progress', '2026-02-28',
     '転職のため退職予定。引き継ぎ中。',
     NULL,
     '[{"item": "社員証返却", "completed": false}, {"item": "PC・備品返却", "completed": false}, {"item": "アカウント削除", "completed": false}, {"item": "引き継ぎ完了", "completed": false}]'::jsonb,
     NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ サーベイデータ
-- ============================================================
INSERT INTO surveys (id, title, description, type, status, is_anonymous, questions, created_by, published_at, closed_at, created_at, updated_at) VALUES
    ('a4000001-0000-0000-0000-000000000001'::uuid, '2025年度 従業員エンゲージメント調査',
     '社員の満足度とエンゲージメントを測定する年次調査です。率直なご回答をお願いします。',
     'engagement', 'closed', true,
     '[{"id": "q1", "text": "仕事にやりがいを感じていますか？", "type": "scale", "min": 1, "max": 5},
       {"id": "q2", "text": "職場の人間関係に満足していますか？", "type": "scale", "min": 1, "max": 5},
       {"id": "q3", "text": "キャリア成長の機会があると感じますか？", "type": "scale", "min": 1, "max": 5},
       {"id": "q4", "text": "会社のビジョンに共感していますか？", "type": "scale", "min": 1, "max": 5},
       {"id": "q5", "text": "ワークライフバランスは取れていますか？", "type": "scale", "min": 1, "max": 5},
       {"id": "q6", "text": "改善してほしいことがあれば記入してください", "type": "text"}]'::jsonb,
     'e2000009-0000-0000-0000-000000000009'::uuid, '2025-10-01 09:00:00', '2025-10-31 23:59:59', NOW(), NOW()),
    ('a4000002-0000-0000-0000-000000000002'::uuid, 'リモートワーク満足度調査',
     'リモートワーク制度の改善のため、現状の満足度を調査します。',
     'satisfaction', 'closed', true,
     '[{"id": "q1", "text": "リモートワークの頻度に満足していますか？", "type": "scale", "min": 1, "max": 5},
       {"id": "q2", "text": "自宅の作業環境は整っていますか？", "type": "scale", "min": 1, "max": 5},
       {"id": "q3", "text": "リモートワーク時のコミュニケーションに問題はありますか？", "type": "choice", "options": ["問題なし", "やや問題あり", "問題あり"]},
       {"id": "q4", "text": "改善点があれば記入してください", "type": "text"}]'::jsonb,
     'e2000009-0000-0000-0000-000000000009'::uuid, '2025-11-15 09:00:00', '2025-12-15 23:59:59', NOW(), NOW()),
    ('a4000003-0000-0000-0000-000000000003'::uuid, '2026年度 従業員エンゲージメント調査（上期）',
     '上期のエンゲージメント状況を把握するための調査です。',
     'engagement', 'active', true,
     '[{"id": "q1", "text": "現在の業務に対するモチベーションは？", "type": "scale", "min": 1, "max": 5},
       {"id": "q2", "text": "上司のサポートに満足していますか？", "type": "scale", "min": 1, "max": 5},
       {"id": "q3", "text": "チームの雰囲気はどうですか？", "type": "scale", "min": 1, "max": 5},
       {"id": "q4", "text": "今後取り組みたいことは？", "type": "text"}]'::jsonb,
     'e2000009-0000-0000-0000-000000000009'::uuid, '2026-02-01 09:00:00', NULL, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- サーベイ回答データ（完了済みサーベイに対して）
INSERT INTO survey_responses (id, survey_id, employee_id, answers, created_at, updated_at)
SELECT
    gen_random_uuid(),
    s.id,
    CASE WHEN s.is_anonymous THEN NULL ELSE e.id END,
    CASE
        WHEN s.id = 'a4000001-0000-0000-0000-000000000001'::uuid THEN
            ('{"q1": ' || (3 + floor(random() * 3)::int) || ', "q2": ' || (3 + floor(random() * 3)::int) ||
             ', "q3": ' || (2 + floor(random() * 4)::int) || ', "q4": ' || (3 + floor(random() * 3)::int) ||
             ', "q5": ' || (2 + floor(random() * 4)::int) ||
             ', "q6": "' || (ARRAY['特になし', '福利厚生の充実を希望', 'リモートワークの拡充', '研修制度の充実', '評価制度の透明性向上'])[1 + (e.rn % 5)::int] || '"}')::jsonb
        WHEN s.id = 'a4000002-0000-0000-0000-000000000002'::uuid THEN
            ('{"q1": ' || (3 + floor(random() * 3)::int) || ', "q2": ' || (2 + floor(random() * 4)::int) ||
             ', "q3": "' || (ARRAY['問題なし', 'やや問題あり', '問題なし', '問題なし', 'やや問題あり'])[1 + (e.rn % 5)::int] ||
             '", "q4": "' || (ARRAY['特になし', 'モニター補助が欲しい', 'オンラインMTGツールの改善', '', '通信費の補助'])[1 + (e.rn % 5)::int] || '"}')::jsonb
    END,
    NOW(),
    NOW()
FROM temp_hr_employees e
CROSS JOIN (SELECT id, is_anonymous FROM surveys WHERE status = 'closed') s
WHERE random() < 0.8  -- 80%の回答率
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 経費ポリシー
-- ============================================================
INSERT INTO expense_policies (id, category, monthly_limit, per_claim_limit, auto_approve_limit, requires_receipt_above, description, is_active, created_at, updated_at) VALUES
    (gen_random_uuid(), 'transportation', 50000, 30000, 3000, 1000, '交通費：月上限5万円、1回3万円まで。1000円以上は領収書必須。3000円以下は自動承認。', true, NOW(), NOW()),
    (gen_random_uuid(), 'meals', 30000, 5000, 1500, 1000, '飲食費：月上限3万円、1回5000円まで。接待費は別途申請。', true, NOW(), NOW()),
    (gen_random_uuid(), 'accommodation', 100000, 15000, 0, 0, '宿泊費：月上限10万円、1泊1.5万円まで。全件承認必要。', true, NOW(), NOW()),
    (gen_random_uuid(), 'supplies', 20000, 10000, 3000, 500, '消耗品費：月上限2万円。500円以上は領収書必須。', true, NOW(), NOW()),
    (gen_random_uuid(), 'communication', 10000, 5000, 2000, 0, '通信費：月上限1万円。', true, NOW(), NOW()),
    (gen_random_uuid(), 'entertainment', 50000, 30000, 0, 0, '交際費：月上限5万円。全件承認必要。事前申請推奨。', true, NOW(), NOW()),
    (gen_random_uuid(), 'other', 30000, 15000, 0, 1000, 'その他：月上限3万円。用途の詳細記載必須。', true, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 経費承認フロー設定
-- ============================================================
INSERT INTO expense_approval_flows (id, name, min_amount, max_amount, required_steps, is_active, auto_approve_below, created_at, updated_at) VALUES
    (gen_random_uuid(), '少額経費（自動承認）', 0, 3000, 0, true, 3000, NOW(), NOW()),
    (gen_random_uuid(), '通常経費（1段階承認）', 3001, 50000, 1, true, 0, NOW(), NOW()),
    (gen_random_uuid(), '高額経費（2段階承認）', 50001, 200000, 2, true, 0, NOW(), NOW()),
    (gen_random_uuid(), '特別経費（3段階承認）', 200001, 0, 3, true, 0, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 経費予算データ（部門×カテゴリ）
-- ============================================================
INSERT INTO expense_budgets (id, department_id, category, fiscal_year, budget_amount, spent_amount, created_at, updated_at)
SELECT
    gen_random_uuid(),
    d.id,
    cat.category,
    2025,
    cat.budget,
    floor(cat.budget * (0.5 + random() * 0.4))::int,
    NOW(),
    NOW()
FROM departments d
CROSS JOIN (VALUES
    ('transportation'::text, 600000), ('meals', 360000), ('accommodation', 1200000),
    ('supplies', 240000), ('communication', 120000), ('entertainment', 600000), ('other', 360000)
) AS cat(category, budget)
ON CONFLICT DO NOTHING;

-- 2026年度の予算も追加
INSERT INTO expense_budgets (id, department_id, category, fiscal_year, budget_amount, spent_amount, created_at, updated_at)
SELECT
    gen_random_uuid(),
    d.id,
    cat.category,
    2026,
    floor(cat.budget * 1.05)::int,  -- 前年比5%増
    floor(cat.budget * 0.05 * random())::int,  -- まだ序盤
    NOW(),
    NOW()
FROM departments d
CROSS JOIN (VALUES
    ('transportation'::text, 600000), ('meals', 360000), ('accommodation', 1200000),
    ('supplies', 240000), ('communication', 120000), ('entertainment', 600000), ('other', 360000)
) AS cat(category, budget)
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 経費申請データ（過去6ヶ月分、各ユーザー月2-5件）
-- ============================================================

-- 一時テーブル：経費申請用のユーザーIDリスト
CREATE TEMP TABLE temp_expense_users AS
SELECT id FROM users WHERE is_active = true;

-- 経費申請（約200件）
INSERT INTO expenses (id, user_id, title, status, notes, total_amount, approved_by, approved_at, rejected_reason, reimbursed_at, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.id,
    (ARRAY[
        '出張旅費精算（大阪出張）', '通勤定期券（4-6月分）', '書籍購入費', 'クライアント接待費',
        '出張旅費精算（名古屋出張）', 'セミナー参加費', '備品購入（マウス・キーボード）', 'タクシー代',
        '通勤定期券（7-9月分）', '出張旅費精算（福岡出張）', '顧客訪問交通費', 'オンライン会議ツール年額',
        '出張旅費精算（札幌出張）', 'チーム懇親会費', '通勤定期券（10-12月分）', '資格試験受験料',
        '出張旅費精算（仙台出張）', '携帯電話通信費', '名刺印刷費', '外部研修参加費'
    ])[1 + (row_number() OVER() % 20)::int],
    (ARRAY['draft', 'pending', 'approved', 'approved', 'approved', 'reimbursed', 'reimbursed',
           'reimbursed', 'rejected', 'approved', 'pending', 'reimbursed'])[1 + (row_number() OVER() % 12)::int],
    CASE WHEN random() < 0.3 THEN '領収書添付済み' ELSE NULL END,
    0,  -- total_amountは後でアイテムから集計
    CASE WHEN (ARRAY['draft', 'pending', 'approved', 'approved', 'approved', 'reimbursed', 'reimbursed',
                      'reimbursed', 'rejected', 'approved', 'pending', 'reimbursed'])[1 + (row_number() OVER() % 12)::int]
         IN ('approved', 'reimbursed', 'rejected')
    THEN 'a0000001-0000-0000-0000-000000000001'::uuid ELSE NULL END,
    CASE WHEN (ARRAY['draft', 'pending', 'approved', 'approved', 'approved', 'reimbursed', 'reimbursed',
                      'reimbursed', 'rejected', 'approved', 'pending', 'reimbursed'])[1 + (row_number() OVER() % 12)::int]
         IN ('approved', 'reimbursed', 'rejected')
    THEN NOW() - (floor(random() * 30) || ' days')::interval ELSE NULL END,
    CASE WHEN (ARRAY['draft', 'pending', 'approved', 'approved', 'approved', 'reimbursed', 'reimbursed',
                      'reimbursed', 'rejected', 'approved', 'pending', 'reimbursed'])[1 + (row_number() OVER() % 12)::int] = 'rejected'
    THEN '金額上限超過のため差し戻し。再申請してください。' ELSE NULL END,
    CASE WHEN (ARRAY['draft', 'pending', 'approved', 'approved', 'approved', 'reimbursed', 'reimbursed',
                      'reimbursed', 'rejected', 'approved', 'pending', 'reimbursed'])[1 + (row_number() OVER() % 12)::int] = 'reimbursed'
    THEN NOW() - (floor(random() * 15) || ' days')::interval ELSE NULL END,
    NOW() - (floor(random() * 180) || ' days')::interval,
    NOW()
FROM temp_expense_users u
CROSS JOIN generate_series(1, 5) n
WHERE random() < 0.85
ON CONFLICT DO NOTHING;

-- 経費明細データ（各経費に1-4アイテム）
INSERT INTO expense_items (id, expense_id, expense_date, category, description, amount, receipt_url, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.id,
    (e.created_at - (floor(random() * 30) || ' days')::interval)::date,
    (ARRAY['transportation', 'meals', 'accommodation', 'supplies', 'communication', 'entertainment', 'other'])[1 + (item_n % 7)],
    (ARRAY[
        '新幹線（東京→大阪 往復）', 'タクシー（新宿→品川）', '昼食（顧客同行）',
        '宿泊費（ビジネスホテル1泊）', 'A4コピー用紙 5束', '電話会議サービス月額',
        'クライアント接待（夕食）', '技術書「Go言語プログラミング」', 'USBメモリ 64GB',
        'JR定期券（渋谷→新宿 3ヶ月分）', 'コワーキングスペース利用料', '名刺印刷 200枚',
        '高速バス（東京→名古屋）', 'カフェ（打ち合わせ利用）', 'レンタカー代（地方出張）'
    ])[1 + ((item_n + e_rn) % 15)],
    CASE
        WHEN (ARRAY['transportation', 'meals', 'accommodation', 'supplies', 'communication', 'entertainment', 'other'])[1 + (item_n % 7)] = 'transportation'
            THEN 500 + floor(random() * 25000)::int
        WHEN (ARRAY['transportation', 'meals', 'accommodation', 'supplies', 'communication', 'entertainment', 'other'])[1 + (item_n % 7)] = 'meals'
            THEN 500 + floor(random() * 5000)::int
        WHEN (ARRAY['transportation', 'meals', 'accommodation', 'supplies', 'communication', 'entertainment', 'other'])[1 + (item_n % 7)] = 'accommodation'
            THEN 5000 + floor(random() * 12000)::int
        WHEN (ARRAY['transportation', 'meals', 'accommodation', 'supplies', 'communication', 'entertainment', 'other'])[1 + (item_n % 7)] = 'entertainment'
            THEN 3000 + floor(random() * 25000)::int
        ELSE 500 + floor(random() * 8000)::int
    END,
    CASE WHEN random() < 0.7 THEN '/receipts/receipt_' || gen_random_uuid()::text || '.jpg' ELSE NULL END,
    NOW(),
    NOW()
FROM (SELECT id, created_at, row_number() OVER() as e_rn FROM expenses) e
CROSS JOIN generate_series(1, 3) item_n
WHERE item_n <= 1 + floor(random() * 3)::int
ON CONFLICT DO NOTHING;

-- 経費合計金額を更新
UPDATE expenses SET total_amount = sub.total
FROM (
    SELECT expense_id, SUM(amount) as total
    FROM expense_items
    GROUP BY expense_id
) sub
WHERE expenses.id = sub.expense_id;

-- ============================================================
-- ■ 経費コメントデータ
-- ============================================================
INSERT INTO expense_comments (id, expense_id, user_id, content, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.id,
    CASE WHEN random() < 0.5 THEN e.user_id ELSE 'a0000001-0000-0000-0000-000000000001'::uuid END,
    (ARRAY[
        '領収書を添付しました。ご確認をお願いします。',
        '承認しました。経理部で処理いたします。',
        '金額の内訳について確認させてください。',
        '宿泊費は社内規定の上限内であることを確認しました。',
        '交通費の詳細（経路）を追記してください。',
        '差し戻しします。事前申請書の添付をお願いします。',
        '再提出しました。ご確認をお願いします。',
        '本件、緊急で処理をお願いいたします。'
    ])[1 + (row_number() OVER() % 8)::int],
    e.created_at + (floor(random() * 5) || ' days')::interval,
    NOW()
FROM expenses e
WHERE e.status IN ('approved', 'rejected', 'reimbursed', 'pending')
  AND random() < 0.6
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 経費変更履歴データ
-- ============================================================
INSERT INTO expense_histories (id, expense_id, user_id, action, old_value, new_value, changed_by, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.id,
    e.user_id,
    (ARRAY['status_change', 'status_change', 'amount_update', 'submitted'])[1 + (row_number() OVER() % 4)::int],
    (ARRAY['draft', 'pending', NULL, 'draft'])[1 + (row_number() OVER() % 4)::int],
    (ARRAY['pending', 'approved', NULL, 'pending'])[1 + (row_number() OVER() % 4)::int],
    CASE WHEN random() < 0.5 THEN '管理者 ユーザー' ELSE '山田 太郎' END,
    e.created_at + (floor(random() * 3) || ' days')::interval,
    NOW()
FROM expenses e
WHERE e.status != 'draft'
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 経費テンプレートデータ
-- ============================================================
INSERT INTO expense_templates (id, user_id, name, title, category, description, amount, is_recurring, recurring_day, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.id,
    tmpl.name,
    tmpl.title,
    tmpl.category,
    tmpl.description,
    tmpl.amount,
    tmpl.is_recurring,
    tmpl.recurring_day,
    NOW(),
    NOW()
FROM (SELECT id, row_number() OVER() as rn FROM users WHERE is_active = true LIMIT 10) u
CROSS JOIN (VALUES
    ('通勤定期券', '通勤定期券（3ヶ月分）', 'transportation', 'JR定期券', 45000, true, 1),
    ('携帯電話', '携帯電話通信費', 'communication', '業務用携帯電話の月額料金', 3000, true, 25),
    ('書籍購入', '技術書購入', 'supplies', '業務関連書籍の購入', 3000, false, 0)
) AS tmpl(name, title, category, description, amount, is_recurring, recurring_day)
WHERE u.rn <= 5 OR (u.rn > 5 AND random() < 0.5)
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 経費通知データ
-- ============================================================
INSERT INTO expense_notifications (id, user_id, expense_id, type, message, is_read, read_at, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.user_id,
    e.id,
    (ARRAY['submitted', 'approved', 'rejected', 'reimbursed', 'reminder'])[1 + (row_number() OVER() % 5)::int],
    (ARRAY[
        '経費申請「' || e.title || '」が提出されました。',
        '経費申請「' || e.title || '」が承認されました。',
        '経費申請「' || e.title || '」が差し戻されました。',
        '経費申請「' || e.title || '」の精算が完了しました。',
        '経費申請「' || e.title || '」の承認待ちです。'
    ])[1 + (row_number() OVER() % 5)::int],
    random() < 0.7,
    CASE WHEN random() < 0.7 THEN NOW() - (floor(random() * 10) || ' days')::interval ELSE NULL END,
    e.created_at + interval '1 hour',
    NOW()
FROM expenses e
WHERE e.status != 'draft'
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 経費通知設定データ
-- ============================================================
INSERT INTO expense_notification_settings (id, user_id, email_enabled, push_enabled, approval_alerts, reimbursement_alerts, policy_alerts, weekly_digest, created_at, updated_at)
SELECT
    gen_random_uuid(),
    u.id,
    true,
    random() < 0.8,
    true,
    true,
    random() < 0.6,
    random() < 0.3,
    NOW(),
    NOW()
FROM users u
WHERE u.is_active = true
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 経費代理承認データ
-- ============================================================
INSERT INTO expense_delegates (id, user_id, delegate_id, start_date, end_date, is_active, created_at, updated_at) VALUES
    (gen_random_uuid(), 'b0000001-0000-0000-0000-000000000002'::uuid, 'a0000001-0000-0000-0000-000000000001'::uuid, '2026-02-01', '2026-02-15', true, NOW(), NOW()),
    (gen_random_uuid(), 'b0000002-0000-0000-0000-000000000003'::uuid, 'b0000001-0000-0000-0000-000000000002'::uuid, '2026-03-01', '2026-03-10', false, NOW(), NOW()),
    (gen_random_uuid(), 'a0000001-0000-0000-0000-000000000001'::uuid, 'b0000001-0000-0000-0000-000000000002'::uuid, '2025-12-25', '2026-01-05', false, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 経費リマインダーデータ
-- ============================================================
INSERT INTO expense_reminders (id, user_id, expense_id, message, due_date, is_dismissed, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.user_id,
    e.id,
    CASE
        WHEN e.status = 'draft' THEN '下書きの経費申請があります。提出をお忘れなく。'
        WHEN e.status = 'pending' THEN '承認待ちの経費申請があります。ご確認ください。'
        ELSE '経費申請の処理が完了していません。'
    END,
    (CURRENT_DATE + (floor(random() * 14) || ' days')::interval)::date,
    random() < 0.3,
    NOW(),
    NOW()
FROM expenses e
WHERE e.status IN ('draft', 'pending')
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ 経費ポリシー違反データ
-- ============================================================
INSERT INTO expense_policy_violations (id, expense_id, policy_id, user_id, reason, severity, created_at, updated_at)
SELECT
    gen_random_uuid(),
    e.id,
    (SELECT id FROM expense_policies ORDER BY random() LIMIT 1),
    e.user_id,
    (ARRAY[
        '月間上限額を超過しています。',
        '1回あたりの上限額を超過しています。',
        '領収書が添付されていません。',
        '事前申請なしの経費です。'
    ])[1 + (row_number() OVER() % 4)::int],
    CASE WHEN random() < 0.7 THEN 'warning' ELSE 'error' END,
    NOW(),
    NOW()
FROM expenses e
WHERE e.total_amount > 10000
  AND random() < 0.15  -- 高額経費の15%に違反
ON CONFLICT DO NOTHING;

-- ============================================================
-- ■ クリーンアップ
-- ============================================================
DROP TABLE IF EXISTS temp_hr_employees;
DROP TABLE IF EXISTS temp_expense_users;

-- 統計情報更新
ANALYZE hr_departments;
ANALYZE hr_employees;
ANALYZE evaluation_cycles;
ANALYZE evaluations;
ANALYZE hr_goals;
ANALYZE training_programs;
ANALYZE training_enrollments;
ANALYZE recruitment_positions;
ANALYZE applicants;
ANALYZE hr_documents;
ANALYZE hr_announcements;
ANALYZE one_on_one_meetings;
ANALYZE employee_skills;
ANALYZE salary_records;
ANALYZE onboarding_templates;
ANALYZE onboardings;
ANALYZE offboardings;
ANALYZE surveys;
ANALYZE survey_responses;
ANALYZE expenses;
ANALYZE expense_items;
ANALYZE expense_comments;
ANALYZE expense_histories;
ANALYZE expense_templates;
ANALYZE expense_policies;
ANALYZE expense_budgets;
ANALYZE expense_notifications;
ANALYZE expense_reminders;
ANALYZE expense_notification_settings;
ANALYZE expense_approval_flows;
ANALYZE expense_delegates;
ANALYZE expense_policy_violations;

-- ============================================================
-- ■ データ件数確認
-- ============================================================
SELECT '=== HR データ ===' as section, '' as table_name, 0 as count WHERE false
UNION ALL SELECT '', 'hr_departments', COUNT(*) FROM hr_departments
UNION ALL SELECT '', 'hr_employees', COUNT(*) FROM hr_employees
UNION ALL SELECT '', 'evaluation_cycles', COUNT(*) FROM evaluation_cycles
UNION ALL SELECT '', 'evaluations', COUNT(*) FROM evaluations
UNION ALL SELECT '', 'hr_goals', COUNT(*) FROM hr_goals
UNION ALL SELECT '', 'training_programs', COUNT(*) FROM training_programs
UNION ALL SELECT '', 'training_enrollments', COUNT(*) FROM training_enrollments
UNION ALL SELECT '', 'recruitment_positions', COUNT(*) FROM recruitment_positions
UNION ALL SELECT '', 'applicants', COUNT(*) FROM applicants
UNION ALL SELECT '', 'hr_documents', COUNT(*) FROM hr_documents
UNION ALL SELECT '', 'hr_announcements', COUNT(*) FROM hr_announcements
UNION ALL SELECT '', 'one_on_one_meetings', COUNT(*) FROM one_on_one_meetings
UNION ALL SELECT '', 'employee_skills', COUNT(*) FROM employee_skills
UNION ALL SELECT '', 'salary_records', COUNT(*) FROM salary_records
UNION ALL SELECT '', 'onboarding_templates', COUNT(*) FROM onboarding_templates
UNION ALL SELECT '', 'onboardings', COUNT(*) FROM onboardings
UNION ALL SELECT '', 'offboardings', COUNT(*) FROM offboardings
UNION ALL SELECT '', 'surveys', COUNT(*) FROM surveys
UNION ALL SELECT '', 'survey_responses', COUNT(*) FROM survey_responses
UNION ALL SELECT '=== 経費 データ ===' , '', 0 WHERE false
UNION ALL SELECT '', 'expense_policies', COUNT(*) FROM expense_policies
UNION ALL SELECT '', 'expense_approval_flows', COUNT(*) FROM expense_approval_flows
UNION ALL SELECT '', 'expense_budgets', COUNT(*) FROM expense_budgets
UNION ALL SELECT '', 'expenses', COUNT(*) FROM expenses
UNION ALL SELECT '', 'expense_items', COUNT(*) FROM expense_items
UNION ALL SELECT '', 'expense_comments', COUNT(*) FROM expense_comments
UNION ALL SELECT '', 'expense_histories', COUNT(*) FROM expense_histories
UNION ALL SELECT '', 'expense_templates', COUNT(*) FROM expense_templates
UNION ALL SELECT '', 'expense_notifications', COUNT(*) FROM expense_notifications
UNION ALL SELECT '', 'expense_notification_settings', COUNT(*) FROM expense_notification_settings
UNION ALL SELECT '', 'expense_delegates', COUNT(*) FROM expense_delegates
UNION ALL SELECT '', 'expense_reminders', COUNT(*) FROM expense_reminders
UNION ALL SELECT '', 'expense_policy_violations', COUNT(*) FROM expense_policy_violations
ORDER BY table_name;
