package aoi

import "fmt"

// Grid 格子
type Grid[T ObjID] struct {
	id                     int              // 格子id
	row, col               int              // 行列
	minX, minY, maxX, maxY int              // 格子范围
	surroundGrids          []*Grid[T]       // 包含自己在内的九个格子
	surroundGridsMap       map[int]struct{} // map用作快速求交集并集

	observers map[T]struct{} // 观察者
	objs      map[T]struct{} // obj
}

func newGrid[T ObjID](id int, gridMinX, gridMinY, gridMaxX, gridMaxY int, row, col int) *Grid[T] {
	return &Grid[T]{
		id:   id,
		minX: gridMinX, minY: gridMinY, maxX: gridMaxX, maxY: gridMaxY,
		objs:             make(map[T]struct{}),
		col:              col,
		row:              row,
		observers:        make(map[T]struct{}),
		surroundGrids:    make([]*Grid[T], 0, GridNum),
		surroundGridsMap: make(map[int]struct{}, GridNum),
	}
}
func (g *Grid[ObjID]) add(obj ObjID, isObserver bool) {
	g.objs[obj] = struct{}{}
	if isObserver {
		g.observers[obj] = struct{}{}
	}
}

func (g *Grid[ObjID]) del(obj ObjID) {
	delete(g.objs, obj)
	delete(g.observers, obj)
}

func (g *Grid[ObjID]) clear() {
	g.objs = make(map[ObjID]struct{})
	g.observers = make(map[ObjID]struct{})
}

func (g *Grid[ObjID]) isSurround(gridID int) bool {
	_, ok := g.surroundGridsMap[gridID]
	return ok
}

func (g *Grid[ObjID]) addSurroundGrid(other *Grid[ObjID]) {
	if _, ok := g.surroundGridsMap[other.id]; ok {
		panic("duplicate grid")
	}
	g.surroundGridsMap[other.id] = struct{}{}
	g.surroundGrids = append(g.surroundGrids, other)
	if len(g.surroundGrids) > GridNum {
		panic("surround grid num is 9")
	}
}

func (g *Grid[ObjID]) invokeEvent(triggerID ObjID, toAll bool, eventType EventType, cb EventCallback[ObjID]) {
	others := g.observers
	if toAll {
		others = g.objs
	}
	for other := range others {
		if triggerID == other {
			continue
		}
		cb(eventType, other)
	}
}

// ID 格子id
func (g *Grid[ObjID]) ID() int {
	return g.id
}

// BoundingBox 范围
func (g *Grid[ObjID]) BoundingBox() (int, int, int, int) {
	return g.minX, g.minY, g.maxX, g.maxY
}

// RowCol 行列
func (g *Grid[ObjID]) RowCol() (int, int) {
	return g.row, g.col
}

// Contains 是否包含obj
func (g *Grid[ObjID]) Contains(obj ObjID) bool {
	_, ok := g.objs[obj]
	return ok
}

// ObjIDs 当前格子的所有obj
func (g *Grid[ObjID]) ObjIDs() map[ObjID]struct{} {
	return g.objs
}

// ObserverIDs 当前格子的所有观察者
func (g *Grid[ObjID]) ObserverIDs() map[ObjID]struct{} {
	return g.observers
}

// SurroundGrids 九宫格(包括自己)
func (g *Grid[ObjID]) SurroundGrids() []*Grid[ObjID] {
	return g.surroundGrids
}

// ForeachObjInSurroundGrids 遍历当前格子包含的obj
func (g *Grid[ObjID]) ForeachObjInSurroundGrids(f func(id ObjID)) {
	for _, v := range g.surroundGrids {
		for k := range v.objs {
			f(k)
		}
	}
}

// HasObserverInSurroundGrids 遍历当前格子包含的obj
func (g *Grid[ObjID]) HasObserverInSurroundGrids() bool {
	for _, v := range g.surroundGrids {
		if len(v.observers) > 0 {
			return true
		}
	}
	return false
}

func (g *Grid[ObjID]) String() string {
	return fmt.Sprintf("(%d:%d,%d)", g.id, g.row, g.col)
}
