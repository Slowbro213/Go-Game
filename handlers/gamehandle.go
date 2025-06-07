package handlers


import (
	"fmt"
	"net/http"
	"log"
	"os"
	"github.com/gorilla/websocket"
	"gametry.com/player"
	"sync"
	"errors"
	"encoding/json"
	"github.com/gorilla/sessions"

)

type AuthStatus int

const (
	AuthOK AuthStatus = iota
	AuthUnauthenticated
	AuthDuplicate
)

type AuthResult struct {
	Status  AuthStatus
	ID      int
	Session *sessions.Session
}

type Event struct{
	Type string
	Data map[string]interface{}
}


type GameHandler struct{
	log *log.Logger
	upgrader *websocket.Upgrader
	playersMu sync.Mutex
	players map[int]*player.Player 
	store *sessions.CookieStore
	nextID int
}

func NewGameHandler(l* log.Logger, u* websocket.Upgrader, s* sessions.CookieStore) *GameHandler {
	return &GameHandler{
		log: l,
		upgrader: u,
		players : make(map[int]*player.Player),
		store: s,
	}
}




func (g *GameHandler) checkAuth(r *http.Request) (AuthResult, error) {
	session, err := g.store.Get(r, "poked-cookie")
	if err != nil {
		return AuthResult{Status: AuthUnauthenticated}, fmt.Errorf("session retrieval failed: %w", err)
	}

	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		return AuthResult{Status: AuthUnauthenticated}, errors.New("unauthenticated access")
	}

	id, ok := session.Values["player_id"].(int)
	if !ok {
		return AuthResult{Status: AuthOK, ID: -1, Session: session}, nil
	}

	g.playersMu.Lock()
	_, exists := g.players[id]
	g.playersMu.Unlock()

	if exists {
		return AuthResult{Status: AuthDuplicate, ID: id, Session: session}, errors.New("duplicate connection")
	}

	return AuthResult{Status: AuthOK, ID: id, Session: session}, nil
}



func (g *GameHandler) Auth(w http.ResponseWriter, r *http.Request) {
	result, err := g.checkAuth(r)
	if err != nil {
		switch result.Status {
		case AuthUnauthenticated:
			http.Error(w, "Unauthenticated", http.StatusUnauthorized)
		case AuthDuplicate:
			http.Error(w, "Duplicate connection", http.StatusForbidden)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		g.log.Println(err)
		return
	}

	id := result.ID
	if id == -1 {
		g.playersMu.Lock()
		id = g.nextID
		g.nextID++
		g.playersMu.Unlock()

		result.Session.Values["player_id"] = id
		if err := result.Session.Save(r, w); err != nil {
			g.log.Println("Failed to save session:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}   

	g.log.Println("Auth Success for ID:", id)
	fmt.Fprintf(w, "%d", id)
}

func (g *GameHandler) Match(w http.ResponseWriter, r *http.Request) {
	result, err := g.checkAuth(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	conn, err := g.upgrader.Upgrade(w, r, nil)
	if err != nil {
		g.log.Println("WebSocket upgrade error:", err)
		return
	}

	Id := result.ID
	l := log.New(os.Stdout, fmt.Sprintf("Player %d: ",Id), log.LstdFlags)
	p := player.NewPlayer(Id, 0, 0, conn,l)

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
		delete(g.players, id)
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
		x, y := player.Position()
		positions[id] = [2]float64{x, y}
	}

	return positions

}

func (g* GameHandler) broadcastMessage(bytes []byte) {
	for _, player := range g.players {
		player.Notify(bytes)
	}
}
