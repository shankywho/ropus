# AWS EKS Kubernetes Cluster Configuration
resource "aws_eks_cluster" "ropus_cluster" {
  name     = var.cluster_name
  role_arn = aws_iam_role.eks_cluster_role.arn
  version  = "1.30"

  vpc_config {
    subnet_ids = var.private_subnet_ids
  }
}

resource "aws_eks_node_group" "ropus_nodes" {
  cluster_name    = aws_eks_cluster.ropus_cluster.name
  node_group_name = "ropus-production-nodes"
  node_role_arn   = aws_iam_role.eks_node_role.arn
  subnet_ids      = var.private_subnet_ids

  scaling_config {
    desired_size = 6
    max_size     = 24
    min_size     = 3
  }

  instance_types = ["c6i.2xlarge"] # High compute for real-time risk decisioning
}

resource "aws_iam_role" "eks_cluster_role" {
  name = "ropus-eks-cluster-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "eks.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role" "eks_node_role" {
  name = "ropus-eks-node-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
}

variable "private_subnet_ids" {
  type    = list(string)
  default = ["subnet-01", "subnet-02", "subnet-03"]
}
