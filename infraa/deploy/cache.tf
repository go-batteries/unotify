resource "aws_security_group" "cache_server" {
  name = "DashDotDashCacheServer"
  description = "AWS ElastiCache for Redis"
  vpc_id = aws_vpc.dashdotdash_vpc.id

  ingress {
    from_port = 6379
    to_port = 6379
    protocol = "tcp"
    security_groups = [aws_security_group.app_sg.id]
  }

  ingress {
    from_port = 6379
    to_port = 6379
    protocol = "tcp"
    cidr_blocks = [aws_subnet.dashdotdash_subnet_b.cidr_block]
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "DashDotDashCacheServer"
  }
}

resource "aws_elasticache_subnet_group" "redis_cache_group" {
  name       = "my-cache-subnet"
  subnet_ids = [aws_subnet.dashdotdash_subnet_b.id]
}


resource "aws_elasticache_cluster" "redis_cache_cluster" {
  cluster_id           = "dashdotdash-redis-cluster"
  engine               = "redis"
  node_type            = "cache.t3.micro"
  num_cache_nodes      = 1
  engine_version       = "7.1"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.redis_cache_group.name
  security_group_ids = [ aws_security_group.cache_server.id ]
}
