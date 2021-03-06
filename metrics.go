package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//timing counters - query time out to prometheus

// ApplicationMetrics enumerates the metrics collected and reported by the trickster application.
type ApplicationMetrics struct {

	// Persist Metrics
	CacheRequestStatus   *prometheus.CounterVec
	CacheRequestElements *prometheus.CounterVec
	ProxyRequestDuration *prometheus.HistogramVec
}

// NewApplicationMetrics returns a ApplicationMetrics object and instantiates an HTTP server for polling them.
func NewApplicationMetrics(config *Config, logger log.Logger) *ApplicationMetrics {

	metrics := ApplicationMetrics{
		// Metrics
		CacheRequestStatus: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trickster_requests_total",
				Help: "Count of ",
			},
			[]string{"origin", "origin_type", "method", "status", "http_status"},
		),
		CacheRequestElements: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trickster_points_total",
				Help: "Count of data points returned in a Prometheus query_range Request",
			},
			[]string{"origin", "origin_type", "status"},
		),
		ProxyRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "trickster_proxy_duration_ms",
				Help:    "Time required (ms) to proxy a given Prometheus query.",
				Buckets: []float64{50, 100, 500, 1000, 5000, 10000, 20000},
			},
			[]string{"origin", "origin_type", "method", "status", "http_status"},
		),
	}

	// Register Metrics
	prometheus.MustRegister(metrics.CacheRequestStatus)
	prometheus.MustRegister(metrics.CacheRequestElements)
	prometheus.MustRegister(metrics.ProxyRequestDuration)

	// Turn up the Metrics HTTP Server
	if config.Metrics.ListenPort > 0 {
		go func() {

			level.Info(logger).Log("event", "metrics http endpoint starting", "port", fmt.Sprintf("%d", config.Metrics.ListenPort))

			http.Handle("/metrics", promhttp.Handler())
			if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Metrics.ListenPort), nil); err != nil {
				level.Error(logger).Log("event", "unable to start metrics http server", "detail", err.Error())
				os.Exit(1)
			}
		}()
	}

	return &metrics

}
