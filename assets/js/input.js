import { socket } from './socket.js';

const pressedKeys = new Set();

const opposites = new Map();

opposites.set("a","d");
opposites.set("d","a");
opposites.set("w","s");
opposites.set("s","w");


const temps = new Set();

export function setupInput() {
  document.addEventListener("keydown", (e) => {
    const key = e.key.toLowerCase();
    if (["w", "a", "s", "d"].includes(key)) {
      let opposite = opposites.get(key)
      if(pressedKeys.has(opposite)){
        pressedKeys.delete(opposite)
        temps.add(opposite)
      }
      pressedKeys.add(key);
    }
  });

  document.addEventListener("keyup", (e) => {
    const key = e.key.toLowerCase();
    const opposite = opposites.get(key);
    if(temps.has(opposite)){
      pressedKeys.add(opposite);
      temps.delete(opposite);
    }
    pressedKeys.delete(key);
    temps.delete(key);
  });
}

export function sendKeys() {
  const keys = Array.from(pressedKeys).sort().join('');
  if (socket.readyState === WebSocket.OPEN && keys.length > 0) {
    socket.send(keys);
  }
  else{
    console.log("Cant send keys");
    
  }
}

let sendKeysIntervalId = null;

export function startSendingKeys() {
  console.log("sending...")
  let lastSent = 0;

  const send = (now) => {
    if (now - lastSent >= 5 && pressedKeys.size > 0) {
      sendKeys();
      lastSent = now;
    }
    sendKeysIntervalId = requestAnimationFrame(send);
  };
  sendKeysIntervalId = requestAnimationFrame(send);
}

export function stopSendingKeys() {
  if (sendKeysIntervalId) clearInterval(sendKeysIntervalId);
}
