# バックエンド

## 起動シーケンス

backendの起点は `backend/cmd/server/main.go` です。

```text
設定読込
  -> logger初期化
  -> PostgreSQL接続（GORM）
  -> 開発時のみAutoMigrateチェック
  -> repository生成
  -> service生成（依存注入）
  -> handler生成
  -> middleware生成
  -> Gin router設定
  -> HTTPサーバ起動
  -> SIGINT/SIGTERMでgraceful shutdown
```

## ディレクトリ責務

- `backend/cmd/server`: エントリーポイント、起動処理
- `backend/internal/router`: ルート統合と委譲
- `backend/internal/middleware`: 認証・制限・ヘッダなど横断処理
- `backend/internal/handler`: HTTP入出力層
- `backend/internal/service`: 業務ロジック層
- `backend/internal/repository`: 永続化・クエリ層
- `backend/internal/model`: モデル・DTO
- `backend/internal/apps/*`: ドメイン分割用パッケージ

## 依存注入の構造

`backend/internal/service/service.go` の `Services` が各ドメインサービスを束ねます。

```text
Repositories --> Services --> Handlers --> Router
```

## ミドルウェア順序

`backend/internal/router/router.go` で以下順に適用されています。

1. `Recovery()`
2. `RequestLogger()`
3. `CORS()`
4. `SecurityHeaders()`
5. `RateLimit()`

保護ルート配下では `Auth()`、管理系では `RequireRole(...)` を追加します。

## 認証フロー

`service.AuthService` の責務:

- `POST /api/v1/auth/login`: 認証、access/refresh token発行
- `POST /api/v1/auth/refresh`: refresh token検証、token再発行
- `POST /api/v1/auth/logout`: refresh token失効

JWTクレーム: `sub`, `email`, `role`, `exp`, `iat`

## 主要ルート群

- shared: auth、notifications、profile、projects、holidays、export
- attendance: 打刻、勤怠一覧、休暇、残業、勤怠修正
- expense: 経費申請、コメント、履歴、テンプレート、ポリシー、通知
- hr: 社員情報、評価、目標、育成、採用、組織、入退社、サーベイ

## 典型的な処理フロー

```text
HTTP JSON request
  -> Handler: bind/validate
  -> Service: use-case実行
  -> Repository: DBアクセス
  -> Service: 結果整形
  -> Handler: HTTP status + payload返却
```

## エラーとHTTPステータス方針

- 入力不正: `400`
- 未認証: `401`
- 権限不足: `403`
- 業務上の未存在/競合: `404/409/422`
- サーバ内部エラー: `500`

共通エラー形式は `model.ErrorResponse` を利用します。

## 設定モデル

`backend/internal/config/config.go` で管理。

- `APP_ENV`, `APP_PORT`
- `DATABASE_URL`, `REDIS_URL`
- `JWT_*`
- `ALLOWED_ORIGINS`
- rate limit設定
- logging/observability設定

本番では開発用JWT秘密鍵のまま起動できないガードがあります。

## データ変更とmigration運用

- ローカル開発時は条件付きAutoMigrateが動作
- 標準運用では `golang-migrate` によるSQL migrationを優先
- 破壊的変更は段階移行（追加 -> 切替 -> 削除）で実施
- ロールバック可能性を事前に設計

## エンドポイント追加プレイブック

1. repositoryメソッドとテストを追加
2. serviceメソッドと業務テストを追加
3. handler実装と入力/応答テストを追加
4. ドメインroute packageに登録
5. frontend API client/UI/testを更新
6. Wikiとルート一覧を更新

## よく使うコマンド

```sh
# backend test
cd backend
go test ./... -v

# lint
cd backend
golangci-lint run ./...

# swagger生成
cd backend
swag init -g cmd/server/main.go -o docs
```

## トラブルシュート観点

- 401多発: JWT秘密鍵不一致、token期限、Authorizationヘッダ確認
- CORS失敗: `ALLOWED_ORIGINS` とfrontend origin整合確認
- 起動失敗: `DATABASE_URL`、DBヘルス、ネットワーク確認
- 応答遅延: repositoryのクエリパスから調査開始
