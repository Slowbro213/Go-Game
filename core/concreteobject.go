package core


import (
	"encoding/binary"
	"math"
//	"fmt"
)

type Concrete struct {
	Object
	Position     Point
	PrevPosition Point
}

func NewConcreteObject(id int,children map[int]GameObject, pos Point) *Concrete{

	return &Concrete{
		Object:       *NewObject(id,children),
		Position:     pos,
		PrevPosition: pos,
	}

}


func (c *Concrete) PositionXY() Point {
	return c.Position
}


func (c *Concrete) Sprite(){

}

//Serializable

func (c *Concrete) ToBytes(buf []byte, start int) int {
	offset := start

	// Write embedded Object
	offset += c.Object.ToBytes(buf, offset)

	// Write Position.X
	binary.LittleEndian.PutUint32(buf[offset:offset+4], math.Float32bits(c.Position.X))
	offset += 4

	// Write Position.Y
	binary.LittleEndian.PutUint32(buf[offset:offset+4], math.Float32bits(c.Position.Y))
	offset += 4

	return offset - start
}

func (c *Concrete) ToDeltaBytes(buf []byte, start int) int {
	if !c.IsDirty(){
		return 0
	}
	offset := start

	// Write embedded Object
	offset += c.Object.ToBytes(buf, offset)

	// Write Position.X
	binary.LittleEndian.PutUint32(buf[offset:offset+4], math.Float32bits(c.Position.X))
	offset += 4

	// Write Position.Y
	binary.LittleEndian.PutUint32(buf[offset:offset+4], math.Float32bits(c.Position.Y))
	offset += 4

//	fmt.Printf("Concrete: %v offset: %d\n",buf,offset)

	return offset - start
}


func (c *Concrete) Size() int {
	return c.Object.Size() + 8 // 4 bytes for X + 4 bytes for Y
}

//Also returning if the object is dirty
func (c *Concrete) DeltaSize() int {
	if c.Position.X == c.PrevPosition.X && c.Position.Y == c.PrevPosition.Y{
		c.Object.MarkClean()
		return c.Object.Size()
	}

	c.Object.MarkDirty()
	c.PrevPosition.X = c.Position.X
	c.PrevPosition.Y = c.Position.Y
	return c.Object.Size() + 8


}

//Helpers



