package pool

import (
	"net"
	"time"
)

const (
	DialTimeout = 5 * time.Second
)

type Options struct {
	Dial func(address string) (net.Conn, error)
	Cap  int
}

var DefaultOptions = Options{
	Dial: Dial,
	Cap:  8,
}

func Dial(address string) (net.Conn, error) {
	return net.DialTimeout("tcp", address, DialTimeout)
}
