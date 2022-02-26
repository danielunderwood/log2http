# log2http

Matches log lines to trigger HTTP requests. A super hacky alternative to ELK or loki.

## Supported HTTP Targets
- Discord

## Caveats

At the moment, the following edge cases apply:
- There is no knowledge of what lines trigger requests, so requests could be duplicated upon restart
- There is no retry mechanism, so failed webhooks will not be retried

## Usage

### Docker
A docker image is available at ghcr.io/danielunderwood/log2http. Really you probably want to use
docker-compose or some other system to manage configuration (ansible, k8s manifests, etc), but this should get you started.

```shell
$ docker run -d \
    -v $LOG_DIR:/logs \
    -e URL=https://example.com/... \
    --name log2http \
    ghcr.io/danielunderwood/log2http -file /log/whatever.log -sourceName "$(hostname)/docker" -regexp "a|b"
```

### Binary

Binaries for most major OS (Linux/MacOS/Windows/BSD) and architecture (x64/arm/arm64) are available under
[releases](https://github.com/danielunderwood/log2http/releases).

```shell
$ log2http -file FILENAME -regexp "^[Rr]egex$" -url https://discordapp.com/...
$ # Or supply URL via environment
$ export URL="https://discordapp.com/..."
$ log2http -file FILENAME -regexp "^[Rr]egex$
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
