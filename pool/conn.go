package pool

import (
	"net"
	"sync/atomic"
	"time"
)

type Conn interface {
	Value() net.Conn
	Close() error
	net.Conn
}

var _ Conn = (*conn)(nil)

type conn struct {
	net.Conn
	pool     *pool
	inflight int32
	lastPick int64
}

func (c *conn) Value() net.Conn {
	atomic.AddInt32(&c.inflight, 1)
	atomic.StoreInt64(&c.lastPick, time.Now().UnixNano())
	return c.Conn
}

func (c *conn) Close() error {
	atomic.AddInt32(&c.inflight, -1)
	go func() {
		if err := connCheck(c); err != nil {
			c.reset()
			c.pool.refresh()
		}
	}()
	return nil
}

func (c *conn) reset() error {
	cc := c.Conn
	c.Conn = nil
	atomic.StoreInt32(&c.inflight, 0)
	atomic.StoreInt64(&c.lastPick, 0)
	if cc != nil {
		return cc.Close()
	}
	return nil
}

func (p *pool) wrapConn(cc net.Conn) *conn {
	return &conn{
		Conn: cc,
		pool: p,
	}
}

func (c *conn) pickElapsed() time.Duration {
	return time.Duration(time.Now().UnixNano() - atomic.LoadInt64(&c.lastPick))
}
