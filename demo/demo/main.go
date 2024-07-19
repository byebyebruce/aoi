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
	objs       map[int]*obj
	player     *obj
	mapW, mapH int
	tickCount  int
	a          *aoi.AOIManager[int]
	pause      bool
	see        map[int]struct{}
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
		see:  map[int]struct{}{},
	}

	for i := 0; i < npcNum; i++ {
		g.objs[i] = &obj{
			id: i,
			x:  rand.Int() % g.mapW,
			y:  rand.Int() % g.mapH,
		}
		if i == 0 {
			g.player = g.objs[i]
		}
		g.a.EnterWithType(i, g.objs[i].x, g.objs[i].y, i == 0, nil)
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
		if !o.isPlayer() {
			//continue
		}
		g.a.Move(o.id, o.x, o.y, func(event aoi.EventType, _ int, other int) {
			if o.id == g.player.id {
				if event == aoi.Enter || event == aoi.Move {
					g.see[other] = struct{}{}
				} else if event == aoi.Leave {
					delete(g.see, other)
				}
			} else if other == g.player.id {
				if event == aoi.Enter || event == aoi.Move {
					g.see[o.id] = struct{}{}
				} else if event == aoi.Leave {
					delete(g.see, o.id)
				}
			}
		})
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
		gridW = 8
		gridH = 4
		mapW  = gridW * 20
		mapH  = gridH * 10
	)
	rand.Seed(time.Now().Unix())
	up2down := func(x1, y1, x2, y2 int) (int, int, int, int) {
		return x1, mapH - y1, x2, mapH - y2
	}

	var (
		init    = false
		g       = newGame(100, mapW, mapH, gridW, gridH)
		ds      = make([]ui.Drawable, 0)
		objGrid = map[int]*widgets.Paragraph{}
	)

	render := func() {

		if !init {
			init = true
			for _, v := range g.a.AllGrids() {
				p := widgets.NewParagraph()
				p.Text = strconv.Itoa(v.ID())
				p.SetRect(up2down(v.BoundingBox()))
				ds = append(ds, p)
			}

			for _, o := range g.objs {
				dot := widgets.NewParagraph()
				dot.SetRect(up2down(o.x, o.y, o.x+1, o.y+1))
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
			dot.SetRect(up2down(o.x-1, o.y-1, o.x+1, o.y+1))
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
					g.player.x, g.player.y = m.X, mapH-m.Y
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
