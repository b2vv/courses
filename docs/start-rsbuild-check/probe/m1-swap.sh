#!/bin/zsh
# usage: m1-swap.sh <image-tag>  — run frontend from that exact image (+ m1-delay), wait healthy, print ids
. <kiri>/environment.sh >/dev/null 2>&1
export COMPOSE_FILE="$COMPOSE_FILE:<scratch>/m1-probe/m1-override.yml"
export M1_FRONT_IMAGE="$1"
docker compose up -d --no-build --no-deps m1-delay frontend >/dev/null 2>&1 || { echo UP_FAILED; exit 1; }
for i in {1..60}; do s=$(docker inspect courses-frontend-1 --format '{{.State.Health.Status}}'); [[ $s == healthy ]] && break; sleep 2; done
echo "health=$s want=$(docker image inspect $1 --format '{{.Id}}') running=$(docker inspect courses-frontend-1 --format '{{.Image}}') arch=$(docker image inspect $1 --format '{{.Architecture}}')"
