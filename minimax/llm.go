package minimax

import (
	"errors"
	"github.com/comqositi/kpllms/minimax/minimaxclientv1"
	"net/http"
	"os"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingToken             = errors.New("缺少GROUP ID 或者 API KEY") //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

func newClient(opts ...Option) (*minimaxclientv1.Client, error) {
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

	return minimaxclientv1.NewClient(minimaxclientv1.WithGroupId(options.groupId),
		minimaxclientv1.WithApiKey(options.apiKey),
		minimaxclientv1.WithBaseUrl(options.baseUrl),
		minimaxclientv1.WithHttpClient(options.httpClient),
		minimaxclientv1.WithModel(options.model),
		minimaxclientv1.WithEmbeddingsModel(options.embeddingModel),
	)

}
