package main

import (
	"context"
	"database/sql"
	"reflect"
)

// DB interface includes methods required for your database operations.
type DB interface {
	Close() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (RowScanner, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (RowScanner, error)
}

// SqlDB wraps a sql.DB and implements DB.
type SqlDB struct {
	*sql.DB
}

func (db *SqlDB) Query(query string, args ...interface{}) (RowScanner, error) {
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return &SqlRows{rows}, nil
}

func (db *SqlDB) QueryContext(ctx context.Context, query string, args ...interface{}) (RowScanner, error) {
	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SqlRows{rows}, nil
}

// SqlRows wraps sql.Rows and implements RowScanner.
type SqlRows struct {
	*sql.Rows
}

// DBFactory interface includes a method to generate new DB instances.
type DBFactory interface {
	New(connStr string) (DB, error)
}

// SqlDBFactory creates new SqlDB instances.
type SqlDBFactory struct{}

func (f *SqlDBFactory) New(connStr string) (DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &SqlDB{db}, nil
}

// MockDBFactory creates mock DB instances.
type MockDBFactory struct {
	openError error
	conn      *MockSQLConn
}

func (m *MockDBFactory) New(connStr string) (DB, error) {
	if m.openError != nil {
		return nil, m.openError
	}
	return &MockDB{conn: m.conn}, nil
}

// MockDB holds the mock implementation of DB for testing.
type MockDB struct {
	conn *MockSQLConn
}

func (m *MockDB) Close() error {
	return nil
}

func (m *MockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return m.conn.ExecContext(context.Background(), query, args...)
}

func (m *MockDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return m.conn.ExecContext(ctx, query, args...)
}

func (m *MockDB) Query(query string, args ...interface{}) (RowScanner, error) {
	return m.conn.QueryContext(context.Background(), query, args...)
}

func (m *MockDB) QueryContext(ctx context.Context, query string, args ...interface{}) (RowScanner, error) {
	return m.conn.QueryContext(ctx, query, args...)
}

// RowScanner interface includes methods for scanning rows of a result.
type RowScanner interface {
	Close() error
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}

// MockSQLConn mocks sql.DB for testing.
type MockSQLConn struct {
	execError  error
	queryError error
	rows       RowScanner
}

func (m *MockSQLConn) Close() error {
	return nil
}

func (m *MockSQLConn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, m.execError
}

func (m *MockSQLConn) QueryContext(ctx context.Context, query string, args ...interface{}) (RowScanner, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}
	return m.rows, nil
}

// MockSQLRows mocks sql.Rows for testing.
type MockSQLRows struct {
	scanError error
	rowsErr   error
	next      bool
	data      [][]interface{}
	current   int
}

func (m *MockSQLRows) Scan(dest ...interface{}) error {
	if m.scanError != nil {
		return m.scanError
	}

	if m.current < len(m.data) {
		row := m.data[m.current]
		for i, v := range row {
			val := reflect.ValueOf(v)
			if val.Kind() == reflect.Ptr {
				reflect.ValueOf(dest[i]).Elem().Set(val.Elem())
			} else {
				reflect.ValueOf(dest[i]).Elem().Set(val)
			}
		}
	}
	return nil
}

func (m *MockSQLRows) Next() bool {
	if m.current < len(m.data) {
		m.current++
		return true
	}
	return false
}

func (m *MockSQLRows) Err() error {
	return m.rowsErr
}

func (m *MockSQLRows) Close() error {
	return nil
}
