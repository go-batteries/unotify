output "aws_lb_name" {
    description = "AWS ALB DNS Name"
    value = "http://${aws_lb.dashdotdash_lb.dns_name}"
}

output "redis_hostname" {
  value = aws_elasticache_cluster.redis_cache_cluster.cache_nodes.0.address
}

output "redis_port" {
  value = aws_elasticache_cluster.redis_cache_cluster.cache_nodes.0.port
}

output "redis_endpoint" {
  value = "${aws_elasticache_cluster.redis_cache_cluster.cache_nodes[0].address}:${aws_elasticache_cluster.redis_cache_cluster.cache_nodes[0].port}"
}
