# Rowdy - CockroachDB row count & size exporter for Prometheus

[![Test and coverage](https://github.com/madworx/rowdy-crdb-exporter/actions/workflows/test.yml/badge.svg)](https://github.com/madworx/rowdy-crdb-exporter/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/madworx/rowdy-crdb-exporter)](https://goreportcard.com/report/github.com/madworx/rowdy-crdb-exporter)
[![codecov](https://codecov.io/gh/madworx/rowdy-crdb-exporter/branch/main/graph/badge.svg?token=4EMXW0RRKU)](https://codecov.io/gh/madworx/rowdy-crdb-exporter)

Rowdy is a tool that connects to a CockroachDB or PostgreSQL database, fetches information about the (estimated) number of rows in each table, the disk space those tables consume, as well as information in index usage and then exports this data to Prometheus. By default, it listens on `0.0.0.0:9612`.

*⚠️ Disclaimer: CockroachDB themselves strongly advise against running this type of tool in production environments. The tables this tool queries are considered internal and experimental, and may change in the future. Also, be prepared for it to consume a considerable amount of resources when queried. Don't say I didn't warn you, risk-taker!*

## Command Line Flags

### `-connstr`

The connection string to connect to your CockroachDB instance.  (Environment Variable `CONNSTR`)

### `-db`

The name of the database you are connecting to. (Environment Variable `DB`)

### `-listen_address`

The address on which the exporter will listen. If not specified, defaults to `:9612`.  (Environment Variable `LISTEN_ADDRESS`)

### `-cache_ttl`

The duration that data should be kept in the cache. This should be a valid Go duration string. If not specified, defaults to 5m (5 minutes). (Environment Variable `CACHE_TTL`)

### `-stale_read_threshold`

The maximum duration statistics gathering SQL queries may take before the query is continued in the background and stale data is returned to the requestor. (Environment variable `STALE_READ_THRESHOLD`)

## Running as a Systemd Service

If you want to run Rowdy as a service, you can create a Systemd service file:

```
[Unit]
Description=Rowdy CockroachDB Exporter
After=network.target

[Service]
ExecStart=/path/to/rowdy -connstr your_conn_str -db your_db_name
User=rowdy
Restart=always

[Install]
WantedBy=multi-user.target
```

Replace `/path/to/rowdy` with the actual path to the `rowdy` binary,
`your_conn_str` with your connection string, and `your_db_name` with your database name.

To install the service:

1. Save the service file to `/etc/systemd/system/rowdy.service`.
2. Run `systemctl enable rowdy` to enable the service.
3. Run `systemctl start rowdy` to start the service.

## Releases

You can download the compiled binaries of each version of Rowdy from the Github [releases page](https://github.com/madworx/rowdy-crdb-exporter/releases/). Each release contains pre-built binaries for various platforms.

After downloading, it is strongly recommended to verify the binary using the SHA256SUM file provided in each release. This helps ensure the integrity and authenticity of the binary you downloaded. Here's how you can do it:

1. Download both the binary and the corresponding `SHA256SUM.txt` file.
2. Run the command `sha256sum -c SHA256SUM.txt` in the terminal.
3. The command should output a message saying that the binary is OK.

If the checksum verification fails, do not use the downloaded binary. It means that the binary may have been tampered with or there was an error in the download.

## Build From Source

If you prefer, you can build the application yourself:

1. Clone the repository.
2. Open a terminal in the repository root directory.
3. Run the make command.

The compiled binaries will be placed in the `dist/` directory.

## How to Contribute

Contributions are most welcome! The project is set up to work with a VSCode dev container, so you should be able to just open the project in VSCode and everything should work.

Follow these steps to contribute:

1. Fork the repository.
2. Open the project in VSCode.
3. Make changes to the code.
4. Commit your changes and push to your fork.
5. Open a pull request on the original repository from your fork.

### License

Rowdy is released under the [MIT](LICENSE) license.