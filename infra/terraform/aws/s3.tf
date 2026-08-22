# AWS S3 Bucket for ML Model Artifacts & Audit Logs
resource "aws_s3_bucket" "ropus_model_store" {
  bucket        = "ropus-enterprise-model-artifacts-prod"
  force_destroy = false
}

resource "aws_s3_bucket_versioning" "model_store_versioning" {
  bucket = aws_s3_bucket.ropus_model_store.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "model_store_encryption" {
  bucket = aws_s3_bucket.ropus_model_store.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}
