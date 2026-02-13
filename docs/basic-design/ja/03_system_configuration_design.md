# システム構成設計書

## 1. 文書情報

- 文書ID: BD-03
- 文書名: システム構成設計書
- 対象システム: 勤怠管理システム（Kintai）
- 版数: 0.1-draft
- 作成日: 2026-02-13
- 関連文書:
- `docs/basic-design/ja/01_basic_design_policy.md`
- `docs/basic-design/ja/02_system_context_design.md`

## 2. 目的

本書は、Kintaiのシステム構成（論理・物理・実行時）を定義し、開発環境とクラウド環境で一貫した構成判断を行うための基準を示す。

## 3. 構成方針

- アプリケーションは frontend / backend を分離し、HTTPで連携する。
- backendはドメイン分割されたモジュラーモノリスとして単一デプロイする。
- データ永続化はPostgreSQL、補助ストアはRedisを使用する。
- 監視は metrics / logs / traces を分離収集する。
- IaCはTerraformを正とし、環境差分を `infrastructure/environments/*` に閉じ込める。

## 4. 論理構成

```text
[Browser]
   |
   v
[Frontend UI Layer]
   |
   v
[Backend API Layer]
   |            |
   v            v
[PostgreSQL]  [Redis]
   |
   +-------> [Monitoring/Observability]
```

## 5. 実行時構成（ローカル）

### 5.1 標準構成

- frontend: `localhost:3000`（nginx配信）
- backend: `localhost:8080`（Gin API）
- postgres: host `localhost:5433` -> container `5432`
- redis: `localhost:6379`

### 5.2 監視プロファイル追加構成

- elasticsearch: `9200`
- logstash: `5044`, `9600`
- kibana: `5601`
- prometheus: `9090`
- grafana: `3001`
- otel-collector: `4317`, `4318`, `8888`

### 5.3 ローカル構成図

```text
Browser
  -> Frontend container (nginx:80, host:3000)
     -> /api/* proxy -> Backend container (:8080, host:8080)
        -> Postgres container (:5432, host:5433)
        -> Redis container (:6379, host:6379)
        -> OTel Collector (monitoring profile時)
```

## 6. 実行時構成（AWS dev）

### 6.1 構成要素

- Network: VPC, public/private subnet, IGW, NAT
- Edge: CloudFront
- Load Balancing: ALB
- Compute: ECS Fargate（frontend task / backend task）
- Data: RDS PostgreSQL, ElastiCache Redis
- Registry: ECR（frontend/backend）
- Static Assets: S3（CloudFront OAC 経由）

### 6.2 クラウド構成図

```text
User
  -> CloudFront
     -> ALB
        -> ECS Frontend (port 80)
        -> ECS Backend (port 8080)
           -> RDS PostgreSQL (5432)
           -> ElastiCache Redis (6379)
```

## 7. コンポーネント責務

- frontend
- 画面描画、ルーティング、i18n、API呼び出し
- nginxでSPA配信し `/api/` をbackendへリバースプロキシ
- backend
- 認証/認可、業務ロジック、DBアクセス、メトリクス提供
- postgres
- 業務データの永続化（正本データ）
- redis
- キャッシュ/補助用途の高速アクセス
- otel-collector
- backendからのテレメトリ集約・転送
- prometheus/grafana
- メトリクス収集・可視化
- elk（elasticsearch/logstash/kibana）
- ログ収集・検索・可視化

## 8. 接続仕様

| From | To | Protocol | Port | 用途 |
|---|---|---|---|---|
| Browser | Frontend | HTTP/HTTPS | 3000(ローカル) | 画面表示 |
| Frontend | Backend | HTTP | 8080 | `/api/*` 呼び出し |
| Backend | PostgreSQL | TCP | 5432 | 永続化 |
| Backend | Redis | TCP | 6379 | キャッシュ等 |
| Backend | OTel Collector | gRPC/HTTP | 4317/4318 | テレメトリ送信 |
| Prometheus | Backend/Collector | HTTP | 8080/8888 | scrape |

## 9. デプロイ単位

- backend: 単一サービスとしてデプロイ（ドメインはアプリ境界で内部分離）
- frontend: 単一SPAとしてデプロイ
- DB/Redis: マネージドサービスまたはコンテナで独立運用

## 10. 可用性・拡張性方針

- frontend/backendはコンテナスケールを前提とする。
- statefulデータはRDS/ElastiCacheに分離し、アプリ層はステートレス運用を基本とする。
- 障害時の切り分けを容易にするため、ログ/メトリクス/トレースをコンポーネント単位で収集する。

## 11. 構成上の制約・留意点

- backend実装のヘルスエンドポイントは `/health` だが、`infrastructure/modules/alb/main.tf` のbackend target group health checkは `/api/v1/health` 指定となっているため統一が必要。
- `infrastructure/modules/alb/main.tf` ではHTTPS listenerがコメントアウトされており、TLS終端の最終設計を別途確定する必要がある。
- frontend配信は環境により Vite dev server（開発）と nginx（コンテナ）で経路が異なるため、CORS/Proxy設定を環境別に整合させる。

## 12. SoT（参照元）

- `docker-compose.yml`
- `frontend/nginx.conf`
- `frontend/vite.config.ts`
- `backend/cmd/server/main.go`
- `backend/internal/config/config.go`
- `backend/internal/router/router.go`
- `infrastructure/environments/dev/resources.tf`
- `infrastructure/modules/vpc/main.tf`
- `infrastructure/modules/alb/main.tf`
- `infrastructure/modules/ecs/main.tf`
- `infrastructure/modules/rds/main.tf`
- `infrastructure/modules/elasticache/main.tf`
- `infrastructure/modules/cdn/main.tf`

## 13. 次文書

次に作成する基本設計書は BD-04「機能一覧・ユースケース設計書」とする。
