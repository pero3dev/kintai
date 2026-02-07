terraform {
  required_version = ">= 1.7"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # リモートバックエンドは本番運用時に有効化
  # backend "s3" {
  #   bucket = "kintai-terraform-state"
  #   key    = "dev/terraform.tfstate"
  #   region = "ap-northeast-1"
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "kintai"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}
