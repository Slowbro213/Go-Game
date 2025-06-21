package game

import (
	"game/core"
)

type MovementEffect struct {
	Direction string
}

func (e *MovementEffect) Apply(obj core.GameObject) {
	if character, IsCharacter := obj.(Character); IsCharacter {
		character.Move(e.Direction)
	}
}

