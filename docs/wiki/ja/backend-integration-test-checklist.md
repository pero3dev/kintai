# バックエンド結合テスト実装チェックリスト

## 目的
このドキュメントは、バックエンド結合テストで実装するべき項目を網羅的に定義し、実装順序を管理するためのチェックリストです。

## スコープ
- 対象: `backend/cmd/server/main.go` から起動される実運用のAPI経路
- 検証レイヤー: Gin router, middleware, handler, service, repository, PostgreSQL（実DB）
- ルート件数（2026-02-11時点）
- `backend/internal/apps/shared/routes.go`: 46
- `backend/internal/apps/attendance_routes/routes.go`: 22
- `backend/internal/apps/expense_routes/routes.go`: 39
- `backend/internal/apps/hr_routes/routes.go`: 90
- `GET /health`, `GET /metrics`: 2
- 合計: 199 endpoint
- 非対象: frontend E2E、mock中心の unit/branch テストのみの確認

## 現状ギャップ
- 既存のbackendテストは `sqlmock` / `DryRun` / `httptest` が中心で、実DBを使った結合テストが不足しています。
- そのため、SQL実行時の制約、関連取得、ミドルウェア連動の回帰検知が弱い状態です。

## 実装チェックリスト

## 1. テスト基盤（共通）
- [x] PostgreSQL実DBで migration を適用して起動するテスト基盤を構築
- [x] テストケースごとにDB状態を初期化（truncate または schema recreate）
- [x] JWT生成ヘルパーを準備（admin / manager / employee / 期限切れ / 不正署名）
- [x] APIリクエストヘルパーを準備（JSON / multipart / download）
- [x] seed投入ヘルパーを準備（必要に応じて `backend/seeds/*.sql` を利用）

## 2. 全endpoint共通のテスト軸
- [x] endpoint分類表を作成（SoT: `backend/internal/router/router.go` + 各 `*/routes.go`）
- [x] 公開endpoint（5件）で未認証アクセスが成功することを確認
- [x] 保護endpoint（194件）で未認証アクセスが401になることを確認
- [x] ロール制御endpoint（34件）で employee は403、admin/manager は403以外になることを確認
- [x] 入力バリデーション異常で4xx + `model.ErrorResponse`（`code`, `message`）が返ることを確認
- [x] 正常系で期待ステータス/レスポンス形式になることを確認
- [x] 書き込み系でDB変更（insert/update/delete）が反映されることを確認

### 2.1 endpoint分類（2026-02-11時点）
- [x] 公開endpoint（5件）: `GET /health`, `GET /metrics`, `POST /api/v1/auth/login`, `POST /api/v1/auth/register`, `POST /api/v1/auth/refresh`
- [x] 保護endpoint（194件）: `Auth()` 必須（`/api/v1/**` の protected group配下）
- [x] ロール制御endpoint（34件）: `RequireRole(admin, manager)` 適用（shared: 24 + attendance: 10）

### 2.2 共通ケースID（全endpointで使い回す）
- [x] `CA-01 PublicAccess`: 公開endpointにAuthorizationなしでアクセスし、`401/403` にならないことを確認
- [x] `CA-02 ProtectedUnauthorized`: 保護endpointにAuthorizationなしでアクセスし、`401` + `model.ErrorResponse` を確認
- [x] `CA-03 RoleForbidden`: ロール制御endpointに employee JWT でアクセスし、`403` を確認
- [x] `CA-04 RoleAllowed`: ロール制御endpointに admin/manager JWT でアクセスし、`403` にならないことを確認
- [x] `CA-05 ValidationError`: 不正入力（body/query/path）を渡し、`4xx` + `model.ErrorResponse`（`code`, `message`）を確認
- [x] `CA-06 HappyPath`: 正常入力で期待HTTPステータス・必須レスポンス項目・Content-Type を確認
- [x] `CA-07 DBMutation`: POST/PUT/PATCH/DELETE 実行後にDBを直接参照し、対象レコードの増減または更新値を確認

