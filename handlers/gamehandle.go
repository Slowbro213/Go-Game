package handlers


import (

	"net/http"

	"fmt"
	"math"
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


	tickInterval time.Duration 
  tickChan     chan struct{}

	store *sessions.CookieStore
	nextID int
}

const (
	maxPlayers = 10
	fps = 24
	pxps = 600 //pixels per second aka player speed
)

func NewGameHandler(l* log.Logger, u* websocket.Upgrader, s* sessions.CookieStore) *GameHandler {
	h := &GameHandler{
		log: l,
		upgrader: u,
		store: s,

		players : make([]*player.Player, maxPlayers),
		users: make(map[string]struct{}),
		pendingTokens: make(map[string]*PendingConnection),

		tickInterval: time.Duration(int(math.Round(1000.0 / fps))) * time.Millisecond,
    tickChan:     make(chan struct{}, 1),

	}

	go h.runGameLoop() 

	return h
}

func (g *GameHandler) runGameLoop() {
	ticker := time.NewTicker(g.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:


			g.updatePositions()

			g.broadcastPositions()

		case <-g.tickChan:
			return // For graceful shutdown
		}
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
	p := player.NewPlayer(Id,playerID, 0, 0,pxps, conn,l)

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
	g.broadcastMessage(jsonData)

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

	g.broadcastMessage(jsonData)

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


	for {
		_, msg, err := p.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				g.log.Println("Read error:", err)
			}
			return
		}

		p.NewInput(msg)
	}
}


func (g* GameHandler) broadcastPositions() {
	positions := g.getPositions()
	if len(positions) == 0 {
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


func (g *GameHandler) getPositions() map[int][2]float32{
	g.playersMu.Lock() // Read lock only
	defer g.playersMu.Unlock()

	var positions map[int][2]float32 // Don't pre-allocate

	for id, player := range g.players {
		if player == nil {
			continue
		}

		if positions == nil {
			positions = make(map[int][2]float32, len(g.players)/2) // Heuristic
		}

		x, y := player.PositionXY()
		positions[id] = [2]float32{x, y}
	}

	return positions // May return nil
}

func (g *GameHandler) updatePositions() {
	g.tokensMu.Lock()
	defer g.tokensMu.Unlock()

	for _, player := range g.players {
		if player == nil {
			continue
		}
		player.ApplySpeed()
	}
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
