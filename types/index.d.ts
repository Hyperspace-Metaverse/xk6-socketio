// Type definitions for k6/x/socketio
// Project: xk6-socketio

declare module 'k6/x/socketio' {
  /**
   * Connect to a Socket.IO server.
   * @param url The server URL.
   */
  export function connect(url: string): void;

  /**
   * Emit an event to the server.
   * @param event The event name.
   * @param data The data to send.
   */
  export function emit(event: string, data: any): void;

  /**
   * Emit an event and wait for an acknowledgement.
   * @param event The event name.
   * @param data The data to send.
   * @param timeout Optional timeout in milliseconds (default: 2000).
   * @returns The acknowledgement response or a timeout error object.
   */
  export function emitWithAck(event: string, data: any, timeout?: number): any;

  /**
   * Disconnect from the server.
   */
  export function disconnect(): void;
}
