package aoi

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAOI_NewAOIManager(t *testing.T) {
	a, err := NewAOIManager[int](0, 0, 100, 100, 10, 5)
	require.Nil(t, err)
	require.EqualValues(t, 10, a.col)
	require.EqualValues(t, 20, a.row)
	require.EqualValues(t, 10*20, len(a.grids))
	fmt.Println(a)

	grid := a.grids[0]
	require.True(t, grid.isSurround(0))
	require.True(t, grid.isSurround(1))
	require.True(t, grid.isSurround(10))
	require.True(t, grid.isSurround(11))

	require.EqualValues(t, 4, len(grid.surroundGrids))
}

type obj struct {
	id   int
	x, y int
}

func TestAOI(t *testing.T) {
	a, err := NewAOIManager[int](0, 0, 100, 100, 10, 10)
	assert.Nil(t, err)
	//fmt.Println(a)
	obj := obj{
		id: 1,
		x:  0,
		y:  0,
	}

	ok := a.Enter(obj.id, obj.x, obj.y, func(event AOIEvent, eventMaker int, eventWatcher int) {
		fmt.Println(event, eventMaker, eventWatcher)
	})
	require.True(t, ok)

	obj.x += 10
	ok = a.Move(obj.id, obj.x, obj.y, func(event AOIEvent, eventMaker int, eventWatcher int) {

	})
	require.True(t, ok)

	/*
		obj.x += 10
		curr, enter, leave = a.Move(obj.id, obj.x, obj.y)
		fmt.Println("curr", curr)
		fmt.Println("enter", enter)
		fmt.Println("leave", leave)
		fmt.Println()

		obj.y += 10
		curr, enter, leave = a.Move(obj.id, obj.x, obj.y)
		fmt.Println("curr", curr)
		fmt.Println("enter", enter)
		fmt.Println("leave", leave)
		fmt.Println()

		obj.y += 10
		curr, enter, leave = a.Move(obj.id, obj.x, obj.y)
		fmt.Println("curr", curr)
		fmt.Println("enter", enter)
		fmt.Println("leave", leave)
		fmt.Println()

		obj.y += 30
		curr, enter, leave = a.Move(obj.id, obj.x, obj.y)
		fmt.Println("curr", curr)
		fmt.Println("enter", enter)
		fmt.Println("leave", leave)
		fmt.Println()
	*/

	ok = a.Leave(obj.id, func(event AOIEvent, eventMaker int, eventWatcher int) {})
	require.True(t, ok)
}

func BenchmarkMove(b *testing.B) {
	rand.Seed(time.Now().Unix())
	const (
		w   = 10000
		h   = 10000
		obj = 1000
	)
	a, _ := NewAOIManager[int](0, 0, w, h, 10, 10)
	for i := 0; i < obj; i++ {
		a.Enter(i, rand.Int()%w, rand.Int()%h, func(event AOIEvent, eventMaker int, eventWatcher int) {})
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x, y := rand.Int()%w, rand.Int()%h
		_ = a.Move(i%100, x, y, func(event AOIEvent, eventMaker int, eventWatcher int) {})
	}
}
