package player

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"

	"gametry.com/core"
)

type Player struct {
	Id          int
	userID      string
	Position    core.Point
	VelocityVec core.Vector
	conn        *websocket.Conn
	log         *log.Logger
	writeMu     sync.Mutex
	pxps        float32

	children map[int]core.GameObject


}

func NewPlayer(id int, userID string, x, y, pxps float32, conn *websocket.Conn, l *log.Logger) *Player {
	p := &Player{
		Id:          id,
		userID:      userID,
		Position:    core.Point{X: x, Y: y},
		VelocityVec: core.Vector{VX: 0, VY: 0},
		conn:        conn,
		log:         l,
		pxps:        pxps,

		children: nil,
	}
	return p
}

func (p *Player) UserID() string {
	return p.userID
}

func (p *Player) ID() int {
	return p.Id
}

func (p *Player) State() string {
	return "Happy"
}

func (p *Player) Children() map[int]core.GameObject {
	return p.children
}

func (p *Player) OnTick(delta float64) {
	p.Position.X += p.VelocityVec.VX * float32(delta) 
	p.Position.Y += p.VelocityVec.VY * float32(delta)
}

func (p *Player) OnFrame(delta float64) {
	return
}

func (p *Player) PositionXY() core.Point {
	return p.Position
}

func (p *Player) Velocity() *core.Vector {
	return &p.VelocityVec
}

func (p *Player) SetVelocity(v *core.Vector) {
	p.VelocityVec.VX = v.VX
	p.VelocityVec.VY = v.VY
}

func (p *Player) ApplyForce(v core.Vector) {
	p.VelocityVec.VX += v.VX
	p.VelocityVec.VY += v.VY
}

func (p *Player) ApplyAcceleration(v core.Vector) {
	p.VelocityVec.VX += v.VX
	p.VelocityVec.VY += v.VY
}

func (p *Player) Conn() *websocket.Conn {
	return p.conn
}

func (p *Player) CloseConn() {
	p.conn.Close()
}

func(p *Player) SetConn(c *websocket.Conn) {
	p.conn = c
}

func (p *Player) Notify(bytes []byte) {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	if err := p.conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
		p.log.Println("Write error to player", p.ID(), ":", err)
	}
}


func (p *Player) GetSpeed() float32 {
	return p.pxps
}



func (p *Player) Move(rawInput string) {
	var baseDirection core.Vector


	switch rawInput {
	case "move_right":
		baseDirection = core.DirRight
	case "move_left":
		baseDirection = core.DirLeft
	case "move_up":
		baseDirection = core.DirUp
	case "move_down":
		baseDirection = core.DirDown
	case "move_up_left":
		baseDirection = core.DirUpLeft
	case "move_up_right":
		baseDirection = core.DirUpRight
	case "move_down_left":
		baseDirection = core.DirDownLeft
	case "move_down_right":
		baseDirection = core.DirDownRight
	case "move_stop":
		baseDirection = core.DirStop
	default:
		p.log.Printf("Unknown input command: %s\n", rawInput)
		return
	}

	p.SetVelocity(&baseDirection) 
	p.Velocity().Scale(p.pxps)
}

