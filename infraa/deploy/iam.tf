## ========================== Policy Attachment and IAM roles ========================== ##
# create sts role and iam policy attachment for EcsTaskExecutionRole
# data "aws_iam_policy_document" "ec2_policy_role" {
#   statement {
#     actions = ["sts:AssumeRole"]
#     principals {
#       type = "Service"
#       identifiers = ["ec2.amazonaws.com"]
#     }
#   }
# }

# data "aws_iam_policy_document" "ecs_policy_role" {
#   statement {
#     actions = ["sts:AssumeRole"]
#     principals {
#       type = "Service"
#       identifiers = ["ecs-tasks.amazonaws.com"]
#     }
#   }
# }

# data "aws_iam_policy_document" "cloudwatch_policy_role" {
#   statement {
#     actions = [
#       "logs:CreateLogGroup",
#       "logs:CreateLogStream",
#       "logs:PutLogEvents"
#     ]

#     resources = ["*"]
#   }
# }

# resource "aws_iam_role" "ecs_service_role" {
#   name = "DashDotDashEC2ServiceRole"
#   assume_role_policy = data.aws_iam_policy_document.ec2_policy_role.json 
# }

# resource "aws_iam_role" "ecs_task_role" {
#   name = "DashDotDashECSServiceRole"
#   assume_role_policy = data.aws_iam_policy_document.ecs_policy_role.json
# }

# resource "aws_iam_role_policy_attachment" "ecs_service_policy_attachment" {
#   for_each = toset([
#       "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy",
#       "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role",
#       "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
#   ])

#   role       = aws_iam_role.ecs_service_role.name
#   policy_arn = each.value
# }


# resource "aws_iam_instance_profile" "ecsInstanceProfile" {
#   name = "DashDotDashInstanceProfile"
#   role = aws_iam_role.ecs_service_role.name
# }

# resource "aws_iam_role_policy_attachment" "ecs_task_policy_attachment" {
#   for_each = toset([
#       "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy",
#       "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role",
#       "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
#   ])

#   role       = aws_iam_role.ecs_task_role.name
#   policy_arn = each.value
# }

