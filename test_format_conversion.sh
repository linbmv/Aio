#!/bin/bash

# LLMIO 格式转换完整测试脚本
# 测试所有 9 种客户端格式 × Provider 类型组合（流式 + 非流式）

set -e

# 配置
BASE_URL="${BASE_URL:-https://llmio.150129.xyz}"
TOKEN="${TOKEN:-sk-LinHome-wo20Fang13145204eVer}"
TEST_INPUT="count to 3"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_test() {
    echo -e "\n${YELLOW}========================================${NC}"
    echo -e "${YELLOW}测试 #$TOTAL_TESTS: $1${NC}"
    echo -e "${YELLOW}========================================${NC}"
}

# 测试函数
test_request() {
    local test_name="$1"
    local endpoint="$2"
    local payload="$3"
    local stream="$4"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_test "$test_name"

    if [ "$stream" = "true" ]; then
        log_info "发送流式请求..."
        response=$(curl -s -N -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$payload" \
            "$BASE_URL$endpoint" 2>&1)
    else
        log_info "发送非流式请求..."
        response=$(curl -s -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$payload" \
            "$BASE_URL$endpoint" 2>&1)
    fi

    # 检查响应
    if echo "$response" | grep -q "error\|Error\|ERROR"; then
        log_error "测试失败！"
        echo "响应内容："
        echo "$response" | head -20
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    else
        log_info "测试通过！"
        echo "响应摘要："
        if [ "$stream" = "true" ]; then
            echo "$response" | grep "data:" | head -5
            echo "..."
            echo "$response" | tail -3
        else
            echo "$response" | jq -r '.output // .choices[0].message.content // .content[0].text' 2>/dev/null || echo "$response" | head -10
        fi
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    fi
}

# ============================================
# 测试矩阵：9 种组合 × 2 种模式（流式/非流式）
# ============================================

log_info "开始 LLMIO 格式转换完整测试"
log_info "测试服务器: $BASE_URL"
echo ""

# --------------------------------------------
# 1. OpenAI 客户端格式测试
# --------------------------------------------

# 1.1 OpenAI → OpenAI (直接)
test_request \
    "OpenAI客户端 → OpenAI Provider (非流式)" \
    "/v1/chat/completions" \
    '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"'"$TEST_INPUT"'"}],"stream":false}' \
    "false"

test_request \
    "OpenAI客户端 → OpenAI Provider (流式)" \
    "/v1/chat/completions" \
    '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"'"$TEST_INPUT"'"}],"stream":true}' \
    "true"

# 1.2 OpenAI → Anthropic (转换)
test_request \
    "OpenAI客户端 → Anthropic Provider (非流式)" \
    "/v1/chat/completions" \
    '{"model":"claude-3-5-sonnet-20241022","messages":[{"role":"user","content":"'"$TEST_INPUT"'"}],"stream":false}' \
    "false"

test_request \
    "OpenAI客户端 → Anthropic Provider (流式)" \
    "/v1/chat/completions" \
    '{"model":"claude-3-5-sonnet-20241022","messages":[{"role":"user","content":"'"$TEST_INPUT"'"}],"stream":true}' \
    "true"

# 1.3 OpenAI → OpenAI-Res (转换)
test_request \
    "OpenAI客户端 → OpenAI-Res Provider (非流式)" \
    "/v1/chat/completions" \
    '{"model":"gpt-5.1-codex-max","messages":[{"role":"user","content":"'"$TEST_INPUT"'"}],"stream":false}' \
    "false"

test_request \
    "OpenAI客户端 → OpenAI-Res Provider (流式)" \
    "/v1/chat/completions" \
    '{"model":"gpt-5.1-codex-max","messages":[{"role":"user","content":"'"$TEST_INPUT"'"}],"stream":true}' \
    "true"

# --------------------------------------------
# 2. Anthropic 客户端格式测试
# --------------------------------------------

# 2.1 Anthropic → OpenAI (转换)
test_request \
    "Anthropic客户端 → OpenAI Provider (非流式)" \
    "/v1/messages" \
    '{"model":"gpt-4o-mini","messages":[{"role":"user","content":[{"type":"text","text":"'"$TEST_INPUT"'"}]}],"stream":false,"max_tokens":1024}' \
    "false"

test_request \
    "Anthropic客户端 → OpenAI Provider (流式)" \
    "/v1/messages" \
    '{"model":"gpt-4o-mini","messages":[{"role":"user","content":[{"type":"text","text":"'"$TEST_INPUT"'"}]}],"stream":true,"max_tokens":1024}' \
    "true"

# 2.2 Anthropic → Anthropic (直接)
test_request \
    "Anthropic客户端 → Anthropic Provider (非流式)" \
    "/v1/messages" \
    '{"model":"claude-3-5-sonnet-20241022","messages":[{"role":"user","content":[{"type":"text","text":"'"$TEST_INPUT"'"}]}],"stream":false,"max_tokens":1024}' \
    "false"

test_request \
    "Anthropic客户端 → Anthropic Provider (流式)" \
    "/v1/messages" \
    '{"model":"claude-3-5-sonnet-20241022","messages":[{"role":"user","content":[{"type":"text","text":"'"$TEST_INPUT"'"}]}],"stream":true,"max_tokens":1024}' \
    "true"

# 2.3 Anthropic → OpenAI-Res (转换)
test_request \
    "Anthropic客户端 → OpenAI-Res Provider (非流式)" \
    "/v1/messages" \
    '{"model":"gpt-5.1-codex-max","messages":[{"role":"user","content":[{"type":"text","text":"'"$TEST_INPUT"'"}]}],"stream":false,"max_tokens":1024}' \
    "false"

test_request \
    "Anthropic客户端 → OpenAI-Res Provider (流式)" \
    "/v1/messages" \
    '{"model":"gpt-5.1-codex-max","messages":[{"role":"user","content":[{"type":"text","text":"'"$TEST_INPUT"'"}]}],"stream":true,"max_tokens":1024}' \
    "true"

# --------------------------------------------
# 3. OpenAI-Res 客户端格式测试（重点测试）
# --------------------------------------------

# 3.1 OpenAI-Res → OpenAI (转换)
test_request \
    "OpenAI-Res客户端 → OpenAI Provider (非流式)" \
    "/v1/responses" \
    '{"model":"gpt-4o-mini","input":"'"$TEST_INPUT"'","stream":false}' \
    "false"

test_request \
    "OpenAI-Res客户端 → OpenAI Provider (流式)" \
    "/v1/responses" \
    '{"model":"gpt-4o-mini","input":"'"$TEST_INPUT"'","stream":true}' \
    "true"

# 3.2 OpenAI-Res → Anthropic (转换)
test_request \
    "OpenAI-Res客户端 → Anthropic Provider (非流式)" \
    "/v1/responses" \
    '{"model":"claude-3-5-sonnet-20241022","input":"'"$TEST_INPUT"'","stream":false}' \
    "false"

test_request \
    "OpenAI-Res客户端 → Anthropic Provider (流式)" \
    "/v1/responses" \
    '{"model":"claude-3-5-sonnet-20241022","input":"'"$TEST_INPUT"'","stream":true}' \
    "true"

# 3.3 OpenAI-Res → OpenAI-Res (直接 - 这是修复的重点！)
test_request \
    "OpenAI-Res客户端 → OpenAI-Res Provider (非流式) ⭐️ 修复重点" \
    "/v1/responses" \
    '{"model":"gpt-5.1-codex-max","input":"'"$TEST_INPUT"'","stream":false}' \
    "false"

test_request \
    "OpenAI-Res客户端 → OpenAI-Res Provider (流式) ⭐️ 修复重点" \
    "/v1/responses" \
    '{"model":"gpt-5.1-codex-max","input":"'"$TEST_INPUT"'","stream":true}' \
    "true"

# ============================================
# 测试结果汇总
# ============================================

echo ""
echo "========================================"
echo "测试完成！"
echo "========================================"
echo "总测试数: $TOTAL_TESTS"
echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "${RED}失败: $FAILED_TESTS${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    log_info "🎉 所有测试通过！格式转换功能完全正常！"
    exit 0
else
    log_error "❌ 有 $FAILED_TESTS 个测试失败，请检查日志"
    exit 1
fi
