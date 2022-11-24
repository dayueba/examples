package watcher

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"sync"
)

// watch包的使用方法为：1）使用watch.Parse(查询参数)生成Plan，2）绑定Plan的handler，3）运行Plan

// 定义watcher
type Watcher struct {
	Address  string                 // consul agent 的地址："127.0.0.1:8500"
	Wp       *watch.Plan            // 总的Services变化对应的Plan
	watchers map[string]*watch.Plan // 对已经进行监控的service作个记录
	RWMutex  *sync.RWMutex
}

// 将consul新增的service加入，并监控
func (w *Watcher) registerServiceWatcher(serviceName string) error {
	// watch endpoint 的请求参数，具体见官方文档：https://www.consul.io/docs/dynamic-app-config/watches#service
	wp, err := watch.Parse(map[string]interface{}{
		"type":    "service",
		"service": serviceName,
	})
	if err != nil {
		return err
	}

	// 定义service变化后所执行的程序(函数)handler
	wp.Handler = func(idx uint64, data interface{}) {
		fmt.Printf("have event: %T\n", data)
		switch d := data.(type) {
		case []*consulapi.ServiceEntry:
			for _, i := range d {
				// 这里是单个service变化时需要做的逻辑，可以自己添加，或在外部写一个类似handler的函数传进来
				fmt.Printf("service %s 已变化 ", i.Service.Service)
				// 打印service的状态
				fmt.Println("service status: ", i.Checks.AggregatedStatus())
			}
}

		
	}
	// 启动监控
	go wp.Run(w.Address)
	// 对已启动监控的service作一个记录
	w.RWMutex.Lock()
	w.watchers[serviceName] = wp
	w.RWMutex.Unlock()

	return nil
}

func NewWatcher(watchType string, opts map[string]string, consulAddr string) (*Watcher, error) {
	var options = map[string]interface{}{
		"type": watchType,
	}
	// 组装请求参数。(监控类型不同，其请求参数不同)
	for k, v := range opts {
		options[k] = v
	}

	wp, err := watch.Parse(options)
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		Address:  consulAddr,
		Wp:       wp,
		watchers: make(map[string]*watch.Plan),
		RWMutex:  new(sync.RWMutex),
	}

	wp.Handler = func(idx uint64, data interface{}) {
		switch d := data.(type) {
		// 这里只实现了对services的监控，其他监控的data类型判断参考：https://github.com/dmcsorley/avast/blob/master/consul.go
		// services
		case map[string][]string:
			for i := range d {
				// 如果该service已经加入到ConsulRegistry的services里监控了，就不再加入 或者i 为 "consul"的字符串
				// 为什么会多一个consul，参考官方文档services监听的返回值：https://www.consul.io/docs/dynamic-app-config/watches#
				if _, ok := w.watchers[i]; ok || i == "consul" {
					continue
				}
				// 这里设置监听的类型
				w.registerServiceWatcher(i)
			}

			// 从总的services变化中找到不再监控的service并停止
			w.RWMutex.RLock()
			watches := w.watchers
			w.RWMutex.RUnlock()

			// remove unknown services from watchers
			for i, svc := range watches {
				if _, ok := d[i]; !ok {
					svc.Stop()
					delete(watches, i)
				}
			}
		default:
			fmt.Printf("不能判断监控的数据类型: %v", &d)
		}
	}

	return w, nil
}

func RegisterWatcher(watchType string, opts map[string]string, consulAddr string) error {
	w, err := NewWatcher(watchType, opts, consulAddr)
	if err != nil {
		panic(err)
	}
	defer w.Wp.Stop()

	if err = w.Wp.Run(consulAddr); err != nil {
		fmt.Println("err: ", err)
		return err
	}

	return nil
}
