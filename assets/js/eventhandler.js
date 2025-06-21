import { createAnimator } from './animation.js';
const eventsMap = new Map();

function PlayerLeft(data,players,game_container){
  players.delete(data.id);

  const leavingPlayer = document.getElementById(`character${data.id}`);
  game_container.removeChild(leavingPlayer);
}


function PlayerJoined(data, players, game_container) {
  const playerId = data.id;
  const type = data.type || "character"; 
  const pos = data.Data?.Position || { X: 0, Y: 0 };
  const x = pos.X;
  const y = pos.Y;

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
  const objects = data.objects;

  Object.entries(objects).forEach(([id, obj]) => {
    const playerId = parseInt(id, 10);
    const x = obj.Data.Position.X;
    const y = obj.Data.Position.Y;
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
