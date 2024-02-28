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
  region = var.AWS_REGION
  profile = var.AWS_PROFILE
}

terraform {
  backend "local" {}
  # backend "s3" {} 
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

# setup subnet

resource "aws_subnet" "dashdotdash_subnet" {
  vpc_id            = aws_vpc.dashdotdash_vpc.id
  cidr_block        = cidrsubnet(aws_vpc.dashdotdash_vpc.cidr_block, 8, 1)
  availability_zone = "ap-south-1b"

  # map_public_ip_on_launch = true
}

# setup security group
resource "aws_security_group" "api_sg" {
  name = "DashDotDashSg"
  vpc_id = aws_vpc.dashdotdash_vpc.id

  ## ssh: port 22
  ## tcp: port 80(http), 443(https)
  ## app: port 9090
  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_block = "0.0.0.0/0"
  }

  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_block = "0.0.0.0/0"
  }

  ingress {
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_block = "0.0.0.0/0"
  }

  ingress {
    from_port = 443
    to_port = 443
    protocol = "tcp"
    cidr_block = "0.0.0.0/0"
  }

  ingress {
    from_port = var.APP_PORT
    to_port = var.APP_PORT
    protocol = "tcp"
    cidr_block = "0.0.0.0/0"
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_block = "0.0.0.0/0"
  }
}


# setup keypair for ssh
resource "aws_key_pair" "tf-key-pair" {
  key_name = var.ECS_KEY_NAME
  public_key = tls_private_key.rsa.public_key_openssh
}

resource "tls_private_key" "rsa" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "local_file" "tf-key" {
  content  = tls_private_key.rsa.private_key_pem
  filename = "dash-tf-key-pair.pem" var.ECS_KEY_NAME
}


## ========================== Internet Gateway to Subnet Association ========================== ##
# setup internet gateway(IG) for external traffic connect to vpc
resource "aws_internet_gateway" "ddd_igw" {
  vpc_id = aws_vpc.dashdotdash_vpc.id
}

# setup route table association with (IG)
resource "aws_route_table" "ddd_rt" {
  vpc_id = aws_vpc.dashdotdash_vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.ddd_igw.id
  }
}

# associate the route table to subnet
resource "aws_route_table_association" "ddd_subnet_route" {
  subnet_id = aws_subnet.dashdotdash_vpc.id
  route_table_id = aws_route_table.ddd_rt.id
}


## ========================== Load Balancer ========================== ##
# setup a load balance (LB) with target group (TG) pointing to 
# application port
resource "aws_lb" "dashdotdash_lb" {
  name = "DashDotDashLB"
  internal = false
  load_balancer_type = "application"

  security_groups = [aws_security_group.app_sg.id]
  subnets = [aws_subnet.dashdotdash_subnet.id]

  enable_deletion_protection = false

  tags = {
    Name = "DashDotDashLB"
  }
}

resource "aws_lb_target_group" "app_lb_tg" {
    name = "${var.APP_NAME}-Tg"
    port = var.APP_PORT
    protocol = "HTTP"
    vpc_id = aws_vpc.dashdotdash_vpc.id

    health_check {    
      healthy_threshold   = 3    
      unhealthy_threshold = 10    
      timeout             = 10
      interval            = 30    
      path                = "/ping"    
      port                = "${var.APP_PORT}"
      matcher             = "200-299" 
  }
}

# setup the LB to listeners to forward requests
data "aws_lb_listener" "https_lb" {
  load_balancer_arn = aws_lb.dashdotdash_lb.arn
  port              = 443
}

# resource "aws_lb_listener_rule" "app_server_https_rule" {
#   listener_arn = aws_lb_listener.https_lb.arn
#   priority = 100
#
#   action {
#       type = "forward"
#       target_group_arn = aws_lb_target_group.app_lb_tg.arn
#     }
#
#   # condition {
#   #   host_header {
#   #     values = [""]
#   #   }
#   # }
# }

data "aws_lb_listener" "http_lb" {
  load_balancer_arn = aws_lb.dashdotdash_lb.arn
  port              = 80
}


