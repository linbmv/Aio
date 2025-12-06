#!/bin/bash

# 快速测试脚本 - 仅测试修复的关键场景

BASE_URL="${BASE_URL:-https://llmio.150129.xyz}"
TOKEN="${TOKEN:-sk-LinHome-wo20Fang13145204eVer}"

echo "=========================================="
echo "快速测试：OpenAI-Res 格式转换修复"
echo "=========================================="
echo ""

echo "测试 1: OpenAI-Res → OpenAI-Res (流式) ⭐️"
echo "------------------------------------------"
curl -N -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  "$BASE_URL/v1/responses" \
  -d '{"model":"gpt-5.1-codex-max","input":"count to 3","stream":true}'
echo ""
echo ""

echo "测试 2: OpenAI-Res → OpenAI-Res (非流式) ⭐️"
echo "------------------------------------------"
curl -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  "$BASE_URL/v1/responses" \
  -d '{"model":"gpt-5.1-codex-max","input":"count to 3","stream":false}' | jq .
echo ""
echo ""

echo "测试 3: OpenAI → OpenAI-Res (流式转换)"
echo "------------------------------------------"
curl -N -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  "$BASE_URL/v1/chat/completions" \
  -d '{"model":"gpt-5.1-codex-max","messages":[{"role":"user","content":"count to 3"}],"stream":true}' | head -20
echo ""
echo ""

echo "=========================================="
echo "✅ 如果看到简化格式的输出，说明修复成功！"
echo "=========================================="
