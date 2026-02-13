# 機能一覧・ユースケース設計書

## 1. 文書情報

- 文書ID: BD-04
- 文書名: 機能一覧・ユースケース設計書
- 対象システム: 勤怠管理システム（Kintai）
- 版数: 0.1-draft
- 作成日: 2026-02-13
- 関連文書:
- `docs/basic-design/ja/01_basic_design_policy.md`
- `docs/basic-design/ja/02_system_context_design.md`
- `docs/basic-design/ja/03_system_configuration_design.md`

## 2. 目的

本書は、Kintaiの機能を業務観点で一覧化し、機能ごとの利用者・責務・主要ユースケースを明確化することを目的とする。

## 3. スコープ

- 対象機能: frontend の有効ルートと backend API の実装済み機能
- 非対象: 画面項目詳細、API項目定義、DBカラム定義（別紙で定義）

参考件数（2026-02-13時点）:

- backend routes: 199 endpoint（shared 46 / attendance 22 / expense 39 / hr 90）
- frontend routes: 48 route

## 4. 機能ID規約

- 形式: `FN-{ドメイン}-{連番2桁}`
- ドメインコード:
- `COMMON`: 認証・共通管理
- `ATT`: 勤怠
- `EXP`: 経費
- `HR`: 人事
- `WIKI`: 社内Wiki

## 5. 機能一覧

