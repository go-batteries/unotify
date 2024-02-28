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

resource "aws_iam_policy" "cloudwatch_logs_policy" {
  name        = "CloudWatchLogsPolicy"
  description = "IAM policy for CloudWatch Logs"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Effect   = "Allow",
        Resource = "arn:aws:logs:${var.AWS_REGION}:${var.AWS_ACCOUNT}:log-group:/ecs/${var.APP_NAME}:*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_task_poliy_attachment" {
  role       = aws_iam_role.ecs_service_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "ecr_access_attachment" {
  role      = aws_iam_role.ecs_service_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role_policy_attachment" "cloudwatch_logs_attachment" {
  role      = aws_iam_role.ecs_service_role.name
  policy_arn = aws_iam_policy.cloudwatch_logs_policy.arn
}

resource "aws_iam_instance_profile" "ecsInstanceProfile" {
  name = "DashDotDashEcsInstanceProfile"
  role = aws_iam_role.ecs_service_role.name
}
