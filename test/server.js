// Basic Socket.IO server for testing
const { Server } = require('socket.io');

const io = new Server(4000, {
  cors: {
    origin: '*',
  },
});

io.on('connection', (socket) => {
  console.log('Client connected:', socket.id);

  // Catch-all for all events
  socket.onAny((event, ...args) => {
    console.log('onAny event:', event, args);
  });

  socket.on('test', (data) => {
    console.log('Received test event:', data);
    // Optionally emit a response
    socket.emit('test_response', { received: true });
  });

  socket.on('ackevent', (data, callback) => {
    console.log('Received ackevent:', data, typeof callback);
    if (typeof callback === 'function') {
      callback({ success: true });
    }
  });

  socket.on('disconnect', () => {
    console.log('Client disconnected:', socket.id);
  });
});

console.log('Socket.IO server running on port 4000');
