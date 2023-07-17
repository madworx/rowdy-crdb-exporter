*** Settings ***
Resource    keywords.robot


*** Test Cases ***
Connect to database
    Connect To Cockroach
    Setup Test Database

Query TTL should be respected
    Setup Test Table
    Start App    -connstr ${CONNECTION_STRING} -db e2e_test -cache_ttl 2s -cache_ttl_indices 1s -request_limit 20
    FOR    ${i}    IN RANGE    0    20
        ${vars}=    Poll And Parse
        Sleep    0.5s
    END
    Log    ${vars}
    ${res}=    Wait For Process    timeout=5
    Log Many    ${res.stdout}    ${res.stderr}    ${res.rc}
    Expect Metric By Selector    stat_query_count{}    5
    Expect Metric By Selector    stat_query_indices_count{}    10
    Expect Metric By Selector    stat_error_query_total{}    0
