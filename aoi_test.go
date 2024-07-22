package aoi

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestAOI_NewAOIManager(t *testing.T) {
	_, err := NewAOIManager[int](100, 100, 10, 5)
	require.Nil(t, err)

	a, err := NewAOIManager[int](1, 1, 2, 2)
	require.Nil(t, err)
	require.EqualValues(t, 1, a.col)
	require.EqualValues(t, 1, a.row)

	a, err = NewAOIManagerWithMinXY[int](-1, -1, 2, 2, 2, 2)
	require.Nil(t, err)
}
func TestAOI_InitManger(t *testing.T) {
	a, err := NewAOIManager[int](100, 100, 10, 5)
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

func TestAOI_PosAtGrid(t *testing.T) {
	a, err := NewAOIManager[int](100, 100, 10, 5)
	require.Nil(t, err)

	require.EqualValues(t, 0, a.PosAtGrid(-100, -100).ID())
	require.EqualValues(t, a.col*a.row-1, a.PosAtGrid(1000, 1000).ID())

	require.EqualValues(t, 0, a.PosAtGrid(9, 4).ID())
	require.EqualValues(t, 9, a.PosAtGrid(99, 1).ID())
	require.EqualValues(t, 10, a.PosAtGrid(0, 5).ID())
	require.EqualValues(t, 11, a.PosAtGrid(10, 5).ID())
	require.EqualValues(t, a.col*a.row-1, a.PosAtGrid(99, 99).ID())

	for i := 0; i < a.row; i++ {
		for j := 0; j < a.col; j++ {
			x, y := j*a.gridW, i*a.gridH
			require.EqualValues(t, i*a.col+j, a.PosAtGrid(x, y).ID())
		}
	}
}

func TestAOI_Enter_Leave(tt *testing.T) {
	testFunc := func(t *testing.T) {
		w := 100 + rand.Int()%1000
		h := 100 + rand.Int()%1000
		gridW := 5 + rand.Int()%10
		gridH := 5 + rand.Int()%10
		a, err := NewAOIManager[int](w, h, gridW, gridH)
		assert.Nil(t, err)

		for i := 0; i < a.row; i++ {
			for j := 0; j < a.col; j++ {
				x, y := j*a.gridW, i*a.gridH
				id := a.gridIndex(i, j)
				ok := a.Enter(id, x, y, true, nil)
				require.True(t, ok)
			}
		}

		testID := -1
		for i := 0; i < a.row; i++ {
			for j := 0; j < a.col; j++ {
				x, y := j*a.gridW, i*a.gridH

				enterID := make(map[int]struct{})
				leaveID := make(map[int]struct{})
				g := a.PosAtGrid(x, y)
				g.ForeachObjInSurroundGrids(func(id int) {
					if id == testID {
						return
					}
					enterID[id] = struct{}{}
					leaveID[id] = struct{}{}
				})

				exists := a.Enter(testID, x, y, true, func(event EventType, observer int) {
					require.Equal(t, EnterView, event)
					_, ok := enterID[observer]
					require.True(t, ok, observer)
					delete(enterID, observer)
				})
				require.True(t, exists)
				require.Len(t, enterID, 0)

				exists = a.Leave(testID, func(event EventType, observer int) {
					require.Equal(t, LeaveView, event)
					_, ok := leaveID[observer]
					require.True(t, ok)
					delete(leaveID, observer)
				})
				require.True(t, exists)
				require.Len(t, leaveID, 0)
			}
		}
	}
	for i := 0; i < 20; i++ {
		tt.Run(fmt.Sprintf("test-%d", i), testFunc)
	}
}

func TestAOI_Move(t *testing.T) {
	w := 100
	h := 100
	gridW := 10
	gridH := 10
	a, err := NewAOIManager[int](w, h, gridW, gridH)
	assert.Nil(t, err)

	for i := 0; i < a.row; i++ {
		for j := 0; j < a.col; j++ {
			x, y := j*a.gridW, i*a.gridH
			id := a.gridIndex(i, j)
			ok := a.Enter(id, x, y, true, nil)
			require.True(t, ok)
		}
	}

	testID := -1
	type args struct {
		fromX, formY, toX, toY int
	}
	tests := []struct {
		args  args
		enter []int
		leave []int
		move  []int
	}{
		{
			args{0, 0, 10, 10},
			[]int{20, 21, 22, 12, 02},
			[]int{},
			[]int{10, 11, 0, 01},
		},

		{
			args{0, 0, 20, 20},
			[]int{31, 32, 33, 21, 22, 23, 12, 13},
			[]int{0, 1, 10},
			[]int{11},
		},

		{
			args{0, 0, 20, 20},
			[]int{31, 32, 33, 21, 22, 23, 12, 13},
			[]int{0, 1, 10},
			[]int{11},
		},

		{
			args{0, 0, 00, 30},
			[]int{40, 41, 30, 31, 20, 21},
			[]int{0, 1, 10, 11},
			[]int{},
		},

		{
			args{100, 40, 89, 40},
			[]int{57, 47, 37},
			[]int{},
			[]int{58, 48, 38, 39, 49, 59},
		},

		{
			args{35, 55, 45, 65},
			[]int{73, 74, 75, 65, 55},
			[]int{62, 52, 42, 43, 44},
			[]int{63, 53, 54, 64},
		},
	}

	/*
	   +----+----+----+----+----+----+----+----+----+----+
	   | 90 | 91 | 92 | 93 | 94 | 95 | 96 | 97 | 98 | 99 |
	   +----+----+----+----+----+----+----+----+----+----+
	   | 80 | 81 | 82 | 83 | 84 | 85 | 86 | 87 | 88 | 89 |
	   +----+----+----+----+----+----+----+----+----+----+
	   | 70 | 71 | 72 | 73 | 74 | 75 | 76 | 77 | 78 | 79 |
	   +----+----+----+----+----+----+----+----+----+----+
	   | 60 | 61 | 62 | 63 | 64 | 65 | 66 | 67 | 68 | 69 |
	   +----+----+----+----+----+----+----+----+----+----+
	   | 50 | 51 | 52 | 53 | 54 | 55 | 56 | 57 | 58 | 59 |
	   +----+----+----+----+----+----+----+----+----+----+
	   | 40 | 41 | 42 | 43 | 44 | 45 | 46 | 47 | 48 | 49 |
	   +----+----+----+----+----+----+----+----+----+----+
	   | 30 | 31 | 32 | 33 | 34 | 35 | 36 | 37 | 38 | 39 |
	   +----+----+----+----+----+----+----+----+----+----+
	   | 20 | 21 | 22 | 23 | 24 | 25 | 26 | 27 | 28 | 29 |
	   +----+----+----+----+----+----+----+----+----+----+
	   | 10 | 11 | 12 | 13 | 14 | 15 | 16 | 17 | 18 | 19 |
	   +----+----+----+----+----+----+----+----+----+----+
	   | 00 | 01 | 02 | 03 | 04 | 05 | 06 | 07 | 08 | 09 |
	   +----+----+----+----+----+----+----+----+----+----+

	*/

	for _, tt := range tests {
		t.Run(fmt.Sprintf("test-%d", tt.args), func(t *testing.T) {
			defer a.Leave(testID, nil)

			a.Enter(testID, tt.args.fromX, tt.args.formY, true, nil)

			enterID := make(map[int]struct{})
			leaveID := make(map[int]struct{})
			moveID := make(map[int]struct{})

			for _, i := range tt.enter {
				enterID[i] = struct{}{}
			}
			for _, i := range tt.move {
				moveID[i] = struct{}{}
			}
			for _, i := range tt.leave {
				leaveID[i] = struct{}{}
			}

			exists := a.Move(testID, tt.args.toX, tt.args.toY, func(event EventType, observer int) {
				if event == EnterView {
					_, ok := enterID[observer]
					require.True(t, ok, observer)
					delete(enterID, observer)
				} else if event == LeaveView {
					_, ok := leaveID[observer]
					require.True(t, ok)
					delete(leaveID, observer)
				} else if event == UpdateView {
					_, ok := moveID[observer]
					require.True(t, ok)
					delete(moveID, observer)
				}
			})
			require.True(t, exists)
			require.Len(t, enterID, 0)
			require.Len(t, leaveID, 0)
			require.Len(t, moveID, 0)
		})
	}
}

type shouldCall map[int]struct{}

func (s shouldCall) callFunc(t *testing.T) func(event EventType, observer int) {
	return func(event EventType, observer int) {
		_, ok := s[observer]
		if !ok {
			require.Fail(t, "should not call")
		}
		require.True(t, ok)
		delete(s, observer)
	}
}
func (s shouldCall) shouldEmpty(t *testing.T) {
	require.Len(t, s, 0)
}

func TestAOI_Observer(tt *testing.T) {
	a, err := NewAOIManager[int](100, 100, 10, 10)
	require.Nil(tt, err)

	shouldNotCall := func(event EventType, observer int) {
		require.Fail(tt, "should not call")
	}

	a.Enter(1, 10, 10, false, shouldNotCall)
	a.Enter(2, 10, 10, false, shouldNotCall)
	a.Move(1, 20, 20, shouldNotCall)
	a.Move(2, 20, 20, shouldNotCall)

	_shouldCall := shouldCall{
		1: {},
		2: {},
	}
	a.Enter(3, 20, 20, true, _shouldCall.callFunc(tt))
	_shouldCall.shouldEmpty(tt)

	_shouldCall = shouldCall{
		1: {},
		2: {},
		3: {},
	}
	a.Enter(4, 20, 20, true, _shouldCall.callFunc(tt))
	_shouldCall.shouldEmpty(tt)

	_shouldCall = shouldCall{
		3: {},
		4: {},
	}
	a.Move(1, 20, 20, _shouldCall.callFunc(tt))
	_shouldCall.shouldEmpty(tt)

	_shouldCall = shouldCall{
		4: {},
	}
	a.Move(3, 20, 20, _shouldCall.callFunc(tt))
	_shouldCall.shouldEmpty(tt)

	_shouldCall = shouldCall{
		3: {},
	}
	a.Move(4, 20, 20, _shouldCall.callFunc(tt))
	_shouldCall.shouldEmpty(tt)

	_shouldCall = shouldCall{
		3: {},
		4: {},
	}
	a.Leave(1, _shouldCall.callFunc(tt))
}

func BenchmarkAOI_Move(b *testing.B) {
	const (
		w   = 10000
		h   = 10000
		obj = 1000000
	)
	a, _ := NewAOIManager[int](w, h, 10, 10)
	for i := 0; i < obj; i++ {
		a.Enter(i, rand.Int()%w, rand.Int()%h, true, nil)
	}
	moveXMin := -2 * a.gridW
	moveYMin := -2 * a.gridH
	moveXLength := 4 * a.gridW
	moveYLength := 4 * a.gridH
	cb := func(event EventType, other int) {}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		x, y := moveXMin+rand.Intn(moveXLength), moveYMin+rand.Intn(moveYLength)
		b.StartTimer()
		_ = a.Move(i%obj, x, y, cb)
	}
}
