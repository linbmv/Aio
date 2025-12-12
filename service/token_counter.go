package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// CountTokensRequest 符合 Anthropic messages/count_tokens 的请求结构
type CountTokensRequest struct {
	Model    string          `json:"model"`
	Messages []MessageParam  `json:"messages"`
	System   json.RawMessage `json:"system,omitempty"`
	Tools    []Tool          `json:"tools,omitempty"`
}

// MessageParam 消息体
type MessageParam struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// Tool 工具定义
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
}

// EstimateCountTokens 本地快速估算 tokens
func EstimateCountTokens(body []byte) (int, error) {
	var req CountTokensRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return 0, fmt.Errorf("解析请求失败: %w", err)
	}
	if req.Model == "" {
		return 0, fmt.Errorf("缺少模型参数")
	}
	if len(req.Messages) == 0 {
		return 0, fmt.Errorf("messages 不能为空")
	}
	if !isValidModel(req.Model) {
		return 0, fmt.Errorf("无效模型: %s", req.Model)
	}
	return estimateTokens(&req), nil
}

func estimateTokens(req *CountTokensRequest) int {
	total := 0

	// system 提示词
	if len(req.System) > 0 {
		var sysText string
		if err := json.Unmarshal(req.System, &sysText); err == nil && sysText != "" {
			total += estimateTextTokens(sysText)
			total += 5
		} else {
			var sysBlocks []json.RawMessage
			if err := json.Unmarshal(req.System, &sysBlocks); err == nil {
				for _, block := range sysBlocks {
					total += estimateContentBlock(block)
				}
				total += 5
			} else {
				total += len(req.System) / 4
			}
		}
	}

	// 消息内容
	for _, msg := range req.Messages {
		total += 10

		if len(msg.Content) == 0 {
			continue
		}

		var text string
		if err := json.Unmarshal(msg.Content, &text); err == nil {
			total += estimateTextTokens(text)
			continue
		}

		var blocks []json.RawMessage
		if err := json.Unmarshal(msg.Content, &blocks); err == nil {
			for _, block := range blocks {
				total += estimateContentBlock(block)
			}
			continue
		}

		total += len(msg.Content) / 4
	}

	// 工具开销
	toolCount := len(req.Tools)
	if toolCount > 0 {
		var baseOverhead, perToolOverhead int
		switch {
		case toolCount == 1:
			baseOverhead, perToolOverhead = 0, 400
		case toolCount <= 5:
			baseOverhead, perToolOverhead = 150, 150
		default:
			baseOverhead, perToolOverhead = 250, 80
		}

		total += baseOverhead

		for _, tool := range req.Tools {
			total += estimateToolName(tool.Name)
			total += estimateTextTokens(tool.Description)

			if len(tool.InputSchema) > 0 {
				schemaBytes := compactJSON(tool.InputSchema)

				var charsPerToken float64
				switch {
				case toolCount == 1:
					charsPerToken = 1.6
				case toolCount <= 5:
					charsPerToken = 1.9
				default:
					charsPerToken = 2.2
				}

				schemaTokens := int(float64(len(schemaBytes)) / charsPerToken)
				if bytes.Contains(schemaBytes, []byte("$schema")) {
					if toolCount == 1 {
						schemaTokens += 15
					} else {
						schemaTokens += 8
					}
				}

				minTokens := 80
				if toolCount > 5 {
					minTokens = 40
				}
				if schemaTokens < minTokens {
					schemaTokens = minTokens
				}

				total += schemaTokens
			}

			total += perToolOverhead
		}
	}

	total += 10
	return total
}

func estimateContentBlock(raw json.RawMessage) int {
	var block map[string]any
	if err := json.Unmarshal(raw, &block); err != nil {
		return max(len(raw)/4, 10)
	}

	blockType, _ := block["type"].(string)
	switch blockType {
	case "text":
		if text, ok := block["text"].(string); ok {
			return estimateTextTokens(text)
		}
		return 10
	case "image":
		return 1500
	case "document":
		return 500
	case "tool_use":
		if input, ok := block["input"]; ok {
			if b, err := json.Marshal(input); err == nil {
				return max(len(b)/4, 50)
			}
		}
		return 50
	case "tool_result":
		if content, ok := block["content"].(string); ok {
			return estimateTextTokens(content)
		}
		return 50
	default:
		if b, err := json.Marshal(block); err == nil {
			return max(len(b)/4, 10)
		}
		return 10
	}
}

func estimateToolName(name string) int {
	if name == "" {
		return 0
	}

	base := len(name) / 2
	underscorePenalty := strings.Count(name, "_")

	camel := 0
	for _, r := range name {
		if r >= 'A' && r <= 'Z' {
			camel++
		}
	}
	camelPenalty := camel / 2

	return max(base+underscorePenalty+camelPenalty, 2)
}

func estimateTextTokens(text string) int {
	if text == "" {
		return 0
	}

	total := 0
	sample := 0
	chinese := 0
	for _, r := range text {
		total++
		if sample < 500 {
			if r >= 0x4E00 && r <= 0x9FFF {
				chinese++
			}
			sample++
		}
	}

	if total == 0 {
		return 0
	}

	if sample == 0 {
		sample = total
	}

	chineseRatio := float64(chinese) / float64(sample)
	charsPerToken := 4.0 - (4.0-1.5)*chineseRatio

	tokens := int(float64(total) / charsPerToken)
	if tokens < 1 {
		tokens = 1
	}
	return tokens
}

func isValidModel(model string) bool {
	if model == "" {
		return false
	}
	model = strings.ToLower(model)
	prefixes := []string{
		"claude-",
		"anthropic.claude",
		"gpt-",
		"chatgpt-",
		"o1",
		"o3",
		"o4",
		"gemini-",
		"text-",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(model, p) {
			return true
		}
	}
	return false
}

func compactJSON(data []byte) []byte {
	var buf bytes.Buffer
	if err := json.Compact(&buf, data); err != nil {
		return data
	}
	return buf.Bytes()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
