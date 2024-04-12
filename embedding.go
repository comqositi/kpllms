package kpllms

import "context"

// Embedder  向量接口
type Embedder interface {
	// EmbedDocuments 文档存储向量：存入数据库的向量，被用于检索
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
	// EmbedQuery 查询向量：对需要用于检索的文本进行想量化
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
}
