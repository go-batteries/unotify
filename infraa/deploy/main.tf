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
  backend "local" {
    path = "./tfstates/dashsotdash.tfstates"
  }
  # backend "s3" {} 
}

# setup vpc
resource "aws_vpc" "dashdotdash_vpc" {
  cidr_block = "192.168.0.0/20"
  enable_dns_support   = true
  enable_dns_hostnames = true
  tags       = {
      Name = var.VPC_NAME
  }
}

# setup subnet
resource "aws_subnet" "dashdotdash_subnet_b" {
  vpc_id            = aws_vpc.dashdotdash_vpc.id
  cidr_block        = cidrsubnet(aws_vpc.dashdotdash_vpc.cidr_block, 8, 1)
  availability_zone = "ap-south-1b"

  # map_public_ip_on_launch = true
  tags = {
    Name = "DDD_SubnetB"
  }
}

resource "aws_subnet" "dashdotdash_subnet" {
  vpc_id            = aws_vpc.dashdotdash_vpc.id
  cidr_block        = cidrsubnet(aws_vpc.dashdotdash_vpc.cidr_block, 8, 2)
  availability_zone = "ap-south-1a"

  map_public_ip_on_launch = true

  tags = {
    Name = "DDD_SubnetA"
  }
}

# setup security group
resource "aws_security_group" "app_sg" {
  name = "DashDotDashSg"
  vpc_id = aws_vpc.dashdotdash_vpc.id

  ## ssh: port 22
  ## tcp: port 80(http), 443(https)
  ## app: port 9090
  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 80
    to_port = 80
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 443
    to_port = 443
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = var.APP_PORT
    to_port = var.APP_PORT
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}


# setup keypair for ssh
resource "aws_key_pair" "dash-tf-key-pair" {
  key_name = var.ECS_KEY_NAME
  public_key = tls_private_key.rsa.public_key_openssh
}

resource "tls_private_key" "rsa" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "local_file" "tf-key" {
  content  = tls_private_key.rsa.private_key_pem
  filename = "${var.ECS_KEY_NAME}.pem"
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
  subnet_id = aws_subnet.dashdotdash_subnet.id
  route_table_id = aws_route_table.ddd_rt.id
}

resource "aws_route_table_association" "ddd_subnet_route_b" {
  subnet_id = aws_subnet.dashdotdash_subnet_b.id
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
  subnets = [aws_subnet.dashdotdash_subnet.id, aws_subnet.dashdotdash_subnet_b.id]

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
      path                = "/arij/ping"    
      port                = "${var.APP_PORT}"
      matcher             = "200-299" 
  }
}

# setup the LB to listeners to forward requests

resource "aws_lb_listener" "http_lb_listener" {
  load_balancer_arn = aws_lb.dashdotdash_lb.arn
  port = "80"
  protocol = "HTTP"

  default_action {
    type = "fixed-response"
 
    fixed_response {
      content_type = "text/plain"
      message_body = "HEALTHY"
      status_code  = "200"
    }
  }
}

resource "aws_lb_listener_rule" "app_server_http_rule" {
  listener_arn = aws_lb_listener.http_lb_listener.arn
  priority = 100

  action {
      type = "forward"
      target_group_arn = aws_lb_target_group.app_lb_tg.arn
  }

  condition {
    path_pattern {
      values = ["/arij/*"]
    }
  }
}


## ========================== ECS Cluster ========================== ##
# create cluster
resource "aws_ecs_cluster" "dashdotdash_cluster" {
  name = var.ECS_CLUSTER_NAME
}

# create sts role and iam policy attachment for EcsTaskExecutionRole
# iam.tf

## ========================== ECS Deploy Service ========================== ##
# create ecs task definition and launch template
# this cluster name should match the cluster name in the ecs.sh file

data "aws_iam_instance_profile" "ecs_instance_profile_arn" {
  name = "ecsInstanceProfile"
}

data "aws_iam_role" "ecsTaskExecutionRole" {
  name = "ecsTaskExecutionRole"
}

# create autoscaling group using the launch template
# this will create an ec2 instance using the template
# run task definition on it
resource "aws_autoscaling_group" "app_server_ecs_asg" {
  name = "${var.APP_NAME}-Asg"
  vpc_zone_identifier = [
    aws_subnet.dashdotdash_subnet.id,
     aws_subnet.dashdotdash_subnet_b.id,
  ]
  target_group_arns = [aws_lb_target_group.app_lb_tg.arn]

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
    arn = data.aws_iam_instance_profile.ecs_instance_profile_arn.arn
  }

  user_data = base64encode(local.ecs_sh_content)
  # user_data = filebase64("${path.module}/templates/ecs/dashdotdash.sh")

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = "${var.APP_NAME}"
    }
  }
}

resource "aws_ecs_task_definition" "server_task_definition" {
    family             = var.APP_NAME
    task_role_arn      = data.aws_iam_role.ecsTaskExecutionRole.arn
    execution_role_arn = data.aws_iam_role.ecsTaskExecutionRole.arn

    container_definitions = templatefile("${path.module}/templates/ecs/task-definition.json", {
      IMAGE: format("%s.dkr.ecr.%s.amazonaws.com/%s:%s", var.AWS_ACCOUNT, var.AWS_REGION, var.APP_NAME, var.APP_VERSION),
      APP_NAME: var.APP_NAME,
      APP_VERSION: var.APP_VERSION,
      APP_PORT: var.APP_PORT
    })
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
