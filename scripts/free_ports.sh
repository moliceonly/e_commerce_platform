#!/usr/bin/env bash
# 释放本仓库常用端口，并停掉本机 mysql/redis（需 sudo）
set -euo pipefail

PORTS=(8080 3306 6379 8081 3307 6380 33060)

echo "==> docker compose down"
cd "$(dirname "$0")/.."
sudo docker compose -f deployments/docker-compose.yml down --remove-orphans 2>/dev/null || true

echo "==> kill listeners on ${PORTS[*]}"
for p in "${PORTS[@]}"; do
  sudo fuser -k "${p}/tcp" 2>/dev/null || true
done

echo "==> stop host mysql / redis (if any)"
sudo service mysql stop 2>/dev/null || sudo systemctl stop mysql 2>/dev/null || true
sudo service mariadb stop 2>/dev/null || true
sudo service redis-server stop 2>/dev/null || true

echo "==> remaining listeners:"
sudo ss -ltnp | grep -E ':8080|:3306|:6379|:8081|:3307|:6380|:33060' || echo "  (none)"
echo "done. then: sudo docker compose -f deployments/docker-compose.yml up --build"
