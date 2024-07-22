package aoi

import (
	"fmt"
)

func Example() {
	a, err := NewAOIManager[int](100, 100, 10, 5)
	if err != nil {
		panic(err)
	}

	gmID := 0
	npcID := 1
	playerID := 2

	// npc enter
	a.Enter(npcID, 10, 10, Trigger, nil)

	notifierFunc := func(id int) func(event EventType, other int) {
		return func(event EventType, other int) {
			// player see npc
			if event == EnterView {
				fmt.Println(id, "see", other)
			} else if event == LeaveView {
				fmt.Print(id, "lost", other)
			} else if event == UpdateView {
				fmt.Print(id, "move", other)
			}
		}
	}

	// player enter
	a.Enter(playerID, 10, 10, TriggerAndObserver, notifierFunc(playerID))

	// gm enter
	a.Enter(gmID, 10, 10, Observer, notifierFunc(gmID))

	// player move
	a.Move(playerID, 11, 11, notifierFunc(playerID))

	// gm move
	a.Move(gmID, 11, 11, notifierFunc(gmID))

	// player leave
	a.Leave(playerID, notifierFunc(playerID))
}
