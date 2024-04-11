package ssekpai

import (
	"context"
	"fmt"
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

type header struct {
}

func (h header) Set(key, value string) {
	fmt.Printf("set header %s %s", key, value)
}

// 实现http.ResponseWriter接口的Header方法
func (m *MockResponseWriter) Header() http.Header {
	return m.headers
}

// 实现http.ResponseWriter接口的Write方法
func (m *MockResponseWriter) Write(b []byte) (int, error) {
	fmt.Println(string(b))
	return 0, nil
}

// 实现http.ResponseWriter接口的WriteHeader方法
func (m *MockResponseWriter) WriteHeader(code int) {
	m.code = code
}

func (m *MockResponseWriter) Flush() {
	fmt.Println("flush")
}

func TestSse(t *testing.T) {

	w := &MockResponseWriter{}
	// 新建 sse 客户端
	c := NewSse(context.Background(), w,
		// contex done 时执行的方法，可不配置
		WithCtxDoneFunc(func(done any) {
			fmt.Println("ctx done ")
		}),
		// 通道数据读取超时执行的方法，可不配置
		WithTimeOutFunc(func() {
			fmt.Println("time out ")
		}),
	)
	var resp Response
	go func(r *Response) {
		// 执行业务逻辑， 外部接收函数执行参数
		result := WriteData(c)
		// 赋值返回值，供外部使用
		r.Success = result.Success
	}(&resp)

	// 此方法会阻塞
	c.SendMsgBlock()

	fmt.Println("=====")
	fmt.Println(resp.Success)
}

type Response struct {
	Success string
}

func WriteData(s *Sse) *Response {
	defer s.Finished()
	for i := 0; i < 10; i++ {
		e := Event{
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
