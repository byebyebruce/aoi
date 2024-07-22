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

const (
	gridW = 8
	gridH = 4
	mapW  = gridW * 20
	mapH  = gridH * 10
)

const playerNumber = 10

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
		g.objs[i] = &obj{
			id:         i,
			x:          rand.Int() % g.mapW,
			y:          rand.Int() % g.mapH,
			seeList:    map[int]struct{}{},
			playerFlag: true,
		}
		g.a.Enter(i, g.objs[i].x, g.objs[i].y, aoi.TriggerAndObserver, func(_ aoi.EventType, other int) {
			g.objs[i].seeList[other] = struct{}{}
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
				o.seeList[other] = struct{}{}
				g.objs[other].seeList[o.id] = struct{}{}
			} else if event == aoi.LeaveView {
				delete(o.seeList, other)
				delete(g.objs[other].seeList, o.id)
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

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

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
	g.choosePlayer(0)

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
				ds = append(ds, dot)
				objGrid[o.id] = dot
			}
		}
		for _, o := range g.objs {
			dot := objGrid[o.id]
			if dot == nil {
				dot = widgets.NewParagraph()
				ds = append(ds, dot)
				objGrid[o.id] = dot
			}
			dot.SetRect(up2down(o.x-1, o.y-1, o.x+1, o.y+1))
			dot.Text = fmt.Sprintf(o.name())
			if o.id == g.currentPlayer.id {
				dot.BorderStyle = ui.NewStyle(ui.ColorRed, ui.ColorRed)
			} else {
				if _, ok := g.currentPlayer.seeList[o.id]; ok {
					if o.isPlayer() {
						dot.BorderStyle = ui.NewStyle(ui.ColorBlue, ui.ColorBlue)
					} else {
						dot.BorderStyle = ui.NewStyle(ui.ColorYellow, ui.ColorYellow)
					}
				} else {
					if o.isPlayer() {
						dot.BorderStyle = ui.NewStyle(ui.ColorBlue, ui.ColorBlack)
					} else {
						dot.BorderStyle = ui.NewStyle(ui.ColorYellow, ui.ColorBlack)
					}
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
					g.currentPlayer.vx = -playerSpeed
				case "s":
					g.currentPlayer.vy = -playerSpeed
				case "w":
					g.currentPlayer.vy = playerSpeed
				case "d":
					g.currentPlayer.vx = playerSpeed
				}
				if e.ID >= "0" && e.ID <= "9" {
					i, _ := strconv.Atoi(e.ID)
					g.choosePlayer(i)
				}
			} else if e.Type == ui.MouseEvent {
				switch e.ID {
				case "<MouseLeft>":
					m := e.Payload.(ui.Mouse)
					g.currentPlayer.setPos(m.X, mapH-m.Y)
					//cg, eg, lg = a.Move(1, x, y)
					//render()
					g.currentPlayer.setVelocity(0, 0)
				}
				//fmt.Println(e.Payload)
			} else {
				g.currentPlayer.setVelocity(0, 0)
			}
		}

	}

}
