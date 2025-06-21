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
		Typed:   Typed{ Type: "Object"},
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

func (o *Object) ToBytes() []byte {
	return encodeIDAndType(o.id, o.Type)
}

func (o *Object) ToDeltaBytes() []byte {
	return encodeIDAndType(o.id, o.Type)
}



//Helpers
func encodeIDAndType(id int, objType string) []byte {
	typeBytes := []byte(objType)
	buf := make([]byte, 4+len(typeBytes)) 
	binary.LittleEndian.PutUint32(buf[:4], uint32(id))
	copy(buf[4:], typeBytes)
	return buf
}

