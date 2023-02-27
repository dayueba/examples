package pool

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrClosed = errors.New("pool is closed")
	forcePick = time.Second * 3
)

type Pool interface {
	Get() (Conn, error)
	Close() error
	//Do(func(net.Conn) error) error
}

type pool struct {
	opts    Options
	conns   []*conn
	address string
	closed  int32
	sync.RWMutex
	r *rand.Rand
}

func New(address string, opts Options) (Pool, error) {
	if address == "" {
		return nil, errors.New("invalid address settings")
	}
	if opts.Dial == nil {
		return nil, errors.New("invalid dial settings")
	}
	if opts.Cap <= 0 {
		return nil, errors.New("invalid maximum settings")
	}

	p := &pool{
		opts:    opts,
		conns:   make([]*conn, opts.Cap),
		address: address,
		closed:  0,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	for i := 0; i < p.opts.Cap; i++ {
		c, err := p.newConn()
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("dial is not able to fill the pool: %s", err)
		}
		p.conns[i] = c
	}

	return p, nil
}

// Close 关闭连接池并释放资源
func (p *pool) Close() error {
	// 如果 pool 已经被关闭，什么也不做
	if atomic.LoadInt32(&p.closed) == 1 {
		return nil
	}
	atomic.StoreInt32(&p.closed, 1)
	for i := 0; i < len(p.conns); i++ {
		conn := p.conns[i]
		if conn.Conn != nil {
			conn.reset()
		}
	}
	return nil
}

func (p *pool) Do(fn func(cc net.Conn) error) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrClosed
	}
	c, err := p.Get()
	if err != nil {
		return err
	}
	defer c.Close()

	err = fn(c)
	if err != nil {
		// 如果 err 是表示连接出问题了, 那么需要关闭连接
		return err
	}

	return nil
}

func (p *pool) Get() (Conn, error) {
	p.RWMutex.RLock()
	defer p.RWMutex.RUnlock()
	if atomic.LoadInt32(&p.closed) == 1 {
		return nil, ErrClosed
	}

	var pc, upc *conn
	c1, c2 := p.prePick()

	if c2.inflight > c1.inflight {
		pc, upc = c2, c1
	} else {
		pc, upc = c1, c2
	}

	if upc.inflight == 0 && upc.pickElapsed() > forcePick {
		pc = upc
	}

	return pc, nil
}

func (p *pool) prePick() (*conn, *conn) {
	p.RWMutex.Lock()
	a := p.r.Intn(len(p.conns))
	b := p.r.Intn(len(p.conns) - 1)
	p.RWMutex.Unlock()
	if b >= a {
		b = b + 1
	}
	c1, c2 := p.conns[a], p.conns[b]
	return c1, c2
}

func (p *pool) newConn() (*conn, error) {
	conn, err := p.opts.Dial(p.address)
	if err != nil {
		return nil, err
	}
	return p.wrapConn(conn), nil
}

func (p *pool) refresh() error {
	p.RWMutex.Lock()
	defer p.RWMutex.Unlock()
	for i := 0; i < len(p.conns); i++ {
		if p.conns[i].Conn == nil {
			conn, err := p.newConn()
			if err != nil {
				return err
			}
			p.conns[i] = conn
			break
		}
	}
	return nil
}
