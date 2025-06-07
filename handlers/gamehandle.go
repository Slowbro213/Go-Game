package handlers


import (

	"net/http"

	"fmt"
	"log"
	"os"
	"time"

	"gametry.com/player"
	"gametry.com/middleware"
	"gametry.com/utils"


	"sync"

	"encoding/json"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"

)



type Event struct{
	Type string
	Data map[string]interface{}
}
type PendingConnection struct {
	PlayerID int
	Expires  time.Time
}

type GameHandler struct{

	log *log.Logger
	upgrader *websocket.Upgrader

	playersMu sync.Mutex
	players []*player.Player 
	users map[string]struct{}

	tokensMu      sync.Mutex
	pendingTokens map[string]*PendingConnection

	store *sessions.CookieStore
	nextID int
}

const (
	maxPlayers = 10
)

func NewGameHandler(l* log.Logger, u* websocket.Upgrader, s* sessions.CookieStore) *GameHandler {
	return &GameHandler{
		log: l,
		upgrader: u,
		players : make([]*player.Player, maxPlayers),
		users: make(map[string]struct{}),
		pendingTokens: make(map[string]*PendingConnection),
		store: s,
	}
}




func (g *GameHandler) Join(w http.ResponseWriter, r *http.Request) {
    // Get authenticated player ID from context
    playerID, ok := r.Context().Value(middleware.ContextPlayerID).(string)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Early check for duplicates without full lock
    g.tokensMu.Lock()
    _, exists := g.users[playerID]
    g.tokensMu.Unlock()
    
    if exists {
        http.Error(w, "Duplicate", http.StatusForbidden)
        return
    }

    // Generate token before taking locks
    token := utils.GenerateToken()

    // Find available slot
    g.playersMu.Lock()
    slot, found := -1, false
    for i := 0; i < maxPlayers; i++ {
        if g.players[g.nextID] == nil {
            slot = g.nextID
            found = true
            g.nextID = (g.nextID + 1) % maxPlayers
            break
        }
        g.nextID = (g.nextID + 1) % maxPlayers
    }
    g.playersMu.Unlock()

    if !found {
        http.Error(w, "Maximum player capacity reached", http.StatusServiceUnavailable)
        return
    }

    // Final reservation with all locks
    g.tokensMu.Lock()
    defer g.tokensMu.Unlock()

    // Double-check after lock
    if _, exists := g.users[playerID]; exists {
        http.Error(w, "Duplicate", http.StatusForbidden)
        return
    }

    g.users[playerID] = struct{}{}
    g.pendingTokens[token] = &PendingConnection{
        PlayerID: slot,
        Expires:  time.Now().Add(30 * time.Second),
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "player_id": slot,
        "token":     token,
    })
}

func (g *GameHandler) Match(w http.ResponseWriter, r *http.Request) {

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}
	playerID, ok := r.Context().Value(middleware.ContextPlayerID).(string)
	if !ok {
		g.log.Println(fmt.Sprintf("UserID: %s",playerID))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Validate and consume token
	g.tokensMu.Lock()
	pending, exists := g.pendingTokens[token]
	delete(g.pendingTokens, token) // One-time use
	g.tokensMu.Unlock()

	if !exists || time.Now().After(pending.Expires) {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	// Proceed with WebSocket upgrade using the validated playerID
	Id := pending.PlayerID
	conn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		g.log.Println("WebSocket upgrade error:", err)
		return
	}

	l := log.New(os.Stdout, fmt.Sprintf("Player %d: ",Id), log.LstdFlags)
	p := player.NewPlayer(Id,playerID, 0, 0, conn,l)

	g.playersMu.Lock()
	g.players[Id] = p
	g.playersMu.Unlock()

	g.log.Println("User Joined:", Id)

	msg := map[string]interface{}{
		"type": "player_joined",
		"data": map[string]interface{}{
			"id": Id,
		},
	}
	jsonData , err := json.Marshal(msg)
	if err != nil{
		g.log.Println("JSON marshal error:", err)
		return
	}
	go g.broadcastMessage(jsonData)

	positions := g.getPositions()
	if err != nil {
		g.log.Println("JSON marshal error:", err)
		return
	}

	event := map[string]interface{}{
		"type": "position_update",
		"data": map[string]interface{}{
			"positions": positions,
		},
	}
	jsonData, err = json.Marshal(event)
	if err != nil {
		g.log.Println("JSON marshal error:", err)
		return
	}

	go g.broadcastMessage(jsonData)

	// Handle player connection
	go g.handlePlayerConnection(p)
}

func (g *GameHandler) handlePlayerConnection(p *player.Player) {
	defer func() {
		msg := map[string]interface{}{
			"type": "player_left",
			"data": map[string]interface{}{
				"id": p.ID,
			},
		}
		jsonBytes, err := json.Marshal(msg)
		if err != nil {
			g.log.Println("JSON marshal error:", err)
			return
		}
		g.broadcastMessage(jsonBytes)

		p.Conn.Close()
		g.playersMu.Lock()
		id := p.ID
		userID := p.UserID
		g.players[id] = nil
		delete(g.users, userID)
		g.StartTokenCleanup()
		g.playersMu.Unlock()
		g.log.Println("User Left:", id)
	}()

	step := 10

	for {
		_, msg, err := p.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				g.log.Println("Read error:", err)
			}
			return
		}

		if len(msg) > 0 {
			g.processMovement(p, msg, step)
		}

		positions := g.getPositions()
		if err != nil {
			g.log.Println("JSON marshal error:", err)
			return
		}

		event := map[string]interface{}{
			"type": "position_update",
			"data": map[string]interface{}{
				"positions": positions,
			},
		}
		jsonData, err := json.Marshal(event)
		if err != nil {
			g.log.Println("JSON marshal error:", err)
			return
		}

		g.broadcastMessage(jsonData)

	}
}

func (g *GameHandler) processMovement(p *player.Player, msg []byte, step int) {
	if len(msg) > 1 {
		key1 := rune(msg[0])
		key2 := rune(msg[1])
		normalizedStep := float64(step) * 0.707
		p.MoveByKey(key1, normalizedStep)
		p.MoveByKey(key2, normalizedStep)
	} else {
		p.MoveByKey(rune(msg[0]), float64(step))
	}
}

func (g *GameHandler) getPositions() map[int][2]float64{
	g.playersMu.Lock()
	defer g.playersMu.Unlock()

	positions := make(map[int][2]float64)
	for id, player := range g.players {
		if player != nil {
			x, y := player.Position()
			positions[id] = [2]float64{x, y}
		}
	}

	return positions

}

func (g* GameHandler) broadcastMessage(bytes []byte) {
	for _, player := range g.players {
		if player != nil{
			player.Notify(bytes)
		}
	}
}



func (g *GameHandler) StartTokenCleanup() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			g.cleanupExpiredTokens()
		}
	}()
}

func (g *GameHandler) cleanupExpiredTokens() {
	g.tokensMu.Lock()
	defer g.tokensMu.Unlock()
	
	now := time.Now()
	for token, pc := range g.pendingTokens {
		if now.After(pc.Expires) {
			delete(g.pendingTokens, token)
		}
	}
}
