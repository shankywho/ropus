# AWS ElastiCache Redis Cluster for Feature Store & Caching
resource "aws_elasticache_replication_group" "ropus_redis" {
  replication_group_id       = "ropus-production-redis"
  description                = "ElastiCache Redis cluster for sub-millisecond feature store"
  node_type                  = "cache.r6g.xlarge"
  num_cache_clusters         = 3
  automatic_failover_enabled = true
  transit_encryption_enabled = true
  at_rest_encryption_enabled = true
  port                       = 6379
}
