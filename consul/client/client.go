package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"fmt"
	"sync"

	"consuldemo/watcher"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/imroc/req/v3"
)

type WeightNode struct {
	weight *int
	addr string
}

// 只保存一个service的
type ServiceMap struct { // 先不考虑并发问题
	l []*WeightNode
	sync.RWMutex
}

var defaultServiceMap = new()

func new() *ServiceMap{
	return &ServiceMap{
	}
} 

func (s *ServiceMap) Add(addr string) {
	s.l = append(s.l, &WeightNode{addr: addr})
}

func (s *ServiceMap) List() []*WeightNode {
	return s.l
}

var serviceMap = make(map[string][]*WeightNode)
var serviceName = "go-consul-test-server"

func main() {
	// consul 配置
	config := consulapi.DefaultConfig()
	config.Address = "localhost:8500" // 写死
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("create consul client: ", err)
	}

	go func() {
		if err := watcher.RegisterWatcher("services", nil, "localhost:8500"); err != nil {
			fmt.Println("启动 consul 的watch监控失败")
		}
	
	}()
	// 获取全部的
	// 	services, _ := client.Agent().Services()
	//  for _, value := range services {
	// 		log.Printf("ServerId: %s,  Address: %s, Port: %d\n", value.ID, value.Address, value.Port)
	//  }

	// service id
	// service, _, _ := client.Agent().Service("1", nil)
	// log.Printf("ServerId: %s,  Address: %s, Port: %d\n", service.ID, service.Address, service.Port)

	// 获取健康的
	// services, _, _ := client.Health().Service("go-consul-test-server", "", true, nil)
	// for _, service := range services {
	// 	value := service.Service
	// 	log.Printf("ServerId: %s,  Address: %s, Port: %d\n", value.ID, value.Address, value.Port)
	// }

	// 根据 service 筛选

	services, err := client.Agent().ServicesWithFilter(`Service == "go-consul-test-server"`)
	if err != nil {
		panic(err)
	}

	for _, value := range services {
		// if serviceMap["go-consul-test-server"] == nil {
		// 	serviceMap["go-consul-test-server"] = make([]*WeightNode, 0)
		// }

		// serviceMap[serviceName] = append(serviceMap[serviceName], &WeightNode{
		// 	addr: fmt.Sprintf("%s:%d", value.Address, value.Port),
		// })
		defaultServiceMap.Add(fmt.Sprintf("%s:%d", value.Address, value.Port))
	}

	go Curl()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func Curl() {
	t := time.NewTicker(5 * time.Second)
	// 一定要调用Stop()，回收资源
	defer t.Stop()
	for {
		<-t.C
		// node := serviceMap[serviceName][0]
		// addr := node.addr

		list := defaultServiceMap.List()
		if len(list) == 0 {
			log.Println("there is no ip")
			return
		}
		addr := list[0].addr

		// req.DevMode() //  Use Client.DevMode to see all details, try and surprise :)
		req.Get("http://" + addr)
	}
}
