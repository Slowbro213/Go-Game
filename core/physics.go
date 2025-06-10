package core

type Point struct{
	X,Y float32
}

func (p *Point) Add(other Vector) {
	p.X += other.VX
	p.Y += other.VY
}

type Vector struct {
	VX,VY float32
}

func (v *Vector) Add(other Vector) {
	v.VX += other.VX
	v.VY += other.VY
}

func (v *Vector) Scale(factor float32) {
	v.VX *= factor
	v.VY *= factor
}


var (
	DirRight     = Vector{VX: 1, VY: 0}
	DirLeft      = Vector{VX: -1, VY: 0}
	DirUp        = Vector{VX: 0, VY: -1}
	DirDown      = Vector{VX: 0, VY: 1}

	DirUpLeft    = Vector{VX: -0.70710678118, VY: -0.70710678118} // Use full precision if possible
	DirUpRight   = Vector{VX: 0.70710678118, VY: -0.70710678118}
	DirDownLeft  = Vector{VX: -0.70710678118, VY: 0.70710678118}
	DirDownRight = Vector{VX: 0.70710678118, VY: 0.70710678118}

	DirStop      = Vector{VX: 0, VY: 0}
)


