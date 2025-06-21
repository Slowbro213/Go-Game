//game.js
import { socket, setupSocket } from './socket.js';
import { setupInput } from './input.js';
import { createAnimator } from './animation.js';
//import  init, { decode } from '../../wasm/decoder/pkg/decoder.js';
import { decode } from './decode.js';
import { HandleEvent } from './eventhandler.js';

const game_container = document.getElementById("game-container");
const log = document.getElementById("log");

const gameState = JSON.parse(window.GAMESTATE);



//async function setupWasmDecoder() {
//  await init();
//}
//
//await setupWasmDecoder()

const players = new Map();

let player_id = window.PLAYERID;
let token = window.TOKEN;
const myID = Number(player_id);


setupInput();

setupSocket( token ,
  (e) => {
    //log.textContent += "Server: " + e.data + "\n";
    const event = decode(e.data);
    HandleEvent(event,players,game_container);

  },
  () => {
    log.textContent += "Status: Connected\n";
    log.textContent += `${myID} ${token}`;

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
