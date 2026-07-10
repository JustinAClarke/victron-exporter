package main

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "victron"

// Prometheus metrics exported by the exporter. These provide insight into
// connection health, runtime resource usage, and the volume of MQTT subscription
// updates handled.
var (
	connectionStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "mqtt_connection_state",
		Help:      "0=Disconnected; 1=Connected",
	}, []string{"client_id"})

	connectionStatusSinceTimeSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "mqtt_connection_state_since_time_seconds",
		Help:      "Time since last change to mqtt_connection_state",
	}, []string{"client_id"})

	goGoroutines = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "go_goroutines",
		Help:      "Number of goroutines currently active in the exporter.",
	})

	goMemAllocBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "go_mem_alloc_bytes",
		Help:      "Bytes of allocated heap objects currently in use.",
	})

	goMemSysBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "go_mem_sys_bytes",
		Help:      "Bytes of memory obtained from the OS by the Go runtime.",
	})

	subscriptionsUpdatesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "mqtt_subscription_updates_total",
		Help:      "MQTT subscriptions updated received",
	})

	subscriptionsUpdatesIgnoredTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "mqtt_subscription_updates_ignored_total",
		Help:      "MQTT subscription updates ignored",
	})
)

func init() {
	prometheus.MustRegister(connectionStatus)
	prometheus.MustRegister(connectionStatusSinceTimeSeconds)
	prometheus.MustRegister(goGoroutines)
	prometheus.MustRegister(goMemAllocBytes)
	prometheus.MustRegister(goMemSysBytes)
	prometheus.MustRegister(subscriptionsUpdatesTotal)
	prometheus.MustRegister(subscriptionsUpdatesIgnoredTotal)
}

// collectRuntimeMetrics updates Prometheus gauges with the current Go runtime
// memory and goroutine statistics. This is useful for diagnosing long-running
// resource leaks or instability in the exporter process.
func collectRuntimeMetrics() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	goGoroutines.Set(float64(runtime.NumGoroutine()))
	goMemAllocBytes.Set(float64(mem.Alloc))
	goMemSysBytes.Set(float64(mem.Sys))
}
