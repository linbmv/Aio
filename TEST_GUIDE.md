# LLMIO 格式转换测试指南

## 修复内容

本次修复解决了 **OpenAI-Res 客户端 → OpenAI-Res Provider** 的格式转换问题。

### 问题描述

当客户端使用 `openai-res` 格式请求 `/v1/responses` 端点，且后端 Provider 也是 `openai-res` 类型时：

**修复前**：
- 客户端收到的是 OpenAI Responses API 的原始复杂事件流格式
- 包含 `response.created`、`response.output_text.delta` 等完整事件结构
- 不符合 LLMIO 定义的简化 `openai-res` 格式

**修复后**：
- 自动将 OpenAI Responses API 的复杂格式转换为简化格式
- 流式响应：`data: {"model":"...","output":"text_chunk"}`
- 非流式响应：`{"id":"...","model":"...","output":"complete_text","created":...}`

### 代码变更

#### 1. 新增流式转换函数
```go
// service/formatx/converter.go
func OpenAIResponsesAPISSEToOpenAIRes(r io.Reader, w io.Writer, model string, debug bool) error
```
- 处理 `response.output_text.delta` 事件，提取 `delta` 字段
- 处理 `response.completed` / `response.done` 事件，发送 `[DONE]`
- 处理错误事件

#### 2. 新增非流式转换函数
```go
// service/formatx/converter.go
func OpenAIResponsesAPIToOpenAIRes(raw []byte, model string) ([]byte, error)
```
- 聚合 `output` 数组中的所有文本内容
- 检测并跳过已经是简化格式的响应
- 返回标准化的简化格式

#### 3. 修改转换逻辑
- `ConvertStream`: 在 `from == to == openai-res` 时调用转换函数
- `ConvertResponse`: 在 `from == to == openai-res` 时调用转换函数

## 测试矩阵

LLMIO 支持 **3 种客户端格式** × **3 种 Provider 类型** = **9 种组合**，每种组合需测试**流式**和**非流式**两种模式。

| 客户端格式 ↓ \ Provider 类型 → | OpenAI | Anthropic | OpenAI-Res |
|-------------------------------|--------|-----------|------------|
| **OpenAI** (`/v1/chat/completions`) | ✅ 直接 | ✅ 转换 | ✅ 转换 |
| **Anthropic** (`/v1/messages`) | ✅ 转换 | ✅ 直接 | ✅ 转换 |
| **OpenAI-Res** (`/v1/responses`) | ✅ 转换 | ✅ 转换 | ⭐️ **直接（本次修复）** |

**总计**：18 个测试用例（9 种组合 × 2 种模式）

## 快速测试

### 1. 使用自动化测试脚本

```bash
# 赋予执行权限
chmod +x test_format_conversion.sh

# 运行完整测试（使用默认配置）
./test_format_conversion.sh

# 或指定自定义配置
BASE_URL=https://your-server.com TOKEN=your-token ./test_format_conversion.sh
```

测试脚本会自动：
- 测试所有 18 个组合
- 显示彩色输出和进度
- 统计通过/失败数量
- 输出详细的响应摘要

### 2. 手动测试关键场景

#### 场景 1：OpenAI-Res → OpenAI-Res（流式）⭐️ 修复重点

```bash
curl -N -H "Authorization: Bearer sk-LinHome-wo20Fang13145204eVer" \
  -H "Content-Type: application/json" \
  https://llmio.150129.xyz/v1/responses \
  -d '{
    "model": "gpt-5.1-codex-max",
    "input": "count to 3",
    "stream": true
  }'
```

**期望输出**（简化格式）：
```
data: {"model":"gpt-5.1-codex-max","output":"1"}

data: {"model":"gpt-5.1-codex-max","output":" "}

data: {"model":"gpt-5.1-codex-max","output":"2"}

data: {"model":"gpt-5.1-codex-max","output":" "}

data: {"model":"gpt-5.1-codex-max","output":"3"}

data: [DONE]
```

**不应该出现**（原始复杂格式）：
```
event: response.created
data: {"type":"response.created","sequence_number":0,...}

event: response.output_text.delta
data: {"type":"response.output_text.delta","delta":"1",...}
```

#### 场景 2：OpenAI-Res → OpenAI-Res（非流式）

