package core



type GameObject interface {
	ID() int 
	Children() map[int]GameObject
}

type Entity interface {
	OnTick(delta float64)
	OnFrame(delta float64)
}

type ConcreteObject interface {
	PositionXY() Point
}

type PhysicsObject interface {
	Velocity() Vector         
	SetVelocity(Vector)       
	ApplyForce(Vector)        
	ApplyAcceleration(Vector)
}


type NetworkObject interface {
	Conn()
	CloseConn()
}

type Notifiable interface {
	Notify()
}

type Character interface {
	//Health()
	//SetHealth()
	//Mana()
	GetSpeed() float32
	Move(string)
}
