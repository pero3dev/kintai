locals {
  name_prefix = "${var.project_name}-${var.environment}"
  az_count    = 2
  azs         = slice(data.aws_availability_zones.available.names, 0, local.az_count)
}

data "aws_availability_zones" "available" {
  state = "available"
}

# ===== VPC Module =====
module "vpc" {
  source = "../../modules/vpc"

  name_prefix = local.name_prefix
  vpc_cidr    = var.vpc_cidr
  azs         = local.azs
}

# ===== ECR =====
module "ecr" {
  source = "../../modules/ecr"

  name_prefix = local.name_prefix
}

# ===== ALB =====
module "alb" {
  source = "../../modules/alb"

  name_prefix       = local.name_prefix
  vpc_id            = module.vpc.vpc_id
  public_subnet_ids = module.vpc.public_subnet_ids
}

# ===== ECS =====
module "ecs" {
  source = "../../modules/ecs"

  name_prefix               = local.name_prefix
  vpc_id                    = module.vpc.vpc_id
  private_subnet_ids        = module.vpc.private_subnet_ids
  backend_target_group_arn  = module.alb.backend_target_group_arn
  frontend_target_group_arn = module.alb.frontend_target_group_arn
  alb_security_group_id     = module.alb.alb_security_group_id
  backend_image             = var.backend_image != "" ? var.backend_image : "${module.ecr.backend_repository_url}:latest"
  frontend_image            = var.frontend_image != "" ? var.frontend_image : "${module.ecr.frontend_repository_url}:latest"

  backend_env_vars = {
    APP_ENV      = var.environment
    APP_PORT     = "8080"
    DATABASE_URL = "postgres://${var.db_username}:${var.db_password}@${module.rds.address}:${module.rds.port}/kintai?sslmode=require"
    REDIS_URL    = "redis://${module.elasticache.endpoint}:${module.elasticache.port}/0"
  }
}

# ===== RDS =====
module "rds" {
  source = "../../modules/rds"

  name_prefix                = local.name_prefix
  vpc_id                     = module.vpc.vpc_id
  private_subnet_ids         = module.vpc.private_subnet_ids
  db_name                    = "kintai"
  db_username                = var.db_username
  db_password                = var.db_password
  allowed_security_group_ids = [module.ecs.ecs_security_group_id]
}

# ===== ElastiCache =====
module "elasticache" {
  source = "../../modules/elasticache"

  name_prefix                = local.name_prefix
  vpc_id                     = module.vpc.vpc_id
  private_subnet_ids         = module.vpc.private_subnet_ids
  allowed_security_group_ids = [module.ecs.ecs_security_group_id]
}

# ===== S3 + CloudFront =====
module "cdn" {
  source = "../../modules/cdn"

  name_prefix  = local.name_prefix
  alb_dns_name = module.alb.alb_dns_name
}
