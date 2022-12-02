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

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	const (
		w    = 5
		h    = 3
		maxw = w * 10
		maxh = h * 10
	)
	rand.Seed(time.Now().Unix())
	a, _ := aoi.NewAOIManager[int](0, 0, maxw, maxh, w, h)

	up2down := func(x1, y1, x2, y2 int) (int, int, int, int) {
		return x1, maxh - y1, x2, maxh - y2
	}
	x, y := 0, 0
	cg, eg, lg := a.Move(1, x, y)
	render := func() {
		ds := make([]ui.Drawable, 0)

		for _, v := range a.AllGrids() {
			p := widgets.NewParagraph()
			p.Text = strconv.Itoa(v.ID())
			p.SetRect(up2down(v.Rectangle()))
			ds = append(ds, p)
		}

		for _, v := range lg {
			p := widgets.NewParagraph()
			p.Text = strconv.Itoa(v.ID())
			p.SetRect(up2down(v.Rectangle()))
			p.BorderStyle = ui.NewStyle(ui.ColorBlack)
			ds = append(ds, p)
		}
		for _, v := range cg {
			p := widgets.NewParagraph()
			p.Text = strconv.Itoa(v.ID())
			p.SetRect(up2down(v.Rectangle()))
			p.BorderStyle = ui.NewStyle(ui.ColorYellow)
			ds = append(ds, p)
		}
		for _, v := range eg {
			p := widgets.NewParagraph()
			p.Text = strconv.Itoa(v.ID())
			p.SetRect(up2down(v.Rectangle()))
			p.BorderStyle = ui.NewStyle(ui.ColorRed)
			ds = append(ds, p)
		}
		dot := widgets.NewParagraph()
		dot.SetRect(up2down(x-2, y-1, x+2, y+1))
		dot.Text = fmt.Sprintf("(%d,%d)", x, y)
		dot.BorderStyle = ui.NewStyle(ui.ColorWhite)
		ds = append(ds, dot)
		ui.Render(ds...)
	}
	render()

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		if e.Type == ui.KeyboardEvent {
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		} else if e.Type == ui.MouseEvent {
			switch e.ID {
			case "<MouseLeft>":
				m := e.Payload.(ui.Mouse)
				x, y = m.X, maxh-m.Y
				cg, eg, lg = a.Move(1, x, y)
				render()
			}
		}
	}

}
