# 部署指南

## 快速部署

### 1. 拉取最新代码

```bash
cd /path/to/llmio
git pull origin fix/security-improvements
```

### 2. 编译

```bash
go build -o llmio .
```

### 3. 重启服务

```bash
# 如果使用 systemd
sudo systemctl restart llmio

# 或者直接运行
./llmio
```

### 4. 验证部署

```bash
# 运行快速测试
./quick_test.sh

# 或运行完整测试
./test_format_conversion.sh
```

## Docker 部署

### 1. 构建镜像

```bash
docker build -t llmio:latest .
```

### 2. 停止旧容器

```bash
docker stop llmio
docker rm llmio
```

### 3. 启动新容器

```bash
docker run -d \
  --name llmio \
  -p 8080:8080 \
  -e TOKEN=your_token \
  -v /path/to/db:/db \
  llmio:latest
```

### 4. 查看日志

```bash
docker logs -f llmio
```

## 验证修复

运行以下命令验证 OpenAI-Res 格式转换是否正常：

```bash
# 应该看到简化格式的输出
curl -N -H "Authorization: Bearer $TOKEN" \
  https://your-server.com/v1/responses \
  -d '{"model":"gpt-5.1-codex-max","input":"test","stream":true}'
```

**期望输出**：
```
data: {"model":"gpt-5.1-codex-max","output":"text"}
...
data: [DONE]
```

**不应该看到**：
```
event: response.created
data: {"type":"response.created",...}
```

## 故障排查

### 问题：仍然看到原始 Responses API 格式

**原因**：代码未正确部署

**解决**：
1. 确认 git pull 成功
2. 重新编译：`go build -o llmio .`
3. 确认进程已重启：`ps aux | grep llmio`
4. 检查日志中是否有编译错误

### 问题：流式响应只有 `: ping`

**原因**：可能是路由判断错误或上游 API 未返回数据

**解决**：
1. 启用 DEBUG 模式：`DEBUG_MODE=true ./llmio`
2. 查看日志中的 "ConvertStream" 相关信息
3. 运行 `debug_response.sh` 查看原始响应

### 问题：非流式响应返回无效 JSON

**原因**：可能是上游 API 返回错误或格式不符

**解决**：
1. 运行 `debug_response.sh` 查看原始响应
2. 检查 HTTP 状态码是否为 200
3. 检查 Provider 配置是否正确

## 回滚

如果新版本有问题，可以快速回滚：

```bash
git checkout HEAD~1
go build -o llmio .
sudo systemctl restart llmio
```
