package metrics

import (
	"log"
	"net/http"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	metricsPrefix = "runai_cli_exposer_"
)

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: metricsPrefix + "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "handler", "status"},
	)
	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    metricsPrefix + "http_request_duration_seconds",
			Help:    "Histogram of the duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "handler", "status"},
	)
)

func ExposeMetrics(metricsHandler http.Handler, port string) {
	router := gin.Default()
	router.Use(gin.Logger())
	router.GET(consts.MetricsPath, gin.WrapH(metricsHandler))
	go func() {
		log.Printf("Expose metrics on port %v\n", port)
		err := router.Run(":" + port)
		if err != nil {
			panic(err)
		}
	}()
}

func InitMetrics() http.Handler {
	promReg := prometheus.NewRegistry()
	promReg.MustRegister(HttpRequestsTotal)
	promReg.MustRegister(HttpRequestDuration)
	handler := promhttp.HandlerFor(promReg, promhttp.HandlerOpts{Registry: promReg})

	return handler
}
