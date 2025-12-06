#!/bin/bash

SERVER="https://llmio.150129.xyz"
TOKEN="sk-LinHome-wo20Fang13145204eVer"

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL=0
PASSED=0
FAILED=0

# 测试函数
test_case() {
    local name="$1"
    local expected="$2"
    local result="$3"

    TOTAL=$((TOTAL + 1))

    if echo "$result" | grep -q "$expected"; then
        echo -e "${GREEN}✓${NC} $name"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $name"
        echo "  期望: $expected"
        echo "  实际: $(echo "$result" | head -c 100)..."
        FAILED=$((FAILED + 1))
        return 1
    fi
}

echo "=========================================="
echo "LLMIO 18种格式转换完整测试"
echo "=========================================="
echo ""

# ============================================
# 第 1 组：OpenAI 客户端格式测试
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "第 1 组：OpenAI 客户端格式 (6 种组合)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 1. OpenAI → OpenAI (流式)
echo "1.1 OpenAI → OpenAI Provider (流式)"
result=$(curl -s -N "$SERVER/v1/chat/completions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hi"}],"stream":true}' 2>&1 | head -5)
test_case "OpenAI → OpenAI (流式)" "data:.*choices.*delta" "$result"
echo ""

# 2. OpenAI → OpenAI (非流式)
echo "1.2 OpenAI → OpenAI Provider (非流式)"
result=$(curl -s "$SERVER/v1/chat/completions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hi"}],"stream":false}')
test_case "OpenAI → OpenAI (非流式)" "choices.*message.*content" "$result"
echo ""

