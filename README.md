# 蜍､諤邂｡逅・す繧ｹ繝・Β (Kintai)

蠕捺･ｭ蜩｡縺ｮ蜍､諤邂｡逅・ｒ陦後≧Web繧｢繝励Μ繧ｱ繝ｼ繧ｷ繝ｧ繝ｳ縺ｧ縺吶・

## 謚陦薙せ繧ｿ繝・け

### 繝舌ャ繧ｯ繧ｨ繝ｳ繝・
- **Go 1.23** + Gin (REST API)
- **GORM** (ORM) + PostgreSQL 16
- **Redis 7** (繧ｭ繝｣繝・す繝･繝ｻ繧ｻ繝・す繝ｧ繝ｳ)
- **JWT** 隱崎ｨｼ (繧｢繧ｯ繧ｻ繧ｹ繝医・繧ｯ繝ｳ + 繝ｪ繝輔Ξ繝・す繝･繝医・繧ｯ繝ｳ)
- **zap** (讒矩蛹悶Ο繧ｰ) + **Swagger** (API 繝峨く繝･繝｡繝ｳ繝・
- **golang-migrate** (DB繝槭う繧ｰ繝ｬ繝ｼ繧ｷ繝ｧ繝ｳ)

### 繝輔Ο繝ｳ繝医お繝ｳ繝・
- **React 19** + TypeScript (Vite)
- **shadcn/ui** + Tailwind CSS
- **TanStack Router** + **TanStack Query**
- **Zustand** (迥ｶ諷狗ｮ｡逅・
- **React Hook Form** + **Zod** (繝輔か繝ｼ繝/繝舌Μ繝・・繧ｷ繝ｧ繝ｳ)
- **react-i18next** (譌･譛ｬ隱・闍ｱ隱・

### 繧､繝ｳ繝輔Λ
- **Docker Compose** (繝ｭ繝ｼ繧ｫ繝ｫ髢狗匱)
- **AWS** (ECS Fargate, RDS, ElastiCache, ALB, CloudFront, ECR)
- **Terraform** (IaC)
- **GitHub Actions** (CI/CD)

### 繝｢繝九ち繝ｪ繝ｳ繧ｰ
- **ELK Stack** (繝ｭ繧ｰ蜿朱寔繝ｻ蜿ｯ隕門喧)
- **Prometheus** + **Grafana** (繝｡繝医Μ繧ｯ繧ｹ)
- **OpenTelemetry** (蛻・淵繝医Ξ繝ｼ繧ｷ繝ｳ繧ｰ)
- **Sentry** (繧ｨ繝ｩ繝ｼ繝医Λ繝・く繝ｳ繧ｰ)

## 讖溯・荳隕ｧ

- 柏 繝ｦ繝ｼ繧ｶ繝ｼ隱崎ｨｼ・医Ο繧ｰ繧､繝ｳ/逋ｻ骭ｲ/JWT・・
- 竢ｰ 蜃ｺ騾蜍､謇灘綾・亥・蜍､/騾蜍､/蜍､蜍呎凾髢楢・蜍戊ｨ育ｮ暦ｼ・
- 投 繝繝・す繝･繝懊・繝会ｼ亥・蜍､迥ｶ豕・谿区･ｭ/驛ｨ鄂ｲ蛻･邨ｱ險茨ｼ・
- 統 莨第嚊逕ｳ隲・謇ｿ隱阪Ρ繝ｼ繧ｯ繝輔Ο繝ｼ
- 套 繧ｷ繝輔ヨ邂｡逅・
- 則 繝ｦ繝ｼ繧ｶ繝ｼ邂｡逅・ｼ育ｮ｡逅・・ｰら畑・・

## 繧ｯ繧､繝・け繧ｹ繧ｿ繝ｼ繝・

### 蜑肴署譚｡莉ｶ
- Docker & Docker Compose
- Go 1.23+
- Node.js 22+
- pnpm 10+
- Make

### 繝ｭ繝ｼ繧ｫ繝ｫ髢狗匱

```bash
# 繝ｪ繝昴ず繝医Μ繧偵け繝ｭ繝ｼ繝ｳ
git clone https://github.com/your-org/kintai.git
cd kintai

# 迺ｰ蠅・､画焚繧定ｨｭ螳・
cp .env.example .env

# Docker Compose縺ｧ襍ｷ蜍・
make up

# DB繝槭う繧ｰ繝ｬ繝ｼ繧ｷ繝ｧ繝ｳ螳溯｡・
make migrate-up

# 繝｢繝九ち繝ｪ繝ｳ繧ｰ繧ｹ繧ｿ繝・け繧りｵｷ蜍輔☆繧句ｴ蜷・
make monitoring-up
```

### 蛟句挨襍ｷ蜍・

```bash
# 繝舌ャ繧ｯ繧ｨ繝ｳ繝峨・縺ｿ (繝帙ャ繝医Μ繝ｭ繝ｼ繝・
make backend-dev

# 繝輔Ο繝ｳ繝医お繝ｳ繝峨・縺ｿ
make frontend-dev
```

### Frontend (direct command)

```bash
cd frontend
pnpm install --frozen-lockfile
pnpm dev
```

### 繝・せ繝・

```bash
# 繝舌ャ繧ｯ繧ｨ繝ｳ繝峨ユ繧ｹ繝・
make backend-test

# 繝輔Ο繝ｳ繝医お繝ｳ繝峨ユ繧ｹ繝・
make frontend-test

# E2E繝・せ繝・
make frontend-e2e

# 繝ｪ繝ｳ繝・
make backend-lint
make frontend-lint
```

## 繧｢繧ｯ繧ｻ繧ｹ蜈・

| 繧ｵ繝ｼ繝薙せ | URL |
|----------|-----|
| 繝輔Ο繝ｳ繝医お繝ｳ繝・| http://localhost:3000 |
| 繝舌ャ繧ｯ繧ｨ繝ｳ繝陰PI | http://localhost:8080 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| Kibana | http://localhost:5601 |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3001 (admin/admin) |

## 繝励Ο繧ｸ繧ｧ繧ｯ繝域ｧ区・

```
kintai/
笏懌楳笏 backend/                  # Go繝舌ャ繧ｯ繧ｨ繝ｳ繝・
笏・  笏懌楳笏 cmd/server/           # 繧ｨ繝ｳ繝医Μ繝ｼ繝昴う繝ｳ繝・
笏・  笏懌楳笏 internal/
笏・  笏・  笏懌楳笏 config/           # 險ｭ螳・
笏・  笏・  笏懌楳笏 handler/          # HTTP繝上Φ繝峨Λ繝ｼ
笏・  笏・  笏懌楳笏 middleware/       # 繝溘ラ繝ｫ繧ｦ繧ｧ繧｢
笏・  笏・  笏懌楳笏 model/            # 繝｢繝・Ν/DTO
笏・  笏・  笏懌楳笏 repository/       # 繝・・繧ｿ繧｢繧ｯ繧ｻ繧ｹ螻､
笏・  笏・  笏懌楳笏 router/           # 繝ｫ繝ｼ繝・ぅ繝ｳ繧ｰ
笏・  笏・  笏披楳笏 service/          # 繝薙ず繝阪せ繝ｭ繧ｸ繝・け
笏・  笏懌楳笏 migrations/           # DB繝槭う繧ｰ繝ｬ繝ｼ繧ｷ繝ｧ繝ｳ
笏・  笏披楳笏 pkg/logger/           # 繝ｭ繧ｰ繝ｦ繝ｼ繝・ぅ繝ｪ繝・ぅ
笏懌楳笏 frontend/                 # React繝輔Ο繝ｳ繝医お繝ｳ繝・
笏・  笏披楳笏 src/
笏・      笏懌楳笏 api/              # API繧ｯ繝ｩ繧､繧｢繝ｳ繝・
笏・      笏懌楳笏 components/       # UI繧ｳ繝ｳ繝昴・繝阪Φ繝・
笏・      笏懌楳笏 i18n/             # 蝗ｽ髫帛喧
笏・      笏懌楳笏 pages/            # 繝壹・繧ｸ繧ｳ繝ｳ繝昴・繝阪Φ繝・
笏・      笏披楳笏 stores/           # 迥ｶ諷狗ｮ｡逅・
笏懌楳笏 infrastructure/           # Terraform
笏・  笏懌楳笏 environments/dev/     # dev迺ｰ蠅・
笏・  笏披楳笏 modules/              # 蜀榊茜逕ｨ蜿ｯ閭ｽ繝｢繧ｸ繝･繝ｼ繝ｫ
笏・      笏懌楳笏 vpc/              # VPC/繧ｵ繝悶ロ繝・ヨ
笏・      笏懌楳笏 ecr/              # 繧ｳ繝ｳ繝・リ繝ｬ繧ｸ繧ｹ繝医Μ
笏・      笏懌楳笏 alb/              # 繝ｭ繝ｼ繝峨ヰ繝ｩ繝ｳ繧ｵ繝ｼ
笏・      笏懌楳笏 rds/              # PostgreSQL
笏・      笏懌楳笏 elasticache/      # Redis
笏・      笏懌楳笏 ecs/              # ECS Fargate
笏・      笏披楳笏 cdn/              # CloudFront + S3
笏懌楳笏 monitoring/               # 繝｢繝九ち繝ｪ繝ｳ繧ｰ險ｭ螳・
笏・  笏懌楳笏 logstash/pipeline/    # Logstash繝代う繝励Λ繧､繝ｳ
笏・  笏懌楳笏 prometheus/           # Prometheus險ｭ螳・
笏・  笏披楳笏 otel/                 # OTel Collector險ｭ螳・
笏懌楳笏 .github/workflows/        # CI/CD
笏懌楳笏 docker-compose.yml
笏懌楳笏 Makefile
笏披楳笏 README.md
```

## API 繧ｨ繝ｳ繝峨・繧､繝ｳ繝・

### 隱崎ｨｼ
- `POST /api/v1/auth/register` - 繝ｦ繝ｼ繧ｶ繝ｼ逋ｻ骭ｲ
- `POST /api/v1/auth/login` - 繝ｭ繧ｰ繧､繝ｳ
- `POST /api/v1/auth/refresh` - 繝医・繧ｯ繝ｳ繝ｪ繝輔Ξ繝・す繝･
- `POST /api/v1/auth/logout` - 繝ｭ繧ｰ繧｢繧ｦ繝・

### 蜍､諤
- `POST /api/v1/attendance/clock-in` - 蜃ｺ蜍､謇灘綾
- `POST /api/v1/attendance/clock-out` - 騾蜍､謇灘綾
- `GET  /api/v1/attendance` - 蜍､諤荳隕ｧ
- `GET  /api/v1/attendance/today` - 譛ｬ譌･縺ｮ蜍､諤
- `GET  /api/v1/attendance/summary` - 蜍､諤繧ｵ繝槭Μ繝ｼ

### 莨第嚊
- `POST /api/v1/leaves` - 莨第嚊逕ｳ隲・
- `GET  /api/v1/leaves/my` - 閾ｪ蛻・・莨第嚊荳隕ｧ
- `GET  /api/v1/leaves/pending` - 謇ｿ隱榊ｾ・■荳隕ｧ
- `PUT  /api/v1/leaves/:id/approve` - 謇ｿ隱・蜊ｴ荳・

### 繧ｷ繝輔ヨ
- `POST /api/v1/shifts` - 繧ｷ繝輔ヨ菴懈・
- `POST /api/v1/shifts/bulk` - 荳諡ｬ菴懈・
- `GET  /api/v1/shifts` - 繧ｷ繝輔ヨ荳隕ｧ
- `DELETE /api/v1/shifts/:id` - 繧ｷ繝輔ヨ蜑企勁

### 縺昴・莉・
- `GET  /api/v1/health` - 繝倥Ν繧ｹ繝√ぉ繝・け
- `GET  /api/v1/users/me` - 繝ｭ繧ｰ繧､繝ｳ繝ｦ繝ｼ繧ｶ繝ｼ諠・ｱ
- `GET  /api/v1/dashboard/stats` - 繝繝・す繝･繝懊・繝臥ｵｱ險・

## 繝・・繝ｭ繧､

### Terraform (AWS)

```bash
cd infrastructure/environments/dev

# 蛻晄悄蛹・
terraform init

# 繝励Λ繝ｳ遒ｺ隱・
terraform plan

# 驕ｩ逕ｨ
terraform apply
```

### Docker 繧､繝｡繝ｼ繧ｸ繝薙Ν繝・& 繝励ャ繧ｷ繝･

```bash
# ECR繝ｭ繧ｰ繧､繝ｳ
aws ecr get-login-password --region ap-northeast-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.ap-northeast-1.amazonaws.com

# 繝薙Ν繝・& 繝励ャ繧ｷ繝･
docker build -t kintai-backend ./backend
docker tag kintai-backend:latest <ecr-url>/kintai-dev-backend:latest
docker push <ecr-url>/kintai-dev-backend:latest
```

## 繝ｩ繧､繧ｻ繝ｳ繧ｹ

MIT



