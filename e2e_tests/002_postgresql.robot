*** Settings ***
Resource          keywords.robot

*** Test Cases ***
Empty database should ok
    Connect To PostgreSQL
    Start App                         -dbtype postgres -connstr ${CONNECTION_STRING_PG} -db e2e_test -request_limit 1
    Poll And Parse
    Wait For Process                  timeout=5
    Expect Metric By Selector         crdb_query_count{}          1
    Expect Metric By Selector         crdb_error_query_total{}    0

Database with data should be ok
    Setup Test Table PostgreSQL
    Start App                         -dbtype postgres -connstr ${CONNECTION_STRING_PG} -db e2e_test -request_limit 1
    Poll And Parse
    Wait For Process                  timeout=5
    Expect Metric By Selector         crdb_error_query_total{}    0
    Expect Metric By Selector         crdb_query_count{}          1
    Expect Metric By Selector         crdb_table_rows{db="e2e_test",schema="public",table_name="mekmitasdi"}    1
