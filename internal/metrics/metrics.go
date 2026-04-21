// Package metrics registers Prometheus metrics for the PhysicsJob operator.
// Import this package with a blank identifier to trigger registration:
//
//	import _ "github.com/KaranSinghDev/data-gravity-operator/internal/metrics"
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ReconcileTotal counts reconcile calls by outcome ("success" or "error").
	ReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "physjob_reconcile_total",
			Help: "Total PhysicsJob reconciliations partitioned by result.",
		},
		[]string{"result"},
	)

	// ResolvedTotal counts successful RSE resolutions by RSE name and scheduling policy.
	ResolvedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "physjob_resolved_total",
			Help: "Total dataset → RSE resolutions by RSE and scheduling policy.",
		},
		[]string{"rse", "policy"},
	)

	// ResolutionFailuresTotal counts RSE resolution failures by reason code.
	ResolutionFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "physjob_resolution_failures_total",
			Help: "Total dataset resolution failures by reason.",
		},
		[]string{"reason"},
	)

	// ReconcileDuration observes the wall-clock time of each reconcile call.
	ReconcileDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "physjob_reconcile_duration_seconds",
			Help:    "Duration of a single PhysicsJob reconcile call in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)

	// DataTransferAvoidedBytes accumulates the estimated bytes NOT moved over
	// the WAN because data-local scheduling placed compute at the storage site.
	DataTransferAvoidedBytes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "physjob_data_transfer_avoided_bytes",
			Help: "Estimated bytes of WAN transfer avoided by data-local scheduling.",
		},
		[]string{"rse"},
	)
)

func init() {
	ctrlmetrics.Registry.MustRegister(
		ReconcileTotal,
		ResolvedTotal,
		ResolutionFailuresTotal,
		ReconcileDuration,
		DataTransferAvoidedBytes,
	)
}
