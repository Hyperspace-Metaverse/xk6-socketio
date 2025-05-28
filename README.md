# xk6-socketio

This is a k6 extension for Socket.IO support, allowing you to load test Socket.IO servers with k6.

## Build Instructions

You can build a custom k6 binary with this extension using [xk6](https://github.com/grafana/xk6):

### Prerequisites

- [Go](https://golang.org/dl/)
- [Docker](https://www.docker.com) or a compatible container runtime
- [xk6](https://github.com/grafana/xk6) (optional. If not present, building with `Make` will default to using a Docker container to build the k6 binary)
- Make (optional, for convenience)

### Build with xk6

```sh
xk6 build --with github.com/Hyperspace-Metaverse/xk6-socketio=.
```

Or use the provided Makefile:

```sh
make build
```

This will produce a `k6` binary in the project directory.

## Test

### Unit Tests

### Integration Tests

Run your k6 test script (see `examples/`) with the custom k6 binary:

```sh
./k6 run examples/basic.js
```

Or:

```sh
make test TEST=examples/basic.js
```

You can also build, run `docker compose` and run a test using Make:

```sh
make run-all
```

For more advanced usage, run:

```sh
make help
```

## Examples

See the `examples/` directory for usage scripts, including `basic.js` and `advanced.js`.

## Contributing

Feel free to open issues or pull requests to improve this extension.

## License

Copyright (c) 2025 Hyperspace Metaverse
[https://hyperspace.mv](https://hyperspace.mv)

This project is licensed under the GNU General Public License v3.0 (GPL-3.0). See the LICENSE file for details.
