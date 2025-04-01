package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func startMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting Prometheus metrics server at %s/metrics", addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("metrics server failed: %v", err)
		}
	}()
}
func InitAndStartMetricsServer() {
	initMetric()
	startMetricsServer(":2112")
}
