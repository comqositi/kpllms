package minimax

import (
	"errors"
	minimaxclientv12 "github.com/comqositi/kpllms/minimax/internal/minimaxclientv1"
	"net/http"
	"os"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingToken             = errors.New("缺少GROUP ID 或者 API KEY") //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

func newClient(opts ...Option) (*minimaxclientv12.Client, error) {
	options := &options{
		groupId:        os.Getenv(groupIdEnvVarName),
		apiKey:         os.Getenv(apiKeyEnvVarName),
		baseUrl:        os.Getenv(baseURLEnvVarName),
		httpClient:     http.DefaultClient,
		embeddingModel: "",
		model:          "",
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.model == "" {
		options.model = defaultModel
	}

	if options.embeddingModel == "" {
		options.embeddingModel = defaultEmbeddingModel
	}

	return minimaxclientv12.NewClient(minimaxclientv12.WithGroupId(options.groupId),
		minimaxclientv12.WithApiKey(options.apiKey),
		minimaxclientv12.WithBaseUrl(options.baseUrl),
		minimaxclientv12.WithHttpClient(options.httpClient),
		minimaxclientv12.WithModel(options.model),
		minimaxclientv12.WithEmbeddingsModel(options.embeddingModel),
	)

}
