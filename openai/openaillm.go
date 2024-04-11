package openai

import (
	"context"
	"errors"
	"fmt"
	"github.com/comqositi/kpllms"
	"github.com/comqositi/kpllms/schema"
	"net/http"
	"os"

	"github.com/comqositi/kpllms/openai/internal/openaiclient"
)

type LLM struct {
	client *openaiclient.Client
}

const (
	OpenaiRoleSystem    = "system"
	OpenaiRoleAssistant = "assistant"
	OpenaiRoleUser      = "user"
	OpenaiRoleTool      = "tool"
)

var (
	_                             kpllms.Model = (*LLM)(nil)
	ErrEmptyResponse                           = errors.New("no response")
	ErrMissingToken                            = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable") //nolint:lll
	ErrMissingAzureModel                       = errors.New("model needs to be provided when using Azure API")
	ErrMissingAzureEmbeddingModel              = errors.New("embeddings model needs to be provided when using Azure API")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

// New 创建大模型 model 的实现
func New(opts ...Option) (*LLM, error) {
	_, c, err := newClient(opts...)
	if err != nil {
		return nil, err
	}
	return &LLM{
		client: c,
	}, err
}

// newClient 创建 openai 客户端
func newClient(opts ...Option) (*options, *openaiclient.Client, error) {
	options := &options{
		token:        os.Getenv(tokenEnvVarName),
		model:        os.Getenv(modelEnvVarName),
		baseURL:      getEnvs(baseURLEnvVarName, baseAPIBaseEnvVarName),
		organization: os.Getenv(organizationEnvVarName),
		apiType:      APIType(openaiclient.APITypeOpenAI),
		httpClient:   http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	// set of options needed for Azure client
	if openaiclient.IsAzure(openaiclient.APIType(options.apiType)) && options.apiVersion == "" {
		options.apiVersion = DefaultAPIVersion
		if options.model == "" {
			return options, nil, ErrMissingAzureModel
		}
		if options.embeddingModel == "" {
			return options, nil, ErrMissingAzureEmbeddingModel
		}
	}

	if len(options.token) == 0 {
		return options, nil, ErrMissingToken
	}

	cli, err := openaiclient.New(options.token, options.model, options.baseURL, options.organization,
		openaiclient.APIType(options.apiType), options.apiVersion, options.httpClient, options.embeddingModel)
	return options, cli, err
}

func getEnvs(keys ...string) string {
	for _, key := range keys {
		val, ok := os.LookupEnv(key)
		if ok {
			return val
		}
	}
	return ""
}

// Chat 实现大模型接口
func (o *LLM) Chat(ctx context.Context, messages []*schema.ChatMessage, options ...kpllms.CallOption) (*schema.ContentResponse, error) {

	opts := kpllms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	chatMsgs := make([]*openaiclient.ChatMessage, 0, len(messages))
	for _, mc := range messages {
		msg := &openaiclient.ChatMessage{
			Role:       "",
			Content:    nil,
			Name:       "",
			ToolCalls:  nil,
			ToolCallID: "",
		}
		msg.Name = mc.Name
		switch mc.Role {
		case schema.RoleSystem:
			msg.Role = OpenaiRoleSystem
			msg.Content = mc.Content
		case schema.RoleAssistant:
			msg.Role = OpenaiRoleAssistant
			msg.Content = mc.Content
			// 如果模型回复的是函数
			if len(mc.ToolCalls) > 0 {
				for _, t := range mc.ToolCalls {
					msg.ToolCalls = append(msg.ToolCalls, openaiclient.ToolCall{
						Index: 0,
						ID:    t.Id,
						Type:  openaiclient.ToolType(t.Type),
						Function: openaiclient.ToolFunction{
							Name:      t.Function.Name,
							Arguments: t.Function.Arguments,
						},
					})
				}
			}
		case schema.RoleUser:
			msg.Role = OpenaiRoleUser
			msg.Content = mc.Content
		case schema.RoleTool:
			msg.Role = OpenaiRoleTool
			msg.ToolCallID = mc.ToolCallId
			msg.Content = mc.Content
		default:
			return nil, fmt.Errorf("role %v not supported", mc.Role)
		}
		chatMsgs = append(chatMsgs, msg)
	}

	req := &openaiclient.ChatRequest{
		Model:         opts.Model,
		Messages:      chatMsgs,
		StreamingFunc: opts.StreamingFunc,
		Temperature:   opts.Temperature,
		MaxTokens:     opts.MaxTokens,
		TopP:          opts.TopP,
		ToolChoice:    opts.ToolChoice,
	}
	if opts.ResponseFormat != nil {
		req.ResponseFormat = ResponseFormatJSON
	}

	for _, tool := range opts.Tools {
		t, err := toolFromTool(tool)
		if err != nil {
			return nil, fmt.Errorf("failed to convert llms tool to openai tool: %w", err)
		}
		req.Tools = append(req.Tools, t)
	}

	result, err := o.client.CreateChat(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	choices := make([]*schema.ContentChoice, len(result.Choices))
	//ss, _ := json.Marshal(result)
	//fmt.Println(string(ss))
	for i, c := range result.Choices {
		choices[i] = &schema.ContentChoice{

			Content:    c.Message.Content,
			StopReason: fmt.Sprint(c.FinishReason),
			Usage: &schema.Usage{
				PromptTokens:     result.Usage.CompletionTokens,
				CompletionTokens: result.Usage.PromptTokens,
				TotalTokens:      result.Usage.TotalTokens,
			},
		}

		if c.FinishReason == "tool_calls" {
			// TODO: we can only handle a single tool call for now, we need to evolve the API to handle multiple tool calls.
			for _, tool := range c.Message.ToolCalls {
				choices[i].ToolCalls = append(choices[i].ToolCalls, &schema.ToolCall{
					Id:   tool.ID,
					Type: string(tool.Type),
					Function: schema.FunctionCall{
						Name:      tool.Function.Name,
						Arguments: tool.Function.Arguments,
					},
				})
			}
			//// populate legacy single-function call field for backwards compatibility
			//if len(choices[i].ToolCalls) > 0 {
			//	choices[i].FuncCall = choices[i].ToolCalls[0].FunctionCall
			//}
		}
	}
	response := &schema.ContentResponse{Choices: choices}

	return response, nil

}

func toolFromTool(t *kpllms.Tool) (openaiclient.Tool, error) {
	tool := openaiclient.Tool{
		Type: openaiclient.ToolType(t.Type),
	}
	switch t.Type {
	case string(openaiclient.ToolTypeFunction):
		tool.Function = openaiclient.FunctionDefinition{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			Parameters:  t.Function.Parameters,
		}
	default:
		return openaiclient.Tool{}, fmt.Errorf("tool type %v not supported", t.Type)
	}
	return tool, nil
}
