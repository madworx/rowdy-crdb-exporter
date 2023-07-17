package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	c                  *cache.Cache
	cacheTTL           time.Duration
	connStr            string
	dbName             string
	dbType             string
	gitCommit          string
	gitTag             string
	listenAddress      string
	requestCount       uint64
	requestLimit       int
	server             *http.Server
	staleReadThreshold time.Duration
)

func init() {
	c = cache.New(time.Second, 10*time.Minute)
	requestCount = 0
}

// Regex to match valid identifiers. Adjust as needed.
var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9\_]+$`).MatchString

func sanitizeIdentifier(identifier string) (string, error) {
	if !isAlphaNumeric(identifier) {
		return "", errors.New("invalid identifier")
	}
	return identifier, nil
}

func checkRequests() {
	if requestLimit > 0 {
		requests := atomic.AddUint64(&requestCount, 1)
		if int(requests) >= requestLimit {
			go func() {
				if err := server.Shutdown(context.Background()); err != nil {
					log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
				}
			}()
		}
	}
}

func updateIndicesMetrics(dbFactory DBFactory) {
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), staleReadThreshold)
	defer cancel()

	start := time.Now()
	doneChan := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		db, err := dbFactory.New(connStr)
		if err != nil {
			log.Println("Failed to open connection:", err)
			queryErrorsCounter.Inc()
			return
		}
		defer db.Close()

		var rows RowScanner

		switch dbType {
		case "cockroachdb":
			rows, err = queryIndices(db, dbName)
		case "postgres":
			rows, err = queryIndicesPostgreSQL(db, dbName)
		default:
			panic(fmt.Sprintf("Assertion failed: Invalid database type: [%s]", dbType))
		}

		if err != nil {
			log.Println("Failed to execute query:", err)
			queryErrorsCounter.Inc()
			return
		}
		defer rows.Close()

		for rows.Next() {
			var schema, table, indexName, indexType, indexUnique string
			var numUsed float64
			if err := rows.Scan(&schema, &table, &indexName, &indexType, &indexUnique, &numUsed); err != nil {
				log.Println("Failed to scan row:", err)
				queryErrorsCounter.Inc()
			} else {
				indexReadCounter.WithLabelValues(dbName, schema, table, indexName, indexType, indexUnique).Set(numUsed)
			}
		}

		if err := rows.Err(); err != nil {
			log.Println("Error fetching rows:", err)
			queryErrorsCounter.Inc()
		}

		queryHistogramIndices.Observe(time.Since(start).Seconds())
		doneChan <- struct{}{}
	}()

	// Wait for the signal from the goroutine or the context timeout
	select {
	case <-ctx.Done():
		// If the context is done (it took more than 3 seconds),
		// update the cache timeout and return a stale read
		c.Set("metrics", true, cache.DefaultExpiration)
		queryStaleReadsCounter.Inc()
		return
	case <-doneChan:
		// If a signal arrives from the channel, the query is done and the metrics are updated
	}

	// Wait for the goroutine to finish
	wg.Wait()
}

func updateMetrics(dbFactory DBFactory) {
	// Create a context that will be cancelled if it takes more than staleReadThreshold
	ctx, cancel := context.WithTimeout(context.Background(), staleReadThreshold)
	defer cancel()

	start := time.Now()

	// Use a WaitGroup to know when the goroutine finishes its execution
	var wg sync.WaitGroup
	wg.Add(1)

	// This channel will receive a signal from the goroutine when the query is done
	doneChan := make(chan struct{})

	// This function will be executed in a goroutine
	go func() {
		defer wg.Done()
		defer cancel()

		db, err := dbFactory.New(connStr)
		if err != nil {
			log.Println("Failed to open connection:", err)
			queryErrorsCounter.Inc()
			return
		}
		defer db.Close()

		var rows RowScanner

		switch dbType {
		case "cockroachdb":
			rows, err = queryTables(db, dbName)
		case "postgres":
			rows, err = queryTablesPostgreSQL(db, dbName)
		default:
			panic(fmt.Sprintf("Assertion failed: Invalid database type: [%s]", dbType))
		}

		if err != nil {
			log.Println("Failed to execute query:", err)
			queryErrorsCounter.Inc()
			return
		}
		defer rows.Close()

		for rows.Next() {
			var schema, tableName string
			var size, estimatedRowCount float64
			if err := rows.Scan(&schema, &tableName, &size, &estimatedRowCount); err != nil {
				log.Println("Failed to scan row:", err)
				queryErrorsCounter.Inc()
			} else {
				tableRowsGauge.WithLabelValues(dbName, schema, tableName).Set(estimatedRowCount)
				tableSizeGauge.WithLabelValues(dbName, schema, tableName).Set(size)
			}
		}

		if err := rows.Err(); err != nil {
			log.Println("Error fetching rows:", err)
			queryErrorsCounter.Inc()
		}

		queryHistogram.Observe(time.Since(start).Seconds())
		doneChan <- struct{}{}
	}()

	// Wait for the signal from the goroutine or the context timeout
	select {
	case <-ctx.Done():
		// If the context is done (it took more than 3 seconds),
		// update the cache timeout and return a stale read
		c.Set("metrics", true, cache.DefaultExpiration)
		queryStaleReadsCounter.Inc()
		return
	case <-doneChan:
		// If a signal arrives from the channel, the query is done and the metrics are updated
	}

	// Wait for the goroutine to finish
	wg.Wait()
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	if _, found := c.Get("metrics"); !found {
		updateMetrics(&SqlDBFactory{})
	}

	updateIndicesMetrics(&SqlDBFactory{})

	promhttp.Handler().ServeHTTP(w, r)
	checkRequests()
}

func main() {
	flag.StringVar(&connStr, "connstr", os.Getenv("CONNSTR"), "Database connection string (environment variable: CONNSTR)")
	flag.StringVar(&dbName, "db", os.Getenv("DB"), "Database name (environment variable: DB)")
	flag.IntVar(&requestLimit, "request_limit", 0, "The maximum number of requests the server will accept before shutting down")
	flag.StringVar(&listenAddress, "listen_address", os.Getenv("LISTEN_ADDRESS"), "Address to listen on (environment variable: LISTEN_ADDRESS)")
	flag.StringVar(&dbType, "dbtype", "cockroachdb", "Database type: cockroachdb or postgres (default: cockroachdb)")

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

	staleReadThresholdStr := os.Getenv("STALE_READ_THRESHOLD")
	if staleReadThresholdStr != "" {
		var err error
		staleReadThreshold, err = time.ParseDuration(staleReadThresholdStr)
		if err != nil {
			log.Fatal("Invalid STALE_READ_THRESHOLD, must be a valid Go duration string: ", err)
		}
	} else {
		staleReadThreshold = time.Duration(3) * time.Second
	}
	flag.DurationVar(&staleReadThreshold, "stale_read_threshold", time.Second*3, "Time for executing the SQL query before stale data is returned (environment variable: STALE_READ_THRESHOLD)")

	flag.Parse()

	if _, err := sanitizeIdentifier(dbName); err != nil {
		log.Fatal("Invalid database name: ", err)
	}

	if dbType != "cockroachdb" && dbType != "postgres" {
		log.Fatal("Invalid database type. Must be 'cockroachdb' or 'postgres'")
	}

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

	server = &http.Server{
		Addr:    listenAddress,
		Handler: mux,
	}
	server.ListenAndServe()
}