| 機能ID | 機能名 | 主アクター | 概要 | 主なルート/API |
|---|---|---|---|---|
| FN-COMMON-01 | 認証管理 | 全ユーザー | ログイン、登録、トークン更新、ログアウト | `/login`, `/api/v1/auth/*` |
| FN-COMMON-02 | ユーザー管理 | 管理者/マネージャー | ユーザー一覧、作成、更新、削除、自己情報 | `/users`, `/api/v1/users*` |
| FN-COMMON-03 | 通知管理 | 全ユーザー | 通知取得、既読、全既読、削除 | `/notifications`, `/api/v1/notifications*` |
| FN-COMMON-04 | 部署管理（共通） | 管理者/マネージャー | 部署の参照・管理 | `/api/v1/departments` |
| FN-COMMON-05 | シフト管理 | 管理者/マネージャー/社員 | シフト参照、作成、削除、一括登録 | `/shifts`, `/api/v1/shifts*` |
| FN-COMMON-06 | プロジェクト管理 | 管理者/マネージャー/社員 | プロジェクト参照・管理、工数参照 | `/projects`, `/api/v1/projects*` |
| FN-COMMON-07 | 工数管理 | 全ユーザー | 工数登録、更新、削除、集計 | `/api/v1/time-entries*` |
| FN-COMMON-08 | 休日管理 | 管理者/マネージャー/社員 | 休日参照、カレンダー、営業日計算、管理 | `/holidays`, `/api/v1/holidays*` |
| FN-COMMON-09 | 承認フロー管理 | 管理者/マネージャー | 承認フロー定義のCRUD | `/approval-flows`, `/api/v1/approval-flows*` |
| FN-COMMON-10 | エクスポート | 管理者/マネージャー | 勤怠/休暇/残業/案件の出力 | `/export`, `/api/v1/export/*` |
| FN-COMMON-11 | ダッシュボード集計 | 管理者/マネージャー | 管理指標の取得と表示 | `/dashboard`, `/api/v1/dashboard/stats` |
| FN-ATT-01 | 打刻管理 | 社員 | 出勤/退勤打刻、当日状態、履歴参照 | `/attendance`, `/api/v1/attendance*` |
| FN-ATT-02 | 休暇申請管理 | 社員/管理者/マネージャー | 休暇申請、承認待ち、承認/却下 | `/leaves`, `/api/v1/leaves*` |
| FN-ATT-03 | 残業申請管理 | 社員/管理者/マネージャー | 残業申請、承認待ち、承認、アラート | `/overtime`, `/api/v1/overtime*` |
| FN-ATT-04 | 勤怠修正管理 | 社員/管理者/マネージャー | 修正申請、承認待ち、承認/却下 | `/corrections`, `/api/v1/corrections*` |
| FN-ATT-05 | 休暇残高管理 | 社員/管理者/マネージャー | 自身/対象者の残高参照・設定・初期化 | `/api/v1/leave-balances*` |
| FN-EXP-01 | 経費申請管理 | 社員/承認者 | 申請の作成、更新、削除、詳細表示 | `/expenses`, `/api/v1/expenses*` |
| FN-EXP-02 | 経費承認管理 | 管理者/承認者 | 通常承認・高度承認、承認待ち一覧 | `/expenses/approve`, `/expenses/advanced-approve` |
| FN-EXP-03 | 経費レポート/分析 | 管理者/承認者 | 統計、レポート、月次推移の表示 | `/expenses/report`, `/api/v1/expenses/report*` |
| FN-EXP-04 | 領収書アップロード | 社員 | 領収書ファイルの登録 | `/api/v1/expenses/receipts/upload` |
| FN-EXP-05 | テンプレート管理 | 社員/承認者 | テンプレートのCRUDと適用 | `/expenses/templates`, `/api/v1/expenses/templates*` |
| FN-EXP-06 | ポリシー/予算管理 | 管理者/承認者 | ポリシー管理、予算、違反検知 | `/expenses/policy`, `/api/v1/expenses/policies*` |
| FN-EXP-07 | コメント/履歴管理 | 社員/承認者 | コメント追加・参照、履歴参照 | `/api/v1/expenses/:id/comments`, `/api/v1/expenses/:id/history` |
| FN-EXP-08 | 経費通知/設定管理 | 全ユーザー | 通知既読、リマインダ、通知設定 | `/expenses/notifications`, `/api/v1/expenses/notifications*` |
| FN-EXP-09 | 承認委任管理 | 承認者 | 委任設定、解除、承認フロー参照 | `/api/v1/expenses/delegates*`, `/api/v1/expenses/approval-flow` |
| FN-HR-01 | HRダッシュボード | 管理者/マネージャー | 統計と活動情報の参照 | `/hr`, `/api/v1/hr/stats`, `/api/v1/hr/activities` |
| FN-HR-02 | 社員管理 | 管理者/マネージャー | 社員情報のCRUD | `/hr/employees`, `/api/v1/hr/employees*` |
| FN-HR-03 | HR部署管理 | 管理者/マネージャー | 部署情報のCRUD | `/hr/departments`, `/api/v1/hr/departments*` |
| FN-HR-04 | 評価管理 | 管理者/マネージャー | 評価・評価サイクルの管理、提出 | `/hr/evaluations`, `/api/v1/hr/evaluations*` |
| FN-HR-05 | 目標管理 | 管理者/マネージャー | 目標のCRUDと進捗更新 | `/hr/goals`, `/api/v1/hr/goals*` |
| FN-HR-06 | 研修管理 | 管理者/マネージャー | 研修のCRUD、受講登録、完了 | `/hr/training`, `/api/v1/hr/training*` |
| FN-HR-07 | 採用管理 | 管理者/マネージャー | ポジション・候補者・選考段階管理 | `/hr/recruitment`, `/api/v1/hr/positions*`, `/api/v1/hr/applicants*` |
| FN-HR-08 | 文書管理 | 管理者/マネージャー | 文書アップロード、削除、ダウンロード | `/hr/documents`, `/api/v1/hr/documents*` |
| FN-HR-09 | お知らせ管理 | 管理者/マネージャー | お知らせのCRUD | `/hr/announcements`, `/api/v1/hr/announcements*` |
| FN-HR-10 | 勤怠連携参照 | 管理者/マネージャー | 連携情報、アラート、推移参照 | `/hr/attendance-integration`, `/api/v1/hr/attendance-integration*` |
| FN-HR-11 | 組織図/シミュレーション | 管理者/マネージャー | 組織図参照、編成シミュレーション | `/hr/org-chart`, `/api/v1/hr/org-chart*` |
| FN-HR-12 | 1on1管理 | 管理者/マネージャー | 1on1のCRUD、アクション管理 | `/hr/one-on-one`, `/api/v1/hr/one-on-ones*` |
| FN-HR-13 | スキル管理 | 管理者/マネージャー | スキルマップ、ギャップ分析、スキル更新 | `/hr/skill-map`, `/api/v1/hr/skill-map*` |
| FN-HR-14 | 給与シミュレーション | 管理者/マネージャー | 給与概要、履歴、予算、シミュレーション | `/hr/salary`, `/api/v1/hr/salary*` |
| FN-HR-15 | オンボーディング管理 | 管理者/マネージャー | 入社プロセス、テンプレート、タスク管理 | `/hr/onboarding`, `/api/v1/hr/onboarding*` |
| FN-HR-16 | オフボーディング管理 | 管理者/マネージャー | 退職プロセス、分析、チェックリスト管理 | `/hr/offboarding`, `/api/v1/hr/offboarding*` |
| FN-HR-17 | サーベイ管理 | 管理者/マネージャー/社員 | サーベイCRUD、公開、回答、結果参照 | `/hr/survey`, `/api/v1/hr/surveys*` |
| FN-WIKI-01 | 技術Wiki閲覧 | 全ユーザー | 言語切替付き技術ドキュメント表示 | `/wiki*` |

