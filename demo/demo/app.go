package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

func runApp(g *game) error {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	up2down := func(x1, y1, x2, y2 int) (int, int, int, int) {
		return x1, mapH - y1, x2, mapH - y2
	}

	var (
		init    = false
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

	for {
		select {
		case <-tick.C:
			g.tick()
			render()
		case e := <-uiEvents:
			if e.Type == ui.KeyboardEvent {
				switch e.ID {
				case "q", "<C-c>":
					return nil
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
					g.currentPlayer.setVelocity(0, 0)
				}
			}
		}

	}

}
