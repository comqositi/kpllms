package kpllms

import "context"

/*
定义大模型调用公共参数
*/
type CallOption func(*CallOptions)

type CallOptions struct {
	// 模型代号， 例如：gpt-4
	Model string
	// 最大输出 token 数
	MaxTokens int
	// 温度 0-2
	Temperature float64
	// 流式输出
	StreamingFunc func(ctx context.Context, chunk []byte, innerErr error) error
	// 采样率 0.1 = 10%
	TopP float64
	/// 是否严格要求返回 json 格式, true: 强制 json 格式返回
	JsonMode bool
	// 函数定义
	Tools []*Tool
	// 函数调用方式  auto， none  指定：{"type":"auto/none/function","function":}, none: 不调用，auto：自动调用，默认是自动调用， functon，指定调用
	ToolChoice ToolChoice
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
	Function ToolChoiceFunction `json:"function,omitempty"`
}

type ToolChoiceFunction struct {
	Name string
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

func WithStreamingFunc(streamingFunc func(ctx context.Context, chunk []byte, innerErr error) error) CallOption {
	return func(o *CallOptions) {
		o.StreamingFunc = streamingFunc
	}
}

func WithTopP(topP float64) CallOption {
	return func(o *CallOptions) {
		o.TopP = topP
	}
}

func WithJsonMode(jsonMode bool) CallOption {
	return func(o *CallOptions) {
		o.JsonMode = jsonMode
	}
}

func WithTools(tools []*Tool) CallOption {
	return func(o *CallOptions) {
		o.Tools = tools
	}
}

func WithToolChoice(choice ToolChoice) CallOption {
	return func(o *CallOptions) {
		o.ToolChoice = choice
	}
}