## 6. 主要ユースケース

| UC-ID | ユースケース名 | 主アクター | 事前条件 | 基本フロー（要約） | 代替/例外 |
|---|---|---|---|---|---|
| UC-01 | ログイン | 全ユーザー | 有効なアカウントを保有 | 認証情報入力 -> token取得 -> ホーム表示 | 認証失敗時はエラー表示 |
| UC-02 | 出勤打刻 | 社員 | ログイン済み、未出勤状態 | 打刻実行 -> 当日勤怠を作成 -> 状態更新 | 二重打刻は業務エラー |
| UC-03 | 退勤打刻 | 社員 | ログイン済み、出勤済み | 打刻実行 -> 勤務時間計算 -> 勤怠確定 | 未出勤時は業務エラー |
| UC-04 | 休暇申請と承認 | 社員/管理者 | ログイン済み | 社員が申請 -> 管理者が承認/却下 -> 結果反映 | 残高不足時は申請不可 |
| UC-05 | 残業申請と承認 | 社員/管理者 | ログイン済み | 社員が申請 -> 管理者が承認 -> 集計反映 | 閾値超過時はアラート表示 |
| UC-06 | 勤怠修正申請と承認 | 社員/管理者 | 対象勤怠が存在 | 修正申請 -> 承認者判定 -> データ更新 | 却下時は理由付きで終了 |
| UC-07 | 経費申請と承認 | 社員/承認者 | ログイン済み | 経費作成 -> 承認者が承認/却下 -> ステータス更新 | ポリシー違反時は警告/却下 |
| UC-08 | 領収書アップロード | 社員 | 経費申請の作成権限あり | ファイル選択 -> アップロード -> 経費に紐付け | 不正形式はアップロード失敗 |
| UC-09 | HR社員情報管理 | 管理者/マネージャー | HR権限あり | 社員一覧参照 -> 登録/更新/削除 | 権限不足時はアクセス拒否 |
| UC-10 | 評価サイクル運用 | 管理者/マネージャー | HR権限あり | サイクル作成 -> 評価作成 -> 提出 -> 参照 | 不正遷移は更新拒否 |
| UC-11 | オンボーディング運用 | 管理者/マネージャー | HR権限あり | 対象者登録 -> タスク管理 -> 進捗更新 | 対象未存在時はエラー |
| UC-12 | 技術Wiki参照 | 全ユーザー | なし | `/wiki`アクセス -> Markdown表示 -> 言語切替 | 対象文書なしは代替ページ表示 |

## 7. 役割別アクセス方針（機能レベル）

| 機能群 | employee | manager | admin |
|---|---|---|---|
| 認証/自己情報 | 可 | 可 | 可 |
| 勤怠（本人操作） | 可 | 可 | 可 |
| 承認系（休暇/残業/修正） | 不可 | 可 | 可 |
| 共通マスタ管理（ユーザー/プロジェクト/休日等） | 一部参照 | 可 | 可 |
| 経費申請 | 可 | 可 | 可 |
| 経費承認・ポリシー管理 | 一部 | 可 | 可 |
| HR機能 | 不可 | 可 | 可 |
| Wiki閲覧 | 可 | 可 | 可 |

補足:

- 詳細なAPI単位の認可は `RequireRole(...)` と各handlerの実装を正とする。
- 画面側は `frontend/src/config/apps.ts` の `requiredRoles` を正とする。

## 8. 受入観点（基本設計レベル）

- 各機能IDに対応する画面導線またはAPIが存在すること。
- 主要ユースケースで正常系と主要異常系が定義されていること。
- 機能群ごとの主アクターと認可方針が矛盾しないこと。

## 9. SoT（参照元）

- `frontend/src/routes.tsx`
- `frontend/src/config/apps.ts`
- `backend/internal/router/router.go`
- `backend/internal/apps/shared/routes.go`
- `backend/internal/apps/attendance_routes/routes.go`
- `backend/internal/apps/expense_routes/routes.go`
- `backend/internal/apps/hr_routes/routes.go`
- `docs/wiki/ja/overview.md`

## 10. 次文書

次に作成する基本設計書は BD-05「業務フロー設計書」とする。
