package main

import (
	"pool"
)

func main() {
	connPool, err := pool.New(":8080", pool.DefaultOptions)
	if err != nil {
		panic(err)
	}
	defer connPool.Close()

	conn, err := connPool.Get()
	defer conn.Close()
}
