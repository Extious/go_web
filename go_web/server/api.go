package server

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	shutdownMaxAge = 15 * time.Second
	shutdownWait   = 1000 * time.Millisecond
)

type ApiServer struct {
	Engine     *gin.Engine
	HttpServer *http.Server
	mu         sync.Mutex
	//钩子函数，用于处理服务器退出时的事务
	Shutdowns []func(*ApiServer)
	Routers   []func(*gin.Engine)
}

//RegisterRouters
//函数接收任意数量的func(engine *gin.Engine)类型函数并存于ApiServer切片中

func (s *ApiServer) RegisterRouters(routers ...func(engine *gin.Engine)) *ApiServer {
	s.Routers = append(s.Routers, routers...)
	return s
}

//创建新的ApiServer

func NewApiServer() *ApiServer {
	s := ApiServer{}
	s.Engine = gin.New()
	//跨域
	s.Engine.Use(s.cors())
	return &s
}

//启动http服务，开始监听

func (s *ApiServer) ListenAndServe() error {
	for _, c := range s.Routers {
		c(s.Engine)
	}
	s.HttpServer = &http.Server{
		Handler:        s.Engine,
		Addr:           "127.0.0.1:8080",
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println(fmt.Sprintf("api-server port run on %s ", s.HttpServer.Addr))
	if err := s.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// 处理操作系统信号，同时优雅退出，注意这里没有设置ctrl+Z信号，ctrl+Z是将进程置于后台，fg命令调出，后台进程无法处理请求但是会占用端口
func (s *ApiServer) setupSignal() {
	go func() {
		var sigChan = make(chan os.Signal, 1)
		signal.Notify(sigChan /*syscall.SIGUSR1,*/, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownMaxAge)
		defer shutdownCancel()
		for sig := range sigChan {
			if sig == syscall.SIGINT || sig == syscall.SIGHUP || sig == syscall.SIGTERM {
				fmt.Println(fmt.Sprintf("Graceful shutdown:signal %v to stop api-server ", sig))
				s.Shutdown(shutdownCtx)
			} else {
				fmt.Println(fmt.Sprintf("Caught signal %v", sig))
			}
		}
		return
	}()
}

func (s *ApiServer) Shutdown(ctx context.Context) {
	//Give priority to business shutdown Hook
	if len(s.Shutdowns) > 0 {
		for _, shutdown := range s.Shutdowns {
			shutdown(s)
		}
	}
	//wait for registry shutdown
	select {
	case <-time.After(shutdownWait):
	}
	// close the HttpServer
	s.HttpServer.Shutdown(ctx)
}
