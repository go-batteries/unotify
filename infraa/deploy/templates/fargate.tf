# providers
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
  }

  required_version = ">= 1.2.0"
}

provider "aws" {
  region  = var.AWS_REGION
  profile = var.AWS_PROFILE
}

# setup vpc
resource "aws_vpc" "dashdotdash_vpc" {
  cidr_block = "10.0.0.0/20"
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags       = {
    Name = var.VPC_NAME
  }
}

# setup security group
resource "aws_security_group" "app_sg" {
  name        = "DashDotDashSg"
  description = "Security group for your application"
  vpc_id      = aws_vpc.dashdotdash_vpc.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = var.APP_PORT
    to_port     = var.APP_PORT
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

## ========================== ECS Cluster ========================== ##
# create cluster
resource "aws_ecs_cluster" "dashdotdash_cluster" {
  name = var.ECS_CLUSTER_NAME
}

## ========================== Policy Attachment and IAM roles ========================== ##
# create sts role and iam policy attachment for EcsTaskExecutionRole
resource "aws_iam_role" "ecs_service_role" {
  name             = "DashDotDashServiceRole"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "ecsTaskExecutionRole" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
  role       = aws_iam_role.ecs_service_role.name
}

## ========================== ECS Deploy Service ========================== ##
# create ecs task definition
resource "aws_ecs_task_definition" "server_task_definition" {
  family                   = var.APP_NAME
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  execution_role_arn = aws_iam_role.ecs_service_role.arn

  container_definitions = jsonencode([{
    name  = var.APP_NAME
    image = format("%s.dkr.ecr.%s.amazonaws.com/%s:%s", var.AWS_ACCOUNT, var.AWS_REGION, var.APP_NAME, var.APP_VERSION)
    portMappings = [{
      containerPort = var.APP_PORT
      hostPort      = var.APP_PORT
    }]
  }])
}

# create fargate service
resource "aws_ecs_service" "app_ecs_service" {
  name            = var.APP_NAME
  cluster         = aws_ecs_cluster.dashdotdash_cluster.id
  task_definition = aws_ecs_task_definition.server_task_definition.arn
  launch_type     = "FARGATE"
  desired_count   = 1

  network_configuration {
    subnets = [aws_subnet.dashdotdash_subnet.id]
    security_groups = [aws_security_group.app_sg.id]
  }
}

