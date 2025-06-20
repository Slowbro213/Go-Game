package game


import (

	"gametry.com/core"
	"gametry.com/player"
)


type State struct{

	Base    *core.State
	Players map[string]*player.Player

}
