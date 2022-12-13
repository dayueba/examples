package pool

import (
	"net"
	"sync/atomic"
	"time"
)

type Conn interface {
	Value() net.Conn
	Close() error
	Reset() error
}

type conn struct {
	cc       net.Conn
	pool     *pool
	inflight int32
	lastPick int64
}

func (c *conn) Value() net.Conn {
	atomic.AddInt32(&c.inflight, 1)
	atomic.StoreInt64(&c.lastPick, time.Now().UnixNano())
	return c.cc
}

func (c *conn) Close() error {
	atomic.AddInt32(&c.inflight, -1)
	return nil
}

func (c *conn) Reset() error {
	cc := c.cc
	c.cc = nil
	atomic.StoreInt32(&c.inflight, 0)
	atomic.StoreInt64(&c.lastPick, 0)
	if cc != nil {
		return cc.Close()
	}
	return c.pool.refresh()
}

func (p *pool) wrapConn(cc net.Conn) *conn {
	return &conn{
		cc:   cc,
		pool: p,
	}
}

func (c *conn) pickElapsed() time.Duration {
	return time.Duration(time.Now().UnixNano() - atomic.LoadInt64(&c.lastPick))
}
