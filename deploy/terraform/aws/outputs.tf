# Outputs consumed by post-apply smoke tests (see docs/LIVE_VERIFICATION_RUNBOOK.md).

output "alb_dns_name" {
  description = "Public ALB DNS name. API base = http(s)://<this> (use HTTPS listener when added)."
  value       = aws_lb.main.dns_name
}

output "rds_endpoint" {
  description = "Aurora cluster endpoint (hostname only)."
  value       = aws_rds_cluster.postgresql.endpoint
}

output "redis_endpoint" {
  description = "ElastiCache primary endpoint (hostname only)."
  value       = aws_elasticache_replication_group.redis.primary_endpoint_address
}
