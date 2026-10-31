# --- Aurora Serverless v2 PostgreSQL 16 (Multi-AZ) ---

resource "aws_db_subnet_group" "rds" {
  name       = "bn-rds-subnet-group"
  subnet_ids = aws_subnet.database[*].id
}

resource "aws_security_group" "rds" {
  name   = "bn-rds-sg"
  vpc_id = aws_vpc.main.id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_rds_cluster" "postgresql" {
  cluster_identifier     = "bn-aurora-cluster-${var.environment}"
  engine                 = "aurora-postgresql"
  engine_version         = "16.1"
  database_name          = "bn_appmarket"
  master_username        = "bn_app"
  master_password        = var.db_password
  db_subnet_group_name   = aws_db_subnet_group.rds.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  skip_final_snapshot    = true
  storage_encrypted      = true

  serverlessv2_scaling_configuration {
    max_capacity = 16.0
    min_capacity = 0.5
  }
}

resource "aws_rds_cluster_instance" "instances" {
  count              = 2
  identifier         = "bn-aurora-instance-${count.index + 1}"
  cluster_identifier = aws_rds_cluster.postgresql.id
  instance_class     = "db.serverless"
  engine             = aws_rds_cluster.postgresql.engine
  engine_version     = aws_rds_cluster.postgresql.engine_version
}

# --- Secrets Manager for Connection Strings ---
resource "aws_secretsmanager_secret" "db_url" {
  name = "bn/appmarket/database_url"
}

resource "aws_secretsmanager_secret_version" "db_url" {
  secret_id     = aws_secretsmanager_secret.db_url.id
  secret_string = "postgres://bn_app:${var.db_password}@${aws_rds_cluster.postgresql.endpoint}:5432/bn_appmarket?sslmode=require"
}

