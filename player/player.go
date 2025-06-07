package player

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Player struct {
	ID   int
	UserID string
	X    float64
	Y    float64
	Conn *websocket.Conn 
	log  *log.Logger
	writeMu sync.Mutex

	movement map[rune]func(step float64)
}



func NewPlayer(id int,userID string, x, y float64, conn *websocket.Conn,l *log.Logger) *Player {
	p := &Player{
		ID:   id,
		UserID: userID,
		X:    x,
		Y:    y,
		Conn: conn,
		log: l,
	}
	p.movement = map[rune]func(step float64){
		'w': func(step float64) { p.Y -= step },
		's': func(step float64) { p.Y += step },
		'a': func(step float64) { p.X -= step },
		'd': func(step float64) { p.X += step },
	}
	return p
}


func (p *Player) MoveByKey(key rune, step float64) {
	if moveFunc, ok := p.movement[key]; ok {
		moveFunc(step)
	}
}

func (p *Player) Position() (float64,float64) {
	return p.X, p.Y
}

func (p *Player) Notify(bytes []byte) {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	if err := p.Conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
		p.log.Println("Write error to player", p.ID, ":", err)
	}
}
