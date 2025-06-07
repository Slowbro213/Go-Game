//socket.js
let socket; // don't export it directly

export function setupSocket(onMessageCallback, onOpenCallback, onCloseCallback) {
  socket = new WebSocket("ws://" + location.host + "/game");

  socket.onmessage = onMessageCallback;
  socket.onopen = onOpenCallback;
  socket.onclose = onCloseCallback;
}

export { socket }; // export it after being set
