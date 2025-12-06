#!/bin/bash

# 调试脚本 - 查看原始响应格式

BASE_URL="${BASE_URL:-https://llmio.150129.xyz}"
TOKEN="${TOKEN:-sk-LinHome-wo20Fang13145204eVer}"

echo "=========================================="
echo "调试：查看原始响应格式"
echo "=========================================="
echo ""

echo "1. 捕获流式响应原始数据"
echo "------------------------------------------"
curl -N -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  "$BASE_URL/v1/responses" \
  -d '{"model":"gpt-5.1-codex-max","input":"count to 3","stream":true}' \
  --output /tmp/stream_response.txt 2>&1

echo "前 50 行："
head -50 /tmp/stream_response.txt
echo ""
echo "总行数："
wc -l /tmp/stream_response.txt
echo ""

echo "2. 捕获非流式响应原始数据"
echo "------------------------------------------"
response=$(curl -s -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  "$BASE_URL/v1/responses" \
  -d '{"model":"gpt-5.1-codex-max","input":"count to 3","stream":false}')

echo "响应长度: ${#response} 字节"
echo "响应内容:"
echo "$response"
echo ""
echo "尝试格式化 JSON:"
echo "$response" | jq . 2>&1 || echo "不是有效的 JSON"
echo ""

echo "3. 检查 HTTP 状态码和头部"
echo "------------------------------------------"
curl -v -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  "$BASE_URL/v1/responses" \
  -d '{"model":"gpt-5.1-codex-max","input":"count to 3","stream":false}' \
  2>&1 | grep -E "^< HTTP|^< Content-Type|^< Content-Length"
echo ""

echo "=========================================="
echo "调试完成"
echo "=========================================="
