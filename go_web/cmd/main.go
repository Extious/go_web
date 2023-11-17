package main

import (
	"fmt"
	"go_web/go_web/internal/handler"
	"go_web/go_web/server"
	"os"
)

func main() {
	server := server.NewApiServer()
	server.RegisterRouters(handler.RegisterRouters)
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(fmt.Sprintf("err to run the httpserver:[%s]", err))
		os.Exit(1)
	}
	os.Exit(0)
}
