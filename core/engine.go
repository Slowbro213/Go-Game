package core

import (
	"math"
	"sync"
	"time"
)

type Engine struct {
	Objects   map[int]GameObject
	objectsMu sync.RWMutex

	Entities   map[int]Entity
	entitiesMu sync.RWMutex

	tickInterval   time.Duration
	fixedTickDelta float64

	done chan struct{}
	wg   sync.WaitGroup

	OnFixedUpdate func(delta float64)
	OnVariableUpdate func(delta float64)

	eventQueue chan *Event
}

func NewEngine(fixedTPS float64, objects ...GameObject) *Engine {
	tickInterval := time.Duration(int(math.Round(1000.0/fixedTPS))) * time.Millisecond

	engine := &Engine{
		Objects:        make(map[int]GameObject),
		Entities:       make(map[int]Entity),
		tickInterval:   tickInterval,
		fixedTickDelta: tickInterval.Seconds(),
		done:           make(chan struct{}),
		eventQueue:     make(chan *Event, 1000),
	}


	for _, obj := range objects {
		engine.Objects[obj.ID()] = obj
		if ent, ok := obj.(Entity); ok { 
			engine.Entities[obj.ID()] = ent
		}
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

			e.entitiesMu.Lock() 
			for _, ent := range e.Entities { 
				ent.OnTick(e.fixedTickDelta)
			}
			e.entitiesMu.Unlock()

			if e.OnFixedUpdate != nil {
				e.OnFixedUpdate(e.fixedTickDelta)
			}
		case <-e.done:
			return
		}
	}
}

func (e *Engine) runVariableUpdateLoop() {
	lastFrameTime := time.Now()
	for {
		select {
		case <-e.done:
			return
		default:
			now := time.Now()
			variableDelta := now.Sub(lastFrameTime).Seconds()
			lastFrameTime = now

			e.entitiesMu.RLock()
			for _, ent := range e.Entities {
				ent.OnFrame(variableDelta)
			}
			e.entitiesMu.RUnlock()

			if e.OnVariableUpdate != nil {
				e.OnVariableUpdate(variableDelta)
			}

			time.Sleep(time.Millisecond)
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
            e.objectsMu.Lock()
            for objID, effects := range ev.Effects {
                obj := e.Objects[objID]
                if obj == nil {
                    continue
                }
                for _, effect := range effects {
                    effect.Apply(obj)
                }
            }
            e.objectsMu.Unlock()
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
    e.objectsMu.Lock()
    e.Objects[obj.ID()] = obj
    e.objectsMu.Unlock() 

    if ent, ok := obj.(Entity); ok {
        e.entitiesMu.Lock() 
        e.Entities[obj.ID()] = ent
        e.entitiesMu.Unlock()
    }
}

func (e *Engine) RemoveObject(id int) {
    e.objectsMu.Lock()
    delete(e.Objects, id)
    e.objectsMu.Unlock()

    e.entitiesMu.Lock()
    delete(e.Entities, id) 
    e.entitiesMu.Unlock()
}


func (e *Engine) GetObject(id int) GameObject {
    e.objectsMu.RLock()
    obj := e.Objects[id]
    e.objectsMu.RUnlock()
    return obj
}
