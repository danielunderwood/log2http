# log2http

Matches log lines to trigger HTTP requests. A super hacky alternative to ELK or loki.

## Supported HTTP Targets
- Discord

## Usage

### Binary

```shell
$ log2http -file FILENAME -expression "^[Rr]egex$" -url https://discordapp.com/...
$ # Or supply URL via environment
$ export URL="https://discordapp.com/..."
$ log2http -file FILENAME -expression "^[Rr]egex$
```

## Development

The development environment is currently set up with nix flakes.

### Development Environment
```shell
$ nix develop
```

### Build Binary
```shell
$ nix build
```

### Build and Run Binary
```shell
$ nix run .#log2http -- -file filename
```
