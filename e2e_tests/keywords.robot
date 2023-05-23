*** Settings ***
Library         Process
Library         DatabaseLibrary
Library         PrometheusLibrary.py  http://localhost:9612/metrics

*** Variables ***
${CONNECTION_STRING}     "postgresql://root@cockroach:26257/rowdy?sslmode=disable"
${CONNECTION_STRING_PG}  "postgresql://root:root@postgresql/rowdy?sslmode=disable"

*** Keywords ***
Connect To Cockroach
    Connect To Database Using Custom Params    psycopg2   db_connect_string=${CONNECTION_STRING}

Connect To PostgreSQL
    Connect To Database Using Custom Params    psycopg2   db_connect_string=${CONNECTION_STRING_PG}

Setup Test Database
    Execute SQL String     DROP DATABASE IF EXISTS e2e_test
    Execute SQL String     CREATE DATABASE e2e_test

Setup Test Table
    Execute SQL String     USE e2e_test
    Execute SQL String     CREATE TABLE mekmitasdi (dier TEXT)
    Execute SQL String     INSERT INTO mekmitasdi VALUES ('goat')
    Execute SQL String     CREATE STATISTICS dankie ON dier FROM mekmitasdi

Setup Test Table PostgreSQL
    Execute SQL String     DROP TABLE IF EXISTS mekmitasdi
    Execute SQL String     CREATE TABLE mekmitasdi (dier TEXT)
    Execute SQL String     INSERT INTO mekmitasdi VALUES ('goat')
    Execute SQL String     ANALYZE mekmitasdi

Start App
    [Arguments]       ${args}    
    ${res}=           Start Process   ./app ${args}   shell=yes   env:GOCOVERDIR=.
    [Return]          ${res}

Expect App Return
    [Arguments]       ${expected}  ${args}    
    ${res}=           Run Process    ./app ${args}        shell=yes   env:GOCOVERDIR=.
    Should Be Equal As Integers    ${res.rc}    ${expected}