### 2.3 実装ルール
- [x] 共通ケースは `backend/internal/integrationtest` のヘルパー（`env.go`, `http.go`, `jwt.go`）で実装
- [x] `CA-02` / `CA-03` はミドルウェアで判定されるため、最小リクエスト（空body・ダミーpath param）で実施
- [x] `CA-04` は「403にならない」ことを共通軸で確認し、200/201等の正常完了はドメイン別ケースで担保
- [x] `CA-07` は before/after のDB状態比較を必須化（件数 or 対象カラム値）
- [x] 各endpointに対して「適用するケースID」の対応表を作成し、未割当を0件にする

### 2.4 endpoint別ケース対応表（2026-02-11時点）
- [x] 下表の全199 endpointで「適用ケースID」実装を完了する
- [x] No欠番・重複なしを確認する

#### router
| No | Method | Endpoint | 種別 | 適用ケースID |
|---:|:---|:---|:---|:---|
| 1 | GET | `/health` | 公開 | CA-01, CA-06 |
| 2 | GET | `/metrics` | 公開 | CA-01, CA-06 |

#### shared
| No | Method | Endpoint | 種別 | 適用ケースID |
|---:|:---|:---|:---|:---|
| 1 | GET | `/api/v1/approval-flows` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 2 | POST | `/api/v1/approval-flows` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 3 | DELETE | `/api/v1/approval-flows/:id` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 4 | GET | `/api/v1/approval-flows/:id` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 5 | PUT | `/api/v1/approval-flows/:id` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 6 | POST | `/api/v1/auth/login` | 公開 | CA-01, CA-05, CA-06, CA-07 |
| 7 | POST | `/api/v1/auth/logout` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 8 | POST | `/api/v1/auth/refresh` | 公開 | CA-01, CA-05, CA-06, CA-07 |
| 9 | POST | `/api/v1/auth/register` | 公開 | CA-01, CA-05, CA-06, CA-07 |
| 10 | GET | `/api/v1/dashboard/stats` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 11 | GET | `/api/v1/departments` | 保護 | CA-02, CA-05, CA-06 |
| 12 | GET | `/api/v1/export/attendance` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 13 | GET | `/api/v1/export/leaves` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 14 | GET | `/api/v1/export/overtime` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 15 | GET | `/api/v1/export/projects` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 16 | GET | `/api/v1/holidays` | 保護 | CA-02, CA-05, CA-06 |
| 17 | POST | `/api/v1/holidays` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 18 | DELETE | `/api/v1/holidays/:id` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 19 | PUT | `/api/v1/holidays/:id` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 20 | GET | `/api/v1/holidays/calendar` | 保護 | CA-02, CA-05, CA-06 |
| 21 | GET | `/api/v1/holidays/working-days` | 保護 | CA-02, CA-05, CA-06 |
| 22 | GET | `/api/v1/notifications` | 保護 | CA-02, CA-05, CA-06 |
| 23 | DELETE | `/api/v1/notifications/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 24 | PUT | `/api/v1/notifications/:id/read` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 25 | PUT | `/api/v1/notifications/read-all` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 26 | GET | `/api/v1/notifications/unread-count` | 保護 | CA-02, CA-05, CA-06 |
| 27 | GET | `/api/v1/projects` | 保護 | CA-02, CA-05, CA-06 |
| 28 | POST | `/api/v1/projects` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 29 | DELETE | `/api/v1/projects/:id` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 30 | GET | `/api/v1/projects/:id` | 保護 | CA-02, CA-05, CA-06 |
| 31 | PUT | `/api/v1/projects/:id` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 32 | GET | `/api/v1/projects/:id/time-entries` | 保護 | CA-02, CA-05, CA-06 |
| 33 | GET | `/api/v1/shifts` | 保護 | CA-02, CA-05, CA-06 |
| 34 | POST | `/api/v1/shifts` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 35 | DELETE | `/api/v1/shifts/:id` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 36 | POST | `/api/v1/shifts/bulk` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 37 | GET | `/api/v1/time-entries` | 保護 | CA-02, CA-05, CA-06 |
| 38 | POST | `/api/v1/time-entries` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 39 | DELETE | `/api/v1/time-entries/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 40 | PUT | `/api/v1/time-entries/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 41 | GET | `/api/v1/time-entries/summary` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 42 | GET | `/api/v1/users` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 43 | POST | `/api/v1/users` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 44 | DELETE | `/api/v1/users/:id` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 45 | PUT | `/api/v1/users/:id` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 46 | GET | `/api/v1/users/me` | 保護 | CA-02, CA-05, CA-06 |

#### attendance
| No | Method | Endpoint | 種別 | 適用ケースID |
|---:|:---|:---|:---|:---|
| 1 | GET | `/api/v1/attendance` | 保護 | CA-02, CA-05, CA-06 |
| 2 | POST | `/api/v1/attendance/clock-in` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 3 | POST | `/api/v1/attendance/clock-out` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 4 | GET | `/api/v1/attendance/summary` | 保護 | CA-02, CA-05, CA-06 |
| 5 | GET | `/api/v1/attendance/today` | 保護 | CA-02, CA-05, CA-06 |
| 6 | GET | `/api/v1/corrections` | 保護 | CA-02, CA-05, CA-06 |
| 7 | POST | `/api/v1/corrections` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 8 | PUT | `/api/v1/corrections/:id/approve` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 9 | GET | `/api/v1/corrections/pending` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 10 | GET | `/api/v1/leave-balances` | 保護 | CA-02, CA-05, CA-06 |
| 11 | GET | `/api/v1/leave-balances/:user_id` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 12 | PUT | `/api/v1/leave-balances/:user_id/:leave_type` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 13 | POST | `/api/v1/leave-balances/:user_id/initialize` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 14 | GET | `/api/v1/leaves` | 保護 | CA-02, CA-05, CA-06 |
| 15 | POST | `/api/v1/leaves` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 16 | PUT | `/api/v1/leaves/:id/approve` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 17 | GET | `/api/v1/leaves/pending` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 18 | GET | `/api/v1/overtime` | 保護 | CA-02, CA-05, CA-06 |
| 19 | POST | `/api/v1/overtime` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 20 | PUT | `/api/v1/overtime/:id/approve` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06, CA-07 |
| 21 | GET | `/api/v1/overtime/alerts` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |
| 22 | GET | `/api/v1/overtime/pending` | ロール制御 | CA-02, CA-03, CA-04, CA-05, CA-06 |

#### expense
| No | Method | Endpoint | 種別 | 適用ケースID |
|---:|:---|:---|:---|:---|
| 1 | GET | `/api/v1/expenses` | 保護 | CA-02, CA-05, CA-06 |
| 2 | POST | `/api/v1/expenses` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 3 | DELETE | `/api/v1/expenses/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 4 | GET | `/api/v1/expenses/:id` | 保護 | CA-02, CA-05, CA-06 |
| 5 | PUT | `/api/v1/expenses/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 6 | PUT | `/api/v1/expenses/:id/advanced-approve` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 7 | PUT | `/api/v1/expenses/:id/approve` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 8 | GET | `/api/v1/expenses/:id/comments` | 保護 | CA-02, CA-05, CA-06 |
| 9 | POST | `/api/v1/expenses/:id/comments` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 10 | GET | `/api/v1/expenses/:id/history` | 保護 | CA-02, CA-05, CA-06 |
| 11 | GET | `/api/v1/expenses/approval-flow` | 保護 | CA-02, CA-05, CA-06 |
| 12 | GET | `/api/v1/expenses/budgets` | 保護 | CA-02, CA-05, CA-06 |
| 13 | GET | `/api/v1/expenses/delegates` | 保護 | CA-02, CA-05, CA-06 |
| 14 | POST | `/api/v1/expenses/delegates` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 15 | DELETE | `/api/v1/expenses/delegates/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 16 | GET | `/api/v1/expenses/export/csv` | 保護 | CA-02, CA-05, CA-06 |
| 17 | GET | `/api/v1/expenses/export/pdf` | 保護 | CA-02, CA-05, CA-06 |
| 18 | GET | `/api/v1/expenses/notifications` | 保護 | CA-02, CA-05, CA-06 |
| 19 | PUT | `/api/v1/expenses/notifications/:id/read` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 20 | PUT | `/api/v1/expenses/notifications/read-all` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 21 | GET | `/api/v1/expenses/notification-settings` | 保護 | CA-02, CA-05, CA-06 |
| 22 | PUT | `/api/v1/expenses/notification-settings` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 23 | GET | `/api/v1/expenses/pending` | 保護 | CA-02, CA-05, CA-06 |
| 24 | GET | `/api/v1/expenses/policies` | 保護 | CA-02, CA-05, CA-06 |
| 25 | POST | `/api/v1/expenses/policies` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 26 | DELETE | `/api/v1/expenses/policies/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 27 | PUT | `/api/v1/expenses/policies/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 28 | GET | `/api/v1/expenses/policy-violations` | 保護 | CA-02, CA-05, CA-06 |
| 29 | POST | `/api/v1/expenses/receipts/upload` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 30 | GET | `/api/v1/expenses/reminders` | 保護 | CA-02, CA-05, CA-06 |
| 31 | PUT | `/api/v1/expenses/reminders/:id/dismiss` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 32 | GET | `/api/v1/expenses/report` | 保護 | CA-02, CA-05, CA-06 |
| 33 | GET | `/api/v1/expenses/report/monthly` | 保護 | CA-02, CA-05, CA-06 |
| 34 | GET | `/api/v1/expenses/stats` | 保護 | CA-02, CA-05, CA-06 |
| 35 | GET | `/api/v1/expenses/templates` | 保護 | CA-02, CA-05, CA-06 |
| 36 | POST | `/api/v1/expenses/templates` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 37 | DELETE | `/api/v1/expenses/templates/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 38 | PUT | `/api/v1/expenses/templates/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 39 | POST | `/api/v1/expenses/templates/:id/use` | 保護 | CA-02, CA-05, CA-06, CA-07 |

