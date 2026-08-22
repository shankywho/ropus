# AWS MSK Managed Apache Kafka Cluster
resource "aws_msk_cluster" "ropus_kafka" {
  cluster_name           = "ropus-production-kafka"
  kafka_version          = "3.6.0"
  number_of_broker_nodes = 6

  broker_node_group_info {
    instance_type   = "kafka.m5.2xlarge"
    client_subnets  = var.private_subnet_ids
    security_groups = [aws_security_group.kafka_sg.id]
  }

  encryption_info {
    encryption_in_transit {
      client_broker = "TLS"
      in_cluster    = true
    }
  }
}

resource "aws_security_group" "kafka_sg" {
  name        = "ropus-kafka-sg"
  description = "Security group for MSK Kafka"
}
