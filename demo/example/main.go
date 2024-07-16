package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/byebyebruce/aoi"
)

const (
	gridW = 10
	gridH = 4
	side  = 10
	w     = gridW * side
	h     = gridH * side
)

func main() {
	rand.Seed(time.Now().Unix())
	// init aoi
	a, err := aoi.NewAOIManager[int](w, h, gridW, gridH)
	if err != nil {
		panic(err)
	}
	playerID := 0

	// add 100 npc
	for i := playerID + 1; i <= 100; i++ {
		randX, randY := rand.Int()%w, rand.Int()%h
		ok := a.Enter(i, randX, randY, nil)
		if !ok {
			panic("add failed")
		}
	}

	var (
		seeList    = map[int]struct{}{}
		fromGridID = 0
		toGridID   = 0
	)

	// player enter
	a.Enter(playerID, w/2, h/2, func(id int) {
		seeList[id] = struct{}{}
	})
	fromGridID = a.ObjGrid(playerID).ID()
	toGridID = a.ObjGrid(playerID).ID()

	fmt.Println("player enter", fromGridID, "seeList:", seeList)

	for i := 0; i < 10; i++ {
		// player move
		randX, randY := rand.Int()%w, rand.Int()%h
		a.Move(playerID, randX, randY, func(event aoi.AOIEvent, id int) {
			if event == aoi.Enter { // npc enters player's view
				seeList[id] = struct{}{}
			} else if event == aoi.Leave { // npc leaves player's view
				delete(seeList, id)
			}
		})
		toGridID = a.ObjGrid(playerID).ID()
		fmt.Printf("player move %d->%d. seeList:%v\n", fromGridID, toGridID, seeList)
		fromGridID = toGridID

		time.Sleep(time.Second)
	}

	// player leave
	a.Leave(playerID, func(id int) {
		delete(seeList, id)
	})
	fmt.Println("player leave. seeList:", seeList)

	if len(seeList) != 0 {
		panic("leave failed")
	}

}