#### hr
| No | Method | Endpoint | 種別 | 適用ケースID |
|---:|:---|:---|:---|:---|
| 1 | GET | `/api/v1/hr/activities` | 保護 | CA-02, CA-05, CA-06 |
| 2 | GET | `/api/v1/hr/announcements` | 保護 | CA-02, CA-05, CA-06 |
| 3 | POST | `/api/v1/hr/announcements` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 4 | DELETE | `/api/v1/hr/announcements/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 5 | GET | `/api/v1/hr/announcements/:id` | 保護 | CA-02, CA-05, CA-06 |
| 6 | PUT | `/api/v1/hr/announcements/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 7 | GET | `/api/v1/hr/applicants` | 保護 | CA-02, CA-05, CA-06 |
| 8 | POST | `/api/v1/hr/applicants` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 9 | PUT | `/api/v1/hr/applicants/:id/stage` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 10 | GET | `/api/v1/hr/attendance-integration` | 保護 | CA-02, CA-05, CA-06 |
| 11 | GET | `/api/v1/hr/attendance-integration/alerts` | 保護 | CA-02, CA-05, CA-06 |
| 12 | GET | `/api/v1/hr/attendance-integration/trend` | 保護 | CA-02, CA-05, CA-06 |
| 13 | GET | `/api/v1/hr/departments` | 保護 | CA-02, CA-05, CA-06 |
| 14 | POST | `/api/v1/hr/departments` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 15 | DELETE | `/api/v1/hr/departments/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 16 | GET | `/api/v1/hr/departments/:id` | 保護 | CA-02, CA-05, CA-06 |
| 17 | PUT | `/api/v1/hr/departments/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 18 | GET | `/api/v1/hr/documents` | 保護 | CA-02, CA-05, CA-06 |
| 19 | POST | `/api/v1/hr/documents` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 20 | DELETE | `/api/v1/hr/documents/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 21 | GET | `/api/v1/hr/documents/:id/download` | 保護 | CA-02, CA-05, CA-06 |
| 22 | GET | `/api/v1/hr/employees` | 保護 | CA-02, CA-05, CA-06 |
| 23 | POST | `/api/v1/hr/employees` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 24 | DELETE | `/api/v1/hr/employees/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 25 | GET | `/api/v1/hr/employees/:id` | 保護 | CA-02, CA-05, CA-06 |
| 26 | PUT | `/api/v1/hr/employees/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 27 | GET | `/api/v1/hr/evaluation-cycles` | 保護 | CA-02, CA-05, CA-06 |
| 28 | POST | `/api/v1/hr/evaluation-cycles` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 29 | GET | `/api/v1/hr/evaluations` | 保護 | CA-02, CA-05, CA-06 |
| 30 | POST | `/api/v1/hr/evaluations` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 31 | GET | `/api/v1/hr/evaluations/:id` | 保護 | CA-02, CA-05, CA-06 |
| 32 | PUT | `/api/v1/hr/evaluations/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 33 | PUT | `/api/v1/hr/evaluations/:id/submit` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 34 | GET | `/api/v1/hr/goals` | 保護 | CA-02, CA-05, CA-06 |
| 35 | POST | `/api/v1/hr/goals` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 36 | DELETE | `/api/v1/hr/goals/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 37 | GET | `/api/v1/hr/goals/:id` | 保護 | CA-02, CA-05, CA-06 |
| 38 | PUT | `/api/v1/hr/goals/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 39 | PUT | `/api/v1/hr/goals/:id/progress` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 40 | GET | `/api/v1/hr/offboarding` | 保護 | CA-02, CA-05, CA-06 |
| 41 | POST | `/api/v1/hr/offboarding` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 42 | GET | `/api/v1/hr/offboarding/:id` | 保護 | CA-02, CA-05, CA-06 |
| 43 | PUT | `/api/v1/hr/offboarding/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 44 | PUT | `/api/v1/hr/offboarding/:id/checklist/:itemKey/toggle` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 45 | GET | `/api/v1/hr/offboarding/analytics` | 保護 | CA-02, CA-05, CA-06 |
| 46 | GET | `/api/v1/hr/onboarding` | 保護 | CA-02, CA-05, CA-06 |
| 47 | POST | `/api/v1/hr/onboarding` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 48 | GET | `/api/v1/hr/onboarding/:id` | 保護 | CA-02, CA-05, CA-06 |
| 49 | PUT | `/api/v1/hr/onboarding/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 50 | PUT | `/api/v1/hr/onboarding/:id/tasks/:taskId/toggle` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 51 | GET | `/api/v1/hr/onboarding/templates` | 保護 | CA-02, CA-05, CA-06 |
| 52 | POST | `/api/v1/hr/onboarding/templates` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 53 | GET | `/api/v1/hr/one-on-ones` | 保護 | CA-02, CA-05, CA-06 |
| 54 | POST | `/api/v1/hr/one-on-ones` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 55 | DELETE | `/api/v1/hr/one-on-ones/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 56 | GET | `/api/v1/hr/one-on-ones/:id` | 保護 | CA-02, CA-05, CA-06 |
| 57 | PUT | `/api/v1/hr/one-on-ones/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 58 | POST | `/api/v1/hr/one-on-ones/:id/actions` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 59 | PUT | `/api/v1/hr/one-on-ones/:id/actions/:actionId/toggle` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 60 | GET | `/api/v1/hr/org-chart` | 保護 | CA-02, CA-05, CA-06 |
| 61 | POST | `/api/v1/hr/org-chart/simulate` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 62 | GET | `/api/v1/hr/positions` | 保護 | CA-02, CA-05, CA-06 |
| 63 | POST | `/api/v1/hr/positions` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 64 | GET | `/api/v1/hr/positions/:id` | 保護 | CA-02, CA-05, CA-06 |
| 65 | PUT | `/api/v1/hr/positions/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 66 | GET | `/api/v1/hr/salary` | 保護 | CA-02, CA-05, CA-06 |
| 67 | GET | `/api/v1/hr/salary/:employeeId/history` | 保護 | CA-02, CA-05, CA-06 |
| 68 | GET | `/api/v1/hr/salary/budget` | 保護 | CA-02, CA-05, CA-06 |
| 69 | POST | `/api/v1/hr/salary/simulate` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 70 | GET | `/api/v1/hr/skill-map` | 保護 | CA-02, CA-05, CA-06 |
| 71 | POST | `/api/v1/hr/skill-map/:employeeId` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 72 | PUT | `/api/v1/hr/skill-map/:employeeId/:skillId` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 73 | GET | `/api/v1/hr/skill-map/gap-analysis` | 保護 | CA-02, CA-05, CA-06 |
| 74 | GET | `/api/v1/hr/stats` | 保護 | CA-02, CA-05, CA-06 |
| 75 | GET | `/api/v1/hr/surveys` | 保護 | CA-02, CA-05, CA-06 |
| 76 | POST | `/api/v1/hr/surveys` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 77 | DELETE | `/api/v1/hr/surveys/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 78 | GET | `/api/v1/hr/surveys/:id` | 保護 | CA-02, CA-05, CA-06 |
| 79 | PUT | `/api/v1/hr/surveys/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 80 | PUT | `/api/v1/hr/surveys/:id/close` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 81 | PUT | `/api/v1/hr/surveys/:id/publish` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 82 | POST | `/api/v1/hr/surveys/:id/respond` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 83 | GET | `/api/v1/hr/surveys/:id/results` | 保護 | CA-02, CA-05, CA-06 |
| 84 | GET | `/api/v1/hr/training` | 保護 | CA-02, CA-05, CA-06 |
| 85 | POST | `/api/v1/hr/training` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 86 | DELETE | `/api/v1/hr/training/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 87 | GET | `/api/v1/hr/training/:id` | 保護 | CA-02, CA-05, CA-06 |
| 88 | PUT | `/api/v1/hr/training/:id` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 89 | PUT | `/api/v1/hr/training/:id/complete` | 保護 | CA-02, CA-05, CA-06, CA-07 |
| 90 | POST | `/api/v1/hr/training/:id/enroll` | 保護 | CA-02, CA-05, CA-06, CA-07 |


