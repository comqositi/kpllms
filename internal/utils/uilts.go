package utils

import (
	"encoding/json"
	"fmt"
)

func PrintlnJson(d any) {
	b, err := json.Marshal(d)
	if err != nil {
		fmt.Println("打印 json 失败：", err.Error())
		return
	}
	fmt.Println(string(b))
}
