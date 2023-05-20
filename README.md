# Rowdy - CockroachDB row count & size exporter for Prometheus

[![Test and coverage](https://github.com/madworx/rowdy-crdb-exporter/actions/workflows/test.yml/badge.svg)](https://github.com/madworx/rowdy-crdb-exporter/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/madworx/rowdy-crdb-exporter/branch/main/graph/badge.svg?token=4EMXW0RRKU)](https://codecov.io/gh/madworx/rowdy-crdb-exporter)

Rowdy is a tool that connects to a CockroachDB database, fetches information about the (estimated) number of rows in each table, and the disk space those tables consume, then exports this data to Prometheus.

*⚠️ Disclaimer: CockroachDB themselves advise against running this type of tool in production environments. The tables that this tool queries are considered experimental and may change in the future. They also consume a considerable amount of resources when queried.*

## Command Line Flags

### `-connstr` : (Environment Variable `CONNSTR`).
The connection string to connect to your CockroachDB instance.

### `-db` : (Environment Variable `DB`).
The name of the database you are connecting to.

### `-listen_address` : (Environment Variable `LISTEN_ADDRESS`).
The address on which the exporter will listen. If not specified, defaults to :9612.

### `-cache_ttl` : (Environment Variable `CACHE_TTL`).
The duration that data should be kept in the cache. This should be a valid Go duration string. If not specified, defaults to 5m (5 minutes).

## Releases

You can download the compiled binaries of each version of Rowdy from the Github [releases page](https://github.com/madworx/rowdy-crdb-exporter/releases/). Each release contains pre-built binaries for various platforms.

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

Rowdy is released under the [MIT](LICENSE).