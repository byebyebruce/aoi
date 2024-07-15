package main

import (
	"math/rand"
	"strconv"
	"time"

	tl "github.com/JoelOtter/termloop"
	"github.com/byebyebruce/aoi"
)

const (
	w    = 10
	h    = 4
	side = 10
	maxw = w * side
	maxh = h * side
)

type game struct {
	*tl.Entity
	level tl.Level
	click *tl.Rectangle

	a          *aoi.AOIManager[int]
	fg         []tl.Drawable
	cg, eg, lg []tl.Drawable
}

func New(level tl.Level, a *aoi.AOIManager[int]) *game {
	g := &game{
		Entity: tl.NewEntity(0, 0, maxw, maxh),
		level:  level,
		a:      a,
	}

	// grids
	for _, v := range a.AllGrids() {
		minx, miny, _, _ := v.Rectangle()
		row, col := v.RowCol()
		var color tl.Attr
		if row%2 == 0 {
			if col%2 == 0 {
				color = tl.ColorBlue
			} else {
				color = tl.ColorWhite
			}
		} else {
			if col%2 == 0 {
				color = tl.ColorWhite
			} else {
				color = tl.ColorBlue
			}
		}
		level.AddEntity(tl.NewRectangle(minx, miny, w, h, color))
		text := tl.NewText(minx+w/2-1, miny+h/2, strconv.Itoa(v.ID()), tl.ColorRed, tl.ColorWhite)
		//level.AddEntity(text)
		g.fg = append(g.fg, text)
	}
	g.click = tl.NewRectangle(0, 0, 1, 1, tl.ColorGreen)
	a.Enter(0, 0, 0, nil)
	g.move(0, 0)

	return g
}
func (g *game) remove(s ...tl.Drawable) {
	for _, v := range s {
		g.level.RemoveEntity(v)
	}
}
func (g *game) add(s ...tl.Drawable) {
	for _, v := range s {
		g.level.AddEntity(v)
	}
}

func (g *game) move(toX, toY int) {
	g.remove(g.fg...)
	g.remove(g.cg...)
	g.remove(g.eg...)
	g.remove(g.lg...)
	g.remove(g.level)
	g.cg, g.eg, g.lg = nil, nil, nil
	_cg, _eg, _lg := g.a.Move(0, toX, toY)
	for _, v := range _cg {
		x, y, _, _ := v.Rectangle()
		r := tl.NewRectangle(x, y, w, h, tl.ColorYellow)
		g.level.AddEntity(r)
		g.cg = append(g.cg, r)
	}
	for _, v := range _eg {
		x, y, _, _ := v.Rectangle()
		r := tl.NewRectangle(x, y, w, h, tl.ColorRed)
		g.level.AddEntity(r)
		g.eg = append(g.eg, r)
	}
	for _, v := range _lg {
		x, y, _, _ := v.Rectangle()
		r := tl.NewRectangle(x, y, w, h, tl.ColorBlack)
		g.level.AddEntity(r)
		g.lg = append(g.lg, r)
	}

	g.add(g.fg...)
	g.click.SetPosition(toX, toY)
	g.add(g.click)
}
func (g *game) Tick(ev tl.Event) {
	if ev.Type != tl.EventMouse {
		return
	}
	switch ev.Key {
	//case tl.MouseLeft:
	case tl.MouseLeft:
		g.move(ev.MouseX, ev.MouseY)
	}
}

func main() {
	g := tl.NewGame()
	g.Screen().SetFps(60)
	l := tl.NewBaseLevel(tl.Cell{
		Bg: tl.ColorBlack | tl.AttrBold,
	})

	rand.Seed(time.Now().Unix())

	g.Screen().SetLevel(l)
	a, _ := aoi.NewAOIManager[int](0, 0, maxw, maxh, w, h)
	game := New(l, a)
	l.AddEntity(game)
	//g.Screen().AddEntity(tl.NewFpsText(0, 0, tl.ColorRed, tl.ColorDefault, 0.5))
	g.Screen().AddEntity(tl.NewText(0, 0, "click to move, yellow:current area, red:enter area, black:leave area", tl.ColorBlue, tl.ColorCyan))
	g.Start()
}
