package minimaxclientv1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/comqositi/kpllms/internal/httputils"
	"net/http"
	"strings"
)

type Client struct {
	groupId         string
	apiKey          string
	baseUrl         string
	model           string
	httpClient      Doer
	embeddingsModel string
}

func NewClient(opts ...Option) (*Client, error) {
	c := &Client{}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	if c.groupId == "" {
		return nil, errors.New("group id 不能为空")
	}
	if c.apiKey == "" {
		return nil, errors.New("api key 不能为空")
	}

	if c.baseUrl == "" {
		c.baseUrl = defaultBaseUrl
	}
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	return c, nil
}

// CreateCompletion creates a completion.
func (c *Client) CreateCompletion(ctx context.Context, r *CompletionRequest) (*Completion, error) {

	if r.Model == "" {
		r.Model = c.model
	}
	url := fmt.Sprintf("%s/text/chatcompletion_pro?GroupId=%s", c.baseUrl, c.groupId)

	//fmt.Println(url)
	//fmt.Println("==========")
	//b, _ := json.Marshal(r)
	//fmt.Println(string(b))

	var streamPayload Completion

	if r.Stream {
		err := httputils.HttpStream(ctx, url, r, c.setHeader(), func(ctx context.Context, line string) error {
			//fmt.Println(line)
			if line == "\n" || line == "" {
				return nil
			}
			var data string
			if !strings.HasPrefix(line, "data: ") {
				return nil
			} else {
				// 错误  {"error_code":6,"error_msg":"No permission to access data"}
				data = strings.TrimPrefix(line, "data: ")
			}
			err := json.Unmarshal([]byte(data), &streamPayload)
			if err != nil {
				return err
			}
			if streamPayload.BaseResp.StatusCode != 0 {
				return errors.New(fmt.Sprintf("statusCode: %d, errMsg: %s", streamPayload.BaseResp.StatusCode, streamPayload.BaseResp.StatusMsg))
			}
			// 用户输入内容命中敏感词
			if streamPayload.InputSensitive {
				return errors.New("模型返回：输入内容违规")
			}
			// 用户输出命中敏感词
			if streamPayload.OutputSensitive {
				return errors.New("模型返回：输出内容违规")
			}
			// 如果调用了 functon， 无需流式返回数据
			if streamPayload.Choices[0].Messages[0].FunctionCall != nil {
				return nil
			}
			if streamPayload.Choices[0].FinishReason == "stop" {
				return nil
			}
			err = r.StreamingFunc(ctx, []byte(streamPayload.Choices[0].Messages[0].Text), nil)
			return err
		})
		if err != nil {
			return nil, err
		}

	} else {
		err := httputils.HttpPost(ctx, url, r, c.setHeader(), &streamPayload)
		if err != nil {
			return nil, err
		}
		if streamPayload.BaseResp.StatusCode != 0 {
			return nil, errors.New(fmt.Sprintf("statusCode: %d, errMsg: %s", streamPayload.BaseResp.StatusCode, streamPayload.BaseResp.StatusMsg))
		}
	}

	return &streamPayload, nil
}

// 设置权限
func (c *Client) setHeader() map[string]string {
	return map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + c.apiKey,
	}
}
