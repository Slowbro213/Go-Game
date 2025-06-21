package core


import (
	"encoding/binary"
	"math"
)

type Concrete struct {
	Object
	Position    Point
}

func NewConcreteObject(id int,children map[int]GameObject, pos Point) *Concrete{

	return &Concrete{
		Object:    *NewObject(id,children),
		Position:  pos,
	}

}


func (c *Concrete) PositionXY() Point {
	return c.Position
}


func (c *Concrete) Sprite(){

}

//Serializable
func (c *Concrete) ToBytes() []byte {
	return encodeConcrete(c)
}

func (c *Concrete) ToDeltaBytes() []byte {
	return encodeConcrete(c)
}


//Helpers
func encodeConcrete(c *Concrete) []byte {
	base := c.Object.ToBytes()
	buf := make([]byte, len(base)+8) 

	copy(buf, base)
	offset := len(base)

	binary.LittleEndian.PutUint32(buf[offset:offset+4], math.Float32bits(c.Position.X))
	binary.LittleEndian.PutUint32(buf[offset+4:offset+8], math.Float32bits(c.Position.Y))

	return buf
}


