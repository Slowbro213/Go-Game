import { socket } from './socket.js';

const pressedKeys = new Set();
const opposites = new Map();
opposites.set("a", "d");
opposites.set("d", "a");
opposites.set("w", "s");
opposites.set("s", "w");
const temps = new Set();

export function setupInput() {
    document.addEventListener("keydown", (e) => {
        const key = e.key.toLowerCase();
        if (["w", "a", "s", "d"].includes(key)) {
            let opposite = opposites.get(key)
            if (pressedKeys.has(opposite)) {
                pressedKeys.delete(opposite)
                temps.add(opposite)
            }
            pressedKeys.add(key);
            sendMovementCommand(); 
    }
    });

    document.addEventListener("keyup", (e) => {
        const key = e.key.toLowerCase();
        const opposite = opposites.get(key);
        if (temps.has(opposite)) {
            pressedKeys.add(opposite);
            temps.delete(opposite);
        }
        pressedKeys.delete(key);
        temps.delete(key);
        sendMovementCommand();  
  });
}

export function sendMovementCommand() {
    let direction = "move_stop";  
  
    if (pressedKeys.has("w") && pressedKeys.has("a")) {
        direction = "move_up_left";  
  } else if (pressedKeys.has("w") && pressedKeys.has("d")) {
        direction = "move_up_right";
    } else if (pressedKeys.has("s") && pressedKeys.has("a")) {
        direction = "move_down_left";
    } else if (pressedKeys.has("s") && pressedKeys.has("d")) {
        direction = "move_down_right";
    } else if (pressedKeys.has("w")) {
        direction = "move_up";
    } else if (pressedKeys.has("s")) {
        direction = "move_down";
    } else if (pressedKeys.has("a")) {
        direction = "move_left";
    } else if (pressedKeys.has("d")) {
        direction = "move_right";
    }

    if (socket.readyState === WebSocket.OPEN) {
        const msg = {
            type: "input_movement",
            data: {
                direction: direction, 
            },
        };
        socket.send(JSON.stringify(msg));
    } else {
        console.log("Cant send movement command: Socket not open.");
    }
}

//let sendKeysIntervalId = null;
//
//export function startSendingKeys() {
//  console.log("sending...")
//  let lastSent = 0;
//
//  const send = (now) => {
//    if (now - lastSent >= 5 && pressedKeys.size > 0) {
//      sendKeys();
//      lastSent = now;
//    }
//    sendKeysIntervalId = requestAnimationFrame(send);
//  };
//  sendKeysIntervalId = requestAnimationFrame(send);
//}
//
//export function stopSendingKeys() {
//  if (sendKeysIntervalId) clearInterval(sendKeysIntervalId);
//}
