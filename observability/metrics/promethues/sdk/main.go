package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	// 自定义接口请求次数自定义指标
	GlobalApiCounter *prometheus.CounterVec
)

func init() {
	// 初始化接口请求次数自定义指标
	GlobalApiCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "demo_api_request_counter",
		Help: "接口请求次数自定义指标",
	},
		[]string{"domain", "uri"}, // 域名和接口
	)
	prometheus.MustRegister(GlobalApiCounter)
}

func main() {
	r := gin.Default()
	go (func() {
		// 创建一个独立的server export暴露Go指标 避免通过业务服务暴露到外网
		metricServer := http.NewServeMux()
		metricServer.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", metricServer)
	})()
	r.GET("/v1/demo", func(c *gin.Context) {
		GlobalApiCounter.WithLabelValues(c.Request.Host, c.Request.RequestURI).Inc()
		c.JSON(200, nil)
	})

	r.Run(":6060")
}