## 3. 認証・ミドルウェア結合
- [x] `register -> login -> refresh -> logout` の一連フロー
- [x] 期限切れJWT / 不正署名JWT / Bearer形式不正
- [x] CORS（許可origin / 不許可origin / OPTIONS）
- [x] SecurityHeaders の付与
- [x] RateLimit で429
- [x] Recovery でpanic時に500

## 4. 勤怠ドメイン（attendance）
- [x] 打刻フロー（出勤 -> 退勤 -> 勤怠一覧/詳細）
- [x] 重複打刻/未出勤退勤/重複退勤の異常系
- [x] 勤怠一覧（期間フィルタ・ページング・日付境界）
- [x] 休暇申請（付与残日数、申請、承認/却下）
- [x] 残業申請（申請、承認、アラート閾値）
- [x] 勤怠修正（管理者修正/本人修正申請/却下）
- [x] 月次集計境界（深夜帯・休日・締め日）

## 5. 共通ドメイン（shared）
- [ ] 通知（一覧/未読件数/既読化/全既読/作成）
- [ ] ユーザー（CRUD/ロール更新/`/users/me`）
- [ ] 部署管理（CRUD）
- [ ] シフト（週次作成/一括作成/個別更新/作成）
- [ ] プロジェクト（CRUD/フィルタ）
- [ ] 休日管理（CRUD/重複検知/プロジェクト横断検索）
- [ ] 祝日管理（CRUD/年度検索/カレンダー/CSV入出力）
- [ ] 修正申請フロー（CRUD/ステータス遷移）
- [ ] エクスポート（attendance/leaves/overtime/projects）
- [ ] ダッシュボード集計

