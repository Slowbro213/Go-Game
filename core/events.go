package core


type IEffect interface {
	Apply(obj GameObject)
}


type ClientEvent struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type Event struct {
	Effects map[int][]IEffect 
	Timestamp int64            
	SourceID  int               
}


type MovementEffect struct {
	Direction string
}

func (e *MovementEffect) Apply(obj GameObject) {
	if character, IsCharacter := obj.(Character); IsCharacter {
		character.Move(e.Direction)
	}
}