# 3. OpenAI → Anthropic (流式)
echo "1.3 OpenAI → Anthropic Provider (流式)"
result=$(curl -s -N "$SERVER/v1/chat/completions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"Hi"}],"stream":true}' 2>&1 | head -5)
test_case "OpenAI → Anthropic (流式)" "data:.*choices.*delta" "$result"
echo ""

# 4. OpenAI → Anthropic (非流式)
echo "1.4 OpenAI → Anthropic Provider (非流式)"
result=$(curl -s "$SERVER/v1/chat/completions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"Hi"}],"stream":false}')
test_case "OpenAI → Anthropic (非流式)" "choices.*message.*content" "$result"
echo ""

# 5. OpenAI → OpenAI-Res (流式)
echo "1.5 OpenAI → OpenAI-Res Provider (流式)"
result=$(curl -s -N "$SERVER/v1/chat/completions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-5.1-codex-max","messages":[{"role":"user","content":"Hi"}],"stream":true}' 2>&1 | head -5)
test_case "OpenAI → OpenAI-Res (流式)" "data:.*choices.*delta" "$result"
echo ""

# 6. OpenAI → OpenAI-Res (非流式)
echo "1.6 OpenAI → OpenAI-Res Provider (非流式)"
result=$(curl -s "$SERVER/v1/chat/completions" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-5.1-codex-max","messages":[{"role":"user","content":"Hi"}],"stream":false}')
test_case "OpenAI → OpenAI-Res (非流式)" "choices.*message.*content" "$result"
echo ""

# ============================================
# 第 2 组：Anthropic 客户端格式测试
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "第 2 组：Anthropic 客户端格式 (6 种组合)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 7. Anthropic → Anthropic (流式)
echo "2.1 Anthropic → Anthropic Provider (流式)"
result=$(curl -s -N "$SERVER/v1/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"Hi"}],"max_tokens":50,"stream":true}' 2>&1 | head -5)
test_case "Anthropic → Anthropic (流式)" "event:.*content_block" "$result"
echo ""

# 8. Anthropic → Anthropic (非流式)
echo "2.2 Anthropic → Anthropic Provider (非流式)"
result=$(curl -s "$SERVER/v1/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"Hi"}],"max_tokens":50,"stream":false}')
test_case "Anthropic → Anthropic (非流式)" "content.*text" "$result"
echo ""

# 9. Anthropic → OpenAI (流式)
echo "2.3 Anthropic → OpenAI Provider (流式)"
result=$(curl -s -N "$SERVER/v1/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hi"}],"max_tokens":50,"stream":true}' 2>&1 | head -5)
test_case "Anthropic → OpenAI (流式)" "event:.*content_block" "$result"
echo ""

# 10. Anthropic → OpenAI (非流式)
echo "2.4 Anthropic → OpenAI Provider (非流式)"
result=$(curl -s "$SERVER/v1/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"Hi"}],"max_tokens":50,"stream":false}')
test_case "Anthropic → OpenAI (非流式)" "content.*text" "$result"
echo ""

# 11. Anthropic → OpenAI-Res (流式)
echo "2.5 Anthropic → OpenAI-Res Provider (流式)"
result=$(curl -s -N "$SERVER/v1/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"gpt-5.1-codex-max","messages":[{"role":"user","content":"Hi"}],"max_tokens":50,"stream":true}' 2>&1 | head -5)
test_case "Anthropic → OpenAI-Res (流式)" "event:.*content_block" "$result"
echo ""

# 12. Anthropic → OpenAI-Res (非流式)
echo "2.6 Anthropic → OpenAI-Res Provider (非流式)"
result=$(curl -s "$SERVER/v1/messages" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "anthropic-version: 2023-06-01" \
  -d '{"model":"gpt-5.1-codex-max","messages":[{"role":"user","content":"Hi"}],"max_tokens":50,"stream":false}')
test_case "Anthropic → OpenAI-Res (非流式)" "content.*text" "$result"
echo ""

# ============================================
# 第 3 组：OpenAI-Res 客户端格式测试
# ============================================
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "第 3 组：OpenAI-Res 客户端格式 (6 种组合)"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 13. OpenAI-Res → OpenAI-Res (流式)
echo "3.1 OpenAI-Res → OpenAI-Res Provider (流式)"
result=$(curl -s -N "$SERVER/v1/responses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-5.1-codex-max","input":"Hi","stream":true}' 2>&1 | head -5)
test_case "OpenAI-Res → OpenAI-Res (流式)" 'data:.*"output"' "$result"
echo ""

# 14. OpenAI-Res → OpenAI-Res (非流式)
echo "3.2 OpenAI-Res → OpenAI-Res Provider (非流式)"
result=$(curl -s "$SERVER/v1/responses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-5.1-codex-max","input":"Hi","stream":false}')
test_case "OpenAI-Res → OpenAI-Res (非流式)" '"output".*[^[]' "$result"
echo ""

# 15. OpenAI-Res → OpenAI (流式)
echo "3.3 OpenAI-Res → OpenAI Provider (流式)"
result=$(curl -s -N "$SERVER/v1/responses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","input":"Hi","stream":true}' 2>&1 | head -5)
test_case "OpenAI-Res → OpenAI (流式)" 'data:.*"output"' "$result"
echo ""

# 16. OpenAI-Res → OpenAI (非流式)
echo "3.4 OpenAI-Res → OpenAI Provider (非流式)"
result=$(curl -s "$SERVER/v1/responses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","input":"Hi","stream":false}')
test_case "OpenAI-Res → OpenAI (非流式)" '"output"' "$result"
echo ""

# 17. OpenAI-Res → Anthropic (流式)
echo "3.5 OpenAI-Res → Anthropic Provider (流式)"
result=$(curl -s -N "$SERVER/v1/responses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-sonnet-4-5","input":"Hi","stream":true}' 2>&1 | head -5)
test_case "OpenAI-Res → Anthropic (流式)" 'data:.*"output"' "$result"
echo ""

# 18. OpenAI-Res → Anthropic (非流式)
echo "3.6 OpenAI-Res → Anthropic Provider (非流式)"
result=$(curl -s "$SERVER/v1/responses" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-sonnet-4-5","input":"Hi","stream":false}')
test_case "OpenAI-Res → Anthropic (非流式)" '"output"' "$result"
echo ""

# ============================================
# 测试总结
# ============================================
echo "=========================================="
echo "测试总结"
echo "=========================================="
echo ""
echo "总计: $TOTAL 个测试"
echo -e "${GREEN}通过: $PASSED${NC}"
echo -e "${RED}失败: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ 所有 18 种格式转换测试通过！${NC}"
    exit 0
else
    echo -e "${RED}✗ 有 $FAILED 个测试失败${NC}"
    exit 1
fi
