// Example usage of k6-socketio extension
import socketio from 'k6/x/socketio';
import { check, sleep } from 'k6';

export default function () {
  // Connect to the Socket.IO server
  socketio.connect('ws://localhost:4000');

  // Emit a test event
  socketio.emit('test', { bool: true, test: 'success' });
  check(true, { 'emit test event': (v) => v === true });

  // Emit with acknowledgement
  let ackResult = socketio.emitWithAck('ackevent', { foo: 'bar' });
  console.log('ackResult:', JSON.stringify(ackResult));
  check(ackResult !== undefined, { 'ack callback received': (v) => v === true });

  // Disconnect
  socketio.disconnect();
  sleep(0.2);
}
