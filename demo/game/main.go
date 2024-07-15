package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/byebyebruce/aoi"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type obj struct {
	id     int
	x, y   int
	vx, vy int
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
	see                    map[int]struct{}
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
		see:  map[int]struct{}{},
	}

	for i := 0; i < npcNum; i++ {
		g.objs[i] = &obj{
			id: i,
			x:  rand.Int() % g.maxx,
			y:  rand.Int() % g.maxy,
		}
		if i == 0 {
			g.player = g.objs[i]
		}
		g.a.Enter(i, g.objs[i].x, g.objs[i].y, nil)
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
		if !o.isPlayer() {
			//continue
		}
		g.a.Move(o.id, o.x, o.y, func(event aoi.AOIEvent, eventMaker int, eventWatcher int) {
			if eventMaker == g.player.id {
				if event == aoi.Enter || event == aoi.Move {
					g.see[eventWatcher] = struct{}{}
				} else if event == aoi.Leave {
					delete(g.see, eventWatcher)
				}
			} else if eventWatcher == g.player.id {
				if event == aoi.Enter || event == aoi.Move {
					g.see[eventMaker] = struct{}{}
				} else if event == aoi.Leave {
					delete(g.see, eventMaker)
				}
			}
		})
		/*
			_, eg, lg :=
			if o.isPlayer() {
				eg.ForeachObj(func(id int) bool {
					o.see[id] = struct{}{}
					return true
				})
				lg.ForeachObj(func(id int) bool {
					delete(o.see, id)
					return true
				})
			} else {
				eg.ForeachObj(func(id int) bool {
					if g.objs[id].isPlayer() {
						g.objs[id].see[o.id] = struct{}{}
					}
					return true
				})
				lg.ForeachObj(func(id int) bool {
					if g.objs[id].isPlayer() {
						delete(g.objs[id].see, o.id)
					}
					return true
				})
			}

		*/
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

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	const (
		w    = 8
		h    = 4
		maxw = w * 20
		maxh = h * 10
	)
	rand.Seed(time.Now().Unix())
	up2down := func(x1, y1, x2, y2 int) (int, int, int, int) {
		return x1, maxh - y1, x2, maxh - y2
	}

	var (
		init    = false
		g       = newGame(100, 0, 0, maxw, maxh, w, h)
		ds      = make([]ui.Drawable, 0)
		objGrid = map[int]*widgets.Paragraph{}
	)

	render := func() {

		if !init {
			init = true
			for _, v := range g.a.AllGrids() {
				p := widgets.NewParagraph()
				p.Text = strconv.Itoa(v.ID())
				p.SetRect(up2down(v.Rectangle()))
				ds = append(ds, p)
			}

			for _, o := range g.objs {
				dot := widgets.NewParagraph()
				dot.SetRect(up2down(o.x, o.y, o.x+4, o.y+2))
				dot.Text = fmt.Sprintf(o.name())
				/*
					if o.isPlayer() {
						dot.BorderStyle = ui.NewStyle(ui.ColorRed, ui.ColorRed)
					} else {
						if _, ok := g.player.see[o.id]; ok {
							dot.BorderStyle = ui.NewStyle(ui.ColorYellow, ui.ColorYellow)
						} else {
							dot.BorderStyle = ui.NewStyle(ui.ColorCyan, ui.ColorCyan)
							//continue
						}
					}
				*/
				ds = append(ds, dot)
				objGrid[o.id] = dot
			}

		}
		for _, o := range g.objs {
			dot := objGrid[o.id]
			dot.SetRect(up2down(o.x-2, o.y-1, o.x+2, o.y+1))
			dot.Text = fmt.Sprintf(o.name())
			if o.isPlayer() {
				dot.BorderStyle = ui.NewStyle(ui.ColorRed, ui.ColorRed)
			} else {
				if _, ok := g.see[o.id]; ok {
					dot.BorderStyle = ui.NewStyle(ui.ColorYellow, ui.ColorYellow)
				} else {
					dot.BorderStyle = ui.NewStyle(ui.ColorCyan, ui.ColorCyan)
					//continue
				}
			}
		}

		ui.Clear()
		ui.Render(ds...)
	}

	tick := time.NewTicker(time.Millisecond * 300)
	defer tick.Stop()
	uiEvents := ui.PollEvents()
	playerSpeed := 3
	//fmt.Println("h,j,k,l,space")
	//time.Sleep(time.Second * 5)
	for {
		select {
		case <-tick.C:
			g.tick()
			render()
		case e := <-uiEvents:
			if e.Type == ui.KeyboardEvent {

				switch e.ID {
				case "q", "<C-c>":
					return
				case "p":
					g.pause = !g.pause
				case "a":
					g.player.vx = -playerSpeed
				case "s":
					g.player.vy = -playerSpeed
				case "w":
					g.player.vy = playerSpeed
				case "d":
					g.player.vx = playerSpeed
				}
			} else if e.Type == ui.MouseEvent {
				switch e.ID {
				case "<MouseLeft>":
					m := e.Payload.(ui.Mouse)
					g.player.x, g.player.y = m.X, maxh-m.Y
					//cg, eg, lg = a.Move(1, x, y)
					//render()
					g.player.vx = 0
					g.player.vy = 0
				}
				//fmt.Println(e.Payload)
			} else {
				g.player.vx = 0
				g.player.vy = 0
			}
		}

	}

}
