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



