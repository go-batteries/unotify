#!/bin/bash

echo "ECS_CLUSTER=${ECS_CLUSTER_NAME}" > /home/ec2-user/ecs.config
sudo echo "ECS_CLUSTER=${ECS_CLUSTER_NAME}" > /etc/ecs/ecs.config


