package main

import (
	"fmt"

	_ "github.com/lib/pq"
)

func queryTables(db DB, dbName string) (RowScanner, error) {
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

func queryIndices(db DB, dbName string) (RowScanner, error) {
	stmt := fmt.Sprintf(`
	SELECT t.schema_name, ti.descriptor_name as table_name,
		   ti.index_name, ti.index_type,
		   ti.is_unique, total_reads
	  FROM %[1]s.crdb_internal.index_usage_statistics us
	  JOIN %[1]s.crdb_internal.table_indexes ti
		ON us.index_id = ti.index_id
	   AND us.table_id = ti.descriptor_id
	  JOIN %[1]s.crdb_internal.tables t
		ON ti.descriptor_id = t.table_id;`, dbName)
	return db.Query(stmt)
}
