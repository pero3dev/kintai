# アーキテクチャ

## アーキテクチャ方針

本システムはモジュラーモノリス構成です。

- backendは単一デプロイ単位
- ドメイン別ルートを `backend/internal/apps/*` で明示
- 共通ルートは `backend/internal/apps/shared` に集約
- frontendはルートprefixでアプリ責務を分離

## コンテキスト図

```text
                        +----------------------+
                        |      利用者          |
                        | 社員 / 管理者 / 人事 |
                        +----------+-----------+
                                   |
                                   v
+----------------------------------------------------------------+
|                    Kintai Web Platform                         |
|  +------------------+  +------------------+  +--------------+  |
|  | 勤怠アプリ         |  | 経費アプリ         |  | 人事アプリ     |  |
|  | /attendance系     |  | /expenses/*      |  | /hr/*        |  |
|  +------------------+  +------------------+  +--------------+  |
|                        +------------------+                    |
|                        | 社内Wiki          |                   |
|                        | /wiki/*           |                   |
|                        +------------------+                    |
+----------------------------------------------------------------+
```

## ランタイム構成図

```text
[Browser]
   |
   | HTTP :3000
   v
[Frontend]
   |   UI表示 + /api/v1 呼び出し
   v
[Backend API :8080]
   |
   +--> [PostgreSQL :5432]  業務データ永続化
   +--> [Redis :6379]       キャッシュ/補助用途
   +--> [OTel Collector] -> [Elasticsearch]
   +--> [Prometheus scrape endpoint]
```

## Backendルーティング責務

`backend/internal/router/router.go` でルート登録を統合し、ドメインごとに委譲しています。

- shared: `backend/internal/apps/shared/routes.go`
- attendance: `backend/internal/apps/attendance_routes/routes.go`
- expense: `backend/internal/apps/expense_routes/routes.go`
- hr: `backend/internal/apps/hr_routes/routes.go`

## 保護APIのリクエスト経路

```text
Client request
  -> Gin Engine
  -> Recovery
  -> RequestLogger
  -> CORS
  -> SecurityHeaders
  -> RateLimit
  -> /api/v1 ルートグループ
  -> Auth(JWT)
  -> 必要に応じて RequireRole
  -> Handler
  -> Service
  -> Repository
  -> Database
  -> Response
```

## Frontendの責務分離

- ルート定義: `frontend/src/routes.tsx`
- 共通レイアウト: `frontend/src/components/layout/Layout.tsx`
- アプリ判定: `frontend/src/config/apps.ts` の `getActiveApp()`
- アプリ切替UI: `frontend/src/components/layout/AppSwitcher.tsx`

## レイヤー責務（backend）

```text
[Handler]
  - リクエスト受付、入力整形、レスポンス整形

[Service]
  - 業務ルール、ユースケース実行、分岐制御

[Repository]
  - DBアクセス、クエリ、永続化詳細

[Model/DTO]
  - 永続・転送データ構造
```

## 認証・権限制御

- `/api/v1` の保護ルートはJWT必須
- 管理系ルートは `RequireRole(...)` で制限
- frontendはZustandでaccess/refresh tokenを保持
- 401時はrefresh -> リトライ -> 失敗時ログアウト

## データドメイン（概念）

```text
users
  |- attendances
  |- leave_requests
  |- overtime_requests
  |- attendance_corrections
  |- notifications
  |- expense_* 系
  |- hr_* 系（評価、目標、育成、入退社など）
```

## 設計上の重要ルール

- `internal/router` はオーケストレーション専用に保つ
- ドメイン責務をルート・サービス単位で明示する
- 仕様変更時はWikiも同時更新する
- migrationは後方互換を意識して段階的に行う

## 拡張時の基本手順

1. ドメイン別 route package へエンドポイントを追加
2. handler/service/repository をドメインに沿って追加
3. frontend ルートとナビゲーションを追加
4. テストとWikiを同時更新
