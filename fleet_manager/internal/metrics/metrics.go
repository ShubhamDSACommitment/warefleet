// Package metrics exposes fleet KPIs to Prometheus (scraped by Grafana).
//
// These are the numbers the benchmark reports and the Grafana board renders:
// throughput, makespan, conflicts, utilization, dropout recovery.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	TasksCompleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "warefleet_tasks_completed_total",
		Help: "Total warehouse tasks completed.",
	})

	TaskMakespan = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "warefleet_task_makespan_seconds",
		Help:    "Time from task intake to completion.",
		Buckets: prometheus.LinearBuckets(5, 5, 20),
	})

	PathConflicts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "warefleet_path_conflicts_total",
		Help: "MAPF conflicts detected requiring a re-plan.",
	})

	RobotsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "warefleet_robots_active",
		Help: "Number of robots currently online.",
	})

	RobotsOffline = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "warefleet_robots_offline",
		Help: "Number of robots currently dropped out.",
	})

	DropoutRecovery = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "warefleet_dropout_recovery_seconds",
		Help:    "Time to re-assign a dropped robot's task.",
		Buckets: prometheus.LinearBuckets(1, 2, 15),
	})
)

// Serve exposes /metrics for Prometheus to scrape.
func Serve(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, mux)
}
