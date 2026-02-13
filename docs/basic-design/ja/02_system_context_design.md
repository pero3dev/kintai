# システムコンテキスト設計書

## 1. 文書情報

- 文書ID: BD-02
- 文書名: システムコンテキスト設計書
- 対象システム: 勤怠管理システム（Kintai）
- 版数: 0.1-draft
- 作成日: 2026-02-13
- 関連方針書: `docs/basic-design/ja/01_basic_design_policy.md`

## 2. 目的

本書は、Kintaiのシステム境界、利用者、外部接続先、ドメイン責務境界を定義し、後続の構成設計・機能設計・I/F設計の前提を統一することを目的とする。

## 3. システム境界

### 3.1 対象システム（内側）

- Webフロントエンド（React, TanStack Router, i18n）
- Backend API（Go, Gin, JWT認証）
- 永続化（PostgreSQL）
- 補助ストア（Redis）
- 監視連携（metrics/logs/traces の出力）
- 社内Wiki表示機能（`docs/wiki` のレンダリング）

### 3.2 システム外部（外側）

- 利用者のブラウザ実行環境
- クラウド基盤（AWS: ALB/ECS/RDS/ElastiCache/CloudFront/S3）
- 監視基盤（Prometheus, ELK, OTel Collector の運用系）
- CI/CD実行基盤（GitHub Actions）

## 4. アクター定義

### 4.1 業務アクター

- 社員（employee）
- 管理者（admin）
- マネージャー（manager）
- 人事担当（HR業務運用者）

### 4.2 システムアクター

- フロントエンドクライアント
- Backend API
- PostgreSQL
- Redis
- 監視コンポーネント（Prometheus/OTel/ELK）

## 5. コンテキスト図（論理）

```text
[社員 / 管理者 / マネージャー / 人事担当]
                |
                v
        [Browser / Frontend]
                |
                | HTTPS or HTTP
                v
         [Kintai Backend API]
             |          |
             |          +----> [Redis]
             |
             +---------------> [PostgreSQL]
             |
             +---------------> [Observability Stack]
```

## 6. 業務コンテキスト境界（アプリ境界）

本システムは単一プロダクト内で以下の業務コンテキストをルートprefixで分離する。

- 勤怠コンテキスト: `/`, `/attendance`, `/leaves`, `/overtime`, `/corrections`
- 経費コンテキスト: `/expenses/*`
- 人事コンテキスト: `/hr/*`
- Wikiコンテキスト: `/wiki/*`

補足:

- `/hr/*` はロール制限対象機能を含む（admin/manager中心）。
- `tasks`, `booking` は `frontend/src/config/apps.ts` 上で `enabled: false` のため本設計の対象外とする。

## 7. インターフェース境界

### 7.1 利用者向け公開境界

- フロントエンド入口: `/`（ログイン前後で表示制御）
- API入口: `/api/v1/*`
- ヘルス/運用入口: `/health`, `/live`, `/ready`, `/metrics`

### 7.2 認証・認可境界

- 公開API: 認証不要（例: `POST /api/v1/auth/login`）
- 保護API: JWT必須（`/api/v1` protected group）
- ロール制御API: `RequireRole(...)` によりアクセス制限

## 8. 信頼境界とデータ境界

### 8.1 信頼境界

- 境界A: ブラウザ <-> Backend（不特定入力を受ける境界）
- 境界B: Backend <-> DB/Redis（内部ネットワーク境界）
- 境界C: Backend <-> 監視基盤（運用データ連携境界）

### 8.2 主要データ境界

- 認証情報: JWT（access/refresh）とユーザーロール
- 業務データ: 勤怠、休暇、残業、経費、人事、通知
- 運用データ: ログ、メトリクス、トレース、ヘルス状態

## 9. 配置コンテキスト（環境別）

### 9.1 ローカル開発

- `docker-compose.yml` により frontend/backend/postgres/redis を統合起動
- 監視系は profile により追加起動（elasticsearch/logstash/kibana/prometheus/grafana/otel-collector）

### 9.2 クラウド（dev）

- `infrastructure/environments/dev/resources.tf` でAWSモジュールを構成
- 主経路: CloudFront -> ALB -> ECS(frontend/backend) -> RDS/ElastiCache

## 10. 制約・前提

- backendはモジュラーモノリスで単一デプロイを維持する。
- ルーティング境界は `backend/internal/apps/*` を正とする。
- 画面側コンテキスト境界は `frontend/src/routes.tsx` と `frontend/src/config/apps.ts` を正とする。
- 認証境界は `backend/internal/router/router.go` と middleware を正とする。

## 11. スコープ外（本書では扱わない詳細）

- API項目レベル仕様（別紙: API基本設計書）
- テーブルカラム定義（別紙: DB設計書）
- 画面項目定義とUI部品仕様（別紙: 画面I/O設計書）
- 詳細運用手順（別紙: 運用設計書）

## 12. SoT（参照元）

- `frontend/src/routes.tsx`
- `frontend/src/config/apps.ts`
- `backend/internal/router/router.go`
- `backend/internal/apps/README.md`
- `docs/wiki/ja/overview.md`
- `docs/wiki/ja/architecture.md`
- `infrastructure/environments/dev/resources.tf`
- `docker-compose.yml`

## 13. 次文書

次に作成する基本設計書は BD-03「システム構成設計書」とする。
