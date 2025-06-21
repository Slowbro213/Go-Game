package gamebase

import (
	"log"
	"sync"
	"time"
	"encoding/json"
	"encoding/binary"
	"bytes"
	//"maps"

	"game/core" 
	"game/player"
)

type Message struct {
	Type string             `json:"type"`
	//Data map[int]core.GameObject `json:"data"`
	Data []byte
}




type Game struct {
	Engine core.Engine

	maxPlayers 		    int
	nextID            int
	State             *State
	PlayerIDs         map[int]string            
	PlayersMu         sync.RWMutex              


	PrevState         *State

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


	msgType := "player_joined"
	msgTypeBytes := []byte(msgType)
	msgTypeLen := uint32(len(msgTypeBytes))

	payloadSize := p.Size()
	totalSize := 4 + len(msgTypeBytes) + payloadSize

	buf := make([]byte, totalSize)

	binary.LittleEndian.PutUint32(buf[0:4], msgTypeLen)
	copy(buf[4:4+len(msgTypeBytes)], msgTypeBytes)

	offset := 4 + len(msgTypeBytes)
	p.ToBytes(buf, offset)

	if g.BroadcastFunc != nil {
		g.BroadcastFunc(buf)
	}
}

func (g *Game) RemovePlayer(p *player.Player) {
	g.PlayersMu.Lock()
	defer g.PlayersMu.Unlock()

	delete(g.State.Players, p.UserID())
	delete(g.PlayerIDs, p.ID())

	g.Engine.RemoveObject(p.ID()) 

	leaveMsg := map[string]any{
		"type": "player_left",
		"data": map[string]any{
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


var (
	bufPool = sync.Pool{
	New: func() any {
		return make([]byte, 4096) // Adjust size based on your needs
	},
}

)

var (
	typeStr      = "position_update"
	typeBytes    = []byte(typeStr)
	typeLen      = len(typeBytes)
	typeLenBytes = make([]byte, 4)
)

func init() {
	binary.LittleEndian.PutUint32(typeLenBytes, uint32(typeLen))
}


func (g *Game) OnFixedUpdate(delta float64) {
	payloadSize := 0

	for _, conc := range g.State.Base.Objects {
		payloadSize += conc.DeltaSize()
	}

	totalSize := 4 + typeLen + payloadSize
	buf := bufPool.Get().([]byte)[:totalSize]
	offset := 0

	copy(buf[offset:offset+4], typeLenBytes)
	offset += 4
	copy(buf[offset:offset+typeLen], typeBytes)
	offset += typeLen

	initOffset := offset
	for _, conc := range g.State.Base.Objects {
		if !conc.IsDirty() {
			continue
		}
		offset += conc.ToDeltaBytes(buf, offset)
	}

	if g.BroadcastFunc != nil && offset > initOffset {
		g.BroadcastFunc(buf[:offset])
	}

	bufPool.Put(buf[:cap(buf)])
}





func (g *Game) OnVariableUpdate(delta float64) {
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
	effect := &MovementEffect{Direction: direction}

	gameEvent := &core.Event{
		Effects: map[int][]core.IEffect{
			playerID: {effect},
		},
		Timestamp: time.Now().UnixNano(),
		SourceID:  playerID,
	}

	g.Engine.HandleEvent(gameEvent)

}
