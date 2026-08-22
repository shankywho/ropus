# AWS Secrets Manager for Enterprise Credentials
resource "aws_secretsmanager_secret" "ropus_app_secrets" {
  name                    = "ropus/production/credentials"
  recovery_window_in_days = 7
}

resource "aws_secretsmanager_secret_version" "ropus_app_secrets_val" {
  secret_id = aws_secretsmanager_secret.ropus_app_secrets.id
  secret_string = jsonencode({
    jwt_secret      = "ropus_prod_jwt_super_secret_key_2026"
    database_url    = "postgres://ropus_admin:ropus_production_secure_db_pass_2026@ropus-production-pg.internal:5432/ropus_db"
    redis_url       = "redis://ropus-production-redis.internal:6379"
    kafka_brokers   = "b-1.ropus-kafka.internal:9092,b-2.ropus-kafka.internal:9092"
    llm_api_key     = "sk-ropus-prod-llm-reasoning-key"
  })
}
