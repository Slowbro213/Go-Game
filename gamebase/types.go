package gamebase

import (
	"game/core"
)



type Character interface {
	core.ConcreteObject
	GetSpeed() float32
	Move(string)
}

// You can define other high-level concepts here
type Wall interface {
	core.ConcreteObject
	IsSolid() bool
}

type Projectile interface {
	core.PhysicsObject
	Damage() int
}



