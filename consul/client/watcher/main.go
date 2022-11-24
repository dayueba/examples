package main

import (
	"context"
	"fmt"
	"time"

	"consuldemo/consul"

	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
)

func main() {
	addr := "127.0.0.1:9002"

	cli, err := api.NewClient(&api.Config{Address: "127.0.0.1:8500", WaitTime: 2 * time.Second})
	if err != nil {
		panic(err)
	}

	instance1 := &registry.ServiceInstance{
		ID:        "1",
		Name:      "server-1",
		Version:   "v0.0.1",
		Endpoints: []string{fmt.Sprintf("tcp://%s?isSecure=false", addr)},
	}


	r := consul.New(cli, consul.WithHealthCheck(false))

	err = r.Register(context.Background(), instance1)
	if err != nil {
		panic(err)
	}

	go func() {
		time.Sleep(time.Second * 10)
		instance2 := &registry.ServiceInstance{
			ID:        "2",
			Name:      "server-1",
			Version:   "v0.0.1",
			Endpoints: []string{fmt.Sprintf("tcp://%s?isSecure=false", "localhost:9096")},
		}
	
		err = r.Register(context.Background(), instance2)
		if err != nil {
			panic(err)
		}

		time.Sleep(time.Second * 10)
		
		err = r.Deregister(context.Background(), instance2)
		if err != nil {
			panic(err)
		}
	}()

	defer func() {
		err = r.Deregister(context.Background(), instance1)
		if err != nil {
			panic(err)
		}
	}()

	watch, err := r.Watch(context.Background(), instance1.Name)
	if err != nil {
		panic(err)
	}

	for {
		service, err := watch.Next()

		if err != nil {
			panic(err)
		}
	
		for _, v := range service {
			fmt.Println(v.Name, v.Endpoints)
		}
		fmt.Println("================================")
	}
}
