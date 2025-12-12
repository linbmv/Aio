package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/tidwall/sjson"
)

type Anthropic struct {
	BaseURL string      `json:"base_url"`
	APIKey  string      `json:"api_key"`
	Keys    []KeyConfig `json:"keys"`
	Version string      `json:"version"`
}

// pickKey 随机抽取状态有效的 key，兼容旧 api_key 配置
func (a *Anthropic) pickKey() string {
	if len(a.Keys) > 0 {
		type activeKey struct {
			term   string
			remark string
		}
		active := make([]activeKey, 0, len(a.Keys))
		for _, key := range a.Keys {
			if key.Status && key.Term != "" {
				active = append(active, activeKey{term: key.Term, remark: key.Remark})
			}
		}
		if len(active) > 0 {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			selected := active[rng.Intn(len(active))]
			keyHint := selected.term[len(selected.term)-4:]
			if selected.remark != "" {
				keyHint = selected.remark
			}
			fmt.Printf("[Anthropic] using key: %s\n", keyHint)
			return selected.term
		}
	}
	return a.APIKey
}

func (a *Anthropic) BuildReq(ctx context.Context, header http.Header, model string, rawBody []byte) (*http.Request, error) {
	body, err := sjson.SetBytes(rawBody, "model", model)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/messages", a.BaseURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if header != nil {
		req.Header = header
	}
	req.Header.Set("content-type", "application/json")
	if req.Header.Get("x-api-key") == "" {
		req.Header.Set("x-api-key", a.pickKey())
	}
	req.Header.Set("anthropic-version", a.Version)
	return req, nil
}

type AnthropicModelsResponse struct {
	Data    []AnthropicModel `json:"data"`
	FirstID string           `json:"first_id"`
	HasMore bool             `json:"has_more"`
	LastID  string           `json:"last_id"`
}

type AnthropicModel struct {
	CreatedAt   time.Time `json:"created_at"`
	DisplayName string    `json:"display_name"`
	ID          string    `json:"id"`
	Type        string    `json:"type"`
}

func (a *Anthropic) Models(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/models", a.BaseURL), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", a.pickKey())
	req.Header.Set("anthropic-version", a.Version)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}
	var anthropicModels AnthropicModelsResponse
	if err := json.NewDecoder(res.Body).Decode(&anthropicModels); err != nil {
		return nil, err
	}

	var modelList ModelList
	for _, model := range anthropicModels.Data {
		modelList.Data = append(modelList.Data, Model{
			ID:      model.ID,
			Created: model.CreatedAt.Unix(),
		})
	}
	return modelList.Data, nil
}

func (a *Anthropic) BuildCountTokensReq(ctx context.Context, header http.Header, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/messages/count_tokens", a.BaseURL), body)
	if err != nil {
		return nil, err
	}
	if header != nil {
		req.Header = header
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", a.pickKey())
	req.Header.Set("anthropic-version", a.Version)
	return req, nil
}
