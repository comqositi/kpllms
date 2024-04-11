package httputils

import (
	"context"
	"fmt"
	"os"
	"testing"
)

type response struct {
	Id    string `json:"id"`
	Model string `json:"model"`
}

func TestHttpPost(t *testing.T) {
	url := "https://api.minimax.chat/v1/text/chatcompletion_v2"
	payload := map[string]any{
		"model": "abab5.5-chat",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "你好",
			},
		},
	}
	header := map[string]string{
		"Authorization": "Bearer " + os.Getenv("MINIMAX_API_KEY"),
		"Content-Type":  "application/json",
	}
	var resp response
	err := HttpPost(context.Background(), url, payload, header, &resp)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v", resp)
}

type response1 struct {
	data  string `json:"data"`
	error string `json:"error"`
}

func TestHttpGet(t *testing.T) {
	url := "http://192.168.1.34:7080/api/notify/"
	header := map[string]string{
		"X-Token":      "ea60f10040f946a2901b260e8dac7b5e",
		"Content-Type": "application/json",
	}
	query := map[string]string{"a": "b", "c": "1"}
	var resp response1
	err := HttpGet(context.Background(), url, query, header, &resp)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v", resp)
}

func TestHttpStream(t *testing.T) {
	url := "https://api.minimax.chat/v1/text/chatcompletion_pro"
	payload := map[string]any{
		"model": "abab5.5-chat",
		"bot_setting": []map[string]string{
			{
				"bot_name": "MM智能助理",
				"content":  "MM智能助理是一款由MiniMax自研的，没有调用其他产品的接口的大型语言模型。MiniMax是一家中国科技公司，一直致力于进行大模型相关的研究。",
			},
		},
		"messages": []map[string]string{
			{
				"sender_type": "USER",
				"sender_name": "小明",
				"text":        "帮我用英文翻译下面这句话：我是谁",
			},
		},
		"stream":            true,
		"reply_constraints": map[string]string{"sender_type": "BOT", "sender_name": "MM智能助理"},
	}
	header := map[string]string{
		"Authorization": "Bearer " + os.Getenv("MINIMAX_API_KEY"),
		"Content-Type":  "application/json",
	}
	for i := 0; i < 200; i++ {
		fmt.Print(" ", i)
	}
	fmt.Println("")
	err := HttpStream(context.Background(), url, payload, header, func(ctx context.Context, line string) error {
		fmt.Println(line)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
