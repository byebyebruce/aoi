package main

import (
	"math/rand"
	"time"
)

const (
	gridW = 8
	gridH = 4
	mapW  = gridW * 20
	mapH  = gridH * 10
)

func main() {
	rand.Seed(time.Now().Unix())

	g := newGame(100, mapW, mapH, gridW, gridH)
	g.choosePlayer(0)

	runApp(g)
}
