# AWS RDS Multi-AZ PostgreSQL Database
resource "aws_db_instance" "ropus_postgres" {
  identifier             = "ropus-production-pg"
  allocated_storage      = 500
  max_allocated_storage  = 2000
  engine                 = "postgres"
  engine_version         = "16.3"
  instance_class         = "db.r6g.2xlarge"
  multi_az               = true
  db_name                = "ropus_db"
  username               = "ropus_admin"
  password               = var.db_password
  publicly_accessible    = false
  skip_final_snapshot    = false
  deletion_protection    = true
  storage_encrypted      = true
}

variable "db_password" {
  type      = string
  sensitive = true
  default   = "ropus_production_secure_db_pass_2026"
}
