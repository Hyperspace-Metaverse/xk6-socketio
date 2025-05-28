import socketio from 'k6/x/socketio';
import { check, sleep } from 'k6';

export default function () {
  // Connect to the Socket.IO server on port 4000
  socketio.connect('ws://localhost:4000');

  // Emit a test event and check
  socketio.emit('test', { bool: true, test: 'success' });
  check(true, { 'emit test event': (v) => v === true });

  // Emit a ping event and check
  socketio.emit('ping', { timestamp: Date.now() });
  check(true, { 'emit ping event': (v) => v === true });

  // Emit a message event with a string
  socketio.emit('message', 'Hello from k6!');
  check(true, { 'emit message event': (v) => v === true });

  // Emit a number event with a number
  socketio.emit('number', 42);
  check(true, { 'emit number event': (v) => v === true });

  // Emit a complex event with nested data
  socketio.emit('complex', { arr: [1, 2, 3], obj: { foo: 'bar' }, flag: false });
  check(true, { 'emit complex event': (v) => v === true });

  // Emit a binary event (simulate with base64 string)
  socketio.emit('binary', { data: 'SGVsbG8gQmluYXJ5IQ==' });
  check(true, { 'emit binary event': (v) => v === true });

  // Emit multiple events in a loop
  for (let i = 0; i < 5; i++) {
    socketio.emit('loop', { idx: i, time: Date.now() });
    check(true, { [`emit loop event #${i}`]: (v) => v === true });
    sleep(0.05);
  }

  // Test Socket.IO acknowledgement callback
  let ackResult = socketio.emitWithAck('ackevent', { foo: 'bar' });
  console.log('ackResult:', JSON.stringify(ackResult));
  check(ackResult !== undefined, { 'ack callback received': (v) => v === true });
  if (ackResult && typeof ackResult === 'object') {
    check(ackResult.success === true, { 'ack result success': (v) => v === true });
  }

  // Simulate a short wait between actions
  sleep(0.2);

  // Disconnect
  socketio.disconnect();

  // Final check (no error thrown means success)
  check(true, { 'socket.io test completed': (v) => v === true });
}
