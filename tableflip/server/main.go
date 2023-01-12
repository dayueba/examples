package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudflare/tableflip"
)

func main() {
	upg, _ := tableflip.New(tableflip.Options{})
	defer upg.Stop()

	time.Sleep(time.Second)
	log.SetPrefix(fmt.Sprintf("[PID: %d] ", os.Getpid()))

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for range sig {
			err := upg.Upgrade()
			if err != nil {
				log.Println("Upgrade failed:", err)
			}
		}
	}()

	// Listen must be called before Ready
	ln, _ := upg.Fds.Listen("tcp", ":20000")
	defer ln.Close()

	go serve(ln)

	if err := upg.Ready(); err != nil {
		panic(err)
	}

	<-upg.Exit()
}

func serve(ln net.Listener) {
	defer ln.Close()

	for {
		conn, err := ln.Accept() // 建立连接
		if err != nil {
			fmt.Println("accept failed, err:", err)
			continue
		}
		go process(conn) // 启动一个goroutine处理连接
	}

}

func process(conn net.Conn) {
	defer conn.Close() // 关闭连接
	for {
		reader := bufio.NewReader(conn)
		var buf [128]byte
		n, err := reader.Read(buf[:]) // 读取数据
		if err == io.EOF {
			// 读完了
			break
		}

		if err != nil {
			fmt.Println("read from client failed, err:", err)
			break
		}
		recvStr := string(buf[:n])
		fmt.Println("收到client端发来的数据：", recvStr)
		conn.Write([]byte(recvStr)) // 发送数据
	}
}
