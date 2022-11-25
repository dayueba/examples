package p2c

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
	"math"
	"context"
)

var (
	penalty uint64 = uint64(time.Second * 10) // 惩罚值，表示初始化没有数据时的默认值
	tau int64 = int64(time.Millisecond * 600) // 衰减系数，衰减系数设置越大，波幅越大（恢复越慢），反之越小（恢复越快）。
)

type WeightedNode struct {
	host string

	// 常量
	weight int // 主要是人为配置的定制，用于给不同的机器按照机器配置分配上不同的权重，权重越高，越容易被pick到。这个值可以做在服务注册与发现里，进行为每个节点分配一个权重值。
	
	// 普通收集指标
	serverCpu int // 对应节点最近500ms 内的cpu使用率，由服务端返回 不过主流实现已经把这个字段去掉了，这样服务端就不需要参与了
	inflight int64 // 代表节点请求拥塞度，代表着当前节点有多少个请求未完成或者正开始请求
	requests int64 // 请求总数
	inflights *list.List //

	// 需要加权指数加权移动平均算法推测的指标
	lag int64 // 加权移动平均算法计算出的请求延迟度, 也就是 ewma 值。用于计算负载
	success uint64 // 加权移动平均算法计算出的请求成功率（只记录grpc内部错误，比如context deadline）。用于判断节点是否健康

	stamp int64 // 最近一次resp时间戳，用于计算 ewma 值
	lastPick int64 // 最近被pick的时间戳，利用该值可以统计被选中后，一次请求的耗时
	predictTs int64
	predict   int64

	rw sync.RWMutex
}

func NewNode(host string, weight int) *WeightedNode {
  return &WeightedNode{
		host: host,
		weight: weight,
		lag:        0,
		success:    1000, // 应该是相当于保留三位小数
		inflight:   1, // 为什么是1：位了在没有请求的时候，计算负载不为0
		inflights:  list.New(),
	}
}

func (n *WeightedNode) valid() bool {
	return n.health() > 500 && n.serverCpu < 900;
}

func (n *WeightedNode) health() uint64 {
	return n.success;
}

// 福啊一点实现方式：go-kratos
func (n *WeightedNode) load() (load uint64) {
	now := time.Now().UnixNano()
	avgLag := atomic.LoadInt64(&n.lag)
	lastPredictTs := atomic.LoadInt64(&n.predictTs)
	predictInterval := avgLag / 5
	if predictInterval < int64(time.Millisecond*5) {
		predictInterval = int64(time.Millisecond * 5)
	}
	if predictInterval > int64(time.Millisecond*200) {
		predictInterval = int64(time.Millisecond * 200)
	}
	if now-lastPredictTs > predictInterval && atomic.CompareAndSwapInt64(&n.predictTs, lastPredictTs, now) {
		var (
			total   int64
			count   int
			predict int64
		)
		n.rw.RLock()
		first := n.inflights.Front()
		for first != nil {
			lag := now - first.Value.(int64)
			if lag > avgLag {
				count++
				total += lag
			}
			first = first.Next()
		}
		if count > (n.inflights.Len()/2 + 1) {
			predict = total / int64(count)
		}
		n.rw.RUnlock()
		atomic.StoreInt64(&n.predict, predict)
	}

	if avgLag == 0 {
		// penalty is the penalty value when there is no data when the node is just started.
		// The default value is 1e9 * 10
		load = penalty * uint64(atomic.LoadInt64(&n.inflight))
		return
	}
	predict := atomic.LoadInt64(&n.predict)
	if predict > avgLag {
		avgLag = predict
	}
	load = uint64(avgLag) * uint64(atomic.LoadInt64(&n.inflight))
	return
}

// 简单一点的实现方式：go-zero
// load = ewma * inflight
// 也可以 * weight
// func (n *WeightedNode) load() uint64 {
// 	// plus one to avoid multiply zero
// 	lag := uint64(math.Sqrt(float64(atomic.LoadInt64(&n.lag) + 1))) // ewma
// 	load := lag * uint64(atomic.LoadInt64(&n.inflight) + 1)
// 	if load == 0 {
// 		// penalty是初始化没有数据时的惩罚值，默认为1e9 * 250
// 		return penalty
// 	}

// 	return load
// }


func (n *WeightedNode) Weight() float64 {
	return float64(n.health()*uint64(time.Second)) / float64(n.load())
}

// 被pick后，完成请求后触发逻辑
func (n *WeightedNode) Pick() DoneFunc {
	// 请求开始
	now := time.Now().UnixNano()
	atomic.StoreInt64(&n.lastPick, now)
	atomic.AddInt64(&n.inflight, 1)
	atomic.AddInt64(&n.requests, 1)
	
	n.rw.Lock()
	e := n.inflights.PushBack(now)
	n.rw.Unlock()
	return func(ctx context.Context, err error) {
		// 请求结束
		// 1 把节点正在处理请求的总数减1
		// 2 保存处理请求结束的时间点，用于计算距离上次节点处理请求的差值，并算出 EWMA 中的 β值
		// 3 计算本次请求耗时，并计算出 EWMA值 保存到节点的 lag 属性里
		// 4 计算节点的健康状态保存到节点的 success 属性中
		n.rw.Lock()
		n.inflights.Remove(e)
		n.rw.Unlock()
		atomic.AddInt64(&n.inflight, -1)

		now := time.Now().UnixNano()
		// get moving average ratio w
		stamp := atomic.SwapInt64(&n.stamp, now)
		td := now - stamp //计算距离上次response的时间差，节点本身闲置越久，这个值越大
		if td < 0 {
			td = 0
		}
		// 实时计算β值，利用衰减函数计算，公式为：β = e^(-t/k)，相比前文给出的衰减公式这里是按照k值的反比计算的，即k值和β值成正比
		w := math.Exp(float64(-td) / float64(tau))

		start := e.Value.(int64)
		lag := now - start //实际耗时
		if lag < 0 {
			lag = 0
		}
		oldLag := atomic.LoadInt64(&n.lag)
		if oldLag == 0 {
			w = 0.0
		}
		lag = int64(float64(oldLag)*w + float64(lag)*(1.0-w))  //计算指数加权移动平均响应时间
		atomic.StoreInt64(&n.lag, lag)

		success := uint64(1000) // 成功为1000，失败为0，两种状态
		if err != nil {
			// 判断错误类型
			success = 0
			// if n.errHandler != nil && n.errHandler(err) {
			// 	success = 0
			// }

			// 如果是 ctx 错误 / server 错误 / gateway超时 则 success = 0
			// if errors.Is(context.DeadlineExceeded, di.Err) || errors.Is(context.Canceled, di.Err) ||
			// 	errors.IsServiceUnavailable(di.Err) || errors.IsGatewayTimeout(di.Err) {
			// 	success = 0
			// }
		}
		oldSuc := atomic.LoadUint64(&n.success)
		success = uint64(float64(oldSuc)*w + float64(success)*(1.0-w)) //计算指数加权移动平均成功率
		atomic.StoreUint64(&n.success, success)
	}
}

func (n *WeightedNode) PickElapsed() time.Duration {
	return time.Duration(time.Now().UnixNano() - atomic.LoadInt64(&n.lastPick))
}
