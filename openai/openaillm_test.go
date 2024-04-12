package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/comqositi/kpllms"
	"github.com/comqositi/kpllms/schema"
	"os"
	"testing"
)

var baseUrl = "https://apiagent.kaopuai.com/v1"
var token = os.Getenv("OPENAI_API_KEY")

// 测试 chat 调用
func TestLLM_Chat(t *testing.T) {
	ctx := context.Background()
	llm, err := New(WithToken(token), WithBaseURL(baseUrl))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	messages := []*schema.ChatMessage{
		&schema.ChatMessage{
			Role:    schema.RoleSystem,
			Content: "你是一个 AI 助手",
		},
		&schema.ChatMessage{
			Role:    schema.RoleUser,
			Content: "查询一下北京的天气？",
		},
	}
	resp, err := llm.Chat(ctx, messages)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	b, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(b))
}

// 测试 stream 返回
func TestLLM_Stream(t *testing.T) {
	ctx := context.Background()
	llm, err := New(WithToken(token), WithBaseURL(baseUrl))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	messages := []*schema.ChatMessage{
		&schema.ChatMessage{
			Role:    schema.RoleSystem,
			Content: "你是一个 AI 助手",
		},
		&schema.ChatMessage{
			Role:    schema.RoleUser,
			Content: "查询一下北京的天气？",
		},
	}
	resp, err := llm.Chat(ctx, messages,
		kpllms.WithStreamingFunc(func(ctx context.Context, chunk []byte, innerErr error) error {
			fmt.Println(string(chunk))
			return nil
		}))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	b, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(b))
}

// 测试函数调用
func TestLLM_Function_Call(t *testing.T) {
	modelName := "gpt-4-1106-preview"
	ctx := context.Background()
	llm, err := New(WithToken(token), WithBaseURL(baseUrl), WithModel(modelName))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	messages := []*schema.ChatMessage{
		&schema.ChatMessage{
			Role:    schema.RoleSystem,
			Content: "你是一个 AI 助手",
		},
		&schema.ChatMessage{
			Role:    schema.RoleUser,
			Content: "查询一下北京的天气？查询一下北京靠谱前程网络技术有限公司的发票信息",
		},
	}
	resp, err := llm.Chat(ctx, messages,
		kpllms.WithStreamingFunc(func(ctx context.Context, chunk []byte, innerErr error) error {
			fmt.Println(string(chunk))
			return nil
		}),
		kpllms.WithTools([]*kpllms.Tool{
			{
				Type: "function",
				Function: &kpllms.FunctionDefinition{
					Name:        "getWeather",
					Description: "根据地区获取天气情况",
					Parameters: schema.Definition{
						Type: schema.Object,
						Properties: map[string]schema.Definition{
							"location": {
								Type:        schema.String,
								Description: "城市或者地区",
							},
							"unit": {
								Type: schema.String,
								Enum: []string{"上海", "武汉"},
							},
						},
						Required: []string{"location"},
					},
				},
			},
			{
				Type: "function",
				Function: &kpllms.FunctionDefinition{
					Name:        "getFapiao",
					Description: "根据企业名称获取企业的发票信息",
					Parameters: schema.Definition{
						Type: schema.Object,
						Properties: map[string]schema.Definition{
							"company_name": {
								Type:        schema.String,
								Description: "企业名称",
							},
						},
						Required: []string{"company_name"},
					},
				},
			},
		}),
	)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	b, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(b))

	if resp.Choices[0].StopReason != "tool_calls" {
		return
	}

	// 如果返回的是 函数， 将返回的函数和执行函数的结果，再次加入上下文，请求回答
	messages = append(messages, &schema.ChatMessage{
		Role:      schema.RoleAssistant,
		ToolCalls: resp.Choices[0].ToolCalls,
	})
	for k, v := range resp.Choices[0].ToolCalls {
		if v.Function.Name == "getWeather" {
			var m map[string]string
			_ = json.Unmarshal([]byte(v.Function.Arguments), &m)
			messages = append(messages, &schema.ChatMessage{
				Role:       schema.RoleTool,
				Content:    getWeather(m["location"]),
				ToolCallId: resp.Choices[0].ToolCalls[k].Id,
			})
		} else if v.Function.Name == "getFapiao" {
			var m map[string]string
			_ = json.Unmarshal([]byte(v.Function.Arguments), &m)
			messages = append(messages, &schema.ChatMessage{
				Role:       schema.RoleTool,
				Content:    getFapiao(m["company_name"]),
				ToolCallId: resp.Choices[0].ToolCalls[k].Id,
			})
		}

	}

	b, _ = json.Marshal(messages)
	fmt.Println(string(b))

	resp, err = llm.Chat(ctx, messages,
		kpllms.WithModel(modelName),
		kpllms.WithStreamingFunc(func(ctx context.Context, chunk []byte, innerErr error) error {
			fmt.Println(string(chunk))
			return nil
		},
		),
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	b, _ = json.Marshal(resp)
	fmt.Println(string(b))

}

func getWeather(area string) string {
	res := map[string]string{
		"area":   area,
		"result": "天气晴，气温 15-18度",
	}
	b, _ := json.Marshal(res)
	return string(b)
}

func getFapiao(corpName string) string {
	res := map[string]string{
		"company_name": corpName,
		"content":      "2023年共开票: 2000万人名币",
	}
	b, _ := json.Marshal(res)
	return string(b)
}

// 测试向量化
func TestEmbedding(t *testing.T) {
	//modelName := "gpt-4-1106-preview"
	ctx := context.Background()
	embedLLM, err := New(WithToken(token), WithBaseURL(baseUrl), WithEmbeddingModel("text-embedding-ada-002"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	texts := []string{
		"hello",
		"你好",
	}
	res, err := embedLLM.EmbedDocuments(ctx, texts)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("res length : %d \n", len(res))
	fmt.Printf("res 向量长度： %d \n", len(res[0]))

	res1, err := embedLLM.EmbedQuery(ctx, "你好")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("res query length : %d \n", len(res1))

}

// 测试对话包含图片
func TestImageContent(t *testing.T) {
	ctx := context.Background()
	llm, err := New(WithToken(token), WithBaseURL(baseUrl), WithModel("gpt-4-vision-preview"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	messages := []*schema.ChatMessage{
		&schema.ChatMessage{
			Role:    schema.RoleSystem,
			Content: "你是一个 AI 助手",
		},
		&schema.ChatMessage{
			Role: schema.RoleUser,
			Content: []any{
				schema.TextContent{
					Type: schema.MultiContentText,
					Text: "图片里描述的是什么？",
				},
				schema.ImageContent{
					Type:     schema.MultiContentImageUrl,
					ImageUrl: "https://www.bangongyi.com/statics/images/bgy/index/index_1.jpg",
				},
			},
		},
	}
	resp, err := llm.Chat(ctx, messages,
		kpllms.WithStreamingFunc(func(ctx context.Context, chunk []byte, innerErr error) error {
			fmt.Println(string(chunk))
			return nil
		}))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	b, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(b))
}
