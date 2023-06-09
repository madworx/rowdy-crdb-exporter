*** Settings ***
Resource          keywords.robot

*** Test Cases ***
No configuration should fail
    ${res}=         Run Process    ./app
    Should Be Equal As Integers    ${res.rc}    1

*** Test Cases ***
Invalid configuration should fail
    Expect App Return     2        -connstr 'v' -db db123 -cache_ttl snigel
    Expect App Return     1        -connstr 'v'
    Expect App Return     1        -db db123

Invalid cache ttl via environment should fail
    ${args}=    Set Variable       -connstr 'v' -db db123
    ${res}=     Run Process        ./app ${args}  shell=yes   env:GOCOVERDIR=.   env:CACHE_TTL=snigel
    Should Be Equal As Integers    ${res.rc}      1
    Should Contain                 ${res.stderr}  Invalid CACHE_TTL

Valid configuration with invalid arguments should err but not fail
    Start App                      -connstr 'v' -db db123 -request_limit 1
    Poll And Parse
    Wait For Process               timeout=5
    Expect Metric By Selector      crdb_error_query_total{}    1

Invalid dbname should fail
    Expect App Return     1        -connstr 'v' -db 'db1;DROP TABLE f00;23' -request_limit 1

Empty database should ok
    Connect To Cockroach
    Setup Test Database
    Start App                      -connstr ${CONNECTION_STRING} -db e2e_test -request_limit 1
    Poll And Parse
    Wait For Process               timeout=5
    Expect Metric By Selector      crdb_query_count{}          1
    Expect Metric By Selector      crdb_error_query_total{}    0

Database with data should be ok
    Setup Test Table
    Start App                      -connstr ${CONNECTION_STRING} -db e2e_test -request_limit 1
    Sleep                          10
    Poll And Parse
    Wait For Process               timeout=5
    Expect Metric By Selector      crdb_query_count{}          1
    Expect Metric By Selector      crdb_error_query_total{}    0
    Expect Metric By Selector      crdb_table_rows{db="e2e_test",schema="public",table_name="mekmitasdi"}    1