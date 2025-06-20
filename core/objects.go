package core

import (
	"github.com/gorilla/websocket"
)


type GameObject interface {
	ID() int 
	Children() map[int]GameObject
}

type Entity interface {
	GameObject
	OnTick(delta float64)
	OnFrame(delta float64)
}

type ConcreteObject interface {
	GameObject
	PositionXY() Point
}

type PhysicsObject interface {
	GameObject
	Entity
	Velocity() Vector         
	SetVelocity(Vector)       
	ApplyForce(Vector)        
	ApplyAcceleration(Vector)
}



type NetworkObject interface {
	GameObject
	Conn() *websocket.Conn
	CloseConn()
	SetConn(*websocket.Conn)
}


type Notifiable interface {
	Entity
	Notify()
}

type Character interface {
	//Health()
	//SetHealth()
	//Mana()
	Entity	
	GetSpeed() float32
	Move(string)
}


type State struct {

	Objects   map[int]GameObject
	Entities   map[int]Entity

}
