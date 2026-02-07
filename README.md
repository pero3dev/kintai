# 勤怠管理システム (Kintai)

従業員の勤怠管理を行うWebアプリケーションです。

## 技術スタック

### バックエンド
- **Go 1.23** + Gin (REST API)
- **GORM** (ORM) + PostgreSQL 16
- **Redis 7** (キャッシュ・セッション)
- **JWT** 認証 (アクセストークン + リフレッシュトークン)
- **zap** (構造化ログ) + **Swagger** (API ドキュメント)
- **golang-migrate** (DBマイグレーション)

### フロントエンド
- **React 19** + TypeScript (Vite)
- **shadcn/ui** + Tailwind CSS
- **TanStack Router** + **TanStack Query**
- **Zustand** (状態管理)
- **React Hook Form** + **Zod** (フォーム/バリデーション)
- **react-i18next** (日本語/英語)

### インフラ
- **Docker Compose** (ローカル開発)
- **AWS** (ECS Fargate, RDS, ElastiCache, ALB, CloudFront, ECR)
- **Terraform** (IaC)
- **GitHub Actions** (CI/CD)

### モニタリング
- **ELK Stack** (ログ収集・可視化)
- **Prometheus** + **Grafana** (メトリクス)
- **OpenTelemetry** (分散トレーシング)
- **Sentry** (エラートラッキング)

## 機能一覧

- 🔐 ユーザー認証（ログイン/登録/JWT）
- ⏰ 出退勤打刻（出勤/退勤/勤務時間自動計算）
- 📊 ダッシュボード（出勤状況/残業/部署別統計）
- 📝 休暇申請/承認ワークフロー
- 📅 シフト管理
- 👥 ユーザー管理（管理者専用）

## クイックスタート

### 前提条件
- Docker & Docker Compose
- Go 1.23+
- Node.js 22+
- Make

### ローカル開発

```bash
# リポジトリをクローン
git clone https://github.com/your-org/kintai.git
cd kintai

# 環境変数を設定
cp .env.example .env

# Docker Composeで起動
make up

# DBマイグレーション実行
make migrate-up

# モニタリングスタックも起動する場合
make monitoring-up
```

### 個別起動

```bash
# バックエンドのみ (ホットリロード)
make backend-dev

# フロントエンドのみ
make frontend-dev
```

### テスト

```bash
# バックエンドテスト
make backend-test

# フロントエンドテスト
make frontend-test

# E2Eテスト
make frontend-e2e

# リント
make backend-lint
make frontend-lint
```

## アクセス先

| サービス | URL |
|----------|-----|
| フロントエンド | http://localhost:3000 |
| バックエンドAPI | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Kibana | http://localhost:5601 |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3001 (admin/admin) |

## プロジェクト構成

```
kintai/
├── backend/                  # Goバックエンド
│   ├── cmd/server/           # エントリーポイント
│   ├── internal/
│   │   ├── config/           # 設定
│   │   ├── handler/          # HTTPハンドラー
│   │   ├── middleware/       # ミドルウェア
│   │   ├── model/            # モデル/DTO
│   │   ├── repository/       # データアクセス層
│   │   ├── router/           # ルーティング
│   │   └── service/          # ビジネスロジック
│   ├── migrations/           # DBマイグレーション
│   └── pkg/logger/           # ログユーティリティ
├── frontend/                 # Reactフロントエンド
│   └── src/
│       ├── api/              # APIクライアント
│       ├── components/       # UIコンポーネント
│       ├── i18n/             # 国際化
│       ├── pages/            # ページコンポーネント
│       └── stores/           # 状態管理
├── infrastructure/           # Terraform
│   ├── environments/dev/     # dev環境
│   └── modules/              # 再利用可能モジュール
│       ├── vpc/              # VPC/サブネット
│       ├── ecr/              # コンテナレジストリ
│       ├── alb/              # ロードバランサー
│       ├── rds/              # PostgreSQL
│       ├── elasticache/      # Redis
│       ├── ecs/              # ECS Fargate
│       └── cdn/              # CloudFront + S3
├── monitoring/               # モニタリング設定
│   ├── logstash/pipeline/    # Logstashパイプライン
│   ├── prometheus/           # Prometheus設定
│   └── otel/                 # OTel Collector設定
├── .github/workflows/        # CI/CD
├── docker-compose.yml
├── Makefile
└── README.md
```

## API エンドポイント

### 認証
- `POST /api/v1/auth/register` - ユーザー登録
- `POST /api/v1/auth/login` - ログイン
- `POST /api/v1/auth/refresh` - トークンリフレッシュ
- `POST /api/v1/auth/logout` - ログアウト

### 勤怠
- `POST /api/v1/attendance/clock-in` - 出勤打刻
- `POST /api/v1/attendance/clock-out` - 退勤打刻
- `GET  /api/v1/attendance` - 勤怠一覧
- `GET  /api/v1/attendance/today` - 本日の勤怠
- `GET  /api/v1/attendance/summary` - 勤怠サマリー

### 休暇
- `POST /api/v1/leaves` - 休暇申請
- `GET  /api/v1/leaves/my` - 自分の休暇一覧
- `GET  /api/v1/leaves/pending` - 承認待ち一覧
- `PUT  /api/v1/leaves/:id/approve` - 承認/却下

### シフト
- `POST /api/v1/shifts` - シフト作成
- `POST /api/v1/shifts/bulk` - 一括作成
- `GET  /api/v1/shifts` - シフト一覧
- `DELETE /api/v1/shifts/:id` - シフト削除

### その他
- `GET  /api/v1/health` - ヘルスチェック
- `GET  /api/v1/users/me` - ログインユーザー情報
- `GET  /api/v1/dashboard/stats` - ダッシュボード統計

## デプロイ

### Terraform (AWS)

```bash
cd infrastructure/environments/dev

# 初期化
terraform init

# プラン確認
terraform plan

# 適用
terraform apply
```

### Docker イメージビルド & プッシュ

```bash
# ECRログイン
aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.ap-northeast-1.amazonaws.com

# ビルド & プッシュ
docker build -t kintai-backend ./backend
docker tag kintai-backend:latest <ecr-url>/kintai-dev-backend:latest
docker push <ecr-url>/kintai-dev-backend:latest
```

## ライセンス

MIT