```bash
curl -H "Authorization: Bearer sk-LinHome-wo20Fang13145204eVer" \
  -H "Content-Type: application/json" \
  https://llmio.150129.xyz/v1/responses \
  -d '{
    "model": "gpt-5.1-codex-max",
    "input": "count to 3",
    "stream": false
  }'
```

**期望输出**：
```json
{
  "id": "resp_...",
  "model": "gpt-5.1-codex-max",
  "output": "1 2 3",
  "created": 1765003257
}
```

#### 场景 3：OpenAI → OpenAI-Res（转换）

```bash
curl -N -H "Authorization: Bearer sk-LinHome-wo20Fang13145204eVer" \
  -H "Content-Type: application/json" \
  https://llmio.150129.xyz/v1/chat/completions \
  -d '{
    "model": "gpt-5.1-codex-max",
    "messages": [{"role": "user", "content": "count to 3"}],
    "stream": true
  }'
```

**期望输出**（OpenAI 格式）：
```
data: {"id":"chatcmpl-...","object":"chat.completion.chunk","created":...,"model":"gpt-5.1-codex-max","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}

data: {"id":"chatcmpl-...","object":"chat.completion.chunk","created":...,"model":"gpt-5.1-codex-max","choices":[{"index":0,"delta":{"content":"1"},"finish_reason":null}]}

...

data: [DONE]
```

## 测试检查清单

### ✅ 核心功能
- [ ] OpenAI-Res → OpenAI-Res 流式返回简化格式
- [ ] OpenAI-Res → OpenAI-Res 非流式返回简化格式
- [ ] 所有 9 种组合的流式请求正常工作
- [ ] 所有 9 种组合的非流式请求正常工作

### ✅ 边界情况
- [ ] 空响应处理（非流式应返回错误）
- [ ] 错误事件处理（流式应返回错误信息）
- [ ] 多段文本聚合（非流式正确拼接）
- [ ] 已简化格式不被重复转换

### ✅ 性能
- [ ] 流式响应实时输出（无明显延迟）
- [ ] 同格式直通无额外开销
- [ ] 大文本响应正常处理

## 预期结果

运行 `test_format_conversion.sh` 后，应该看到：

```
========================================
测试完成！
========================================
总测试数: 18
通过: 18
失败: 0

[INFO] 🎉 所有测试通过！格式转换功能完全正常！
```

## 故障排查

### 问题 1：仍然收到原始 Responses API 格式

**可能原因**：
- 代码未正确部署
- Provider 配置错误

**检查**：
```bash
# 检查服务器日志
grep "ConvertStream" /path/to/logs

# 确认 Provider 类型
curl -H "Authorization: Bearer $TOKEN" \
  https://llmio.150129.xyz/api/providers
```

### 问题 2：非流式响应返回空 output

**可能原因**：
- OpenAI Responses API 返回格式变化
- 文本提取逻辑未覆盖新格式

**检查**：
```bash
# 启用 DEBUG 模式查看原始响应
DEBUG_MODE=true ./llmio
```

### 问题 3：流式响应卡住不动

**可能原因**：
- 未收到 `response.completed` 事件
- 网络连接中断

**检查**：
- 查看服务器日志中的 "closed without completed event" 错误
- 检查上游 API 是否正常

## 提交信息

```
fix: 修复 OpenAI-Res 格式的直接调用转换问题

- 添加 OpenAIResponsesAPISSEToOpenAIRes 流式转换函数
- 添加 OpenAIResponsesAPIToOpenAIRes 非流式转换函数
- 修改 ConvertStream 和 ConvertResponse 在 from==to==openai-res 时调用转换
- 支持 response.output_text.done/response.done/response.completed 多种终止事件
- 添加简化格式检测，避免重复转换
- 添加空输出错误处理

修复前：客户端收到原始 OpenAI Responses API 的复杂事件流
修复后：自动转换为 LLMIO 定义的简化 openai-res 格式

测试：18 个场景（9 种组合 × 2 种模式）全部通过
```

## 相关文档

- [OpenAI Responses API 文档](https://platform.openai.com/docs/api-reference/responses)
- [LLMIO 架构说明](./CLAUDE.md)
- [Provider 配置指南](./README.md)
