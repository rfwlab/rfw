# rfw-cli

`rfw-cli` is the official command-line interface (CLI) tool for the **rfw** framework. It allows you to create, build, and run **rfw** projects from the command line.

## Installation

Ensure you have Go installed on your machine. Then, install `rfw-cli` with the following command:

```bash
go install github.com/rfwlab/rfw-cli@latest
```

## Usage

To create a new **rfw** project, run the following command:

```bash
rfw-cli init github.com/username/project-name
```

## Server

To start the **rfw** server, run the following command:

```bash
rfw-cli dev
```

### Flags

Two flags are available for the `dev` command:

- `--port XXXX`: specify the port number for the server (default is 8080, replace `XXXX` with the desired port number)
- `--host`: to expose the server to the network

For example

```bash
rfw-cli dev --port 3000 --host
```
