import { createAnimator } from './animation.js';
const eventsMap = new Map();

function PlayerLeft(data,players,game_container){
  players.delete(data.id);

  const leavingPlayer = document.getElementById(`character${data.id}`);
  game_container.removeChild(leavingPlayer);
}


function PlayerJoined(data, players, game_container) {
  data = data[0];
  const playerId = data.id;
  const type = data.type || "character"; 
  const pos = data.position || { x: 0, y: 0 };
  const x = pos.x;
  const y = pos.y;

  if (!players.has(playerId)) {
    const newPlayer = document.createElement('div');
    newPlayer.id = type + playerId;
    newPlayer.classList.add(type);
    newPlayer.classList.add('other');
    game_container.appendChild(newPlayer);

    const newAnimation = createAnimator(newPlayer);
    players.set(playerId, newAnimation);
    newAnimation.setPosition(x, y);
  }
}



function PositionUpdate(data, players, game_container) {

    data.forEach((obj) => {
    const playerId = obj.id
    const x = obj.position.x;
    const y = obj.position.y;
    const type = obj.type;

    if (!players.has(playerId)) {
      const newPlayer = document.createElement('div');
      newPlayer.id = type + playerId;
      newPlayer.classList.add(type);
      newPlayer.classList.add('other');
      game_container.appendChild(newPlayer);

      const newAnimator = createAnimator(newPlayer);
      players.set(playerId, newAnimator);
      newAnimator.setPosition(x, y);
    } else {
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
