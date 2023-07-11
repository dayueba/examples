package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"net/http"
)

var meter metric.Meter

func init() {
	// 初始化指标meter
	mexp, err := prometheus.New()
	if err != nil {
		panic(err)
	}
	meter = metricsdk.NewMeterProvider(metricsdk.WithReader(mexp)).Meter("http-demo")
}

func main() {
	// 集成指标
	// https://github.com/open-telemetry/opentelemetry-go/blob/main/example/prometheus/main.go
	// 创建一个接口访问计数器
	urlCouter, _ := meter.Int64Counter("demo_api_request_counter", metric.WithDescription("QPS"))

	go (func() {
		// 创建一个独立的server export暴露Go指标 避免通过业务服务暴露到外网
		metricServer := http.NewServeMux()
		metricServer.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", metricServer)
	})()

	r := gin.Default()
	r.GET("/v1/demo", func(c *gin.Context) {
		opt := metric.WithAttributes(attribute.Key("domain").String(c.Request.Host), attribute.Key("uri").String(c.Request.RequestURI))
		urlCouter.Add(context.Background(), 1, opt) // 计数

		c.JSON(200, nil)
	})

	r.Run(":6060")
}
