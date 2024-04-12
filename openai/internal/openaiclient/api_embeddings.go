package openaiclient

import (
	"context"
	"github.com/comqositi/kpllms/internal/httputils"
)

const (
	defaultEmbeddingModel = "text-embedding-ada-002"
)

// openai embedding 接口实现

type embeddingPayload struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponsePayload struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// nolint:lll
func (c *Client) createEmbedding(ctx context.Context, payload *embeddingPayload) (*embeddingResponsePayload, error) {
	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
	}

	if c.apiType == APITypeOpenAI {
		payload.Model = c.EmbeddingsModel
	}

	var response embeddingResponsePayload
	err := httputils.HttpPost(ctx, c.buildURL("/embeddings", c.EmbeddingsModel), payload, c.setHeaders(), &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
