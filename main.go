package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
			Name: "crdb_table_rows",
			Help: "Estimated row count",
		},
		[]string{"db", "schema", "table_name"},
	)
	tableSizeGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "crdb_table_size",
			Help: "Consumed disk space",
		},
		[]string{"db", "schema", "table_name"},
	)
	queryHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "crdb_query",
			Help:    "Time taken to execute the SQL query",
			Buckets: prometheus.LinearBuckets(0, 0.2, 10),
		},
	)
	connectErrorsCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "crdb_error_connect",
			Help: "Number of connection errors encountered",
		},
	)
	queryErrorsCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "crdb_error_query",
			Help: "Number of query errors encountered",
		},
	)
)

var (
	c             *cache.Cache
	cacheTTL      time.Duration
	connStr       string
	dbName        string
	gitCommit     string
	gitTag        string
	listenAddress string
)

func init() {
	prometheus.MustRegister(tableRowsGauge, tableSizeGauge, queryHistogram, connectErrorsCounter, queryErrorsCounter, info)
	info.WithLabelValues(gitCommit, gitTag).Set(1)
	c = cache.New(time.Second, 10*time.Minute)
}

func queryTables(db *sql.DB, dbName string) (*sql.Rows, error) {
	_, err := db.Exec(`USE $1`, dbName)
	if err != nil {
		return nil, err
	}

	return db.Query(`
	SELECT
		size.namespace,
		size.table_name, 
		size.size AS size,
		rows.rows AS rows
	FROM
		(SELECT schema_name AS namespace, table_name, SUM(range_size) AS size 
			FROM crdb_internal.ranges 
			WHERE database_name = $1
			GROUP BY namespace, table_name) AS size
	LEFT JOIN 
		(SELECT stats.table_name, 
			pg_namespace.nspname AS namespace,
			stats.estimated_row_count AS rows
		FROM crdb_internal.table_row_statistics AS stats, pg_class, pg_namespace
			WHERE pg_class.relnamespace=pg_namespace.oid
				AND pg_class.oid=stats.table_id
				AND nspname NOT IN ('crdb_internal', 'information_schema', 'pg_catalog', 'pg_extension')
		) AS rows 
	ON size.namespace=rows.namespace AND size.table_name = rows.table_name
`, dbName)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	if _, found := c.Get("metrics"); !found {
		start := time.Now()
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			connectErrorsCounter.Inc()
			log.Println("Failed to open connection:", err)
		} else {
			defer db.Close()
			rows, err := queryTables(db, dbName)
			if err != nil {
				queryErrorsCounter.Inc()
				log.Println("Failed to execute query:", err)
			} else {
				defer rows.Close()
				for rows.Next() {
					var schema, tableName string
					var size, estimatedRowCount float64
					if err := rows.Scan(&schema, &tableName, &size, &estimatedRowCount); err != nil {
						queryErrorsCounter.Inc()
						log.Println("Failed to scan row:", err)
					} else {
						tableRowsGauge.WithLabelValues(dbName, schema, tableName).Set(estimatedRowCount)
						tableSizeGauge.WithLabelValues(dbName, schema, tableName).Set(size)
					}
				}
				if err := rows.Err(); err != nil {
					queryErrorsCounter.Inc()
					log.Println("Error fetching rows:", err)
				}
				queryHistogram.Observe(time.Since(start).Seconds())
				c.Set("metrics", true, cache.DefaultExpiration)
			}
		}
	}
	promhttp.Handler().ServeHTTP(w, r)
}

func main() {
	flag.StringVar(&connStr, "connstr", os.Getenv("CONNSTR"), "Database connection string (environment variable: CONNSTR)")
	flag.StringVar(&dbName, "db", os.Getenv("DB"), "Database name (environment variable: DB)")
	cacheTTLStr := os.Getenv("CACHE_TTL")

	if cacheTTLStr != "" {
		var err error
		cacheTTL, err = time.ParseDuration(cacheTTLStr)
		if err != nil {
			log.Fatal("Invalid CACHE_TTL, must be a valid Go duration string: ", err)
		}
	} else {
		cacheTTL = time.Duration(5) * time.Minute
	}
	flag.DurationVar(&cacheTTL, "cache_ttl", cacheTTL, "Cache TTL (environment variable: CACHE_TTL)")
	flag.StringVar(&listenAddress, "listen_address", os.Getenv("LISTEN_ADDRESS"), "Address to listen on (environment variable: LISTEN_ADDRESS)")
	flag.Parse()
	if listenAddress == "" {
		listenAddress = ":9612" // Default port
	}
	if connStr == "" || dbName == "" {
		log.Fatal("Database connection string and name must be provided via command line arguments or environment variables")
	}
	c = cache.New(cacheTTL, 10*time.Minute)

	log.Printf("Rowdy - CockroachDB table rows/size statistics "+
		"exporter for Prometheus. (git:%s version:%s)\n",
		gitCommit, gitTag)
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", metricsHandler)
	log.Fatal(http.ListenAndServe(listenAddress, mux))
}
