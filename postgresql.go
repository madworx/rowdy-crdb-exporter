package main

import (
	_ "github.com/lib/pq"
)

func queryTablesPostgreSQL(db DB, dbName string) (RowScanner, error) {
	return db.Query(`
        SELECT
            schemaname AS namespace,
            relname AS table_name,
            pg_total_relation_size(schemaname || '.' || relname) AS size,
            n_live_tup AS rows
        FROM
            pg_stat_user_tables;
    `)
}

func queryIndicesPostgreSQL(db DB, dbName string) (RowScanner, error) {
	return db.Query(`
	SELECT
		n.nspname AS schema_name, t.relname AS table_name,
		i.relname AS index_name,
		CASE
			WHEN ic.indisprimary THEN 'primary'
			ELSE 'secondary'
		END AS index_type,
		ic.indisunique AS is_unique,
		pg_stat_get_numscans(i.oid) AS stat_total_number_of_reads
	FROM
		pg_class t, pg_class i, pg_index ic,  pg_namespace n
	WHERE
		ic.indrelid = t.oid AND ic.indexrelid = i.oid AND
		n.oid = t.relnamespace AND t.relkind = 'r' AND
		n.nspname NOT LIKE 'pg_%' AND n.nspname != 'information_schema';
`)
}
