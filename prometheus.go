package main

import (
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	info = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rowdy_info",
			Help: "Information about the Rowdy build.",
		},
		[]string{"commit", "version"},
	)
	tableRowsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "table_rows",
			Help: "Estimated row count",
		},
		[]string{"db", "schema", "table_name"},
	)
	indexReadCounter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "index_reads",
			Help: "Total number of index reads",
		},
		[]string{"db", "schema", "table", "name", "type", "unique"},
	)
	tableSizeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "table_size",
			Help: "Consumed disk space",
		},
		[]string{"db", "schema", "table_name"},
	)
	queryHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "stat_query",
			Help:    "Time taken to execute the SQL query",
			Buckets: prometheus.LinearBuckets(0, 0.2, 10),
		},
	)
	queryHistogramIndices = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "stat_query_indices",
			Help:    "Time taken to execute the SQL query",
			Buckets: prometheus.LinearBuckets(0, 0.2, 10),
		},
	)
	queryErrorsCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "stat_error_query",
			Help: "Number of query errors encountered",
		},
	)
	queryStaleReadsCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "stat_stale_reads",
			Help: "Number of stale reads returned",
		},
	)
)

func RegisterPrometheusMetrics() {
	metrics := []prometheus.Collector{
		indexReadCounter,
		info,
		queryErrorsCounter,
		queryHistogram,
		queryHistogramIndices,
		queryStaleReadsCounter,
		tableRowsGauge,
		tableSizeGauge,
	}

	for _, metric := range metrics {
		prometheus.Unregister(metric)
	}

	// re-register the metrics
	prometheus.MustRegister(tableRowsGauge, tableSizeGauge, queryHistogram,
		queryErrorsCounter, queryStaleReadsCounter, info, indexReadCounter,
		queryHistogramIndices)

	// re-apply any required initial states
	info.WithLabelValues(gitCommit, gitTag).Set(1)
}

func init() {
	RegisterPrometheusMetrics()
}
