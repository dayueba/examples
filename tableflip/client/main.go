package main

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"
)

type Pool struct {
	conns  []net.Conn
	size   uint64
	cursor uint64
}

func NewPool(size int) *Pool {
	p := Pool{conns: make([]net.Conn, size), size: uint64(size)}
	for i := 0; i < size; i++ {
		conn, err := net.Dial("tcp", "127.0.0.1:20000")
		if err != nil {
			panic(err)
		}
		p.conns[i] = conn
	}
	return &p
}

func (p *Pool) Cleanup() {
	for i := 0; i < len(p.conns); i++ {
		p.conns[i].Close()
	}
}

func (p *Pool) Do(fn func(c net.Conn) error) error {
	var err error
	for {
		i := atomic.AddUint64(&p.cursor, 1) % p.size
		conn := p.conns[i]
		err = fn(conn)

		if err == nil {
			return nil
		}

		// 先忽略不等于nil的情况
	}

}

func main() {
	pool := NewPool(4)
	defer pool.Cleanup()

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		<-t.C
		err := pool.Do(func(conn net.Conn) error {
			_, err := conn.Write([]byte("hello world"))
			if err != nil {
				return err
			}
			buf := [512]byte{}
			n, err := conn.Read(buf[:])
			if err != nil {
				return err
			}
			fmt.Println("recv: ", string(buf[:n]))
			return nil
		})

		if err != nil {
			fmt.Println(err)
		}
	}

}
