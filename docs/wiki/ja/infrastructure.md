# インフラ

## ローカル構成（Docker Compose）

ローカル開発はリポジトリ直下の `docker-compose.yml` を中心に構成します。

```text
+----------------------+        +----------------------+
| frontend (nginx)     |        | backend (gin)        |
| localhost:3000       | -----> | localhost:8080       |
+----------------------+        +----------+-----------+
                                           |
                        +------------------+------------------+
                        |                                     |
                        v                                     v
                +---------------+                     +---------------+
                | postgres:16   |                     | redis:7       |
                | localhost:5433|                     | localhost:6379|
                +---------------+                     +---------------+
```

monitoring profileを有効化すると、Elasticsearch/Logstash/Kibana/Prometheus/Grafana/OTel Collectorが追加されます。

## ローカル接続先

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`
- Swagger: `http://localhost:8080/swagger/index.html`
- Postgres(host): `localhost:5433`
- Redis(host): `localhost:6379`
- Kibana: `http://localhost:5601`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3001`

## Compose / Dockerビルド上の注意

- frontendイメージのbuild contextはルート（`.`）です。
- `frontend/Dockerfile` は `frontend/` ソースに加えて `docs/wiki` をコピーします。
- backendにはDB/Redis/OTel接続情報が環境変数で注入されます。

## Makeターゲット（主要）

```sh
make up            # compose起動
make down          # compose停止
make build         # compose build
make logs          # ログ追跡
make backend-test
make frontend-test
```

## Terraform構成（AWS）

dev環境ルート: `infrastructure/environments/dev`

```text
env/dev resources.tf
  -> module.vpc
  -> module.ecr
  -> module.alb
  -> module.ecs
  -> module.rds
  -> module.elasticache
  -> module.cdn
```

### 各モジュールの責務

- `vpc`: VPC、public/private subnet、IGW、NAT、route table
- `alb`: 公開ALB、backend/frontend target group、listener
- `ecs`: cluster、task definition、service、CloudWatch logs
- `rds`: PostgreSQL本体、subnet group、security group
- `elasticache`: Redis配置と接続制御
- `ecr`: backend/frontend のコンテナレジストリ
- `cdn`: CloudFront + S3 連携

## クラウド実行時の経路

```text
利用者
  -> CloudFront
  -> ALB
     -> ECS frontend task
     -> ECS backend task
        -> RDS PostgreSQL
        -> ElastiCache Redis
```

## 環境変数とシークレット

参照元: `.env.example`

特に機密扱いする値:

- `JWT_SECRET_KEY`
- `DATABASE_URL`（またはDB認証情報）
- `AWS_*` 認証情報
- `SENTRY_DSN`

## 監視・ログ・トレース経路

```text
backend metrics/logs/traces
  -> OTel Collector (4317/4318)
     -> Prometheus exporter
     -> Elasticsearch exporter
     -> logging exporter
```

Prometheusはbackendとcollectorのメトリクスをscrapeし、
ログはLogstash経由でElasticsearchに取り込みます。

## ローカル運用ランブック

1. `make up`
2. `docker compose ps` で状態確認
3. `http://localhost:3000` にアクセス
4. 必要に応じてmigration/seedを実行
5. `make logs` でログ確認

## 障害切り分けチェック

- 画面が開かない: frontendコンテナ状態と静的配信確認
- APIタイムアウト: backendログとDB接続確認
- ログイン失敗: JWT設定とrefresh挙動確認
- レスポンス遅延: DB/Redis/重いクエリを優先調査
- メトリクス未取得: Prometheus targetとOTel endpoint確認
