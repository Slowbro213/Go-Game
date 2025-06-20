package game

import (
	"log"
	"sync"
	"time"
	"encoding/json"
	"bytes"

	"game/core" 
	"game/player"
)

type PositionUpdateMessage struct {
	Type string             `json:"type"`
	Data PositionUpdateData `json:"data"`
}

type PositionUpdateData struct {
	Positions map[int][2]float32 `json:"positions"`
}

type Game struct {
	Engine core.Engine

	maxPlayers 		int
	nextID        int
	State         *State
	PlayerIDs     map[int]string            
	PlayersMu     sync.RWMutex              


	CurrentPositions     map[int][2]float32
	PrevPositions map[int][2]float32

	UpdateMsg     PositionUpdateMessage
	jsonBuffer        *bytes.Buffer
	jsonEncoder       *json.Encoder

	log *log.Logger

	BroadcastFunc func([]byte)
}

func NewGame(state *State,fixedTPS float64, targetFPS, maxPlayers int, l *log.Logger) *Game {
	g := &Game{
		State:         state,
		maxPlayers:    maxPlayers,
		nextID:        0,
		PlayerIDs:     make(map[int]string),
		log:           l,
		CurrentPositions : make(map[int][2]float32),
		PrevPositions : make(map[int][2]float32),
		UpdateMsg: PositionUpdateMessage{
			Type: "position_update",
			Data: PositionUpdateData{},
		},
		jsonBuffer: bytes.NewBuffer(make([]byte, 0, 2048)),
	}
	g.jsonEncoder = json.NewEncoder(g.jsonBuffer)
	g.Engine = *core.NewEngine(state.Base,fixedTPS,targetFPS)

	g.Engine.OnFixedUpdate = g.OnFixedUpdate
	g.Engine.OnVariableUpdate = g.OnVariableUpdate

	return g
}

func (g *Game) Start() {
	g.Engine.Run()
}

func (g *Game) Shutdown() {
	g.Engine.Shutdown()
}

func (g *Game) ReserveSpot() (int, bool) {
	slot, found := -1, false
	for i := 0; i < g.maxPlayers; i++ {
		if _, ok := g.PlayerIDs[g.nextID]; !ok { 
			slot = g.nextID
			found = true
			g.nextID = (g.nextID + 1) % g.maxPlayers 
			break
		}
		g.nextID = (g.nextID + 1) % g.maxPlayers
	}

	return slot, found
}

func (g *Game) AddPlayer(p *player.Player) {
	g.PlayersMu.Lock()
	defer g.PlayersMu.Unlock()

	g.State.Players[p.UserID()] = p
	g.PlayerIDs[p.ID()] = p.UserID()

	g.Engine.AddObject(p) 

	joinMsg := map[string]interface{}{
		"type": "player_joined",
		"data": map[string]interface{}{
			"id":       p.ID(),
		},
	}
	jsonBytes, err := json.Marshal(joinMsg)
	if err != nil {
		g.log.Println("JSON marshal error:", err)
	} else {
		g.BroadcastFunc(jsonBytes)
	}

}

func (g *Game) RemovePlayer(p *player.Player) {
	g.PlayersMu.Lock()
	defer g.PlayersMu.Unlock()

	delete(g.State.Players, p.UserID())
	delete(g.PlayerIDs, p.ID())

	g.Engine.RemoveObject(p.ID()) 

	leaveMsg := map[string]interface{}{
		"type": "player_left",
		"data": map[string]interface{}{
			"id":     p.ID(),
			"userID": p.UserID(),
		},
	}
	jsonBytes, err := json.Marshal(leaveMsg)
	if err != nil {
		g.log.Println("JSON marshal error:", err)
	} else {
		g.BroadcastFunc(jsonBytes)
	}

}

func (g *Game) GetPlayerByUserID(userID string) *player.Player {
	g.PlayersMu.RLock()
	defer g.PlayersMu.RUnlock()
	return g.State.Players[userID]
}

func (g *Game) GetPlayerByID(playerID int) *player.Player {
    g.PlayersMu.RLock()
    defer g.PlayersMu.RUnlock()
    userID, exists := g.PlayerIDs[playerID]
    if !exists {
        return nil
    }
    return g.State.Players[userID]
}

func (g *Game) OnFixedUpdate(delta float64) {
	for k := range g.CurrentPositions {
		delete(g.CurrentPositions, k)
	}


	g.PlayersMu.RLock()
	if len(g.State.Players) == 0 {
		g.PlayersMu.RUnlock()
		for k := range g.PrevPositions {
			delete(g.PrevPositions, k)
		}
		return
	}
	for _, p := range g.State.Players {
		if p != nil {
			pos := p.PositionXY()
			g.CurrentPositions[p.ID()] = [2]float32{pos.X, pos.Y}
		}
	}
	g.PlayersMu.RUnlock()

	deltaPositions := make(map[int][2]float32)
	var removedIDs []int

	for prevID := range g.PrevPositions {
		if _, exists := g.CurrentPositions[prevID]; !exists {
			removedIDs = append(removedIDs, prevID)
		}
	}

	for currentID, currentPos := range g.CurrentPositions {
		prevPos, existsInPrev := g.PrevPositions[currentID]
		if !existsInPrev || prevPos != currentPos {
			deltaPositions[currentID] = currentPos
		}
	}

	if len(deltaPositions) == 0 && len(removedIDs) == 0 {
		for k := range g.PrevPositions {
			delete(g.PrevPositions, k)
		}
		for id, pos := range g.CurrentPositions {
			g.PrevPositions[id] = pos
		}
		return
	}

	g.UpdateMsg.Data.Positions = deltaPositions

	g.jsonBuffer.Reset()
	err := g.jsonEncoder.Encode(g.UpdateMsg)
	if err != nil {
		g.log.Println("JSON marshal error for position update:", err)
		for k := range g.PrevPositions {
			delete(g.PrevPositions, k)
		}
		for id, pos := range g.CurrentPositions {
			g.PrevPositions[id] = pos
		}
		return
	}

	jsonData := g.jsonBuffer.Bytes()

	if g.BroadcastFunc != nil {
		g.BroadcastFunc(jsonData)
	}

	for k := range g.PrevPositions {
		delete(g.PrevPositions, k)
	}
	for id, pos := range g.CurrentPositions {
		g.PrevPositions[id] = pos
	}

}


func (g *Game) OnVariableUpdate(delta float64) {
	return
}

func (g *Game) HandleInputEvent(clientEv *core.ClientEvent, p *player.Player) {

	switch clientEv.Type {
	case "input_movement":
		g.HandleInputMovement(clientEv,p)
	case "chat_message":
		g.log.Println("Not Yet Implemented")

	default:
		g.log.Println("Unknown client event type:", clientEv.Type, "from player", p.ID())
	}



}


func (g *Game) HandleInputMovement(clientEv *core.ClientEvent, p *player.Player){
	direction, ok := clientEv.Data["direction"].(string)
	if !ok {
		g.log.Println("Invalid direction in input_movement event from player", p.ID())
		return
	}
	playerID := p.ID()
	effect := &core.MovementEffect{Direction: direction}

	gameEvent := &core.Event{
		Effects: map[int][]core.IEffect{
			playerID: {effect},
		},
		Timestamp: time.Now().UnixNano(),
		SourceID:  playerID,
	}

	g.Engine.HandleEvent(gameEvent)

}
