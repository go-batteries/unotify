output "aws_lb_name" {
    description = "AWS ALB DNS Name"
    value = aws_lb.dashdotdash_lb.dns_name
}