## 6. 経費ドメイン（expense）
- [ ] 経費申請作成（項目、税区分、カテゴリ計算）
- [ ] 経費申請更新（ステータス制約）と削除
- [ ] 承認・却下（権限、ステータス遷移、通知）
- [ ] 一覧/検索/レポート/月次集計
- [ ] CSV/PDFエクスポート
- [ ] 領収書アップロード（multipart）
- [ ] コメント追加/取得/削除、履歴取得
- [ ] テンプレートCRUD/テンプレート適用
- [ ] ポリシーCRUD/超過検知アラート
- [ ] 通知設定/リマインダー/通知履歴
- [ ] 修正申請フロー連携（作成、承認、CRUD）

## 7. HRドメイン
- [ ] HRダッシュボード（stats/activities）
- [ ] 社員CRUD + フィルタ検索
- [ ] HR部署CRUD + 階層
- [ ] 評価CRUD + submit + サイクルCRUD
- [ ] 目標CRUD + 進捗更新
- [ ] 研修CRUD + enroll/complete
- [ ] 採用（ポジション/候補者/ステージ更新）
- [ ] 文書（upload/list/delete/download）
- [ ] 入退社CRUD
- [ ] 勤怠連携（integration/alerts/trend）
- [ ] 人員計画（作成、シミュレーション）
- [ ] 1on1（CRUD/アクション登録/トグル）
- [ ] スキル（map/gap/add/update）
- [ ] 給与予算（overview/simulate/history/budget）
- [ ] オンボーディング（CRUD/テンプレートCRUD/タスクトグル）
- [ ] オフボーディング（CRUD/完了処理/チェックリストトグル）
- [ ] サーベイ（CRUD/publish/close/results/respond）

