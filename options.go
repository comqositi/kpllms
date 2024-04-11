package kpllms

import "context"

/*
定义大模型调用公共参数
*/
type CallOption func(*CallOptions)

type CallOptions struct {
	// 模型代号， 例如：gpt-4
	Model string `json:"model"`
	// 最大输出 token 数
	MaxTokens int `json:"max_tokens"`
	// 温度 0-2
	Temperature float64 `json:"temperature"`
	// 流式输出
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
	// 采样率 0.1 = 10%
	TopP float64 `json:"top_p"`
	// 返回格式，例如：{ "type": "json_object" }
	ResponseFormat *ResponseFormat `json:"response_format"`
	// 函数定义
	Tools []*Tool `json:"tools,omitempty"`
	// 函数调用方式  auto， 指定：{"type":"","function":""}
	ToolChoice any `json:"tool_choice"`
}

type ResponseFormat struct {
	//  text 或者 json_object.
	Type string `json:"type"`
}

type Tool struct {
	Type     string              `json:"type"`
	Function *FunctionDefinition `json:"function,omitempty"`
}

type FunctionDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters,omitempty"`
}

type ToolChoice struct {
	Type     string             `json:"type"`
	Function *FunctionReference `json:"function,omitempty"`
}

type FunctionReference struct {
	Name string `json:"name"`
}

func WithModel(model string) CallOption {
	return func(o *CallOptions) {
		o.Model = model
	}
}

func WithMaxTokens(maxTokens int) CallOption {
	return func(o *CallOptions) {
		o.MaxTokens = maxTokens
	}
}

func WithTemperature(temperature float64) CallOption {
	return func(o *CallOptions) {
		o.Temperature = temperature
	}
}

func WithStreamingFunc(streamingFunc func(ctx context.Context, chunk []byte) error) CallOption {
	return func(o *CallOptions) {
		o.StreamingFunc = streamingFunc
	}
}

func WithTopP(topP float64) CallOption {
	return func(o *CallOptions) {
		o.TopP = topP
	}
}

func WithResponseFormat(format *ResponseFormat) CallOption {
	return func(o *CallOptions) {
		o.ResponseFormat = format
	}
}

func WithTools(tools []*Tool) CallOption {
	return func(o *CallOptions) {
		o.Tools = tools
	}
}

func WithToolChoice(choice any) CallOption {
	return func(o *CallOptions) {
		o.ToolChoice = choice
	}
}
