package openaiclient

import (
	"encoding/json"
	"fmt"
	"testing"
)

type Image struct {
	Name string
}

func TestName(t *testing.T) {
	var a, b, c, d, e, f, g, h, i interface{}
	a = 123
	t.Log("a: \n")
	if aa, ok := a.(int); ok {
		fmt.Printf("%#v \n", aa)
	} else {
		fmt.Println("aa, ok := a.(int);ok : false")
	}
	t.Log("\n")

	b = "123"
	t.Log("a: \n")
	if bb, ok := b.(string); ok {
		fmt.Printf("%#v \n", bb)
	} else {
		fmt.Println("aa, ok := a.(int);ok : false")
	}
	t.Log("\n")

	c = b
	t.Log("c: \n")
	if cc, ok := c.(string); ok {
		fmt.Printf("%#v \n", cc)

	} else {
		fmt.Println("aa, ok := a.(*string);ok : false")
	}

	d = Image{Name: "张三"}
	t.Log("d: \n")
	if dd, ok := d.(Image); ok {
		fmt.Printf("%#v \n", dd)
	} else {
		fmt.Println("aa, ok := a.(int);ok : false")
	}
	t.Log("\n")

	e = &Image{Name: "lisi"}
	t.Log("e: \n")
	if ee, ok := e.(*Image); ok {
		fmt.Printf("%#v \n", ee)
	} else {
		fmt.Println("aa, ok := a.(int);ok : false")
	}
	t.Log("\n")

	f = []string{"123"}
	t.Log("e: \n")
	if ff, ok := f.([]string); ok {
		fmt.Printf("%#v \n", ff)
	} else {
		fmt.Println("aa, ok := a.(int);ok : false")
	}
	t.Log("\n")

	g = []Image{{"444"}}
	t.Log("e: \n")
	if gg, ok := g.([]Image); ok {
		fmt.Printf("%#v \n", gg)
	} else {
		fmt.Println("aa, ok := a.(int);ok : false")
	}
	t.Log("\n")

	h = []*Image{}
	t.Log("e: \n")
	if hh, ok := h.([]*Image); ok {
		fmt.Printf("%#v \n", hh)
	} else {
		fmt.Println("aa, ok := a.(int);ok : false")
	}
	t.Log("\n")

	i = &[]Image{}
	t.Log("e: \n")
	if ii, ok := i.(*[]Image); ok {
		fmt.Printf("%#v \n", ii)
	} else {
		fmt.Println("aa, ok := a.(int);ok : false")
	}
	t.Log("\n")

}

type Msg struct {
	Content any `json:"content"`
}

func TestName1(t *testing.T) {
	str := "{\"content\":\"aaa\"}"
	var c Msg
	err := json.Unmarshal([]byte(str), &c)
	if err != nil {

	}
	fmt.Printf("%#v \n", c)
	//if c.Content == nil {
	//	fmt.Println("content is nil")
	//}
	s, ok := c.Content.(string)
	if !ok {
		fmt.Println(" content.(string) false")
	}
	fmt.Printf("%#v \n", s)

}
