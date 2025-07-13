package service

import (
	"github.com/gin-gonic/gin"
	"log"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// requests counter for prometheus
var httpRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of HTTP requests",
	},
	[]string{"method", "path"},
)

// Timer for prometheus
var httpDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "path"},
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration)
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// log event
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		log.Printf("%s %s %s", method, path, start.Format(time.RFC3339))

		// counter increase
		httpRequests.WithLabelValues(method, path).Inc()

		// next handler
		c.Next()

		// set request time
		duration := time.Since(start).Seconds()
		httpDuration.WithLabelValues(method, path).Observe(duration)
	}
}

func RequestLoggerToFileMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Логирование в stdout
		log.Printf("%s %s started", method, path)

		// Логирование в файл
		logFile.WriteString(time.Now().Format(time.RFC3339) + " " +
			method + " " + path + " " +
			"started\n")

		// Метрики
		httpRequests.WithLabelValues(method, path).Inc()

		c.Next()

		// После обработки запроса
		duration := time.Since(start)
		status := c.Writer.Status()

		// Логирование в stdout
		log.Printf("%s %s completed in %v with status %d",
			method, path, duration, status)

		// Логирование в файл
		logFile.WriteString(time.Now().Format(time.RFC3339) + " " +
			method + " " + path + " " +
			"completed in " + duration.String() + " " +
			"status " + strconv.Itoa(status) + "\n")

		httpDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	}
}
