#/bin/bash

echo "ECS_CLUSTER=${ECS_CLUSTER_NAME}" >> /etc/ecs/ecs.config

echo 300000 | sudo tee /proc/sys/fs/nr_open
