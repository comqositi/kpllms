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
	resp, err := llm.Chat(ctx, messages, kpllms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
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
		kpllms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
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
		kpllms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
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
