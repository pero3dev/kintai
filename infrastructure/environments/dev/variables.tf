variable "aws_region" {
  description = "AWSリージョン"
  type        = string
  default     = "ap-northeast-1"
}

variable "environment" {
  description = "環境名"
  type        = string
  default     = "dev"
}

variable "project_name" {
  description = "プロジェクト名"
  type        = string
  default     = "kintai"
}

variable "vpc_cidr" {
  description = "VPCのCIDRブロック"
  type        = string
  default     = "10.0.0.0/16"
}

variable "db_username" {
  description = "RDSマスターユーザー名"
  type        = string
  default     = "kintai"
  sensitive   = true
}

variable "db_password" {
  description = "RDSマスターパスワード"
  type        = string
  sensitive   = true
}

variable "backend_image" {
  description = "バックエンドDockerイメージ"
  type        = string
  default     = ""
}

variable "frontend_image" {
  description = "フロントエンドDockerイメージ"
  type        = string
  default     = ""
}
