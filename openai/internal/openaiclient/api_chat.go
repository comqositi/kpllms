package openaiclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/comqositi/kpllms/internal/httputils"
	"strings"
)

const (
	defaultChatModel = "gpt-3.5-turbo"
)

// openai chat接口实现

// ChatRequest chat接口请求， 请求参数参考文档：https://platform.openai.com/docs/api-reference/chat
type ChatRequest struct {
	// 模型
	Model string `json:"model"`
	// 上下文和用户提问
	Messages []*ChatMessage `json:"messages"`
	// 调整温度
	Temperature      float64  `json:"temperature"`
	TopP             float64  `json:"top_p,omitempty"`
	MaxTokens        int      `json:"max_tokens,omitempty"`
	N                int      `json:"n,omitempty"`
	StopWords        []string `json:"stop,omitempty"`
	Stream           bool     `json:"stream,omitempty"`
	FrequencyPenalty float64  `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64  `json:"presence_penalty,omitempty"`
	Seed             int      `json:"seed,omitempty"`

	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
	LogProbs       bool            `json:"logprobs,omitempty"`

	TopLogProbs int `json:"top_logprobs,omitempty"`

	Tools []Tool `json:"tools,omitempty"`

	// 指定工具调用的方式，string 或者 ToolChoice， 例如：auto，自动调用，指定调用
	ToolChoice any `json:"tool_choice,omitempty"`

	// 流式放回的回调函数
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

type Tool struct {
	Type     ToolType           `json:"type"`
	Function FunctionDefinition `json:"function,omitempty"`
}

type ToolChoice struct {
	Type     ToolType     `json:"type"`
	Function ToolFunction `json:"function,omitempty"`
}

type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCall is a call to a tool.
type ToolCall struct {
	Index    int          `json:"index,omitempty"`
	ID       string       `json:"id,omitempty"`
	Type     ToolType     `json:"type"`
	Function ToolFunction `json:"function,omitempty"`
}

// ResponseFormat is the format of the response.
type ResponseFormat struct {
	Type string `json:"type"`
}

// ChatMessage is a message in a chat request.
type ChatMessage struct { //nolint:musttag
	Role    string `json:"role"`
	Content any    `json:"content,omitempty"`
	Name    string `json:"name,omitempty"`

	// 函数列表
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID 是本次函数调用的 id
	// 只在 role 为 tool 时才有
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ChatMessageResponse is a message in a chat request.
type ChatMessageResponse struct { //nolint:musttag
	Role    string `json:"role"`
	Content string `json:"content,omitempty"`
	Name    string `json:"name,omitempty"`

	// 函数列表
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID 是本次函数调用的 id
	// 只在 role 为 tool 时才有
	ToolCallID string `json:"tool_call_id,omitempty"`
}

type FinishReason string

const (
	// 正常停止
	FinishReasonStop FinishReason = "stop"
	// 达到maxtoken 的长度
	FinishReasonLength FinishReason = "length"
	// 调用函数
	FinishReasonToolCalls FinishReason = "tool_calls"
	// 内容过滤器导致结束
	FinishReasonContentFilter FinishReason = "content_filter"
	// 未结束
	FinishReasonNull FinishReason = "null"
)

func (r FinishReason) MarshalJSON() ([]byte, error) {
	if r == FinishReasonNull || r == "" {
		return []byte("null"), nil
	}
	return []byte(`"` + string(r) + `"`), nil // best effort to not break future API changes
}

// ChatCompletionChoice is a choice in a chat response.
type ChatCompletionChoice struct {
	Index        int                 `json:"index"`
	Message      ChatMessageResponse `json:"message"`
	FinishReason FinishReason        `json:"finish_reason"`
}

// ChatUsage is the usage of a chat completion request.
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse is a response to a chat request.
type ChatCompletionResponse struct {
	ID                string                  `json:"id,omitempty"`
	Created           int64                   `json:"created,omitempty"`
	Choices           []*ChatCompletionChoice `json:"choices,omitempty"`
	Model             string                  `json:"model,omitempty"`
	Object            string                  `json:"object,omitempty"`
	Usage             ChatUsage               `json:"usage,omitempty"`
	SystemFingerprint string                  `json:"system_fingerprint"`
}

// StreamedChatResponsePayload is a chunk from the stream.
type StreamedChatResponsePayload struct {
	ID      string  `json:"id,omitempty"`
	Created float64 `json:"created,omitempty"`
	Model   string  `json:"model,omitempty"`
	Object  string  `json:"object,omitempty"`
	Choices []struct {
		Index float64 `json:"index,omitempty"`
		Delta struct {
			Role      string     `json:"role,omitempty"`
			Content   string     `json:"content,omitempty"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"delta,omitempty"`
		FinishReason FinishReason `json:"finish_reason,omitempty"`
	} `json:"choices,omitempty"`
}

// FunctionDefinition is a definition of a function that can be called by the model.
type FunctionDefinition struct {
	// Name is the name of the function.
	Name string `json:"name"`
	// Description is a description of the function.
	Description string `json:"description,omitempty"`
	// Parameters is a list of parameters for the function.
	Parameters any `json:"parameters"`
}

// FunctionCallBehavior is the behavior to use when calling functions.
type FunctionCallBehavior string

const (
	// FunctionCallBehaviorUnspecified is the empty string.
	FunctionCallBehaviorUnspecified FunctionCallBehavior = ""
	// FunctionCallBehaviorNone will not call any functions.
	FunctionCallBehaviorNone FunctionCallBehavior = "none"
	// FunctionCallBehaviorAuto will call functions automatically.
	FunctionCallBehaviorAuto FunctionCallBehavior = "auto"
)

// FunctionCall is a call to a function.
type FunctionCall struct {
	// Name is the name of the function to call.
	Name string `json:"name"`
	// Arguments is the set of arguments to pass to the function.
	Arguments string `json:"arguments"`
}

func (c *Client) createChat(ctx context.Context, payload *ChatRequest) (*ChatCompletionResponse, error) {
	if payload.StreamingFunc != nil {
		payload.Stream = true
	}
	var response ChatCompletionResponse
	// 处理流式返回
	if payload.StreamingFunc != nil {
		// 流式返回初始化一下， 避免赋值时报空指针
		response.Choices = []*ChatCompletionChoice{
			{},
		}
		err := httputils.HttpStream(ctx, c.buildURL("/chat/completions", c.Model), payload, c.setHeaders(), func(ctx context.Context, line string) error {
			//fmt.Println(line)
			// func 内会返回流式返回的每行数据，对每行数据逐行处理
			if line == "" {
				// 空行不处理，空行是数据间隔行
				return nil
			}
			if !strings.HasPrefix(line, "data:") {
				return errors.New(fmt.Sprintf("unexpected line: %v", line))
			}
			data := strings.TrimPrefix(line, "data: ")
			// 传输结束
			if data == "[DONE]" {
				return nil
			}
			// 解析有用的数据
			var streamResponse StreamedChatResponsePayload
			err := json.Unmarshal([]byte(data), &streamResponse)
			if err != nil {
				return err
			}
			if len(streamResponse.Choices) == 0 {
				return nil
			}
			// 如果是非函数调用
			if streamResponse.Choices[0].Delta.ToolCalls == nil {
				// 非函数调用
				chunk := []byte(streamResponse.Choices[0].Delta.Content)
				// 拼接所有内容
				response.Choices[0].Message.Content += streamResponse.Choices[0].Delta.Content
				// 写入最后一个结束标识
				response.Choices[0].FinishReason = streamResponse.Choices[0].FinishReason
				// 调用用户 func
				return payload.StreamingFunc(ctx, []byte(chunk))
			}

			// 如果是函数调用， 遇到 type=function 加入一个函数,  openai有并行返回函数的功能
			// 返回第几个函数
			toolCallIndex := streamResponse.Choices[0].Delta.ToolCalls[0].Index
			if streamResponse.Choices[0].Delta.ToolCalls[0].Type == "function" {
				response.Choices[0].Message.ToolCalls = append(response.Choices[0].Message.ToolCalls, ToolCall{})
				response.Choices[0].Message.ToolCalls[toolCallIndex].Index = toolCallIndex
				response.Choices[0].Message.ToolCalls[toolCallIndex].ID = streamResponse.Choices[0].Delta.ToolCalls[0].ID
				response.Choices[0].Message.ToolCalls[toolCallIndex].Type = streamResponse.Choices[0].Delta.ToolCalls[0].Type
				response.Choices[0].Message.ToolCalls[toolCallIndex].Function.Name = streamResponse.Choices[0].Delta.ToolCalls[0].Function.Name
			}
			response.Choices[0].Message.ToolCalls[toolCallIndex].Function.Arguments += streamResponse.Choices[0].Delta.ToolCalls[0].Function.Arguments

			// 如果是 function则无需 stream 流式返回，避免输出错误
			// 写入最后一个结束标识
			response.Choices[0].FinishReason = streamResponse.Choices[0].FinishReason
			return nil
		})
		if err != nil {
			return nil, err
		}
		// TODO openai stream 模式没有返回消耗的 token，此处自己计算
		response.Usage = ChatUsage{
			PromptTokens:     100,
			CompletionTokens: 100,
			TotalTokens:      200,
		}

	} else {
		// 处理非流式返回
		err := httputils.HttpPost(ctx, c.buildURL("/chat/completions", c.Model), payload, c.setHeaders(), &response)
		if err != nil {
			return nil, err
		}
	}
	return &response, nil
}
