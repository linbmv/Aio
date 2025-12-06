#!/bin/bash

# 检查部署状态

BASE_URL="${BASE_URL:-https://llmio.150129.xyz}"
TOKEN="${TOKEN:-sk-LinHome-wo20Fang13145204eVer}"

echo "=========================================="
echo "检查服务器部署状态"
echo "=========================================="
echo ""

echo "1. 检查服务器是否在线"
echo "------------------------------------------"
curl -s -o /dev/null -w "HTTP状态码: %{http_code}\n" "$BASE_URL/api/providers" \
  -H "Authorization: Bearer $TOKEN"
echo ""

echo "2. 测试基本的 OpenAI 端点（应该正常工作）"
echo "------------------------------------------"
response=$(curl -s -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  "$BASE_URL/v1/chat/completions" \
  -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}],"stream":false}')

if echo "$response" | grep -q "error code: 520"; then
    echo "❌ 服务器返回 520 错误 - 服务可能未正常运行"
    echo "响应: $response"
elif echo "$response" | grep -q "choices"; then
    echo "✅ OpenAI 端点正常工作"
    echo "响应摘要: $(echo "$response" | jq -r '.choices[0].message.content' 2>/dev/null || echo '无法解析')"
else
    echo "⚠️  收到意外响应"
    echo "响应: $response"
fi
echo ""

echo "3. 测试 OpenAI-Res 端点（修复的重点）"
echo "------------------------------------------"
response=$(curl -s -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  "$BASE_URL/v1/responses" \
  -d '{"model":"gpt-5.1-codex-max","input":"hi","stream":false}')

if echo "$response" | grep -q "error code: 520"; then
    echo "❌ OpenAI-Res 端点返回 520 错误"
    echo "响应: $response"
    echo ""
    echo "可能的原因："
    echo "1. 服务器未部署最新代码"
    echo "2. 服务器进程崩溃或重启中"
    echo "3. Provider 配置有问题"
    echo "4. 上游 API 不可用"
elif echo "$response" | grep -q "output"; then
    echo "✅ OpenAI-Res 端点正常工作"
    echo "响应: $response"
else
    echo "⚠️  收到意外响应"
    echo "响应: $response"
fi
echo ""

echo "=========================================="
echo "诊断建议"
echo "=========================================="
echo ""
echo "如果看到 520 错误，请执行以下步骤："
echo ""
echo "1. SSH 到服务器并检查服务状态："
echo "   systemctl status llmio"
echo "   # 或"
echo "   ps aux | grep llmio"
echo ""
echo "2. 查看服务器日志："
echo "   journalctl -u llmio -n 50"
echo "   # 或"
echo "   tail -f /path/to/llmio.log"
echo ""
echo "3. 确认代码已更新："
echo "   cd /path/to/llmio"
echo "   git log --oneline -1"
echo "   # 应该看到: 758a54f fix: 修复 OpenAI-Res 格式的直接调用转换问题"
echo ""
echo "4. 重新编译并重启："
echo "   go build -o llmio ."
echo "   systemctl restart llmio"
echo ""
