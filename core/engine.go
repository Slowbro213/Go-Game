package core

import (
	"math"
	"sync"
	"time"
)

type Engine struct {
	
	State     *State
	stateMu   sync.RWMutex

	tickInterval   time.Duration
	fixedTickDelta float64
	targetFPS int

	done chan struct{}
	wg   sync.WaitGroup

	OnFixedUpdate func(delta float64)
	OnVariableUpdate func(delta float64)

	eventQueue chan *Event
}

func NewEngine(state *State,fixedTPS float64, targetFPS int) *Engine {
	tickInterval := time.Duration(int(math.Round(1000.0/fixedTPS))) * time.Millisecond

	engine := &Engine{
		State:          state,
		tickInterval:   tickInterval,
		fixedTickDelta: tickInterval.Seconds(),
		targetFPS:      targetFPS,
		done:           make(chan struct{}),
		eventQueue:     make(chan *Event, 1000),
	}


	return engine
}

func (e *Engine) Run() {
	e.wg.Add(3)

	go func() {
		defer e.wg.Done()
		e.runFixedUpdateLoop()
	}()

	go func() {
		defer e.wg.Done()
		e.runVariableUpdateLoop()
	}()

	go func() {
		defer e.wg.Done()
		e.eventConsumer()
	}()
}

func (e *Engine) runFixedUpdateLoop() {
	ticker := time.NewTicker(e.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			e.stateMu.Lock() 
			for _, ent := range e.State.Entities { 
				ent.OnTick(e.fixedTickDelta)
			}
			e.stateMu.Unlock()

			if e.OnFixedUpdate != nil {
				e.OnFixedUpdate(e.fixedTickDelta)
			}
		case <-e.done:
			return
		}
	}
}


func (e *Engine) runVariableUpdateLoop() {
	targetFrameDuration := time.Second / time.Duration(e.targetFPS)

	lastFrameTime := time.Now()

	for {
		select {
		case <-e.done:
			return
		default:
			now := time.Now()
			delta := now.Sub(lastFrameTime).Seconds()
			lastFrameTime = now

			e.stateMu.RLock()
			for _, ent := range e.State.Entities {
				ent.OnFrame(delta)
			}
			e.stateMu.RUnlock()

			if e.OnVariableUpdate != nil {
				e.OnVariableUpdate(delta)
			}

			// Sleep to maintain target frame rate
			frameElapsed := time.Since(now)
			sleepDuration := targetFrameDuration - frameElapsed
			if sleepDuration > 0 {
				time.Sleep(sleepDuration)
			}
		}
	}
}




func (e *Engine) eventConsumer() {
	for {
		select {
		case ev, ok := <-e.eventQueue:
			if !ok {
				return
			}

			e.stateMu.Lock()
			for objID, effects := range ev.Effects {
				obj, exists := e.State.Objects[objID]
				if !exists || obj == nil {
					continue
				}

				for _, effect := range effects {
					effect.Apply(obj)
				}
			}
			e.stateMu.Unlock()

		case <-e.done:
			return
		}
	}
}



func (e *Engine) HandleEvent(ev *Event) {
	select {
	case e.eventQueue <- ev:
	default:
	}
}

func (e *Engine) Shutdown() {
	close(e.done)
	close(e.eventQueue)
	e.wg.Wait()
}


func (e *Engine) AddObject(obj GameObject) {
	id := obj.ID()

	e.stateMu.Lock()
	defer e.stateMu.Unlock()

	e.State.Objects[id] = obj

	if ent, ok := obj.(Entity); ok {
		e.State.Entities[id] = ent
	}
	if phys, ok := obj.(PhysicsObject); ok {
		e.State.PhysicsObjects[id] = phys
	}
	if con, ok := obj.(ConcreteObject); ok {
		e.State.ConcreteObjects[id] = con
	}
}

func (e *Engine) RemoveObject(id int) {
    e.stateMu.Lock()
		defer e.stateMu.Unlock()

    delete(e.State.Objects, id)

    delete(e.State.Entities, id) 

		delete(e.State.ConcreteObjects,id)

		delete(e.State.PhysicsObjects,id)
}


func (e *Engine) GetObject(id int) GameObject {
    e.stateMu.RLock()
    obj := e.State.Objects[id]
    e.stateMu.RUnlock()
    return obj
}
