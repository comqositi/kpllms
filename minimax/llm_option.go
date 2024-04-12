package minimax

import (
	"github.com/comqositi/kpllms/minimax/minimaxclientv1"
)

const (
	groupIdEnvVarName = "MINIMAX_GROUP_ID" //nolint:gosec
	apiKeyEnvVarName  = "MINIMAX_API_KEY"  //nolint:gosec
	baseURLEnvVarName = "OPENAI_BASE_URL"  //nolint:gosec

	defaultModel          = "abab5.5-chat"
	defaultEmbeddingModel = "embo-01"

	defaultSendType       = "BOT"
	defaultSendName       = "用户"
	defaultBotName        = "靠谱智能助理"
	defaultBotDescription = "靠谱大语言模型是一款由靠谱AI智能科技自研的，没有调用其他产品的接口的大型语言模型。靠谱AI智能科技是一家中国科技公司，一直致力于进行大模型相关的研究。"
)

type options struct {
	groupId        string
	apiKey         string
	baseUrl        string
	httpClient     minimaxclientv1.Doer
	embeddingModel string
	model          string
}

type Option func(*options)

func WithGroupId(value string) Option {
	return func(o *options) {
		o.groupId = value
	}
}

func WithApiKey(value string) Option {
	return func(o *options) {
		o.apiKey = value
	}
}

func WithBaseUrl(value string) Option {
	return func(o *options) {
		o.baseUrl = value
	}
}

func WithHttpClient(value minimaxclientv1.Doer) Option {
	return func(o *options) {
		o.httpClient = value
	}
}

func WithEmbeddingModel(value string) Option {
	return func(o *options) {
		o.embeddingModel = value
	}
}

func WithModel(value string) Option {
	return func(o *options) {
		o.model = value
	}
}

func SensitiveTypeToValue(code int64) string {
	re := ""
	switch code {
	case 1:
		re = "严重违规"
	case 2:
		re = "色情"
	case 3:
		re = "广告"
	case 4:
		re = "违禁"
	case 5:
		re = "谩骂"
	case 6:
		re = "暴恐"
	case 7:
		re = "其他"
	}
	return re
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
