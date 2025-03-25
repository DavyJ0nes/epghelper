# Embedded Postgres Helper

> Lightweight CLI helper for managing databases created with the awesome project
> https://github.com/fergusstrange/embedded-postgres

## Installation

To install the `epghelper` CLI tool, ensure you have Go installed on your
system. You can download it from
[the official Go website](https://golang.org/dl/).

Once Go is installed, you can install `epghelper` using the following command:

```sh
go install github.com/davyj0nes/epghelper@latest
```

This command will download and install the `epghelper` binary into your
`$GOPATH/bin` directory. Make sure that `$GOPATH/bin` is included in your
system's `PATH` environment variable so you can run `epghelper` from anywhere in
your terminal.

### Verify Installation

To verify that the installation was successful, run:

```sh
epghelper --help
```

This should display the help information for the `epghelper` CLI tool,
confirming that it is installed and accessible.

## Commands

### `epghelper ls`

- **Description**: Lists all existing databases.
- **Usage**: `epghelper ls`
- **Output**: Displays a table with the port, creation date, and size of each
  database.

```
❯ epghelper ls
┌───────┬─────────────────────┬─────────┐
│ PORT  │ CREATED AT          │ SIZE    │
├───────┼─────────────────────┼─────────┤
│ 57720 │ 2025-03-24 11:00:53 │ 40.8 MB │
├───────┼─────────────────────┼─────────┤
│       │ TOTAL               │ 40.8 MB │
└───────┴─────────────────────┴─────────┘
```

### `epghelper rm`

- **Description**: Removes a specified database or all databases.
- **Usage**:
  - Remove a specific database: `epghelper rm <port>`
  - Remove all databases: `epghelper rm --all`
- **Flags**:
  - `--all`, `-a`: Remove all databases.

### `epghelper connect`

- **Description**: Connects to a specified database or the latest created
  database.
- **Usage**:
  - Connect to a specific database: `epghelper connect <port>`
  - Connect to the latest database: \`epghelper connect --latest
- **Flags**:
  - `--latest`, `-l`: Connect to the latest created database.
