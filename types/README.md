# xk6-socketio TypeScript Types

This package provides TypeScript type definitions for the [xk6-socketio](../README.md) API, enabling autocompletion and type safety when writing k6 test scripts in TypeScript.

## Installation

Copy or link this `@types/xk6-socketio` directory into your project, or publish it to npm and install via:

```sh
npm install @types/xk6-socketio
```

## Usage

Reference the types in your TypeScript k6 scripts:

```ts
// @ts-ignore
import { connect, SocketIOClient } from '@types/xk6-socketio';

export default function () {
  const socket: SocketIOClient = connect('ws://localhost:3000');
  socket.emit('event', { foo: 'bar' });
  socket.disconnect();
}
```

## Contributing

Feel free to submit issues or PRs to improve the type definitions.
