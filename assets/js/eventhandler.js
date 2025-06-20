import { createAnimator } from './animation.js';
const eventsMap = new Map();

function PlayerLeft(data,players,game_container){
  players.delete(data.id);
  const leavingPlayer = document.getElementById(`character${data.id}`);
  game_container.removeChild(leavingPlayer);
}

function PlayerJoined(data,players,game_container){

  const playerId = data.id;

  if(!players.has(playerId)){
    const newPlayer = document.createElement('div');
    newPlayer.id="character" + playerId;
    newPlayer.classList.add("character");
    newPlayer.classList.add('other');
    game_container.appendChild(newPlayer);
    const newAnimation = createAnimator(newPlayer);
    players.set(playerId,newAnimation);
  }
}

function PositionUpdate(data,players,game_container){
    const positions = data.positions;
    Object.entries(positions).forEach(([id, [x, y]]) => {
      const playerId = parseInt(id, 10);
      if(!players.has(playerId)){
        const newPlayer = document.createElement('div');
        newPlayer.id="character" + playerId;
        newPlayer.classList.add("character");
        newPlayer.classList.add('other');
        game_container.appendChild(newPlayer);
        const newAnimation = createAnimator(newPlayer);
        players.set(playerId,newAnimation);
        newAnimation.setPosition(x,y);
      }else {
        const currAnimator = players.get(playerId);
        currAnimator.updatePosition(x, y);
      }

    });
}


eventsMap.set('player_left',PlayerLeft);
eventsMap.set('player_joined',PlayerJoined);
eventsMap.set('position_update',PositionUpdate);


export function HandleEvent(e,players,game_container){
  const type = e.type;
  const data = e.data;

  eventsMap.get(type)(data,players,game_container);
}
