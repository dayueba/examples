package main

import (
	"fmt"

	"github.com/hashicorp/consul/api"

	"consuldemo/config/watcher"
)

func main() {
	consulClient, err := api.NewClient(&api.Config{
			Address: "127.0.0.1:8500",
	})
	if err != nil {
			panic(err)
	}
	cs, err := consul.New(consulClient,  consul.WithPath("test/"))
	// consul中需要标注文件后缀，kratos读取配置需要适配文件后缀
	// The file suffix needs to be marked, and kratos needs to adapt the file suffix to read the configuration.
	if err != nil {
			panic(err)
	}

	w, err := cs.Watch()
	if err != nil {
		panic(err)
	}

	for {
		kvs, err := w.Next()
		if err != nil {
			panic(err)
		}
		for _, kv := range kvs {
			fmt.Println(kv)
		}
	}
}