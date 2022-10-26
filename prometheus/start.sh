docker build -t my-prometheus .
docker run -p 9090:9090 my-prometheus

# or
# docker run -p 9090:9090 -v ./prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus