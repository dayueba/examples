package p2c

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
	"sync/atomic"
)

const (
	forcePick = time.Second * 3
	// Name is balancer name
	Name = "p2c"
)

type DoneFunc func(ctx context.Context, err error)


type Balancer interface {
	Pick(ctx context.Context, nodes []*WeightedNode) (*WeightedNode, DoneFunc, error)
}


type P2CLoadBalancer struct {
	mu     sync.Mutex
	r      *rand.Rand
	picked int64
}

var _ Balancer = (*P2CLoadBalancer)(nil)


func (s *P2CLoadBalancer) prePick(nodes []*WeightedNode) (nodeA *WeightedNode, nodeB *WeightedNode) {
	s.mu.Lock()
	a := s.r.Intn(len(nodes))
	b := s.r.Intn(len(nodes) - 1)
	s.mu.Unlock()
	if b >= a {
		b = b + 1
	}
	nodeA, nodeB = nodes[a], nodes[b]
	// todo 这里最好判断下 node.healthy()
	return
}

func (s *P2CLoadBalancer) Pick(ctx context.Context, nodes []*WeightedNode) (*WeightedNode, DoneFunc, error) {
	if len(nodes) == 0 {
		return nil, nil, errors.New("no available node")
	}

	if len(nodes) == 1 {
		return nodes[0], nodes[0].Pick(), nil
	}

	var pc, upc *WeightedNode
	nodeA, nodeB := s.prePick(nodes)
	// meta.Weight is the weight set by the service publisher in discovery
	if nodeB.Weight() > nodeA.Weight() {
		pc, upc = nodeB, nodeA
	} else {
		pc, upc = nodeA, nodeB
	}

	// 如果选中的节点，在 forceGap 期间内没有被选中一次，那么强制一次
	// 利用强制的机会，来触发成功率、延迟的衰减
	if upc.PickElapsed() > forcePick && atomic.CompareAndSwapInt64(&s.picked, 0, 1) {
		pc = upc
		atomic.StoreInt64(&s.picked, 0)
	}
	done := pc.Pick()
	return pc, done, nil
}

func NewP2CLoadBalancer() Balancer {
	return &P2CLoadBalancer{
    r: rand.New(rand.NewSource(time.Now().UnixNano())),
  }
}