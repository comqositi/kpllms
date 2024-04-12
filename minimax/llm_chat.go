package minimax

import (
	"context"
	"errors"
	"fmt"
	"github.com/comqositi/kpllms"
	"github.com/comqositi/kpllms/minimax/minimaxclientv1"
	"github.com/comqositi/kpllms/schema"
)

type Chat struct {
	client    *minimaxclientv1.Client
	usage     []minimaxclientv1.Usage
	chatError error // 每次模型调用的错误信息
}

const (
	RoleAssistant = "assistant"
	RoleUser      = "user"
)

var (
	_ kpllms.Model = (*Chat)(nil)
)

// NewChat returns a new OpenAI chat LLM.
func NewChat(opts ...Option) (*Chat, error) {
	c, err := newClient(opts...)
	return &Chat{
		client: c,
	}, err
}

func (o *Chat) Chat(ctx context.Context, messageSets []*schema.ChatMessage, options ...kpllms.CallOption) (*schema.ContentResponse, error) {
	opts := kpllms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	clientMsg, setting, reply := messagesToClientMessages(messageSets)
	req := &minimaxclientv1.CompletionRequest{
		Model:            opts.Model,
		Messages:         clientMsg,
		TokensToGenerate: int64(opts.MaxTokens),
		Temperature:      float32(opts.Temperature),
		TopP:             float32(opts.TopP),
		BotSetting:       []minimaxclientv1.BotSetting{setting},
		ReplyConstraints: reply,
		//RequestId : opts.RequestId,
		StreamingFunc:     opts.StreamingFunc,
		Stream:            opts.StreamingFunc != nil,
		MaskSensitiveInfo: false, // 对输出中易涉及隐私问题的文本信息进行打码，目前包括但不限于邮箱、域名、链接、证件号、家庭住址等，默认true，即开启打码
		//FunctionCallSetting   自动模式等
	}

	for _, tool := range opts.Tools {
		t, err := toolFromTool(tool)
		if err != nil {
			return nil, fmt.Errorf("failed to convert llms tool to openai tool: %w", err)
		}
		req.Functions = append(req.Functions, t)
	}

	// opts.ToolChoice 默认自动调用

	result, err := o.client.CreateCompletion(ctx, req)
	if err != nil {
		return nil, err
	}
	if result.InputSensitive {
		return nil, errors.New(fmt.Sprintf("输入命中敏感词：%s", SensitiveTypeToValue(result.InputSensitiveType)))
	}
	if result.OutputSensitive {
		return nil, errors.New(fmt.Sprintf("输出命中敏感词：%s", SensitiveTypeToValue(result.OutputSensitiveType)))
	}
	if result.BaseResp.StatusCode == 0 && len(result.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	resp := &schema.ContentResponse{Choices: make([]*schema.ContentChoice, 1)}
	resp.Choices[0].Usage.PromptTokens = int(result.Usage.PromptTokens)
	resp.Choices[0].Usage.CompletionTokens = int(result.Usage.CompletionTokens)
	resp.Choices[0].Usage.TotalTokens = int(result.Usage.TotalTokens)
	resp.Choices[0].Content = result.Choices[0].Messages[0].Text
	resp.Choices[0].StopReason = result.Choices[0].FinishReason

	if result.Choices[0].Messages[0].FunctionCall != nil {
		resp.Choices[0].ToolCalls = []*schema.ToolCall{
			{
				Type: schema.ToolCallTypeFunction,
				Function: schema.FunctionCall{
					Name:      result.Choices[0].Messages[0].FunctionCall.Name,
					Arguments: result.Choices[0].Messages[0].FunctionCall.Arguments,
				},
			},
		}
	}

	return resp, nil

}
func toolFromTool(t *kpllms.Tool) (*minimaxclientv1.FunctionDefinition, error) {

	tool := &minimaxclientv1.FunctionDefinition{
		Name:        t.Function.Name,
		Description: t.Function.Description,
		Parameters:  t.Function.Parameters,
	}
	return tool, nil
}

func messagesToClientMessages(messages []*schema.ChatMessage) ([]*minimaxclientv1.Message, minimaxclientv1.BotSetting, minimaxclientv1.ReplyConstraints) {

	setting := minimaxclientv1.BotSetting{
		BotName: defaultBotName,
		Content: defaultBotDescription,
	}
	replyConstraints := minimaxclientv1.ReplyConstraints{
		SenderType: defaultSendType,
		SenderName: defaultBotName,
	}
	msglen := len(messages)
	// 第一个system信息放入bot_setting
	if len(messages) > 0 {
		if messages[0].Role == schema.RoleSystem {
			systemContent := messages[0].Content.(string)
			setting.Content = systemContent
			messages = messages[1:]
			msglen -= 1
		}
	}

	msgs := make([]*minimaxclientv1.Message, msglen)
	for i, m := range messages {
		typ := m.Role
		msg := &minimaxclientv1.Message{}
		// 如果是字符串，先赋值
		if content, ok := m.Content.(string); ok {
			msg.Text = content
		}

		switch typ {
		// ai 回答，可能是文本答案，可能是 function
		case schema.RoleAssistant:
			msg.SenderType = "BOT"
			msg.SenderName = defaultBotName
			if m.ToolCalls != nil {
				msg.FunctionCall = toolToFunction(m.ToolCalls)
			}
		case schema.RoleUser:
			msg.SenderType = "USER"
			msg.SenderName = defaultSendName
		case schema.RoleTool:
			msg.SenderType = "FUNCTION"
			msg.SenderName = defaultSendName

			//case schema.ChatMessageTypeFunction:
			//	msg.Role = "function"
		}
		msgs[i] = msg
	}

	return msgs, setting, replyConstraints
}

func toolToFunction(call []*schema.ToolCall) *minimaxclientv1.FunctionCall {
	return &minimaxclientv1.FunctionCall{
		Name:      call[0].Function.Name,
		Arguments: call[0].Function.Arguments,
	}
}

// EmbedDocuments  实现 embedder 接口 文档存储向量：存入数据库的向量，被用于检索
func (o *Chat) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {

	result, err := o.client.CreateEmbedding(ctx, &minimaxclientv1.EmbeddingPayload{
		Texts: texts,
		Type:  "db", //db query
	})
	if err != nil {
		return nil, err
	}
	o.usage = append(o.usage, minimaxclientv1.Usage{TotalTokens: result.TotalTokens})
	return result.Vectors, nil
}

// EmbedQuery  实现 embedder 接口 查询向量：对需要用于检索的文本进行想量化
func (o *Chat) EmbedQuery(ctx context.Context, text string) ([]float32, error) {

	result, err := o.client.CreateEmbedding(ctx, &minimaxclientv1.EmbeddingPayload{
		Texts: []string{text},
		Type:  "query", //db query
	})
	if err != nil {
		return nil, err
	}
	o.usage = append(o.usage, minimaxclientv1.Usage{TotalTokens: result.TotalTokens})
	return result.Vectors[0], nil

}
