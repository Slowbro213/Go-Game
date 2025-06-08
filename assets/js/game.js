//game.js
import { socket, setupSocket } from './socket.js';
import { setupInput } from './input.js';
import { createAnimator } from './animation.js';
import { HandleEvent } from './eventhandler.js';

const game_container = document.getElementById("game-container");
const log = document.getElementById("log");



const players = new Map();
let myID = null;

const response = await fetch('/game/join');
let player_id = null;
let token = null;
if (response.ok) {
  const data = await response.json();
  player_id = data.player_id;
  token = data.token;
  console.log(player_id);
  console.log(token);
  myID = player_id
} else {
  switch(response.status){
    case 401: window.location.href = "/error/unauth"; break;
    case 403: window.location.href = "/error/duplicate"; break;
    case 503: window.location.href = "/error/max"; break;
  }
  throw new Error("Couldnt Join")
}



setupInput();

setupSocket( token ,
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
    //startSendingKeys();
  },
  () => {
    //stopSendingKeys();
  }
);
