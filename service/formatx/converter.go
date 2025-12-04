package formatx

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIToAnthropic 将 OpenAI 请求转换为 Anthropic 格式
func OpenAIToAnthropic(raw []byte) ([]byte, error) {
	var req struct {
		Model       string `json:"model"`
		Messages    []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		MaxTokens   int     `json:"max_tokens,omitempty"`
		Temperature float64 `json:"temperature,omitempty"`
		Stream      bool    `json:"stream,omitempty"`
	}

	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}

	// 提取 system 消息
	var system string
	var messages []map[string]interface{}

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			system = msg.Content
		} else {
			messages = append(messages, map[string]interface{}{
				"role": msg.Role,
				"content": []map[string]string{
					{"type": "text", "text": msg.Content},
				},
			})
		}
	}

	anthReq := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
		"stream":   req.Stream,
	}

	if system != "" {
		anthReq["system"] = system
	}
	if req.MaxTokens > 0 {
		anthReq["max_tokens"] = req.MaxTokens
	}
	if req.Temperature > 0 {
		anthReq["temperature"] = req.Temperature
	}

	return json.Marshal(anthReq)
}

// AnthropicToOpenAI 将 Anthropic 响应转换为 OpenAI 格式
func AnthropicToOpenAI(raw []byte, model string) ([]byte, error) {
	var resp struct {
		ID      string `json:"id"`
		Model   string `json:"model"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, err
	}

	var content string
	if len(resp.Content) > 0 {
		content = resp.Content[0].Text
	}

	openaiResp := map[string]interface{}{
		"id":      resp.ID,
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]string{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]int{
			"prompt_tokens":     resp.Usage.InputTokens,
			"completion_tokens": resp.Usage.OutputTokens,
			"total_tokens":      resp.Usage.InputTokens + resp.Usage.OutputTokens,
		},
	}

	return json.Marshal(openaiResp)
}

// AnthropicSSEToOpenAI 将 Anthropic SSE 流转换为 OpenAI SSE 格式
func AnthropicSSEToOpenAI(r io.Reader, w io.Writer, model string) error {
	scanner := bufio.NewScanner(r)
	var eventType string
	chunkID := fmt.Sprintf("chatcmpl-%d", time.Now().Unix())

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}

		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

			switch eventType {
			case "content_block_delta":
				var delta struct {
					Delta struct {
						Text string `json:"text"`
					} `json:"delta"`
				}
				if err := json.Unmarshal([]byte(data), &delta); err == nil && delta.Delta.Text != "" {
					chunk := map[string]interface{}{
						"id":      chunkID,
						"object":  "chat.completion.chunk",
						"created": time.Now().Unix(),
						"model":   model,
						"choices": []map[string]interface{}{
							{
								"index": 0,
								"delta": map[string]string{
									"content": delta.Delta.Text,
								},
							},
						},
					}
					chunkBytes, _ := json.Marshal(chunk)
					fmt.Fprintf(w, "data: %s\n\n", chunkBytes)
					if flusher, ok := w.(http.Flusher); ok {
						flusher.Flush()
					}
				}

			case "message_stop":
				fmt.Fprintf(w, "data: [DONE]\n\n")
				if flusher, ok := w.(http.Flusher); ok {
					flusher.Flush()
				}
				return nil
			}
		}
	}

	return scanner.Err()
}
