package game


import (

	"game/core"
	"game/player"
)


type State struct{

	Base    *core.State
	Players map[string]*player.Player

}
