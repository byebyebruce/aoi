package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/byebyebruce/aoi"
	"github.com/byebyebruce/varmon/dashboard"
	"github.com/gizak/termui/v3"
)

type obj struct {
	id     int
	x, y   int
	vx, vy int
	see    map[int]struct{}
}

func (o *obj) name() string {
	return strconv.Itoa(o.id)
}
func (o *obj) isPlayer() bool {
	return o.id == 0
}

type game struct {
	objs                   map[int]*obj
	player                 *obj
	minx, miny, maxx, maxy int
	tickCount              int
	a                      *aoi.AOIManager[int]
	pause                  bool
}

func newGame(npcNum int, minx, miny, maxx, maxy, w, h int) *game {
	a, err := aoi.NewAOIManager[int](minx, miny, maxx, maxy, w, h)
	if err != nil {
		panic(err)
	}
	g := &game{
		minx: minx,
		miny: miny,
		maxx: maxx,
		maxy: maxy,
		a:    a,
		objs: map[int]*obj{},
	}

	g.player = &obj{
		id:  0,
		x:   g.maxx / 2,
		y:   g.maxy / 2,
		see: map[int]struct{}{},
	}
	g.objs[g.player.id] = g.player
	for i := 1; i < npcNum; i++ {
		g.objs[i] = &obj{
			id: i,
			x:  rand.Int() % g.maxx,
			y:  rand.Int() % g.maxy,
		}
	}

	return g
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
			o.x = g.maxx
		}
		if o.x > g.maxx {
			o.x = 0
		}
		o.y += o.vy
		if o.y < 0 {
			o.y = g.maxy
		}
		if o.y > g.maxy {
			o.y = 0
		}
	}
	for _, o := range g.objs {
		_, eg, lg := g.a.Move(o.id, o.x, o.y)
		if o.isPlayer() {
			eg.Foreach(func(id int) bool {
				o.see[id] = struct{}{}
				return true
			})
			lg.Foreach(func(id int) bool {
				delete(o.see, id)
				return true
			})
		} else {
			eg.Foreach(func(id int) bool {
				if g.objs[id].isPlayer() {
					g.objs[id].see[o.id] = struct{}{}
				}
				return true
			})
			lg.Foreach(func(id int) bool {
				if g.objs[id].isPlayer() {
					delete(g.objs[id].see, o.id)
				}
				return true
			})
		}
	}

	g.tickCount++
	if g.tickCount%10 == 0 {
		for _, o := range g.objs {
			if o.isPlayer() {
				continue
			}
			o.vx = -1 + rand.Int()%3
			o.vy = -1 + rand.Int()%3
		}
	}

}

var (
	g          = flag.Int("game", 256, "games instance")
	n          = flag.Int("npc", 200, "npc number")
	w          = flag.Int("w", 1000, "map width")
	h          = flag.Int("h", 1000, "map height")
	gridLength = flag.Int("gl", 10, "grid length")
	tick       = flag.Int64("tick", 33, "tick mill")
)

func main() {
	rand.Seed(time.Now().Unix())

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < *g; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g := newGame(*n, 0, 0, *w, *h, *gridLength, *gridLength)
			tick := time.NewTicker(time.Millisecond * time.Duration(*tick))
			defer tick.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-tick.C:
					g.tick()
				}
			}
		}()
	}

	ui, err := dashboard.NewDefault(os.Args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	tick := time.NewTicker(time.Second * 2)

	for {
		select {
		case <-tick.C:
			ui.Update()
		case e := <-termui.PollEvents():
			switch e.Type {
			case termui.KeyboardEvent:
				if e.ID == "q" || e.ID == "<C-c>" {
					return
				}
			case termui.ResizeEvent:
				ui.Relayout()
			}
		}
	}
	cancel()
	wg.Wait()
}
