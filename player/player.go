package player

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Vector struct {
	X float32
	Y float32
}

func (v *Vector) Add(other Vector) {
	v.X += other.X
	v.Y += other.Y
}

func (v *Vector) Scale(factor float32) Vector {
	return Vector{
		X: v.X * factor,
		Y: v.Y * factor,
	}
}



type Player struct {
	ID         int
	UserID     string
	Position   Vector
	Speed      Vector
	lastUpdate time.Time // Track last update time
	Conn       *websocket.Conn
	log        *log.Logger
	writeMu    sync.Mutex
	pxps float32

	movement map[rune]func()
}


func NewPlayer(id int, userID string, x, y,step float32, conn *websocket.Conn, l *log.Logger) *Player {
	now := time.Now()
	p := &Player{
		ID:         id,
		UserID:     userID,
		Position:   Vector{X: x, Y: y},
		Conn:       conn,
		log:        l,
		lastUpdate: now,
		pxps: step,
	}
	p.movement = map[rune]func(){
		'w': func() { p.Speed.Y = -1 },
		's': func() { p.Speed.Y = 1 },
		'a': func() { p.Speed.X = -1 },
		'd': func() { p.Speed.X = 1 },
	}
	return p
}


// Called to handle input and update speed vector
func (p *Player) NewInput(inputs []byte) {

	p.Speed.Y = 0
	p.Speed.X = 0

	sum := 0
	for _ , input := range inputs {
		key := rune(input)
		if moveFunc, ok := p.movement[key]; ok {
			moveFunc()
			sum++
		}
	}

	if sum > 1{
		p.Speed.Y = p.Speed.Y * 0.7071
		p.Speed.X = p.Speed.X * 0.7071
	}

}

// Applies speed vector to the current position

func (p *Player) ApplySpeed() {
	now := time.Now()
	delta := (now.Sub(p.lastUpdate).Seconds()) // delta in seconds
	p.lastUpdate = now

	scale := float32(delta) * p.pxps
	move := p.Speed.Scale(scale)
	p.Position.Add(move)
}


func (p *Player) PositionXY() (float32, float32) {
	return p.Position.X, p.Position.Y
}

func (p *Player) Notify(bytes []byte) {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	if err := p.Conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
		p.log.Println("Write error to player", p.ID, ":", err)
	}
}
