package core 

import (
	"encoding/binary"
)


type Object struct {
	Typed
	id       int
	children map[int]GameObject
	dirty    bool
}

func NewObject(id int,children map[int]GameObject) *Object {
	return &Object{
		id:  id,
		children: children,
		dirty: false,
		Typed:   Typed{ Type: TypeObject},
	}
}


func (o *Object) ID() int {
	return o.id
}

func (o *Object) Children() map[int]GameObject {
	return o.children
}

func (o *Object) AddChild(child GameObject) {
	if o.children == nil {
		o.children = make(map[int]GameObject)
	}
	o.children[child.ID()] = child
}

func (o *Object) RemoveChild(id int) GameObject {
	child := o.children[id]
	delete(o.children, id)
	return child
}

func (o *Object) SetChild(id int, child GameObject) {
	if o.children == nil {
		o.children = make(map[int]GameObject)
	}
	o.children[id] = child
}

func (o *Object) MarkClean() {
	o.dirty = false
}

func (o *Object) MarkDirty(){
	o.dirty = true
}

func (o *Object) IsDirty() bool {
	return o.dirty
}

//Serializable

func (o *Object) ToBytes(buf []byte, start int) int {
	return writeIDAndType(buf, start, o.id, o.Type)
}


func (o *Object) ToDeltaBytes(buf []byte, start int) int {
	return writeIDAndType(buf, start, o.id, o.Type)
}


func (o *Object) Size() int {
	return 4 + 1
}


//Helpers
func writeIDAndType(buf []byte, start int, id int, objType ObjectType) int {
	binary.LittleEndian.PutUint32(buf[start:start+4], uint32(id))
	buf[start+4] = byte(objType)
	return 5 // 4 bytes ID + 1 byte type
}


