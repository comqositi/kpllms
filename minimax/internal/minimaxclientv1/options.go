package minimaxclientv1

import (
	"context"
	"errors"
	"net/http"
)

const (
	defaultBaseUrl = "https://api.minimax.chat/v1"
)

var (
	ErrNotSetAuth      = errors.New("both accessToken and apiKey secretKey are not set")
	ErrCompletionCode  = errors.New("completion API returned unexpected status code")
	ErrAccessTokenCode = errors.New("get access_token API returned unexpected status code")
	ErrEmbeddingCode   = errors.New("embedding API returned unexpected status code")
)

// Option is an option for the ERNIE client.
type Option func(*Client) error

// Doer performs a HTTP request.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

func WithGroupId(value string) Option {
	return func(c *Client) error {
		c.groupId = value
		return nil
	}
}

func WithApiKey(value string) Option {
	return func(c *Client) error {
		c.apiKey = value
		return nil
	}
}

func WithBaseUrl(value string) Option {
	return func(c *Client) error {
		c.baseUrl = value
		return nil
	}
}

func WithHttpClient(value Doer) Option {
	return func(c *Client) error {
		c.httpClient = value
		return nil
	}
}

func WithModel(value string) Option {
	return func(c *Client) error {
		c.model = value
		return nil
	}
}

func WithEmbeddingsModel(value string) Option {
	return func(c *Client) error {
		c.embeddingsModel = value
		return nil
	}
}

type CompletionRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream,omitempty"`
	// 最大输出token
	TokensToGenerate int64   `json:"tokens_to_generate,omitempty"`
	Temperature      float32 `json:"temperature,omitempty"`
	TopP             float32 `json:"top_p,omitempty"`
	// 对输出中易涉及隐私问题的文本信息进行打码，目前包括但不限于邮箱、域名、链接、证件号、家庭住址等，默认true，即开启打码
	MaskSensitiveInfo bool             `json:"mask_sensitive_info,omitempty"`
	Messages          []*Message       `json:"messages"`          //长度影响接口性能
	BotSetting        []BotSetting     `json:"bot_setting"`       //对每一个机器人的设定
	ReplyConstraints  ReplyConstraints `json:"reply_constraints"` //模型回复要求

	StreamingFunc func(ctx context.Context, chunk []byte, err error) error `json:"-"`

	SampleMessages []*SampleMessage `json:"sample_messages,omitempty"`

	Functions []*FunctionDefinition `json:"functions,omitempty"`

	FunctionCallSetting *FunctionCallSetting `json:"function_call,omitempty"` // function call setting

	Plugins []string `json:"plugins,omitempty"` // ["plugin_web_search"] 目前支持：plugin_web_search；会增加除token以外的计费，按照search次数计费，单价为0.03元/次，请注意：因大模型有自动判断当次搜索结果是否准确的能力，所以1次请求可能会自动触发多次搜索引擎调用
}

type Message struct {
	SenderType string `json:"sender_type"` //需要为以下三个合法值之一USER：用户发送的内容，BOT：模型生成的内容，FUNCTION：详见下文中的函数调用部分
	SenderName string `json:"sender_name"` // 必填
	Text       string `json:"text"`
	// FunctionCall represents a function call to be made in the message.
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type BotSetting struct {
	BotName string `json:"bot_name"` // 具体机器人的名字
	Content string `json:"content"`  // 具体机器人的设定, 长度影响接口性能
}

type ReplyConstraints struct {
	SenderType string `json:"sender_type,omitempty"` // 指定回复的角色类型,当前只支持 BOT机器人
	SenderName string `json:"sender_name,omitempty"` //指定回复的机器人名称
	Glyph      *Glyph `json:"glyph,omitempty"`       // 限制返回格式功能（glyph）
}

type SampleMessage struct {
	SenderType string `json:"sender_type,omitempty"`
	SenderName string `json:"sender_name,omitempty"`
	Text       string `json:"text,omitempty"`
}

type FunctionDefinition struct {
	// Name is the name of the function.
	Name string `json:"name"`
	// Description is a description of the function.
	Description string `json:"description,omitempty"`
	// Parameters is a list of parameters for the function.
	Parameters any `json:"parameters"`
}

// 限制的返回格式配置
type Glyph struct {
	Type           string           `json:"type,omitempty"`            //使用什么模板功能当前仅支持1、raw2、json_value
	RawGlyph       string           `json:"raw_glyph,omitempty"`       // type: raw 需要限制的格式要求，使用 glpyh 语法.   这句话的翻译是：{{gen 'content'}}
	JsonProperties map[string]any   `json:"json_properties,omitempty"` // type:json_value
	PropertyList   []map[string]any `json:"property_list,omitempty"`   // 当需要生成数据的key和输入顺序一致时，需要使用 property_list
}

type FunctionCallSetting struct {
	Type string `json:"type,omitempty"` // functions 使用模式，有三个枚举值可选 auto,specific,none
	Name string `json:"name,omitempty"` // 通过该字段指定要使用的 function，仅当 function_call.type = specific 时生效
}

type Completion struct {
	Created             int64         `json:"created,omitempty"`               //请求发起时间
	Model               string        `json:"model,omitempty"`                 //请求指定的模型名称
	Reply               string        `json:"reply,omitempty"`                 //回复内容
	InputSensitive      bool          `json:"input_sensitive,omitempty"`       //输入命中敏感词
	InputSensitiveType  int64         `json:"input_sensitive_type,omitempty"`  //输入命中敏感词类型，当input_sensitive为true时返回  取值为以下其一 1 严重违规 2 色情 3 广告 4 违禁 5 谩骂 6 暴恐 7 其他
	OutputSensitive     bool          `json:"output_sensitive,omitempty"`      // 输出命中敏感词
	OutputSensitiveType int64         `json:"output_sensitive_type,omitempty"` //输出命中敏感词类型，当output_sensitive为true时返回,  取值同input_sensitive_type
	Choices             []Choice      `json:"choices,omitempty"`               // 所有结果
	Usage               Usage         `json:"usage,omitempty"`                 // 流式场景下，增量数据包不含该字段；全量（最后一个）数据包含有该字段
	Id                  string        `json:"id,omitempty"`                    //
	BaseResp            BaseResp      `json:"base_resp,omitempty"`             // 错误状态码和详情
	FunctionCall        *FunctionCall `json:"function_call,omitempty"`         // 调用的函数
}

type Choice struct {
	Messages     []Message   `json:"messages,omitempty"`      //回复结果的具体内容
	Index        int64       `json:"index,omitempty"`         //
	FinishReason string      `json:"finish_reason,omitempty"` //
	GlyphResult  GlyphResult `json:"glyph_result,omitempty"`  // glyph的结果
}

type GlyphResult struct {
	Content string `json:"content,omitempty"`
}

type Usage struct {
	CompletionTokens      int64 `json:"completion_tokens,omitempty"`
	PromptTokens          int64 `json:"prompt_tokens,omitempty"`
	TotalTokens           int64 `json:"total_tokens,omitempty"`             //消耗tokens总数，包括输入和输出,
	TokensWithAddedPlugin int64 `json:"tokens_with_added_plugin,omitempty"` // 调用搜多插件，额外消耗的token
}

type BaseResp struct {
	StatusCode int64  `json:"status_code,omitempty"` //状态码 1000，未知错误 1001，超时 1002，触发RPM限流 1004，鉴权失败 1008，余额不足 1013，服务内部错误 1027，输出内容错误 1039，触发TPM限流 2013，输入格式信息不正常
	StatusMsg  string `json:"status_msg,omitempty"`  //
}
