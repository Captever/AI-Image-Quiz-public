provider "aws" {
  region = "ap-northeast-2"  # 서울 리전
}

variable "vpc_id" {
  description = "The ID of the existing VPC"
  type        = string
  default     = "vpc-xxxxxxxxxxxxxxxxxx"  # 사용할 VPC ID로 변경
}

variable "security_group_id" {
  description = "The ID of the existing security group"
  type        = string
  default     = "sg-xxxxxxxxxxxxxxxxx"  # 사용할 보안 그룹 ID로 변경
}

variable "subnet_ids" {
  description = "The IDs of the existing subnets"
  type        = list(string)
  default     = ["subnet-xxxxxxxxxxxxxxxxx", "subnet-xxxxxxxxxxxxxxxxxx"]  # 사용할 서브넷 ID로 변경
}

resource "aws_db_subnet_group" "rds_subnet_group" {
  name       = "rds-subnet-group"
  subnet_ids = var.subnet_ids

  tags = {
    Name = "rds-subnet-group"
  }
}


resource "aws_rds_cluster" "aurora_cluster" {
  cluster_identifier      = "<RDS-cluster-name>"
  engine                  = "aurora-mysql"
  engine_version          = "8.0.mysql_aurora.3.05.2"
  master_username         = "<master-name>"
  master_password         = "<master-pw>"
  database_name           = "<init-DB-name>"
  db_subnet_group_name    = aws_db_subnet_group.rds_subnet_group.name
  vpc_security_group_ids  = [var.security_group_id]
  backup_retention_period = 1
  storage_encrypted       = false

  serverlessv2_scaling_configuration {
    min_capacity = 2
    max_capacity = 8
  }

  tags = {
    Name = "<name>"
  }
}

resource "aws_rds_cluster_instance" "aurora_instance" {
  count                = 1
  identifier           = "${aws_rds_cluster.aurora_cluster.cluster_identifier}-instance-${count.index}"
  cluster_identifier   = aws_rds_cluster.aurora_cluster.id
  instance_class       = "db.serverless"
  engine               = aws_rds_cluster.aurora_cluster.engine
  engine_version       = aws_rds_cluster.aurora_cluster.engine_version
  db_subnet_group_name = aws_db_subnet_group.rds_subnet_group.name
  publicly_accessible  = false

  tags = {
    Name = "<name>"
  }
}

output "rds_cluster_endpoint" {
  value = aws_rds_cluster.aurora_cluster.endpoint
}

output "rds_cluster_reader_endpoint" {
  value = aws_rds_cluster.aurora_cluster.reader_endpoint
}
