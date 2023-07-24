package main

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestPost(t *testing.T) {
	imgByte, err := os.ReadFile("1.gif")
	if err != nil {
		t.Fatal(err)
	}
	//c.PostForm("data")
	res, err := http.Post("http://localhost:8800/img", "image/gif", bytes.NewReader(imgByte))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	// 读取res.Body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	// 写入文件
	err = os.WriteFile("1.webp", body, 0666)
	if err != nil {
		t.Fatal(err)
	}
}
