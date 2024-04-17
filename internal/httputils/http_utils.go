package httputils

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/comqositi/kpllms/schema"
	"io"
	"log"
	"net/http"
	netUrl "net/url"
)

// HttpPost 发送 http post 请求
// resp 为返回值
func HttpPost(ctx context.Context, baseUrl string, payload any, headers map[string]string, resp any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}

	// Build request
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseUrl, body)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	for k, v := range headers {
		fmt.Println(k, ":", v)
		req.Header.Set(k, v)
	}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	if r.StatusCode != http.StatusOK {
		return schema.NewHttpError(r.StatusCode, string(b))
	}

	err = json.Unmarshal(b, resp)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	return nil
}

// HttpGet 发送 http get 请求， resp 为返回值
func HttpGet(ctx context.Context, baseUrl string, query map[string]string, headers map[string]string, resp any) error {

	params := netUrl.Values{}
	for s, s2 := range query {
		params.Add(s, s2)
	}
	u, err := netUrl.Parse(baseUrl)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	u.RawQuery = params.Encode()
	baseUrl = u.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseUrl, nil)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	if r.StatusCode != http.StatusOK {
		return schema.NewHttpError(r.StatusCode, string(b))
	}

	err = json.Unmarshal(b, resp)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	return nil
}

type streamFunc = func(ctx context.Context, line string) error

// HttpStream http sse 默认 post 提交
func HttpStream(ctx context.Context, baseUrl string, payload any, headers map[string]string, sfunc streamFunc) error {
	payloadBytes, err := json.Marshal(payload)
	//fmt.Println(string(payloadBytes))
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	// Build request
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseUrl, body)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return schema.NewHttpError(0, err.Error())
	}
	defer func(Body io.ReadCloser) {
		err1 := Body.Close()
		if err1 != nil {
			fmt.Println("关闭 r.body 异常：", err1.Error())
		}
	}(r.Body)

	if r.StatusCode != http.StatusOK {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return schema.NewHttpError(r.StatusCode, err.Error())
		}
		return schema.NewHttpError(r.StatusCode, string(b))
	}

	return parseStreaming(ctx, r, sfunc)
}

// parseStreaming 处理流式返回
func parseStreaming(ctx context.Context, r *http.Response, sfunc streamFunc) error { //nolint:cyclop,lll
	scanner := bufio.NewScanner(r.Body)
	for scanner.Scan() {
		line := scanner.Text()
		//fmt.Println(line)
		err := sfunc(ctx, line)
		if err != nil {
			return schema.NewHttpError(0, err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("issue scanning response:", err)
		return schema.NewHttpError(0, err.Error())
	}
	return nil

}
