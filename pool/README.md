## 多路复用连接池

一个连接可以被多个`goroutine`持有

useage
```go
package main

func main() {
	connPool, err := pool.New(":8080", pool.DefaultOptions)
	if err != nil {
		panic(err)
	}
	defer connPool.Close()

	conn, err := connPool.Get()
	defer conn.Close()
}

```