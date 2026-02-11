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
- [ ] PostgreSQL実DBで migration を適用して起動するテスト基盤を構築
- [ ] テストケースごとにDB状態を初期化（truncate または schema recreate）
- [ ] JWT生成ヘルパーを準備（admin / manager / employee / 期限切れ / 不正署名）
- [ ] APIリクエストヘルパーを準備（JSON / multipart / download）
- [ ] seed投入ヘルパーを準備（必要に応じて `backend/seeds/*.sql` を利用）

## 2. 全endpoint共通のテスト軸
- [ ] endpoint分類表を作成（SoT: `backend/internal/router/router.go` + 各 `*/routes.go`）
- [ ] 公開endpoint（5件）で未認証アクセスが成功することを確認
- [ ] 保護endpoint（194件）で未認証アクセスが401になることを確認
- [ ] ロール制御endpoint（34件）で employee は403、admin/manager は403以外になることを確認
- [ ] 入力バリデーション異常で4xx + `model.ErrorResponse`（`code`, `message`）が返ることを確認
- [ ] 正常系で期待ステータス/レスポンス形式になることを確認
- [ ] 書き込み系でDB変更（insert/update/delete）が反映されることを確認

### 2.1 endpoint分類（2026-02-11時点）
- [ ] 公開endpoint（5件）: `GET /health`, `GET /metrics`, `POST /api/v1/auth/login`, `POST /api/v1/auth/register`, `POST /api/v1/auth/refresh`
- [ ] 保護endpoint（194件）: `Auth()` 必須（`/api/v1/**` の protected group配下）
- [ ] ロール制御endpoint（34件）: `RequireRole(admin, manager)` 適用（shared: 24 + attendance: 10）

### 2.2 共通ケースID（全endpointで使い回す）
- [ ] `CA-01 PublicAccess`: 公開endpointにAuthorizationなしでアクセスし、`401/403` にならないことを確認
- [ ] `CA-02 ProtectedUnauthorized`: 保護endpointにAuthorizationなしでアクセスし、`401` + `model.ErrorResponse` を確認
- [ ] `CA-03 RoleForbidden`: ロール制御endpointに employee JWT でアクセスし、`403` を確認
- [ ] `CA-04 RoleAllowed`: ロール制御endpointに admin/manager JWT でアクセスし、`403` にならないことを確認
- [ ] `CA-05 ValidationError`: 不正入力（body/query/path）を渡し、`4xx` + `model.ErrorResponse`（`code`, `message`）を確認
- [ ] `CA-06 HappyPath`: 正常入力で期待HTTPステータス・必須レスポンス項目・Content-Type を確認
- [ ] `CA-07 DBMutation`: POST/PUT/PATCH/DELETE 実行後にDBを直接参照し、対象レコードの増減または更新値を確認

### 2.3 実装ルール
- [ ] 共通ケースは `backend/internal/integrationtest` のヘルパー（`env.go`, `http.go`, `jwt.go`）で実装
- [ ] `CA-02` / `CA-03` はミドルウェアで判定されるため、最小リクエスト（空body・ダミーpath param）で実施
- [ ] `CA-04` は「403にならない」ことを共通軸で確認し、200/201等の正常完了はドメイン別ケースで担保
- [ ] `CA-07` は before/after のDB状態比較を必須化（件数 or 対象カラム値）
- [ ] 各endpointに対して「適用するケースID」の対応表を作成し、未割当を0件にする

## 3. 認証・ミドルウェア結合
- [ ] `register -> login -> refresh -> logout` の一連フロー
- [ ] 期限切れJWT / 不正署名JWT / Bearer形式不正
- [ ] CORS（許可origin / 不許可origin / OPTIONS）
- [ ] SecurityHeaders の付与
- [ ] RateLimit で429
- [ ] Recovery でpanic時に500

## 4. 勤怠ドメイン（attendance）
- [ ] 打刻フロー（出勤 -> 退勤 -> 勤怠一覧/詳細）
- [ ] 重複打刻/未出勤退勤/重複退勤の異常系
- [ ] 勤怠一覧（期間フィルタ・ページング・日付境界）
- [ ] 休暇申請（付与残日数、申請、承認/却下）
- [ ] 残業申請（申請、承認、アラート閾値）
- [ ] 勤怠修正（管理者修正/本人修正申請/却下）
- [ ] 月次集計境界（深夜帯・休日・締め日）

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
