download promethues: https://github.com/docker/awesome-compose/tree/master/prometheus-grafana

原生sdk 与 otel对比
1. 相同点：仍然单独创建一个server使用github.com/prometheus/client_golang/prometheus/promhttp暴露指标
2. 不同点：使用go.opentelemetry.io/otel/exporters/prometheus初始化一个指标对象meter
3. 不同点：使用meter.Int64Counter初始化计数器、直方图等
4. 不同点：metric.WithAttributes打标签

