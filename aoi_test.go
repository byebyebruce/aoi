package aoi

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

	fmt.Println(a.Enter(obj.id, obj.x, obj.y))
	fmt.Println()

	obj.x += 10
	curr, enter, leave := a.Move(obj.id, obj.x, obj.y)
	fmt.Println("curr", curr)
	fmt.Println("enter", enter)
	fmt.Println("leave", leave)
	fmt.Println()

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

	fmt.Println(a.Leave(obj.id))
}

func BenchmarkMove(b *testing.B) {
	rand.Seed(time.Now().Unix())
	const (
		w   = 10000
		h   = 10000
		obj = 100
	)
	a, _ := NewAOIManager[int](-w/2, -h/2, w, w, 10, 10)
	for i := 0; i < obj; i++ {
		a.Enter(i, 0, 0)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x, y := rand.Int()%30, rand.Int()%30
		_, _, _ = a.Move(i%100, x, y)
	}
}
