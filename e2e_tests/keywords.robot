*** Settings ***
Library     DatabaseLibrary
Library     OperatingSystem
Library     Process
Library     PrometheusLibrary    http://localhost:9612/metrics


*** Variables ***
${CONNECTION_STRING}        "postgresql://root@cockroach:26257/rowdy?sslmode=disable"
${CONNECTION_STRING_PG}     "postgresql://root:root@postgresql/rowdy?sslmode=disable"


*** Keywords ***
Connect To Cockroach
    Connect To Database Using Custom Params    psycopg2    db_connect_string=${CONNECTION_STRING}

Connect To PostgreSQL
    Connect To Database Using Custom Params    psycopg2    db_connect_string=${CONNECTION_STRING_PG}

Setup Test Database
    Execute SQL String    DROP DATABASE IF EXISTS e2e_test
    Execute SQL String    CREATE DATABASE e2e_test

Common Test Table Setup
    Execute SQL String    DROP TABLE IF EXISTS mekmitasdi;
    Execute SQL String    CREATE TABLE mekmitasdi (dier TEXT, kangoeroe INT, PRIMARY KEY(dier));
    Execute SQL String    INSERT INTO mekmitasdi VALUES ('goat', 42);
    Execute SQL String    CREATE INDEX ON mekmitasdi (dier, kangoeroe);

Setup Test Table
    Execute SQL String    USE e2e_test;
    Common Test Table Setup
    Execute SQL String    CREATE STATISTICS dankie ON dier FROM mekmitasdi;

Setup Test Table PostgreSQL
    Execute SQL String    SET SESSION enable_seqscan TO off;
    Common Test Table Setup
    Execute SQL String    ANALYZE mekmitasdi;

Start App
    [Arguments]    ${args}
    Create Directory    results/coverage/
    ${res}=    Start Process    ./app ${args}    shell=yes    env:GOCOVERDIR=results/coverage/
    RETURN    ${res}

Expect App Return
    [Arguments]    ${expected}    ${args}    &{env_vars}
    Create Directory    results/coverage/
    ${res}=    Run Process    ./app ${args}    shell=yes    env:GOCOVERDIR=results/coverage/    &{env_vars}
    Should Be Equal As Integers    ${res.rc}    ${expected}
    RETURN    ${res}
