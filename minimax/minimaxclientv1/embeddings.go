package minimaxclientv1

import (
	"context"
	"errors"
	"fmt"
	"github.com/comqositi/kpllms/internal/httputils"
)

type EmbeddingPayload struct {
	Model string   `json:"model"`
	Texts []string `json:"texts"`
	Type  string   `json:"type"` // db: 存储，query：检索
}

type EmbeddingResponsePayload struct {
	Vectors     [][]float32 `json:"vectors"` // 一个文本对应一个float32数组，长度为1536
	TotalTokens int64       `json:"total_tokens"`
	BaseResp    BaseResp    `json:"base_resp"`
}

// nolint:lll
func (c *Client) CreateEmbedding(ctx context.Context, payload *EmbeddingPayload) (*EmbeddingResponsePayload, error) {
	if payload.Model == "" {
		payload.Model = c.embeddingsModel
	}
	if c.baseUrl == "" {
		c.baseUrl = defaultBaseUrl
	}
	if payload.Type == "" {
		return nil, errors.New("type 参数不能为空，db/query 二选一")
	}

	url := fmt.Sprintf("%s/embeddings?GroupId=%s", c.baseUrl, c.groupId)
	var resp EmbeddingResponsePayload
	err := httputils.HttpPost(ctx, url, payload, c.setHeader(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