resource "aws_lb_listener_rule" "app_server_http_rule" {
  listener_arn = aws_lb_listener.http_lb.arn
  priority = 100

  action {
      type = "forward"
      target_group_arn = aws_lb_target_group.app_lb_tg.arn
    }

  # condition {
  #   host_header {
  #     values = [""]
  #   }
  # }
}


## ========================== ECS Cluster ========================== ##
# create cluster
reosurce "aws_ecs_cluster" "dashdotdash_cluster" {
  name = var.ECS_CLUSTER_NAME
}

## ========================== Policy Attachment and IAM roles ========================== ##
# create sts role and iam policy attachment for EcsTaskExecutionRole
resource "aws_iam_role" "ecs_service_role" {
  name = "DashDotDashServiceRole"
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "ecsInstanceRole" {
  role       = aws_iam_role.ecs_service_role.name
  policy_arn =
  "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_instance_profile" "ecsInstanceProfile" {
  name = "DashDotDashEcsInstanceProfile"
  role = aws_iam_role.ecsInstanceRole.name
}


## ========================== ECS Deploy Service ========================== ##
# create ecs task definition and launch template
# this cluster name should match the cluster name in the ecs.sh file
locals {
  ecs_sh_content = templatefile("${path.module}/templates/ecs/ecs.sh", {
    ECS_CLUSTER_NAME = var.ECS_CLUSTER_NAME
  })
}

resource "aws_launch_template" "app_server_launch_configuration" {
  name_prefix = var.APP_NAME
  image_id      = "ami-027a0367928d05f3e"
  instance_type = "t2.micro"
  key_name      = var.ECS_KEY_NAME
  vpc_security_group_ids = [aws_security_group.app_sg.id]

  iam_instance_profile {
    arn = aws_iam_instance_profile.ecsInstanceProfile.arn
  }

  user_data = base64encode(local.ecs_sh_content)
}

resource "aws_ecs_task_definition" "server_task_definition" {
    family             = var.APP_NAME
    task_role_arn      = aws_iam_role.ecsInstanceRole.arn
    execution_role_arn = aws_iam_role.ecsInstanceRole.arn

    container_definitions = templatefile("${path.module}/templates/ecs/task-definition.json", {
      IMAGE: format("%s.dkr.ecr.%s.amazonaws.com/%s", var.AWS_ACCOUNT, var.AWS_REGION, var.APP_VERSION),
      APP_NAME: var.APP_NAME,
      APP_VERSION: var.APP_VERSION,
      APP_PORT: var.APP_PORT
    })
}

# create autoscaling group using the launch template
# this will create an ec2 instance using the template
# run task definition on it
resource "aws_autoscaling_group" "app_server_ecs_asg" {
  name = "${vap.APP_NAME}-Asg"
  vpc_zone_identifier = [aws_subnet.dashdotdash_vpc.id]
  target_group_arns = [aws_lb_target_group.app_lb_tg]

   launch_template {
    id = aws_launch_template.app_server_launch_configuration.id
    version = "$Latest"
  }

  min_size         = 1
  max_size         = 1
  desired_capacity = 1

  health_check_type         = "EC2"
  
  lifecycle {
    create_before_destroy = true
  }
  
  tag {
    key                 = "Name"
    value               = var.APP_NAME
    propagate_at_launch = true
  }

  tag {
   key                 = "AmazonECSManaged"
   value               = true
   propagate_at_launch = true
 }

 tag {
  key = "ForceRedeploy"
  value = 1
   propagate_at_launch = true
 }
}

# create aws_ecs_service and associate to cluster 
resource "aws_ecs_service" "app_ecs_service" {
    name = var.APP_NAME
    cluster = aws_ecs_cluster.dashdotdash_cluster.id 
    task_definition = aws_ecs_task_definition.server_task_definition.arn
    desired_count = 1

    force_new_deployment = true
    triggers = {
      redeployment = plantimestamp()
  }
}

# create loadbalancer with target group and listener
