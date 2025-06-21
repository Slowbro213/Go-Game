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
	Sprite()
	Type() string
	PositionXY() Point
}

type PhysicsObject interface {
	ConcreteObject
	Entity
	Velocity() *Vector         
	SetVelocity(*Vector)       
	ApplyForce(*Vector)        
	ApplyAcceleration(*Vector)
}



type NetworkObject interface {
	GameObject
	Conn() *websocket.Conn
	CloseConn() error
	SetConn(*websocket.Conn)
}


type Notifiable interface {
	Entity
	Notify(any)
}



type State struct {
	Objects   map[int]GameObject
	ConcreteObjects map[int]ConcreteObject
	Entities   map[int]Entity
	PhysicsObjects map[int]PhysicsObject
}
