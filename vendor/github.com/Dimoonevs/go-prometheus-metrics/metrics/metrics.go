package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	UserRegistered = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "user_registered_total",
		Help: "Total number of registered users",
	})

	UserVerified = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "user_verified_total",
		Help: "Total number of verified users",
	})
)

func initMetric() {
	prometheus.MustRegister(UserRegistered)
	prometheus.MustRegister(UserVerified)
}
