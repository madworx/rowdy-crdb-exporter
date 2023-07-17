*** Settings ***
Resource    keywords.robot


*** Test Cases ***
No configuration should fail
    Create Directory    results/coverage/
    ${res}=    Run Process    ./app    env:GOCOVERDIR=results/coverage
    Should Be Equal As Integers    ${res.rc}    1

Invalid configuration should fail
    Expect App Return    2    -connstr 'v' -db db123 -cache_ttl snigel
    Expect App Return    2    -connstr 'v' -db db123 -cache_ttl_indices snigel
    Expect App Return    1    -connstr 'v' -db db123 -dbtype snigel
    Expect App Return    1    -connstr 'v'
    Expect App Return    1    -db db123

Invalid cache ttl via environment should fail
    ${res}=    Expect App Return    1    -connstr 'v' -db db123    env:CACHE_TTL=snigel
    Should Contain    ${res.stderr}    Invalid CACHE_TTL

Invalid indices cache ttl via environment should fail
    ${res}=    Expect App Return    1    -connstr 'v' -db db123    env:CACHE_TTL_INDICES=snigel
    Should Contain    ${res.stderr}    Invalid CACHE_TTL_INDICES

Invalid stale read ttl via environment should fail
    ${res}=    Expect App Return    1    -connstr 'v' -db db123    env:STALE_READ_THRESHOLD=snigel
    Should Contain    ${res.stderr}    Invalid STALE_READ_THRESHOLD

Valid configuration with invalid arguments should err but not fail
    Start App    -connstr 'v' -db db123 -request_limit 1
    Poll And Parse
    ${res}=    Wait For Process    timeout=5
    Log Many    ${res.stdout}    ${res.stderr}    ${res.rc}
    Expect Metric By Selector    stat_error_query_total{}    2

Invalid dbname should fail
    Expect App Return    1    -connstr 'v' -db 'db1;DROP TABLE f00;23' -request_limit 1
