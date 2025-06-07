//game.js
import { socket, setupSocket } from './socket.js';
import { setupInput, startSendingKeys, stopSendingKeys } from './input.js';
import { createAnimator } from './animation.js';
import { HandleEvent } from './eventhandler.js';

const request = fetch('/game/auth');
const game_container = document.getElementById("game-container");
const log = document.getElementById("log");



const players = new Map();
let myID = null;

const response = await request;
if (response.ok) {
  const idText = await response.text(); // Reads response body as plain text
  const id = parseInt(idText, 10);      // Convert to integer (optional)
  myID = id;
  console.log("Auth ID:", id);
} else {
  console.error("Auth failed with status:", response.status);
  switch(response.status){
    case 401: window.location.href="/error/unauth"; break;
    case 403: window.location.href="/error/duplicate"; break;


  }

}
setupInput();

setupSocket(
  (e) => {
    //log.textContent += "Server: " + e.data + "\n";
    const event = JSON.parse(e.data);
    HandleEvent(event,players,game_container);

  },
  () => {
    log.textContent += "Status: Connected\n";

    const newPlayer = document.createElement('div');
    newPlayer.id="character" + myID;
    newPlayer.classList.add("character");
    newPlayer.classList.add("me");
    game_container.appendChild(newPlayer);
    const newAnimation = createAnimator(newPlayer);
    players.set(myID,newAnimation);
    startSendingKeys();
  },
  () => {
    stopSendingKeys();
  }
);
