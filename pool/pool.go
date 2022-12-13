package pool

import (
	"errors"
	"fmt"
	"math/rand"
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
}

type pool struct {
	opts    Options
	conns   []*conn
	address string
	closed  int32
	sync.RWMutex
	r *rand.Rand
}

func New(address string, option Options) (Pool, error) {
	if address == "" {
		return nil, errors.New("invalid address settings")
	}
	if option.Dial == nil {
		return nil, errors.New("invalid dial settings")
	}
	if option.Cap <= 0 {
		return nil, errors.New("invalid maximum settings")
	}

	p := &pool{
		opts:    option,
		conns:   make([]*conn, option.Cap),
		address: address,
		closed:  0,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	for i := 0; i < p.opts.Cap; i++ {
		c, err := p.opts.Dial(address)
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("dial is not able to fill the pool: %s", err)
		}
		p.conns[i] = p.wrapConn(c)
	}

	return p, nil
}

func (p *pool) Close() error {
	atomic.StoreInt32(&p.closed, 1)
	for i := 0; i < len(p.conns); i++ {
		conn := p.conns[i]
		if conn != nil {
			conn.Reset()
			p.conns[i] = nil
		}
	}
	return nil
}

func (p *pool) Get() (Conn, error) {
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
		if p.conns[i].cc == nil {
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
