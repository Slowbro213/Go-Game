package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"gametry.com/middleware" // Assuming your middleware is here
	"gametry.com/player"     // Player type
	"gametry.com/utils"      // Utility functions
	"gametry.com/core"      // Utility functions


	"gametry.com/game" // NEW: Import your game package

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)


type PendingConnection struct {
	PlayerID int
	Expires  time.Time
}

const (
	maxPlayers = 10
	fixedTPS   = 30
	playerBasePxPs = 800 )

type GameHandler struct {
	log *log.Logger

	upgrader *websocket.Upgrader

	game *game.Game 

	tokensMu      sync.Mutex
	pendingTokens map[string]*PendingConnection

	store *sessions.CookieStore
	nextID int 
}

func NewGameHandler(l *log.Logger, s *sessions.CookieStore) *GameHandler {
	handler := &GameHandler{
		log:      l,
		upgrader: &websocket.Upgrader{}, 
		store:    s,

		pendingTokens: make(map[string]*PendingConnection),
		nextID:        0, 
	}


	handler.game = game.NewGame(fixedTPS, l)

	handler.game.BroadcastFunc = handler.broadcastMessage

	handler.game.Start()

	handler.StartTokenCleanup() 

	return handler
}

func (g *GameHandler) Join(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.ContextPlayerID).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if g.game.GetPlayerByUserID(userID) != nil {		http.Error(w, "Duplicate", http.StatusForbidden)
		return
	}

		g.tokensMu.Lock()
		slot, found := -1, false
		for i := 0; i < maxPlayers; i++ {
		if _, ok := g.game.PlayerIDs[g.nextID]; !ok { 
			slot = g.nextID
			found = true
			g.nextID = (g.nextID + 1) % maxPlayers 
			break
		}
		g.nextID = (g.nextID + 1) % maxPlayers
		}
		g.tokensMu.Unlock()

		if !found {
			http.Error(w, "Maximum player capacity reached", http.StatusServiceUnavailable)
			return
		}

		token := utils.GenerateToken()

		g.tokensMu.Lock()
		defer g.tokensMu.Unlock()

		if g.game.GetPlayerByUserID(userID) != nil {
			http.Error(w, "Duplicate", http.StatusForbidden)
			return
		}

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
		userID, ok := r.Context().Value(middleware.ContextPlayerID).(string)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		g.tokensMu.Lock()
		pending, exists := g.pendingTokens[token]
		delete(g.pendingTokens, token) 
		g.tokensMu.Unlock()

		if !exists || time.Now().After(pending.Expires) {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}

		conn, err := g.upgrader.Upgrade(w, r, nil)
		if err != nil {
			g.log.Println("WebSocket upgrade error:", err)
			return
		}

		playerLogger := log.New(os.Stdout, fmt.Sprintf("Player %d [%s]: ", pending.PlayerID, userID), log.LstdFlags)
		p := player.NewPlayer(pending.PlayerID, userID, 0, 0, playerBasePxPs, conn, playerLogger)

		g.game.AddPlayer(p) 

		g.log.Println("User Joined:", p.ID(), "UserID:", p.UserID)

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
			g.broadcastMessage(jsonBytes)
		}

		g.handlePlayerConnection(p)
	}

	func (g *GameHandler) handlePlayerConnection(p *player.Player) {
		defer func() {
			g.game.RemovePlayer(p) 
			g.log.Println("User Left:", p.ID(), "UserID:", p.UserID)

			leaveMsg := map[string]interface{}{
				"type": "player_left",
				"data": map[string]interface{}{
					"id":     p.ID(),
					"userID": p.UserID,
				},
			}
			jsonBytes, err := json.Marshal(leaveMsg)
			if err != nil {
				g.log.Println("JSON marshal error:", err)
			} else {
				g.broadcastMessage(jsonBytes)
			}

			p.Conn().Close() 
		}()

		for {
			_, msg, err := p.Conn().ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					g.log.Println("Read error:", err)
				}
				return 
			}

			var clientEv core.ClientEvent
			if err := json.Unmarshal(msg, &clientEv); err != nil {
				g.log.Println("JSON unmarshal error for client event:", err)
				continue
			}

			g.game.HandleInputEvent(&clientEv,p)

		}
	}

	func (g *GameHandler) broadcastMessage(bytes []byte) {

		for _, p := range g.game.OnlinePlayers {
			if p != nil {
				p.Notify(bytes)
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
