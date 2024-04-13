package ssekpai

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// 默认超时时间
var defaultTimeOut = 30

type Sse struct {
	// ctx 必传
	ctx context.Context
	// http response
	w http.ResponseWriter
	// 消息通道
	eventChain chan Event
	// 消息通道是否已关闭 false：未关闭， true：已关闭
	isClosed bool

	// 处理 ctx.done
	doneFunc func(done any)
	// 处理 timeout
	timeOutFunc func()
	// timeout 秒
	timeOutSecond int
}

// NewSse 创建 sse 示例
func NewSse(ctx context.Context, w http.ResponseWriter, opts ...Option) *Sse {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	sse := &Sse{
		ctx:        ctx,
		w:          w,
		eventChain: make(chan Event),
	}
	if o.ctxDoneFunc != nil {
		sse.doneFunc = o.ctxDoneFunc
	}
	if o.timeOutFunc != nil {
		sse.timeOutFunc = o.timeOutFunc
	}

	sse.timeOutSecond = o.timeOutSecond
	if sse.timeOutSecond == 0 {
		sse.timeOutSecond = defaultTimeOut
	}
	return sse
}

// StreamData 向通道写入消息
func (s *Sse) StreamData(e Event) {
	s.eventChain <- e
}

func (s *Sse) Finished() {
	close(s.eventChain)
}

// SendMsgBlock 持续读取通道消息, 并发送给客户端
func (s *Sse) SendMsgBlock() {

	defer func() {
		if !s.isClosed {
			close(s.eventChain)
		}
	}()

	flush, ok := s.w.(http.Flusher)
	if !ok {
		fmt.Println("flush faild")
		return
	}

	// sse 响应头
	s.w.Header().Set("Content-Type", "text/event-stream")
	// sse 不缓存
	s.w.Header().Set("Cache-Control", "no-cache")
	s.w.Header().Set("Connection", "keep-alive")
	//s.w.Header().Set("Transfer-Encoding", "chunked")
	// 通知nginx反向代理，不要缓存数据
	s.w.Header().Set("X-Accel-Buffering", "no")

	for {
		select {
		case done := <-s.ctx.Done():
			// 上下文结束时停止监听，例如：请求结束了
			fmt.Printf("s.ctx.Done : %#v \n", done)
			if s.doneFunc != nil {
				s.doneFunc(done)
			}
			return
		case event, hasData := <-s.eventChain:
			if !hasData {
				// 通道关闭，结束读取
				s.isClosed = true
				fmt.Println("close chain")
				return
			}
			Encode(s.w, event)
			flush.Flush()
		case <-time.After(time.Duration(s.timeOutSecond) * time.Second):
			// 单次等待超时
			if s.timeOutFunc != nil {
				s.timeOutFunc()
			}
			return
		}
	}
}
