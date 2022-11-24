//go:build go1.8
// +build go1.8

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"flag"

	"consuldemo/pkg"

	"github.com/gin-gonic/gin"
	consulapi "github.com/hashicorp/consul/api"
)

var (
	port int // http server 端口
	name string // http server name
	serverId string
)

func init() {
	id, _ := os.Hostname()
  flag.IntVar(&port, "port", 8080, "监听的端口")
  flag.StringVar(&name, "name", "go-consul-test-server", "服务名")
	flag.StringVar(&serverId, "id", id, "全局唯一id")
}

func main() {
	flag.Parse()

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome Gin Server")
	})

	router.GET("/check", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "ok",
		})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	ip, err := pkg.GetLocalIP() // 获取本地ip
	if err != nil {
		log.Fatalf("listen: %s\n", err)
	}

	// consul 配置
	config := consulapi.DefaultConfig()
	config.Address = "localhost:8500" // 写死
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("create consul client: ", err)
	}

	// 服务注册
	registration := &consulapi.AgentServiceRegistration{}
	registration.Address = ip
	registration.ID = serverId
	registration.Port = port
	registration.Name = name
	registration.Tags = []string{"tag01", "tag02"}

	check := &consulapi.AgentServiceCheck{}
	check.HTTP = fmt.Sprintf("http://%s:%d/check", registration.Address, registration.Port)
	check.CheckID = serverId
	check.Timeout = "5s"
	check.Interval = "5s"
	check.DeregisterCriticalServiceAfter = "30s"
	registration.Check = check

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatal("ConsulRegister:", err)
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 服务注销
	client.Agent().ServiceDeregister(serverId)

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
