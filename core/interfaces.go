package core

import (
	"github.com/gorilla/websocket"
)

type ObjectType uint8

const (
	TypeObject ObjectType = iota
	TypeConcreteObject
	TypePlayer
)

type Typed struct {
	Type 	ObjectType
}

func (ty *Typed) SetType(t ObjectType){
	ty.Type = t
}

func (ty *Typed) GetType() ObjectType{
	return ty.Type
}

type Serializable interface {
	ToBytes(buf []byte, start int) int
	Size() int
	DeltaSize() int
	ToDeltaBytes(buf []byte, start int) int
}

type GameObject interface {
	Serializable
	ID() int 
	Children() map[int]GameObject
	AddChild(GameObject)
	RemoveChild(int) GameObject
	SetChild(int,GameObject)   
	MarkClean()
	MarkDirty()
	IsDirty() bool
}

type Entity interface {
	GameObject
	OnTick(delta float64)
	OnFrame(delta float64)
}

type ConcreteObject interface {
	GameObject
	Sprite()
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
	Notify([]byte)
}



type State struct {
	Objects   map[int]GameObject
	ConcreteObjects map[int]ConcreteObject
	Entities   map[int]Entity
	PhysicsObjects map[int]PhysicsObject
}
