#!/bin/bash

# 夸克网盘 CLI 工具构建脚本
# 用于为多个平台编译二进制文件

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="kuake"
BUILD_DIR="dist"

# 从 main.go 中提取版本号（使用 sed 以兼容 macOS/BSD）
VERSION=$(sed -n 's/.*var Version = "\([^"]*\)".*/\1/p' cmd/main.go | head -1)
if [ -z "$VERSION" ]; then
    echo -e "${RED}Error: Failed to extract version from cmd/main.go${NC}"
    exit 1
fi

# 清理并创建输出目录
if [ -d "$BUILD_DIR" ]; then
    echo -e "${YELLOW}Cleaning ${BUILD_DIR} directory...${NC}"
    rm -rf "${BUILD_DIR}"/*
fi
mkdir -p "$BUILD_DIR"

echo -e "${GREEN}Starting build for ${PROJECT_NAME}...${NC}"
echo "Version: $VERSION (extracted from cmd/main.go)"
echo "Output directory: $BUILD_DIR"
echo ""

# 构建函数
build() {
    local os=$1
    local arch=$2
    local ext=$3
    local tags=$4  # 构建标签（debug 或空）
    
    # 构建输出文件名（包含版本号）
    local output_name="${PROJECT_NAME}-${VERSION}-${os}-${arch}${ext}"
    local output_path="${BUILD_DIR}/${output_name}"
    
    echo -e "${YELLOW}Building ${os}/${arch} (version: ${VERSION})${tags:+ (tags: $tags)}...${NC}"
    
    local build_cmd="CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath -ldflags=\"-s -w\""
    if [ -n "$tags" ]; then
        build_cmd="$build_cmd -tags=\"$tags\""
    fi
    build_cmd="$build_cmd -o \"$output_path\" ./cmd"
    
    eval $build_cmd
    
    if [ -f "$output_path" ]; then
        # 为 macOS 和 Linux 二进制文件添加执行权限
        if [ "$os" != "windows" ]; then
            chmod +x "$output_path"
        fi
        # 获取文件大小
        local size=$(du -h "$output_path" | cut -f1)
        echo -e "${GREEN}✓ Build successful: ${output_name} (${size})${NC}"
    else
        echo -e "${RED}✗ Build failed: ${output_name}${NC}"
        exit 1
    fi
    echo ""
}

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go compiler not found, please install Go first${NC}"
    exit 1
fi

echo -e "${GREEN}Go version: $(go version)${NC}"
echo ""

# 构建各平台（Release 版本，不包含调试信息）
build "linux" "amd64" "" ""
build "linux" "arm64" "" ""
build "darwin" "amd64" "" ""
build "windows" "amd64" ".exe" ""

# 复制环境变量模板到 dist（与仓库根 .env.example 一致）
echo -e "${YELLOW}Copying .env.example to ${BUILD_DIR}/...${NC}"
if [ -f ".env.example" ]; then
    cp ".env.example" "${BUILD_DIR}/.env.example"
    echo -e "${GREEN}✓ Example env template: ${BUILD_DIR}/.env.example${NC}"
else
    echo -e "${RED}Warning: .env.example not found at repo root, skipped${NC}"
fi
echo ""

# 打包 OpenClaw 技能目录（与 OS 无关，随版本号命名，便于 Release 附件分发）
echo -e "${YELLOW}Packaging OpenClaw skill (openclaw/kuake_skill)...${NC}"
if [ ! -f "openclaw/kuake_skill/SKILL.md" ]; then
    echo -e "${RED}Error: openclaw/kuake_skill/SKILL.md not found${NC}"
    exit 1
fi
SKILL_ZIP="${BUILD_DIR}/kuake_skill-${VERSION}.zip"
SKILL_TGZ="${BUILD_DIR}/kuake_skill-${VERSION}.tar.gz"
rm -f "$SKILL_ZIP" "$SKILL_TGZ"
if command -v zip >/dev/null 2>&1; then
    (cd openclaw && zip -qr "../${SKILL_ZIP}" kuake_skill)
    SKILL_ARCHIVE_NAME="${SKILL_ZIP#$BUILD_DIR/}"
    echo -e "${GREEN}✓ OpenClaw skill: ${SKILL_ARCHIVE_NAME}${NC}"
elif command -v tar >/dev/null 2>&1; then
    tar -czf "$SKILL_TGZ" -C openclaw kuake_skill
    SKILL_ARCHIVE_NAME="${SKILL_TGZ#$BUILD_DIR/}"
    echo -e "${GREEN}✓ OpenClaw skill: ${SKILL_ARCHIVE_NAME} (no zip in PATH; install zip for .zip)${NC}"
else
    echo -e "${RED}Error: need 'zip' or 'tar' to package openclaw/kuake_skill${NC}"
    exit 1
fi
echo ""

echo -e "${GREEN}All platforms built successfully!${NC}"
echo ""
echo "Built files:"
ls -lhA "$BUILD_DIR" | grep -E "(${PROJECT_NAME}|kuake_skill|\\.env\\.example)" | awk '{print "  " $9 " (" $5 ")"}'
echo ""
echo -e "${YELLOW}Usage instructions:${NC}"
echo "  Linux/macOS:"
echo "    ./dist/kuake-${VERSION}-linux-amd64 <command>"
echo "    ./dist/kuake-${VERSION}-darwin-amd64 <command>"
echo "  Windows:"
echo "    dist\\kuake-${VERSION}-windows-amd64.exe <command>"
echo "  OpenClaw:"
echo "    Extract ${SKILL_ARCHIVE_NAME}, use the inner kuake_skill folder as the skill directory, and install a kuake binary from ${BUILD_DIR}/ onto PATH (see README)."
echo ""
echo -e "${YELLOW}Note: Binaries, .env.example, and OpenClaw skill archive are under ${BUILD_DIR}/; copy .env.example to .env and fill credentials (see README).${NC}"

