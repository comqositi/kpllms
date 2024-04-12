package schema

const (
	// system 人格设定
	RoleSystem = "system"
	// 用户提问
	RoleUser = "user"
	// 大模型回复
	RoleAssistant = "assistant"
	// 工具调用的结果
	RoleTool = "tool"

	// 输入内容类型， 文本
	MultiContentText = "text"
	// 输入内容类型，image_url
	MultiContentImageUrl = "image_url"

	// json mod 格式
	JsonMode = "json_object"

	ToolCallTypeFunction = "function"
)

// ChatMessage 以 openai 的接口为标准定义统一参数
type ChatMessage struct {
	// 必填字段 system  user  assistant  tool
	Role string `json:"role"`
	// 可选字段
	Name string `json:"name,omitempty"`
	// role： system、user、assistant  ，Content: string或者是数组。 Content/ToolCalls/ToolCallId三选其一
	Content any `json:"content,omitempty"`
	// role：assistant 返回调用了工具时才有此字段
	ToolCalls []*ToolCall `json:"tool_calls,omitempty"`
	// role：tool 才有， 携带函数调用结果
	ToolCallId string `json:"tool_call_id,omitempty"`
}

type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ImageContent struct {
	Type     string `json:"type"`
	ImageUrl string `json:"image_url"`
}

type ToolCall struct {
	Id       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolMessage struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	ToolCallId string `json:"tool_call_id"`
}

// 大模型 response
type ContentResponse struct {
	Choices []*ContentChoice
}

type ContentChoice struct {
	Content string

	StopReason string

	GenerationInfo map[string]any

	ToolCalls []*ToolCall

	Usage *Usage
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
