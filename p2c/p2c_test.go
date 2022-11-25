package p2c

import (
	"testing"
	"log"
	"context"
	"time"
	"math"
)

func TestP2C(t *testing.T) {
	log.Println("testing...")

	nodes := []*WeightedNode{
		NewNode("1", 0),
    NewNode("2", 0),
    NewNode("3", 0),
		NewNode("4", 0),
    NewNode("5", 0),
	}

	p2cBalancer := NewP2CLoadBalancer()
	for i := 0; i < 10; i++ {
		node, doneFn, err :=  p2cBalancer.Pick(context.Background(), nodes)
		// 期望连续两次选择的不会相同	
		if err!= nil {
			log.Fatal("error:", err)
		}
		log.Println("select node:", node.host)
		time.Sleep(time.Second)
		doneFn(context.Background(), nil)	
	}
  log.Println("done")
}

func TestW(t *testing.T) {
	start := time.Now().UnixNano()
	time.Sleep(time.Second)
	td := time.Now().UnixNano() - start
	log.Println("w: ", math.Exp(float64(-td) / float64(tau)))
}