// 在业务中上报接口请求量：
package main

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	MyCounter prometheus.Counter
)

// 注册指标
func init() {
	// 1 定义指标（类型 名字 帮助信息）
	MyCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "my_counter_total",
		Help: "自定义total",
	})

	// 2 注册指标
	prometheus.MustRegister(MyCounter)
}

func SayHello(w http.ResponseWriter, r *http.Request) {
	MyCounter.Inc()
	fmt.Fprint(w, "Hello world")
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/counter", SayHello)
	http.ListenAndServe(":8080", nil)
}