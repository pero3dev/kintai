-- 000002_add_new_features.up.sql
-- 新機能追加マイグレーション

-- ===== 勤怠テーブルにGPS位置情報カラム追加 =====
ALTER TABLE attendances ADD COLUMN IF NOT EXISTS clock_in_latitude DECIMAL(10,8);
ALTER TABLE attendances ADD COLUMN IF NOT EXISTS clock_in_longitude DECIMAL(11,8);
ALTER TABLE attendances ADD COLUMN IF NOT EXISTS clock_out_latitude DECIMAL(10,8);
ALTER TABLE attendances ADD COLUMN IF NOT EXISTS clock_out_longitude DECIMAL(11,8);

-- ===== 残業申請テーブル =====
CREATE TABLE IF NOT EXISTS overtime_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    planned_minutes INT NOT NULL,
    actual_minutes INT,
    reason VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ,
    rejected_reason VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_overtime_requests_user_id ON overtime_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_overtime_requests_status ON overtime_requests(status);
CREATE INDEX IF NOT EXISTS idx_overtime_requests_date ON overtime_requests(date);

-- ===== 有給休暇残日数テーブル =====
CREATE TABLE IF NOT EXISTS leave_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fiscal_year INT NOT NULL,
    leave_type VARCHAR(20) NOT NULL,
    total_days DECIMAL(5,1) NOT NULL DEFAULT 0,
    used_days DECIMAL(5,1) NOT NULL DEFAULT 0,
    carried_over DECIMAL(5,1) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(user_id, fiscal_year, leave_type)
);

CREATE INDEX IF NOT EXISTS idx_leave_balances_user_id ON leave_balances(user_id);
CREATE INDEX IF NOT EXISTS idx_leave_balances_fiscal_year ON leave_balances(fiscal_year);

-- ===== 勤怠修正申請テーブル =====
CREATE TABLE IF NOT EXISTS attendance_corrections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attendance_id UUID REFERENCES attendances(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    original_clock_in TIMESTAMPTZ,
    original_clock_out TIMESTAMPTZ,
    corrected_clock_in TIMESTAMPTZ,
    corrected_clock_out TIMESTAMPTZ,
    reason VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ,
    rejected_reason VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_attendance_corrections_user_id ON attendance_corrections(user_id);
CREATE INDEX IF NOT EXISTS idx_attendance_corrections_status ON attendance_corrections(status);

-- ===== 通知テーブル =====
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(30) NOT NULL,
    title VARCHAR(200) NOT NULL,
    message VARCHAR(1000) NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    link_url VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(user_id, is_read);

-- ===== プロジェクトテーブル =====
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(1000),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    manager_id UUID REFERENCES users(id) ON DELETE SET NULL,
    budget_hours DECIMAL(10,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_projects_code ON projects(code);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

-- ===== 工数記録テーブル =====
CREATE TABLE IF NOT EXISTS time_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    minutes INT NOT NULL,
    description VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_time_entries_user_id ON time_entries(user_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_project_id ON time_entries(project_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_date ON time_entries(date);
CREATE INDEX IF NOT EXISTS idx_time_entries_user_date ON time_entries(user_id, date);

-- ===== 祝日・会社カレンダーテーブル =====
CREATE TABLE IF NOT EXISTS holidays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL,
    name VARCHAR(200) NOT NULL,
    holiday_type VARCHAR(20) NOT NULL,
    is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_holidays_date ON holidays(date);
CREATE INDEX IF NOT EXISTS idx_holidays_type ON holidays(holiday_type);

-- ===== 承認フローテーブル =====
CREATE TABLE IF NOT EXISTS approval_flows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    flow_type VARCHAR(30) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_approval_flows_type ON approval_flows(flow_type);

-- ===== 承認ステップテーブル =====
CREATE TABLE IF NOT EXISTS approval_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flow_id UUID NOT NULL REFERENCES approval_flows(id) ON DELETE CASCADE,
    step_order INT NOT NULL,
    step_type VARCHAR(20) NOT NULL,
    approver_role VARCHAR(20),
    approver_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_approval_steps_flow_id ON approval_steps(flow_id);

-- ===== 日本の祝日マスタデータ（2026年） =====
INSERT INTO holidays (date, name, holiday_type, is_recurring) VALUES
    ('2026-01-01', '元日', 'national', true),
    ('2026-01-12', '成人の日', 'national', false),
    ('2026-02-11', '建国記念の日', 'national', true),
    ('2026-02-23', '天皇誕生日', 'national', true),
    ('2026-03-20', '春分の日', 'national', false),
    ('2026-04-29', '昭和の日', 'national', true),
    ('2026-05-03', '憲法記念日', 'national', true),
    ('2026-05-04', 'みどりの日', 'national', true),
    ('2026-05-05', 'こどもの日', 'national', true),
    ('2026-05-06', '振替休日', 'national', false),
    ('2026-07-20', '海の日', 'national', false),
    ('2026-08-11', '山の日', 'national', true),
    ('2026-09-21', '敬老の日', 'national', false),
    ('2026-09-22', '国民の休日', 'national', false),
    ('2026-09-23', '秋分の日', 'national', false),
    ('2026-10-12', 'スポーツの日', 'national', false),
    ('2026-11-03', '文化の日', 'national', true),
    ('2026-11-23', '勤労感謝の日', 'national', true),
    ('2026-12-29', '年末休暇', 'company', true),
    ('2026-12-30', '年末休暇', 'company', true),
    ('2026-12-31', '年末休暇', 'company', true)
ON CONFLICT DO NOTHING;
