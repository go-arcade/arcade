#!/bin/bash

# ================================================================
# 插件校验和生成工具
# 用途: 为插件文件生成 SHA256 校验和，用于安全验证
# 使用: ./generate_plugin_checksum.sh <plugin_file>
# ================================================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查参数
if [ $# -eq 0 ]; then
    echo -e "${RED}错误: 请提供插件文件路径${NC}"
    echo "使用方法: $0 <plugin_file>"
    echo "示例: $0 plugins/notify/slack.so"
    exit 1
fi

PLUGIN_FILE="$1"

# 检查文件是否存在
if [ ! -f "$PLUGIN_FILE" ]; then
    echo -e "${RED}错误: 文件不存在: $PLUGIN_FILE${NC}"
    exit 1
fi

# 计算 SHA256
echo -e "${YELLOW}正在计算 SHA256 校验和...${NC}"
CHECKSUM=$(sha256sum "$PLUGIN_FILE" | awk '{print $1}')

# 获取文件信息
FILE_SIZE=$(ls -lh "$PLUGIN_FILE" | awk '{print $5}')
FILE_NAME=$(basename "$PLUGIN_FILE")

# 输出结果
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}插件文件信息${NC}"
echo -e "${GREEN}========================================${NC}"
echo "文件路径: $PLUGIN_FILE"
echo "文件名称: $FILE_NAME"
echo "文件大小: $FILE_SIZE"
echo ""
echo -e "${YELLOW}SHA256 校验和:${NC}"
echo -e "${GREEN}$CHECKSUM${NC}"
echo ""

# 生成 SQL 更新语句
echo -e "${YELLOW}SQL 更新语句:${NC}"
echo "UPDATE \`t_plugin\` SET \`checksum\` = '$CHECKSUM' WHERE \`entry_point\` = '$PLUGIN_FILE' OR \`install_path\` = '$PLUGIN_FILE';"
echo ""

# 生成验证命令
echo -e "${YELLOW}验证命令:${NC}"
echo "sha256sum -c <<< \"$CHECKSUM  $PLUGIN_FILE\""
echo ""

# 保存到文件（可选）
CHECKSUM_FILE="${PLUGIN_FILE}.sha256"
echo "$CHECKSUM  $PLUGIN_FILE" > "$CHECKSUM_FILE"
echo -e "${GREEN}✓ 校验和已保存到: $CHECKSUM_FILE${NC}"

