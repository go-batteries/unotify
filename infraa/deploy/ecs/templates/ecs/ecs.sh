#!/bin/bash

sudo echo "ECS_CLUSTER=${ECS_CLUSTER_NAME}" > /etc/ecs/ecs.config
export ECS_REDIS_URL=${REDIS_URL}

echo "APP_VERSION=${APP_VERSION}" > /etc/ecs/app_version