## 8. DB制約・SQL挙動
- [ ] UNIQUE制約検証（email、department.name、project.code など）
- [ ] 複合UNIQUE検証（attendance user+date、shift user+date など）
- [ ] FK制約検証（CASCADE/SET NULL）
- [ ] Preload関連取得の検証
- [ ] 集計SQL検証（dashboard/overtime/expense report/salary budget/turnover）
- [ ] soft delete 後の検索/復元挙動

## 9. 回帰重点
- [ ] 日付境界（UTC/JST、月跨ぎ、締め日）での4xx/正常系
- [ ] 複数ロール混在時のデータ漏洩防止
- [ ] 打刻修正・経費承認の同時更新で整合性維持

## 10. 推奨実装順序（P0 -> P3）
- [ ] P0: テスト基盤 + 認証/ミドルウェア + 全endpoint共通軸
- [ ] P1: attendance + shared（主要業務）
- [ ] P2: expense + HR（中長期機能）
- [ ] P3: DB制約・同時更新・高負荷寄り検証

## 完了条件
- [ ] 全199 endpointが共通軸テストで検証済み
- [ ] 各ドメインの主要シナリオが最低1件以上実装済み
- [ ] 主要永続化処理に実DB結合テストが存在
- [ ] CI実行で再現可能（ローカル/CIとも同等）

## 参照
- `backend/internal/apps/shared/routes.go`
- `backend/internal/apps/attendance_routes/routes.go`
- `backend/internal/apps/expense_routes/routes.go`
- `backend/internal/apps/hr_routes/routes.go`
- `backend/internal/router/router.go`
