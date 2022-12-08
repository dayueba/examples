// 配置中心demo
package main

import (
	"log"

	consulapi "github.com/hashicorp/consul/api"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Db DBConfig
}

type DBConfig struct {
	Dsn string
}

func main() {
	config := consulapi.DefaultConfig()
	config.Address = "localhost:8500" // 写死
	client, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal("create consul client: ", err)
	}

	key := "config"

	c1 := Config{
		Db: DBConfig{
			Dsn: "localhost:3307",
		},
	}
	b, err := yaml.Marshal(&c1)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// set
	meta, err := client.KV().Put(&consulapi.KVPair{Key: key, Flags: 0, Value: b}, nil)
	if err != nil {
		log.Fatalln(err)
		return
	}
	log.Println("put:", meta)
	log.Println("=================")

	// get
	data, _, err := client.KV().Get(key, nil)
	if err != nil {
		log.Fatalln(err)
		return
	}
	var c Config
	err = yaml.Unmarshal(data.Value, &c)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("dsn value：", c.Db.Dsn)

	// list
	/*datas, _, _ := client.KV().List("key_", nil)
	  for _, value := range datas {
	     fmt.Println("val:", string(value.Value))
	  }
	  keys, _, _ := client.KV().Keys("key_", "", nil)
	  fmt.Println("key:", keys)*/

}
