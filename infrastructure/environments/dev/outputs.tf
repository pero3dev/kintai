output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "alb_dns_name" {
  description = "ALB DNS名"
  value       = module.alb.alb_dns_name
}

output "cloudfront_domain" {
  description = "CloudFrontドメイン"
  value       = module.cdn.cloudfront_domain_name
}

output "ecr_backend_url" {
  description = "バックエンドECRリポジトリURL"
  value       = module.ecr.backend_repository_url
}

output "ecr_frontend_url" {
  description = "フロントエンドECRリポジトリURL"
  value       = module.ecr.frontend_repository_url
}

output "rds_endpoint" {
  description = "RDSエンドポイント"
  value       = module.rds.endpoint
  sensitive   = true
}

output "redis_endpoint" {
  description = "Redisエンドポイント"
  value       = module.elasticache.endpoint
}
