package sse_example

import (
	"context"
	"fmt"
	"github.com/comqositi/kpllms"
	"github.com/comqositi/kpllms/internal/utils"
	"github.com/comqositi/kpllms/minimax"
	"github.com/comqositi/kpllms/schema"
	"github.com/comqositi/kpllms/ssekpai"
	"net/http"
	"strconv"
	"testing"
	"time"
)

// 模拟的ResponseWriter
// sse 使用示例
type MockResponseWriter struct {
	headers http.Header
	code    int
}

// 实现http.ResponseWriter接口的Header方法
func (m *MockResponseWriter) Header() http.Header {
	return m.headers
}

// 实现http.ResponseWriter接口的Write方法
func (m *MockResponseWriter) Write(b []byte) (int, error) {
	fmt.Print(string(b))
	return 0, nil
}

// 实现http.ResponseWriter接口的WriteHeader方法
func (m *MockResponseWriter) WriteHeader(code int) {
	m.code = code
}

func (m *MockResponseWriter) Flush() {
	//fmt.Print("flush")
}

// 调用大模型，并将大模型结果通过 sse 返回给前端
func TestSse(t *testing.T) {

	w := &MockResponseWriter{}
	w.headers = http.Header{}
	// 新建 sse 客户端
	c := ssekpai.NewSse(context.Background(), w,
		// contex done 时执行的方法，可不配置
		ssekpai.WithCtxDoneFunc(func(done any) {
			fmt.Println("ctx done ")
		}),
		// 通道数据读取超时执行的方法，可不配置
		ssekpai.WithTimeOutFunc(func() {
			fmt.Println("time out ")
		}),
	)
	//var resp Response
	var resllm schema.ContentResponse
	go func() {
		//// 执行业务逻辑， 外部接收函数执行参数
		//result := WriteData(c)
		//// 赋值返回值，供外部使用
		//resp.Success = result.Success

		res, err := writeByLlm(c)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			resllm.Choices = res.Choices
		}
	}()

	// 此方法会阻塞， 先使用 goroutine 执行业务逻辑
	c.SendMsgBlock()

	fmt.Println("=====")
	utils.PrintlnJson(resllm)
}

type Response struct {
	Success string
}

func WriteData(s *ssekpai.Sse) *Response {
	// 输出完成，或者报错，必须调用 finished 方法关闭通道，如果不调用 finished 方式，会一直阻塞，直到60秒超时
	defer s.Finished()
	for i := 0; i < 10; i++ {
		e := ssekpai.Event{
			Event: "success",
			Id:    "",
			Retry: 0,
			Data:  "msg : " + strconv.Itoa(i),
		}
		s.StreamData(e)
		time.Sleep(1 * time.Second)
	}
	return &Response{Success: "over"}
}

var baseUrl = "https://api.minimax.chat/v1"
var modelName = "abab5.5-chat"

// 调用大模型，并将大模型返回的结果通过 sse 输出给前端
func writeByLlm(s *ssekpai.Sse) (*schema.ContentResponse, error) {

	defer s.Finished()

	llm, err := minimax.NewChat(minimax.WithBaseUrl(baseUrl), minimax.WithModel(modelName))
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	messages := []*schema.ChatMessage{
		{
			Role:    schema.RoleSystem,
			Content: "你是一个 AI助手",
		},
		{
			Role:    schema.RoleUser,
			Content: "今天天气怎么样",
		},
	}
	resp, err := llm.Chat(context.Background(), messages, kpllms.WithStreamingFunc(func(ctx context.Context, chunk []byte, innerErr error) error {
		evt := ssekpai.Event{
			Event: "message",
			Data:  string(chunk),
		}
		s.StreamData(evt)
		return nil
	}))
	return resp, nil
}
