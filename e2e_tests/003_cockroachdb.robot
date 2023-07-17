*** Settings ***
Resource    keywords.robot


*** Test Cases ***
Connect to database
    Connect To Cockroach
    Setup Test Database

Empty database should ok
    Start App    -connstr ${CONNECTION_STRING} -db e2e_test -request_limit 1
    Poll And Parse
    ${res}=    Wait For Process    timeout=5
    Log Many    ${res.stdout}    ${res.stderr}    ${res.rc}
    Expect Metric By Selector    stat_query_count{}    1
    Expect Metric By Selector    stat_error_query_total{}    0

Database with data should be ok
    Setup Test Table
    Start App    -connstr ${CONNECTION_STRING} -db e2e_test -request_limit 1
    Sleep    10
    ${vars}=    Poll And Parse
    Log    ${vars}
    ${res}=    Wait For Process    timeout=5
    Log Many    ${res.stdout}    ${res.stderr}    ${res.rc}
    Expect Metric By Selector    stat_query_count{}    1
    Expect Metric By Selector    stat_error_query_total{}    0
    Expect Metric By Selector    table_rows{db="e2e_test",schema="public",table_name="mekmitasdi"}    1

Indices should be used
    Setup Test Table
    Start App    -connstr ${CONNECTION_STRING} -db e2e_test -request_limit 1
    Execute SQL String    SELECT * FROM mekmitasdi WHERE dier='lekker' AND kangoeroe=2;
    Execute SQL String    ANALYZE mekmitasdi;
    Sleep    10s
    ${vars}=    Poll And Parse
    Log    ${vars}
    ${res}=    Wait For Process    timeout=5
    Log Many    ${res.stdout}    ${res.stderr}    ${res.rc}
    Expect Metric By Selector    index_reads{name="mekmitasdi_dier_kangoeroe_idx"}    1
