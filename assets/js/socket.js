//socket.js
let socket; // don't export it directly

export function setupSocket(token,onMessageCallback, onOpenCallback, onCloseCallback) {
  if(token === null)
    return;
  socket = new WebSocket("ws://" + location.host + `/game?token=${token}`);
  socket.binaryType = "arraybuffer";

  socket.onmessage = onMessageCallback;
  socket.onopen = onOpenCallback;
  socket.onclose = onCloseCallback;
}

export { socket }; // export it after being set
