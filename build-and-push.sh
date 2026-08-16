#!/bin/bash
# ================================================
# SwiftLead Backend - Build & Push Docker Image
# ================================================
# Script ini digunakan oleh PEMILIK codebase untuk
# build dan push Docker image ke registry.
#
# Usage:
#   ./build-and-push.sh                    # default: ghcr.io/swiftlead/swiftlead-backend:latest
#   ./build-and-push.sh v1.0.0             # dengan tag version
#   REGISTRY=docker.io/username ./build-and-push.sh  # registry lain
# ================================================

set -euo pipefail

# Configuration
REGISTRY="${REGISTRY:-docker.io/mfuadfakhruzzaki}"
IMAGE_NAME="${IMAGE_NAME:-swiftlead-backend}"
TAG="${1:-latest}"
FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}:${TAG}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  SwiftLead Backend - Build & Push${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Step 1: Login check
echo -e "${YELLOW}[1/4] Checking registry login...${NC}"
if echo "${REGISTRY}" | grep -q "ghcr.io"; then
    echo -e "  Registry: GitHub Container Registry (ghcr.io)"
    echo -e "  ${YELLOW}Pastikan sudah login:${NC} echo \$GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin"
elif echo "${REGISTRY}" | grep -q "docker.io"; then
    echo -e "  Registry: Docker Hub"
    echo -e "  ${YELLOW}Pastikan sudah login:${NC} docker login"
else
    echo -e "  Registry: ${REGISTRY}"
fi
echo ""

# Step 2: Build image
echo -e "${YELLOW}[2/4] Building Docker image...${NC}"
echo -e "  Image: ${FULL_IMAGE}"
echo ""

docker build -t "${FULL_IMAGE}" -f Dockerfile .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}  ✓ Build berhasil!${NC}"
else
    echo -e "${RED}  ✗ Build gagal!${NC}"
    exit 1
fi
echo ""

# Step 3: Tag as latest (if version tag given)
if [ "${TAG}" != "latest" ]; then
    LATEST_IMAGE="${REGISTRY}/${IMAGE_NAME}:latest"
    echo -e "${YELLOW}[3/4] Tagging juga sebagai latest...${NC}"
    docker tag "${FULL_IMAGE}" "${LATEST_IMAGE}"
    echo -e "${GREEN}  ✓ Tagged: ${LATEST_IMAGE}${NC}"
else
    echo -e "${YELLOW}[3/4] Skipping extra tag (already latest)${NC}"
fi
echo ""

# Step 4: Push to registry
echo -e "${YELLOW}[4/4] Pushing ke registry...${NC}"
docker push "${FULL_IMAGE}"

if [ "${TAG}" != "latest" ]; then
    docker push "${LATEST_IMAGE}"
fi

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}================================================${NC}"
    echo -e "${GREEN}  ✓ Push berhasil!${NC}"
    echo -e "${GREEN}================================================${NC}"
    echo ""
    echo -e "Image tersedia di: ${BLUE}${FULL_IMAGE}${NC}"
    if [ "${TAG}" != "latest" ]; then
        echo -e "                   ${BLUE}${LATEST_IMAGE}${NC}"
    fi
    echo ""
    echo -e "${YELLOW}Langkah selanjutnya:${NC}"
    echo -e "  1. Kirimkan folder ${BLUE}dist/${NC} ke teman Anda"
    echo -e "  2. Teman Anda cukup jalankan:"
    echo -e "     ${GREEN}cp .env.example .env${NC}"
    echo -e "     ${GREEN}docker compose -f docker-compose.local.yml up -d${NC}"
else
    echo -e "${RED}  ✗ Push gagal! Pastikan sudah login ke registry.${NC}"
    exit 1
fi
