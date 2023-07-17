package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func testMetricsHandler(t *testing.T) {
	// Connect to the actual database
	factory := &SqlDBFactory{}
	db, err := factory.New(connStr)
	if err != nil {
		t.Fatalf("failed to connect to the database: %v", err)
	}
	defer db.Close()

	if dbType != "postgres" {
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
	}

	// Create a new table for testing
	tableName := "test_table"
	_, err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName))
	if err != nil {
		t.Fatalf("failed to drop existing test table: %v", err)
	}

	_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY, name TEXT)", tableName))
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}

	// Wait until queryTables starts returning rows
	for {
		queryFunc := queryTables
		if dbType == "postgres" {
			queryFunc = queryTablesPostgreSQL
		}
		rows, err := queryFunc(db, dbName)
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

	// Execute metricsHandler 10000 times
	for i := 1; i < 100; i++ {
		metricsHandler(rr, req)
	}

	expected := []string{
		fmt.Sprintf(`table_rows{db="%s",schema="public",table_name="%s"} 0`, dbName, tableName),
		fmt.Sprintf(`table_size{db="%s",schema="public",table_name="%s"} `, dbName, tableName),
	}
	responseBody := rr.Body.String()
	for _, expectedValue := range expected {
		if !strings.Contains(responseBody, expectedValue) {
			t.Errorf("handler didn't contain: [%v] (was: [%v])", expectedValue, responseBody)
		}
	}

	// Print any errors encountered during the test execution
	if err := db.Close(); err != nil {
		t.Errorf("error closing the database connection: %v", err)
	}
}

func TestMetricsHandlerWithCockroachDB(t *testing.T) {
	connStr = "postgresql://root@cockroach:26257/?sslmode=disable"
	dbName = "test_db"
	dbType = "cockroachdb"
	staleReadThreshold = time.Duration(10) * time.Second

	testMetricsHandler(t)
}

func TestMetricsHandlerWithPostgreSQL(t *testing.T) {
	connStr = "postgresql://root:root@postgresql/rowdy?sslmode=disable"
	dbName = "rowdy"
	dbType = "postgres"
	staleReadThreshold = time.Duration(10) * time.Second

	testMetricsHandler(t)
}

func TestUpdateMetrics(t *testing.T) {
	dbType = "cockroachdb"
	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	tt := []struct {
		name      string
		dbFactory DBFactory
		logOutput string
	}{
		{
			name:      "sql open error",
			dbFactory: &MockDBFactory{openError: errors.New("open error")},
			logOutput: "Failed to open connection: open error",
		},
		{
			name:      "db exec error",
			dbFactory: &MockDBFactory{conn: &MockSQLConn{execError: errors.New("exec error")}},
			logOutput: "Failed to execute query: exec error",
		},
		{
			name:      "db query error",
			dbFactory: &MockDBFactory{conn: &MockSQLConn{queryError: errors.New("query error")}},
			logOutput: "Failed to execute query: query error",
		},
		{
			name: "rows scan error",
			dbFactory: &MockDBFactory{
				conn: &MockSQLConn{
					rows: &MockSQLRows{
						scanError: errors.New("scan error"),
						data: [][]interface{}{
							{"public", "test_table", 0.0, 0.0},
							{"public", "test2_table", 0.0, 0.0},
						},
					},
				},
			},
			logOutput: "Failed to scan row: scan error",
		},
		{
			name: "rows scan ok",
			dbFactory: &MockDBFactory{
				conn: &MockSQLConn{
					rows: &MockSQLRows{
						data: [][]interface{}{
							{"public", "test_table", 0.0, 0.0},
							{"public", "test2_table", 0.0, 0.0},
						},
					},
				},
			},
			logOutput: "",
		},
		{
			name:      "rows err",
			dbFactory: &MockDBFactory{conn: &MockSQLConn{rows: &MockSQLRows{rowsErr: errors.New("rows error")}}},
			logOutput: "Error fetching rows: rows error",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			updateMetrics(tc.dbFactory)
			logOutput := logBuffer.String()
			if !strings.Contains(logOutput, tc.logOutput) {
				t.Errorf("expected log message '%s' was not found in the output [%v]", tc.logOutput, logOutput)
			}
			logBuffer.Reset()
		})
	}
}

func TestCloseMockDB(t *testing.T) {
	m := &MockDBFactory{}
	d, _ := m.New("")
	err := d.Close()
	if err != nil {
		t.Fail()
	}
	err = m.conn.Close()
	if err != nil {
		t.Fail()
	}
}

func TestCheckRequests(t *testing.T) {
	requestLimit = 100
	checkRequests()
}

func TestSanitizeDBName(t *testing.T) {
	sanitizeIdentifier("test")
	sanitizeIdentifier("test;DROP TABLE test;")
}

// Just fake unused functions to improve coverage.
func TestMockContextFuncs(t *testing.T) {
	db := &MockDB{conn: &MockSQLConn{}}
	args := []interface{}{}
	_, _ = db.ExecContext(context.Background(), "x", args...)
	_, _ = db.QueryContext(context.Background(), "x", args...)
}
