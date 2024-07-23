package main

import (
	"math/rand"
	"strconv"

	"github.com/byebyebruce/aoi"
)

type obj struct {
	id         int
	x, y       int
	vx, vy     int
	playerFlag bool
	seeList    map[int]struct{}
}

func (o *obj) name() string {
	return strconv.Itoa(o.id)
}
func (o *obj) isPlayer() bool {
	return o.playerFlag
}
func (o *obj) setPos(x, y int) {
	o.x, o.y = x, y
}
func (o *obj) setVelocity(x, y int) {
	o.vx, o.vy = x, y
}
func (o *obj) enterView(id int) {
	o.seeList[id] = struct{}{}
}

func (o *obj) leaveView(id int) {
	delete(o.seeList, id)
}

type game struct {
	objs          map[int]*obj
	currentPlayer *obj
	mapW, mapH    int
	tickCount     int
	a             *aoi.AOIManager[int]
	pause         bool
}

func newGame(npcNum int, mapW, mapH, w, h int) *game {
	a, err := aoi.NewAOIManager[int](mapW, mapH, w, h)
	if err != nil {
		panic(err)
	}
	g := &game{
		mapW: mapW,
		mapH: mapH,
		a:    a,
		objs: map[int]*obj{},
	}

	for i := 0; i < npcNum; i++ {
		o := &obj{
			id:      1000 + i,
			x:       rand.Int() % g.mapW,
			y:       rand.Int() % g.mapH,
			seeList: map[int]struct{}{},
		}
		g.objs[o.id] = o
		g.a.Enter(o.id, o.x, o.y, aoi.Trigger, nil)
	}

	return g
}

func (g *game) choosePlayer(i int) {
	if g.objs[i] == nil {
		o := &obj{
			id:         i,
			x:          rand.Int() % g.mapW,
			y:          rand.Int() % g.mapH,
			seeList:    map[int]struct{}{},
			playerFlag: true,
		}
		g.objs[i] = o
		g.a.Enter(i, o.x, o.y, aoi.TriggerAndObserver, func(_ aoi.EventType, other int) {
			o.seeList[other] = struct{}{}
		})
	}
	if oldPlayer := g.currentPlayer; oldPlayer != nil {
		oldPlayer.setVelocity(0, 0)
	}
	g.currentPlayer = g.objs[i]
}

func (g *game) tick() {
	for _, o := range g.objs {
		if g.pause {
			if !o.isPlayer() {
				continue
			}
		}
		o.x += o.vx
		if o.x < 0 {
			o.x = g.mapW
		}
		if o.x > g.mapW {
			o.x = 0
		}
		o.y += o.vy
		if o.y < 0 {
			o.y = g.mapH
		}
		if o.y > g.mapH {
			o.y = 0
		}
	}
	for _, o := range g.objs {
		g.a.Move(o.id, o.x, o.y, func(event aoi.EventType, other int) {
			if event == aoi.EnterView {
				o.enterView(other)
				g.objs[other].enterView(o.id)
			} else if event == aoi.LeaveView {
				o.leaveView(other)
				g.objs[other].leaveView(o.id)
			} else if event == aoi.UpdateView {
			}
		})
	}

	g.tickCount++
	if g.tickCount%10 == 0 {
		for _, o := range g.objs {
			if o.isPlayer() {
				continue
			}
			o.setVelocity(-1+rand.Int()%3, -1+rand.Int()%3)
		}
	}

}
