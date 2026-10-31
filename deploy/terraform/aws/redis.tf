# --- ElastiCache Redis 7 (Multi-AZ with Replication) ---

resource "aws_elasticache_subnet_group" "redis" {
  name       = "bn-redis-subnet-group"
  subnet_ids = aws_subnet.database[*].id
}

resource "aws_security_group" "redis" {
  name   = "bn-redis-sg"
  vpc_id = aws_vpc.main.id

  ingress {
    from_port       = 6379
    to_port         = 6379
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

resource "aws_elasticache_replication_group" "redis" {
  replication_group_id       = "bn-redis-group"
  description                = "BookingNetwork Redis Cluster"
  node_type                  = "cache.t4g.medium"
  num_cache_clusters         = 2
  parameter_group_name       = "default.redis7"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.redis.name
  security_group_ids         = [aws_security_group.redis.id]
  automatic_failover_enabled = true
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  auth_token                 = var.redis_auth_token
}

resource "aws_secretsmanager_secret" "redis_url" {
  name = "bn/appmarket/redis_url"
}

resource "aws_secretsmanager_secret_version" "redis_url" {
  secret_id     = aws_secretsmanager_secret.redis_url.id
  secret_string = "rediss://:${var.redis_auth_token}@${aws_elasticache_replication_group.redis.primary_endpoint_address}:6379/0"
}

