# xk6-socketio TypeScript Types

This package provides TypeScript type definitions for the [xk6-socketio](../README.md) API, enabling autocompletion and type safety when writing k6 test scripts in TypeScript.

## Installation

Install via npm:

```sh
npm install xk6-socketio-types
```

## Usage

Import types in your k6 scripts:

```ts
// @ts-ignore
import { connect, emit, emitWithAck, disconnect } from 'xk6-socketio-types';

export default function () {
  connect('ws://localhost:3000');
  emit('event', { foo: 'bar' });
  const ack = emitWithAck('ackevent', { foo: 'bar' });
  disconnect();
}
```

## API

### `connect(url: string): void`

Connect to a Socket.IO server.

### `emit(event: string, data: any): void`

Emit an event to the server.

### `emitWithAck(event: string, data: any, timeout?: number): any`

Emit an event and wait for an acknowledgement. Returns the acknowledgement response or a timeout error object.

### `disconnect(): void`

Disconnect from the server.

## Contributing

Feel free to submit issues or PRs to improve the type definitions.
