package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsHandlerWithDatabase(t *testing.T) {
	connStr = "postgresql://root@cockroach:26257/rowdy?sslmode=disable"
	dbName = "test_db"

	// Connect to the actual database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to the database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		t.Fatalf("failed to drop test database: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("USE %s", dbName))
	if err != nil {
		t.Fatalf("failed to use test database: %v", err)
	}

	// Create a new table for testing
	tableName := "test_table"
	_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY, name TEXT)", tableName))
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	// Wait until queryTables starts returning rows
	for {
		rows, err := queryTables(db, dbName)
		if err != nil {
			t.Fatalf("failed to query tables: %v", err)
		}
		defer rows.Close()
		if rows.Next() {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}

	// Create http request and response writer
	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// Execute metricsHandler
	metricsHandler(rr, req)

	// Check the response body for the expected table name
	expected := fmt.Sprintf(`# HELP crdb_table_rows Estimated row count
# TYPE crdb_table_rows gauge
crdb_table_rows{db="%s",schema="public",table_name="%s"} 0
# HELP crdb_table_size Consumed disk space
# TYPE crdb_table_size gauge
crdb_table_size{db="%s",schema="public",table_name="%s"} 0`,
		dbName, tableName, dbName, tableName)
	// Check if rr.Body.String() contains the expected string
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	// Print any errors encountered during the test execution
	if err := db.Close(); err != nil {
		t.Errorf("error closing the database connection: %v", err)
	}
}
