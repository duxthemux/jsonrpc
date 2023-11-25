package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/duxthemux/jsonrpc"
)

type InA struct {
}

type OutA struct {
	Msg string `json:"msg"`
}

type InB struct {
}

type OutB struct {
}

type AHandler struct {
}

func (a *AHandler) Name() string {
	return "a"
}

func (a *AHandler) SomeA(ctx context.Context, in *InA, out *OutA) error {
	req := jsonrpc.HttpRequest(ctx)
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Request URL: %s", req.URL.String()))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Request Method: %s", req.Method))

	out.Msg = sb.String()
	println(sb.String())
	return nil
}

type BHandler struct {
}

func (b *BHandler) Name() string {
	return "b"
}

func (b *BHandler) SomeB(ctx context.Context, in *InB, out *OutB) error {
	log.Printf("SomeB")
	return nil
}
func main() {
	jsonHandler := jsonrpc.New()

	if err := jsonHandler.Register(&AHandler{}, &BHandler{}); err != nil {
		panic(err)
	}
	// Try this
	//  curl http://localhost:8080/a/somea  -X POST -d '{}'
	//  curl http://localhost:8080/b/someb  -X POST -d '{}'
	if err := http.ListenAndServe(":8080", jsonHandler); err != nil {
		panic(err)
	}
}
