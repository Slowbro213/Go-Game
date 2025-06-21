package handlers

import (
	"encoding/json"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"html/template"

	"game/middleware"
	"game/player"     
	"game/utils"      
	"game/game"
	"game/core"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)


type TemplateData struct {
	GameState template.JS
	PlayerID  int 
	Token     string
	Binary    []byte
}

type PendingConnection struct {
	PlayerID int
	Expires  time.Time
}

const (
	maxPlayers = 10
	fixedTPS   = 30
	targetFPS   = 120
	playerBasePxPs = 800 )

type GameHandler struct {
	log *log.Logger

	upgrader *websocket.Upgrader

	game *game.Game 

	tokensMu      sync.Mutex
	pendingTokens map[string]*PendingConnection

	store *sessions.CookieStore
}

func NewGameHandler(l *log.Logger, s *sessions.CookieStore) *GameHandler {
	handler := &GameHandler{
		log:      l,
		upgrader: &websocket.Upgrader{}, 
		store:    s,

		pendingTokens: make(map[string]*PendingConnection),
	}

	base := core.State{
		Objects:   make(map[int]core.GameObject),
		ConcreteObjects: make(map[int]core.ConcreteObject),
		Entities:  make(map[int]core.Entity),
		PhysicsObjects: make(map[int]core.PhysicsObject),
	}

	gameState := game.State{
		Base: &base,
		Players: make(map[string]*player.Player),
	}

	handler.game = game.NewGame(&gameState,fixedTPS, targetFPS, maxPlayers ,l)

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

	if g.game.GetPlayerByUserID(userID) != nil {		
		utils.RenderMessage(w, utils.MessageData{

			Type:     "error",
			Title:    "Duplicates are Forbidden",
			Message:  "You cannot play from multiple tabs",
			Link:     "/logout",
			LinkText: "Logout",
		})
		return
	}


	slot , found := g.game.ReserveSpot()


	if !found {
		http.Error(w, "Maximum player capacity reached", http.StatusServiceUnavailable)
		return
	}

	token := utils.GenerateToken()

	g.tokensMu.Lock()

	g.pendingTokens[token] = &PendingConnection{
		PlayerID: slot,
		Expires:  time.Now().Add(30 * time.Second),
	}

	g.tokensMu.Unlock()


	playerLogger := log.New(os.Stdout, fmt.Sprintf("Player %d [%s]: ", slot, userID), log.LstdFlags)
	p := player.NewPlayer(slot, userID, 0, 0, playerBasePxPs, nil, playerLogger)


	g.game.AddPlayer(p) 

	var combined []byte
	
	for _, conc := range g.game.State.Base.Objects {
		serialized := conc.ToBytes()
		lengthBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(lengthBuf, uint32(len(serialized)))	

		combined = append(combined, lengthBuf...)   // [4 bytes: length]
		combined = append(combined, serialized...)  // [n bytes: object]
	}
	

	jsonBytes, err := json.Marshal(g.game.State.Base.Objects)

	if err != nil {
		panic(err)
	}

	templateData := TemplateData{
		GameState: template.JS(jsonBytes),
		PlayerID: slot,
		Token: token,
		Binary: combined,
	}

	tmpl, err := template.ParseFiles("views/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = tmpl.Execute(w, templateData)
	if err != nil {
		utils.RenderMessage(w, utils.MessageData{
			Type: "error",
			Title: "Rendering Error",
			Message: "There was an unexpected error while rendering the game",
			Link: "/home",
			LinkText: "Home",
		})
	}
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

	p := g.game.State.Players[userID]

	p.SetConn(conn)


	g.log.Println("User Joined:", p.ID(), "UserID:", p.UserID())

	g.handlePlayerConnection(p)
}

func (g *GameHandler) handlePlayerConnection(p *player.Player) {
	defer func() {
		g.game.RemovePlayer(p) 
		g.log.Println("User Left:", p.ID(), "UserID:", p.UserID())
		err := p.Conn().Close()

		if err != nil {
			g.log.Println("Error")
		}

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

	for _, p := range g.game.State.Players {
		if p != nil && p.Conn() != nil {
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
