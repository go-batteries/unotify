#!/bin/bash

sudo echo "ECS_CLUSTER=${ECS_CLUSTER_NAME}" > /etc/ecs/ecs.config
export ECS_REDIS_URL=${REDIS_URL}

sudo yum -y install logrotate.x86_64

server_logrotate_config="/etc/logrotate.d/server"
worker_logrotate_config="/etc/logrotate.d/worker"

cat << EOF | sudo tee $server_logrotate_config > /dev/null
/var/log/server.log {
    daily
    rotate 2
    missingok
    notifempty
    compress
    delaycompress
    create 0640 root root
}
EOF

cat << EOF | sudo tee $worker_logrotate_config > /dev/null
/var/log/worker.log {
    daily
    rotate 2
    missingok
    notifempty
    compress
    delaycompress
    create 0640 root root
}
EOF

# Force log rotation
sudo logrotate -f $server_logrotate_config
sudo logrotate -f $worker_logrotate_config
